package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SQLite条件固定の原子保存と冪等性。
func TestStoreFixExperimentConditions(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	draft := testExperimentPreparationDraft("draft-request", "experiment-1", "固定する目的")
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	conditions := fixedConditionsFromDraft(draft)

	fixed, err := store.FixExperimentConditions(context.Background(), conditions)
	if err != nil {
		t.Fatalf("FixExperimentConditions() error = %v", err)
	}
	if fixed.FixedConditionID == "" || fixed.OperationID == "" || fixed.FixedAt.IsZero() {
		t.Errorf("fixed result = %+v, want identifiers and time", fixed)
	}
	var state string
	var fixedConditionID string
	if err := store.db.QueryRow("SELECT state, fixed_condition_id FROM experiments WHERE id = ?", draft.ExperimentID).Scan(&state, &fixedConditionID); err != nil {
		t.Fatalf("fixed experiment query error = %v", err)
	}
	if state != "ready" || fixedConditionID != fixed.FixedConditionID {
		t.Errorf("experiment = (%q, %q), want (ready, %q)", state, fixedConditionID, fixed.FixedConditionID)
	}
	var promptCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_fixed_condition_prompts WHERE fixed_condition_id = ?", fixed.FixedConditionID).Scan(&promptCount); err != nil {
		t.Fatalf("fixed prompts query error = %v", err)
	}
	if promptCount != len(draft.Prompts) {
		t.Errorf("fixed prompt count = %d, want %d", promptCount, len(draft.Prompts))
	}

	resent, err := store.FixExperimentConditions(context.Background(), conditions)
	if err != nil {
		t.Fatalf("second FixExperimentConditions() error = %v", err)
	}
	if resent.FixedConditionID != fixed.FixedConditionID || resent.OperationID != fixed.OperationID {
		t.Errorf("resent = %+v, want fixed snapshot %+v", resent, fixed)
	}
}

// SQLite条件固定の競合再試行とキャンセル。
func TestStoreFixExperimentConditionsRetriesBusyDatabase(t *testing.T) {
	dataDirectory := t.TempDir()
	firstStore, err := Open(dataDirectory)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := firstStore.Close(); err != nil {
			t.Errorf("first Close() error = %v", err)
		}
	})
	secondStore, err := Open(dataDirectory)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := secondStore.Close(); err != nil {
			t.Errorf("second Close() error = %v", err)
		}
	})
	seedExperimentPreparationDraftExperiment(t, firstStore, "experiment-1")
	draft := testExperimentPreparationDraft("draft-request", "experiment-1", "固定する目的")
	if _, err := firstStore.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	transaction, err := firstStore.db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	if _, err := transaction.Exec("UPDATE experiments SET purpose = purpose WHERE id = ?", draft.ExperimentID); err != nil {
		t.Fatalf("lock database update error = %v", err)
	}
	t.Cleanup(func() {
		if err := transaction.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("lock transaction rollback error = %v", err)
		}
	})
	conditions := fixedConditionsFromDraft(draft)
	if _, err := secondStore.FixExperimentConditions(context.Background(), conditions); err == nil {
		t.Fatal("FixExperimentConditions() error = nil, want busy database error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Millisecond, cancel)
	defer cancel()
	if _, err := secondStore.FixExperimentConditions(ctx, conditions); !errors.Is(err, context.Canceled) {
		t.Errorf("FixExperimentConditions() error = %v, want error wrapping %v", err, context.Canceled)
	}
}

// SQLite条件固定の複数DB接続からの同時冪等保存。
func TestStoreFixExperimentConditionsConvergesConcurrentRequestsAcrossStores(t *testing.T) {
	dataDirectory := t.TempDir()
	firstStore, err := Open(dataDirectory)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := firstStore.Close(); err != nil {
			t.Errorf("first Close() error = %v", err)
		}
	})
	secondStore, err := Open(dataDirectory)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := secondStore.Close(); err != nil {
			t.Errorf("second Close() error = %v", err)
		}
	})
	seedExperimentPreparationDraftExperiment(t, firstStore, "experiment-1")
	draft := testExperimentPreparationDraft("draft-request", "experiment-1", "固定する目的")
	if _, err := firstStore.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}

	conditions := fixedConditionsFromDraft(draft)
	stores := []*Store{
		firstStore,
		secondStore,
	}
	ready := make(chan struct{})
	entered := make(chan struct{}, len(stores))
	results := make(chan conditionFixResult, len(stores))
	var waitGroup sync.WaitGroup
	for _, fixingStore := range stores {
		waitGroup.Add(1)
		go func(store *Store) {
			defer waitGroup.Done()
			entered <- struct{}{}
			<-ready
			fixed, err := store.FixExperimentConditions(context.Background(), conditions)
			results <- conditionFixResult{
				conditions: fixed,
				err:        err,
			}
		}(fixingStore)
	}
	for range stores {
		<-entered
	}
	close(ready)
	waitGroup.Wait()
	close(results)

	var snapshots []domain.ExperimentFixedConditions
	for result := range results {
		if result.err != nil {
			t.Fatalf("FixExperimentConditions() error = %v", result.err)
		}
		snapshots = append(snapshots, result.conditions)
	}
	if gotCount := len(snapshots); gotCount != len(stores) {
		t.Fatalf("concurrent snapshots length = %d, want %d", gotCount, len(stores))
	}
	if !reflect.DeepEqual(snapshots[0], snapshots[1]) {
		t.Errorf("concurrent snapshots = %+v, want identical snapshots", snapshots)
	}
	if gotCount := countConditionFixOperations(t, firstStore, draft.ExperimentID); gotCount != 1 {
		t.Errorf("condition fix operations = %d, want 1", gotCount)
	}
}

// SQLite条件固定の競合と固定済み拒否。
func TestStoreFixExperimentConditionsRejectsConflictAndAlreadyFixed(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*domain.ExperimentFixedConditions)
		wantCode apperr.Code
	}{
		{
			name:     "下書きと異なる目的を拒否する",
			mutate:   func(conditions *domain.ExperimentFixedConditions) { conditions.Purpose = "古い目的" },
			wantCode: apperr.CodeExperimentConditionsConflict,
		},
		{
			name:     "固定済み実験を拒否する",
			mutate:   func(conditions *domain.ExperimentFixedConditions) { conditions.RequestID = "second-request" },
			wantCode: apperr.CodeExperimentConditionsAlreadyFixed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentPreparationDraftStore(t)
			draft := testExperimentPreparationDraft("draft-request", "experiment-1", "固定する目的")
			if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
				t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
			}
			conditions := fixedConditionsFromDraft(draft)
			if tt.wantCode == apperr.CodeExperimentConditionsAlreadyFixed {
				if _, err := store.FixExperimentConditions(context.Background(), conditions); err != nil {
					t.Fatalf("FixExperimentConditions() error = %v", err)
				}
			}
			tt.mutate(&conditions)
			_, err := store.FixExperimentConditions(context.Background(), conditions)
			if !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("FixExperimentConditions() error = %v, want code %q", err, tt.wantCode)
			}
		})
	}
}

// 条件固定比較helperの境界。
func TestSameExperimentConditions(t *testing.T) {
	value := "仮説"
	base := domain.ExperimentFixedConditions{
		Purpose:               "目的",
		Hypothesis:            &value,
		EnvironmentConditions: "環境",
		InitialInput:          "入力",
		EvaluationAxes:        "評価",
		Prompts: []domain.ExperimentPreparationPrompt{
			{
				SequenceNo: 1,
				Content:    "prompt",
			},
		},
	}
	tests := []struct {
		name       string
		conditions domain.ExperimentFixedConditions
		want       bool
	}{
		{
			name:       "同じ条件を一致とする",
			conditions: base,
			want:       true,
		},
		{
			name: "prompt内容の違いを検出する",
			conditions: domain.ExperimentFixedConditions{
				Purpose:               base.Purpose,
				Hypothesis:            base.Hypothesis,
				EnvironmentConditions: base.EnvironmentConditions,
				InitialInput:          base.InitialInput,
				EvaluationAxes:        base.EvaluationAxes,
				Prompts: []domain.ExperimentPreparationPrompt{
					{
						SequenceNo: 1,
						Content:    "別prompt",
					},
				},
			},
			want: false,
		},
		{
			name: "nil仮説の違いを検出する",
			conditions: domain.ExperimentFixedConditions{
				Purpose:               base.Purpose,
				EnvironmentConditions: base.EnvironmentConditions,
				InitialInput:          base.InitialInput,
				EvaluationAxes:        base.EvaluationAxes,
				Prompts:               base.Prompts,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameExperimentConditions(base, tt.conditions); got != tt.want {
				t.Errorf("sameExperimentConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

// SQLite条件固定失敗時の関連record不変性。
func TestStoreFixExperimentConditionsRollsBackOnOperationFailure(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	draft := testExperimentPreparationDraft("draft-request", "experiment-1", "固定する目的")
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	before := readConditionFixState(t, store, draft.ExperimentID)
	if _, err := store.db.Exec("CREATE TRIGGER fail_condition_operation BEFORE INSERT ON experiment_condition_fix_operations BEGIN SELECT RAISE(ABORT, 'forced operation failure'); END"); err != nil {
		t.Fatalf("create failure trigger error = %v", err)
	}
	if _, err := store.FixExperimentConditions(context.Background(), fixedConditionsFromDraft(draft)); err == nil {
		t.Fatal("FixExperimentConditions() error = nil, want operation failure")
	}
	after := readConditionFixState(t, store, draft.ExperimentID)
	if before != after {
		t.Errorf("state after rollback = %+v, want %+v", after, before)
	}
}

// SQLite条件固定のtransaction開始前と検証失敗。
func TestStoreFixExperimentConditionsTransactionAndValidationFailures(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) (*Store, domain.ExperimentFixedConditions)
		wantCode apperr.Code
	}{
		{
			name: "transaction開始失敗を返す",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				store := newExperimentPreparationDraftStore(t)
				conditions := fixedConditionsFromDraft(testExperimentPreparationDraft("draft-request", "experiment-1", "目的"))
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
				return store, conditions
			},
		},
		{
			name: "操作読込失敗を返す",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				return &Store{db: newConditionFixRawDatabase(t)}, domain.ExperimentFixedConditions{
					RequestID:    "request-1",
					ExperimentID: "experiment-1",
				}
			},
		},
		{
			name: "別実験のrequest IDを拒否する",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				store, conditions := fixedExperimentPreparationStore(t)
				conditions.ExperimentID = "experiment-2"
				return store, conditions
			},
			wantCode: apperr.CodeFixConditionsRequestInvalid,
		},
		{
			name: "実験がない場合は未検出を返す",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				store := newExperimentPreparationDraftStore(t)
				conditions := fixedConditionsFromDraft(testExperimentPreparationDraft("draft-request", "unknown", "目的"))
				return store, conditions
			},
			wantCode: apperr.CodeExperimentPreparationNotFound,
		},
		{
			name: "実験読込失敗を返す",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				db := newConditionFixRawDatabase(t)
				execConditionFixSQL(t, db, "CREATE TABLE experiment_condition_fix_operations (request_id TEXT PRIMARY KEY, experiment_id TEXT, fixed_condition_id TEXT, operation_id TEXT, fixed_at TEXT)")
				execConditionFixSQL(t, db, "CREATE TABLE experiment_fixed_conditions (id TEXT PRIMARY KEY, experiment_id TEXT, purpose TEXT, hypothesis TEXT, environment_conditions TEXT, initial_input TEXT, evaluation_axes TEXT)")
				return &Store{db: db}, domain.ExperimentFixedConditions{
					RequestID:    "request-1",
					ExperimentID: "experiment-1",
				}
			},
		},
		{
			name: "準備中以外を固定済みとして拒否する",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				store := newExperimentPreparationDraftStore(t)
				execConditionFixSQL(t, store.db, "UPDATE experiments SET state = 'draft' WHERE id = 'experiment-1'")
				return store, fixedConditionsFromDraft(testExperimentPreparationDraft("draft-request", "experiment-1", "目的"))
			},
			wantCode: apperr.CodeExperimentConditionsAlreadyFixed,
		},
		{
			name: "下書きがない場合は未検出を返す",
			setup: func(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
				store := newExperimentPreparationDraftStore(t)
				execConditionFixSQL(t, store.db, "DELETE FROM experiment_preparations WHERE experiment_id = 'experiment-1'")
				return store, fixedConditionsFromDraft(testExperimentPreparationDraft("draft-request", "experiment-1", "目的"))
			},
			wantCode: apperr.CodeExperimentPreparationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, conditions := tt.setup(t)
			_, err := store.fixExperimentConditions(context.Background(), conditions)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("fixExperimentConditions() error = %v, want code %q", err, tt.wantCode)
				}
				return
			}
			if err == nil {
				t.Fatal("fixExperimentConditions() error = nil, want transaction error")
			}
		})
	}
}

// SQLite条件固定の識別子生成失敗。
func TestStoreFixExperimentConditionsIdentifierFailures(t *testing.T) {
	tests := []struct {
		name        string
		failAt      int
		wantMessage string
	}{
		{
			name:        "固定条件IDの生成失敗を返す",
			failAt:      1,
			wantMessage: "generate fixed condition ID",
		},
		{
			name:        "操作IDの生成失敗を返す",
			failAt:      2,
			wantMessage: "generate condition fix operation ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentPreparationDraftStore(t)
			draft := testExperimentPreparationDraft("draft-request", "experiment-1", "目的")
			if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
				t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
			}
			calls := 0
			replaceBriefingRandom(t, func(bytes []byte) (int, error) {
				calls++
				if calls == tt.failAt {
					return 0, errors.New("random unavailable")
				}
				for index := range bytes {
					bytes[index] = byte(calls)
				}

				return len(bytes), nil
			})
			_, err := store.fixExperimentConditions(context.Background(), fixedConditionsFromDraft(draft))
			if err == nil || !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("fixExperimentConditions() error = %v, want message %q", err, tt.wantMessage)
			}
		})
	}
}

// SQLite条件固定の各record書込失敗。
func TestStoreFixExperimentConditionsRecordWriteFailures(t *testing.T) {
	tests := []struct {
		name      string
		trigger   string
		wantError string
	}{
		{
			name:      "固定prompt書込失敗を返す",
			trigger:   "CREATE TRIGGER fail_fixed_prompt BEFORE INSERT ON experiment_fixed_condition_prompts BEGIN SELECT RAISE(ABORT, 'fixed prompt failure'); END",
			wantError: "insert fixed condition prompt",
		},
		{
			name:      "実験状態遷移失敗を返す",
			trigger:   "CREATE TRIGGER fail_fixed_experiment BEFORE UPDATE ON experiments BEGIN SELECT RAISE(ABORT, 'experiment update failure'); END",
			wantError: "transition experiment conditions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentPreparationDraftStore(t)
			draft := testExperimentPreparationDraft("draft-request", "experiment-1", "目的")
			if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
				t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
			}
			execConditionFixSQL(t, store.db, tt.trigger)
			_, err := store.fixExperimentConditions(context.Background(), fixedConditionsFromDraft(draft))
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("fixExperimentConditions() error = %v, want message %q", err, tt.wantError)
			}
		})
	}
}

// fixedExperimentPreparationStore は固定済み条件を持つSQLite storeを生成する。
func fixedExperimentPreparationStore(t *testing.T) (*Store, domain.ExperimentFixedConditions) {
	t.Helper()
	store := newExperimentPreparationDraftStore(t)
	draft := testExperimentPreparationDraft("draft-request", "experiment-1", "目的")
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	conditions := fixedConditionsFromDraft(draft)
	if _, err := store.FixExperimentConditions(context.Background(), conditions); err != nil {
		t.Fatalf("FixExperimentConditions() error = %v", err)
	}

	return store, conditions
}

// 固定前下書き読込の成功とSQLite境界エラー。
func TestFindCurrentExperimentPreparationDraft(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T) (context.Context, conditionFixQueryer, string)
		want  apperr.Code
	}{
		{
			name: "下書きがない場合は未検出を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				return context.Background(), newExperimentPreparationDraftStore(t).db, "unknown"
			},
			want: apperr.CodeExperimentPreparationNotFound,
		},
		{
			name: "主record読込失敗を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				db := newConditionFixRawDatabase(t)
				return context.Background(), db, "experiment-1"
			},
		},
		{
			name: "prompt表読込失敗を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				db := newConditionFixRawDatabase(t)
				execConditionFixSQL(t, db, "CREATE TABLE experiments (id TEXT PRIMARY KEY, purpose TEXT NOT NULL)")
				execConditionFixSQL(t, db, "CREATE TABLE experiment_preparations (experiment_id TEXT PRIMARY KEY, hypothesis TEXT, environment_conditions TEXT NOT NULL, initial_input TEXT NOT NULL, evaluation_criteria TEXT NOT NULL)")
				execConditionFixSQL(t, db, "INSERT INTO experiments (id, purpose) VALUES ('experiment-1', '目的')")
				execConditionFixSQL(t, db, "INSERT INTO experiment_preparations (experiment_id, environment_conditions, initial_input, evaluation_criteria) VALUES ('experiment-1', '環境', '入力', '評価')")
				return context.Background(), db, "experiment-1"
			},
		},
		{
			name: "prompt走査失敗を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				store := newExperimentPreparationDraftStore(t)
				draft := testExperimentPreparationDraft("draft-request", "experiment-1", "目的")
				if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
					t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
				}
				execConditionFixSQL(t, store.db, "UPDATE experiment_preparation_prompts SET sequence_no = 'not-a-number' WHERE experiment_id = 'experiment-1' AND sequence_no = 1")
				return context.Background(), store.db, "experiment-1"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, queryer, experimentID := tt.setup(t)
			got, err := findCurrentExperimentPreparationDraft(ctx, queryer, experimentID)
			if tt.want != "" {
				if !apperr.IsCode(err, tt.want) {
					t.Errorf("findCurrentExperimentPreparationDraft() error = %v, want code %q", err, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatal("findCurrentExperimentPreparationDraft() error = nil, want read error")
			}
			if !reflect.DeepEqual(got, domain.ExperimentFixedConditions{}) {
				t.Errorf("findCurrentExperimentPreparationDraft() = %+v, want zero value", got)
			}
		})
	}
}

// 固定操作snapshot読込の成功とSQLite境界エラー。
func TestFindExperimentConditionFixOperation(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T) (context.Context, conditionFixQueryer, string)
		want  apperr.Code
	}{
		{
			name: "操作がない場合は未検出を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				return context.Background(), newExperimentPreparationDraftStore(t).db, "unknown"
			},
		},
		{
			name: "操作record読込失敗を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				return context.Background(), newConditionFixRawDatabase(t), "request-1"
			},
		},
		{
			name: "固定日時の形式不正を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				store := newExperimentPreparationDraftStore(t)
				seedConditionFixSnapshot(t, store, "request-1", "fixed-1", "invalid-time")
				return context.Background(), store.db, "request-1"
			},
		},
		{
			name: "固定prompt表読込失敗を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				store := newExperimentPreparationDraftStore(t)
				seedConditionFixSnapshot(t, store, "request-1", "fixed-1", time.Now().UTC().Format(time.RFC3339Nano))
				execConditionFixSQL(t, store.db, "DROP TABLE experiment_fixed_condition_prompts")
				return context.Background(), store.db, "request-1"
			},
		},
		{
			name: "固定prompt走査失敗を返す",
			setup: func(t *testing.T) (context.Context, conditionFixQueryer, string) {
				store := newExperimentPreparationDraftStore(t)
				seedConditionFixSnapshot(t, store, "request-1", "fixed-1", time.Now().UTC().Format(time.RFC3339Nano))
				execConditionFixSQL(t, store.db, "INSERT INTO experiment_fixed_condition_prompts (fixed_condition_id, sequence_no, content) VALUES ('fixed-1', 'not-a-number', 'prompt')")
				return context.Background(), store.db, "request-1"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, queryer, requestID := tt.setup(t)
			got, found, err := findExperimentConditionFixOperation(ctx, queryer, requestID)
			if tt.want != "" {
				if !apperr.IsCode(err, tt.want) {
					t.Errorf("findExperimentConditionFixOperation() error = %v, want code %q", err, tt.want)
				}
				return
			}
			if requestID == "unknown" {
				if err != nil {
					t.Errorf("findExperimentConditionFixOperation() error = %v, want nil", err)
				}
				if found {
					t.Errorf("found = %v, want false", found)
				}
				if !reflect.DeepEqual(got, domain.ExperimentFixedConditions{}) {
					t.Errorf("findExperimentConditionFixOperation() = %+v, want zero value", got)
				}
				return
			}
			if err == nil {
				t.Fatal("findExperimentConditionFixOperation() error = nil, want read error")
			}
		})
	}
}

// newConditionFixRawDatabase はスキーマ不足を再現するSQLite接続。
func newConditionFixRawDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	return db
}

// execConditionFixSQL はfixture用SQLを実行する。
func execConditionFixSQL(t *testing.T, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.Exec(statement); err != nil {
		t.Fatalf("Exec(%q) error = %v", statement, err)
	}
}

// seedConditionFixSnapshot は固定操作読込用recordを登録する。
func seedConditionFixSnapshot(t *testing.T, store *Store, requestID, fixedConditionID, fixedAt string) {
	t.Helper()
	execConditionFixSQL(t, store.db, "INSERT INTO experiment_fixed_conditions (id, experiment_id, purpose, environment_conditions, initial_input, evaluation_axes, artifact_payload, fixed_at) VALUES ('"+fixedConditionID+"', 'experiment-1', '目的', '環境', '入力', '評価', '{}', '"+fixedAt+"')")
	execConditionFixSQL(t, store.db, "INSERT INTO experiment_condition_fix_operations (request_id, experiment_id, fixed_condition_id, operation_id, fixed_at) VALUES ('"+requestID+"', 'experiment-1', '"+fixedConditionID+"', 'operation-1', '"+fixedAt+"')")
}

// conditionFixState は固定条件transactionの全関連状態。
type conditionFixState struct {
	state            string
	fixedConditionID string
	fixedConditions  int
	fixedPrompts     int
	operations       int
}

// conditionFixResult は同時条件固定の結果。
type conditionFixResult struct {
	conditions domain.ExperimentFixedConditions
	err        error
}

// countConditionFixOperations は実験に紐づく固定操作数を返す。
func countConditionFixOperations(t *testing.T, store *Store, experimentID string) int {
	t.Helper()

	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_condition_fix_operations WHERE experiment_id = ?", experimentID).Scan(&count); err != nil {
		t.Fatalf("condition operations query error = %v", err)
	}

	return count
}

// readConditionFixState は条件固定関連recordを取得する。
func readConditionFixState(t *testing.T, store *Store, experimentID string) conditionFixState {
	t.Helper()
	var result conditionFixState
	if err := store.db.QueryRow("SELECT state, COALESCE(fixed_condition_id, '') FROM experiments WHERE id = ?", experimentID).Scan(&result.state, &result.fixedConditionID); err != nil {
		t.Fatalf("experiment state query error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_fixed_conditions WHERE experiment_id = ?", experimentID).Scan(&result.fixedConditions); err != nil {
		t.Fatalf("fixed conditions query error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_fixed_condition_prompts WHERE fixed_condition_id IN (SELECT id FROM experiment_fixed_conditions WHERE experiment_id = ?)", experimentID).Scan(&result.fixedPrompts); err != nil {
		t.Fatalf("fixed prompts query error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_condition_fix_operations WHERE experiment_id = ?", experimentID).Scan(&result.operations); err != nil {
		t.Fatalf("condition operations query error = %v", err)
	}

	return result
}

// fixedConditionsFromDraft は下書きと一致する固定条件を返す。
func fixedConditionsFromDraft(draft domain.ExperimentPreparationDraft) domain.ExperimentFixedConditions {
	return domain.ExperimentFixedConditions{
		RequestID:             "fix-request",
		ExperimentID:          draft.ExperimentID,
		Purpose:               draft.Purpose,
		Hypothesis:            draft.Hypothesis,
		EnvironmentConditions: draft.EnvironmentConditions,
		InitialInput:          draft.InitialInput,
		Prompts:               draft.Prompts,
		EvaluationAxes:        draft.EvaluationAxes,
	}
}
