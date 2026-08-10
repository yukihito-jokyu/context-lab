package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SQLite run評価の生成、冪等性、状態更新。
func TestStoreRunEvaluation(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	t.Cleanup(func() { _ = store.Close() })
	start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "安全なrun要約"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}

	evaluation, created, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() error = %v", err)
	}
	if !created || evaluation.State != domain.ExperimentEvaluationStateStarting || evaluation.EvaluationID == "" || evaluation.OperationID == "" {
		t.Errorf("evaluation = %+v, created = %v, want persisted starting evaluation", evaluation, created)
	}
	if err := store.CompleteRunEvaluation(context.Background(), evaluation.EvaluationID, "評価要約"); err != nil {
		t.Fatalf("CompleteRunEvaluation() error = %v", err)
	}
	replayed, created, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("replayed BeginRunEvaluation() error = %v", err)
	}
	if created || replayed.State != domain.ExperimentEvaluationStateCompleted || replayed.Summary == nil || *replayed.Summary != "評価要約" {
		t.Errorf("replayed = %+v, created = %v, want completed snapshot", replayed, created)
	}
	if _, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[1].ID); !apperr.IsCode(err, apperr.CodeRunEvaluationRequestInvalid) {
		t.Errorf("different run BeginRunEvaluation() error = %v, want code %q", err, apperr.CodeRunEvaluationRequestInvalid)
	}
	if _, _, err := store.BeginRunEvaluation(context.Background(), "another-request", start.Runs[0].ID); !apperr.IsCode(err, apperr.CodeRunEvaluationAlreadyExists) {
		t.Errorf("duplicate run BeginRunEvaluation() error = %v, want code %q", err, apperr.CodeRunEvaluationAlreadyExists)
	}
}

// 同一request IDを別々のdatabase/sql poolから同時に送っても一つの評価snapshotへ収束する。
func TestStoreBeginRunEvaluationConvergesConcurrentRequestsAcrossStores(t *testing.T) {
	for range 20 {
		t.Run("attempt", func(t *testing.T) {
			dataDirectory := t.TempDir()
			firstStore, err := Open(dataDirectory)
			if err != nil {
				t.Fatalf("first Open() error = %v", err)
			}
			t.Cleanup(func() { _ = firstStore.Close() })
			secondStore, err := Open(dataDirectory)
			if err != nil {
				t.Fatalf("second Open() error = %v", err)
			}
			t.Cleanup(func() { _ = secondStore.Close() })
			seedExperimentPreparationDraftExperiment(t, firstStore, "experiment-1")
			draft := testExperimentPreparationDraft("draft-request", "experiment-1", "目的")
			if _, err := firstStore.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
				t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
			}
			fixed := fixedConditionsFromDraft(draft)
			if _, err := firstStore.FixExperimentConditions(context.Background(), fixed); err != nil {
				t.Fatalf("FixExperimentConditions() error = %v", err)
			}
			start, _, err := firstStore.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
			if err != nil {
				t.Fatalf("BeginExperiment() error = %v", err)
			}
			if err := firstStore.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "安全なrun要約"); err != nil {
				t.Fatalf("CompleteExperimentRun() error = %v", err)
			}

			type result struct {
				evaluation domain.ExperimentRunEvaluation
				err        error
			}
			ready := make(chan struct{})
			entered := make(chan struct{}, 2)
			results := make(chan result, 2)
			var waitGroup sync.WaitGroup
			for _, store := range []*Store{
				firstStore,
				secondStore,
			} {
				waitGroup.Add(1)
				go func(store *Store) {
					defer waitGroup.Done()
					entered <- struct{}{}
					<-ready
					evaluation, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID)
					results <- result{
						evaluation: evaluation,
						err:        err,
					}
				}(store)
			}
			<-entered
			<-entered
			close(ready)
			waitGroup.Wait()
			close(results)

			var evaluations []domain.ExperimentRunEvaluation
			for result := range results {
				if result.err != nil {
					t.Fatalf("BeginRunEvaluation() error = %v", result.err)
				}
				evaluations = append(evaluations, result.evaluation)
			}
			if !reflect.DeepEqual(evaluations[0], evaluations[1]) {
				t.Errorf("concurrent snapshots = %+v, want identical", evaluations)
			}
		})
	}
}

// SQLite run評価の準備不足と失敗状態更新。
func TestStoreRunEvaluationFailurePaths(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	t.Cleanup(func() { _ = store.Close() })
	start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if _, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID); !apperr.IsCode(err, apperr.CodeRunEvaluationNotReady) {
		t.Errorf("queued BeginRunEvaluation() error = %v, want code %q", err, apperr.CodeRunEvaluationNotReady)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "安全なrun要約"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}
	evaluation, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() error = %v", err)
	}
	if err := store.FailRunEvaluation(context.Background(), evaluation.EvaluationID, string(apperr.CodeOperationTimeout)); err != nil {
		t.Fatalf("FailRunEvaluation() error = %v", err)
	}
	replayed, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("replayed BeginRunEvaluation() error = %v", err)
	}
	if replayed.State != domain.ExperimentEvaluationStateFailed || replayed.FailureCode != string(apperr.CodeOperationTimeout) {
		t.Errorf("replayed = %+v, want persisted failure", replayed)
	}
}

// SQLite run評価のtrigger失敗時は途中状態を残さない。
func TestStoreRunEvaluationTriggerFailures(t *testing.T) {
	tests := []struct {
		name    string
		trigger string
	}{
		{
			name:    "評価insert失敗を返す",
			trigger: "CREATE TRIGGER fail_evaluation_insert BEFORE INSERT ON experiment_evaluations BEGIN SELECT RAISE(ABORT, 'evaluation insert failed'); END",
		},
		{
			name:    "操作insert失敗を返す",
			trigger: "CREATE TRIGGER fail_evaluation_operation_insert BEFORE INSERT ON experiment_evaluation_operations BEGIN SELECT RAISE(ABORT, 'operation insert failed'); END",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, fixed := fixedExperimentPreparationStore(t)
			t.Cleanup(func() { _ = store.Close() })
			start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
			if err != nil {
				t.Fatalf("BeginExperiment() error = %v", err)
			}
			if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run要約"); err != nil {
				t.Fatalf("CompleteExperimentRun() error = %v", err)
			}
			if _, err := store.db.Exec(tt.trigger); err != nil {
				t.Fatalf("create trigger error = %v", err)
			}
			if _, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID); err == nil {
				t.Error("BeginRunEvaluation() error = nil, want trigger failure")
			}
			var evaluations int
			if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_evaluations WHERE run_id = ?", start.Runs[0].ID).Scan(&evaluations); err != nil {
				t.Fatalf("count evaluations error = %v", err)
			}
			if evaluations != 0 {
				t.Errorf("evaluations = %d, want 0", evaluations)
			}
		})
	}
}

// run評価の操作競合判定。
func TestIsRunEvaluationRequestConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "request競合を検出する",
			err:  errors.New("UNIQUE constraint failed: experiment_evaluation_operations.request_id"),
			want: true,
		},
		{
			name: "nilを競合にしない",
			want: false,
		},
		{
			name: "別エラーを競合にしない",
			err:  errors.New("other error"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRunEvaluationRequestConflict(tt.err); got != tt.want {
				t.Errorf("isRunEvaluationRequestConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

// SQLite run評価の更新、読込、transaction境界エラー。
func TestStoreRunEvaluationDatabaseFailures(t *testing.T) {
	t.Run("閉じたDBの評価transaction開始失敗を返す", func(t *testing.T) {
		store, _ := fixedExperimentPreparationStore(t)
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if _, _, err := store.beginRunEvaluation(context.Background(), "request-1", "run-1"); err == nil {
			t.Error("beginRunEvaluation() error = nil, want begin failure")
		}
		if err := store.updateRunEvaluation(context.Background(), "evaluation-1", domain.ExperimentEvaluationStateCompleted, "summary", ""); err == nil {
			t.Error("updateRunEvaluation() error = nil, want begin failure")
		}
	})
	t.Run("評価更新trigger失敗を返す", func(t *testing.T) {
		store, fixed := fixedExperimentPreparationStore(t)
		t.Cleanup(func() { _ = store.Close() })
		start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
		if err != nil {
			t.Fatalf("BeginExperiment() error = %v", err)
		}
		if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run要約"); err != nil {
			t.Fatalf("CompleteExperimentRun() error = %v", err)
		}
		evaluation, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID)
		if err != nil {
			t.Fatalf("BeginRunEvaluation() error = %v", err)
		}
		if _, err := store.db.Exec("CREATE TRIGGER fail_evaluation_update BEFORE UPDATE ON experiment_evaluations BEGIN SELECT RAISE(ABORT, 'evaluation update failed'); END"); err != nil {
			t.Fatalf("create trigger error = %v", err)
		}
		if err := store.CompleteRunEvaluation(context.Background(), evaluation.EvaluationID, "評価要約"); err == nil {
			t.Error("CompleteRunEvaluation() error = nil, want update failure")
		}
		if _, err := store.db.Exec("DROP TRIGGER fail_evaluation_update"); err != nil {
			t.Fatalf("drop trigger error = %v", err)
		}
		if _, err := store.db.Exec("CREATE TRIGGER fail_evaluation_operation_update BEFORE UPDATE ON experiment_evaluation_operations BEGIN SELECT RAISE(ABORT, 'operation update failed'); END"); err != nil {
			t.Fatalf("create trigger error = %v", err)
		}
		if err := store.FailRunEvaluation(context.Background(), evaluation.EvaluationID, string(apperr.CodeOperationTimeout)); err == nil {
			t.Error("FailRunEvaluation() error = nil, want operation update failure")
		}
	})
	t.Run("評価snapshotの時刻不正を返す", func(t *testing.T) {
		store, fixed := fixedExperimentPreparationStore(t)
		t.Cleanup(func() { _ = store.Close() })
		start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
		if err != nil {
			t.Fatalf("BeginExperiment() error = %v", err)
		}
		if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run要約"); err != nil {
			t.Fatalf("CompleteExperimentRun() error = %v", err)
		}
		if _, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID); err != nil {
			t.Fatalf("BeginRunEvaluation() error = %v", err)
		}
		if _, err := store.db.Exec("UPDATE experiment_evaluations SET updated_at = 'invalid-time'"); err != nil {
			t.Fatalf("update timestamp error = %v", err)
		}
		if _, _, err := store.findRunEvaluation(context.Background(), "evaluation-request"); err == nil {
			t.Error("findRunEvaluation() error = nil, want timestamp parse failure")
		}
	})
	t.Run("driver読込失敗を返す", func(t *testing.T) {
		database := openMigrationFailureDatabase(t, "query")
		t.Cleanup(func() { _ = database.Close() })
		store := &Store{db: database}
		if _, _, err := store.findRunEvaluation(context.Background(), "request-1"); err == nil {
			t.Error("findRunEvaluation() error = nil, want query failure")
		}
		transaction, err := database.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("BeginTx() error = %v", err)
		}
		defer func() { _ = transaction.Rollback() }()
		if _, err := findEvaluableRun(context.Background(), transaction, "run-1"); err == nil {
			t.Error("findEvaluableRun() error = nil, want query failure")
		}
	})
	t.Run("driver更新commit失敗を返す", func(t *testing.T) {
		database := openMigrationFailureDatabase(t, "commit")
		t.Cleanup(func() { _ = database.Close() })
		if err := (&Store{db: database}).updateRunEvaluation(context.Background(), "evaluation-1", domain.ExperimentEvaluationStateCompleted, "summary", ""); err == nil {
			t.Error("updateRunEvaluation() error = nil, want commit failure")
		}
	})
	t.Run("評価対象がない場合は準備不足を返す", func(t *testing.T) {
		store, _ := fixedExperimentPreparationStore(t)
		t.Cleanup(func() { _ = store.Close() })
		if _, _, err := store.beginRunEvaluation(context.Background(), "request-1", "unknown-run"); !apperr.IsCode(err, apperr.CodeRunEvaluationNotReady) {
			t.Errorf("beginRunEvaluation() error = %v, want code %q", err, apperr.CodeRunEvaluationNotReady)
		}
	})
	t.Run("評価ID生成失敗を返す", func(t *testing.T) {
		store, fixed := fixedExperimentPreparationStore(t)
		t.Cleanup(func() { _ = store.Close() })
		start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
		if err != nil {
			t.Fatalf("BeginExperiment() error = %v", err)
		}
		if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run要約"); err != nil {
			t.Fatalf("CompleteExperimentRun() error = %v", err)
		}
		replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random failed") })
		if _, _, err := store.beginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID); err == nil {
			t.Error("beginRunEvaluation() error = nil, want ID failure")
		}
	})
	t.Run("操作ID生成失敗を返す", func(t *testing.T) {
		store, fixed := fixedExperimentPreparationStore(t)
		t.Cleanup(func() { _ = store.Close() })
		start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
		if err != nil {
			t.Fatalf("BeginExperiment() error = %v", err)
		}
		if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run要約"); err != nil {
			t.Fatalf("CompleteExperimentRun() error = %v", err)
		}
		calls := 0
		replaceBriefingRandom(t, func(bytes []byte) (int, error) {
			calls++
			if calls > 1 {
				return 0, errors.New("random failed")
			}
			for index := range bytes {
				bytes[index] = byte(index)
			}
			return len(bytes), nil
		})
		if _, _, err := store.beginRunEvaluation(context.Background(), "evaluation-request", start.Runs[0].ID); err == nil {
			t.Error("beginRunEvaluation() error = nil, want operation ID failure")
		}
	})
}

// SQLite run評価の競合とtransaction境界をdriverで再現する。
func TestStoreBeginRunEvaluationDriverFailures(t *testing.T) {
	tests := []runEvaluationDriverTestCase{
		runEvaluationDriverCase("初期読込失敗", runEvaluationInitialReadError),
		runEvaluationDriverCase("transaction開始失敗", runEvaluationBeginError),
		runEvaluationDriverCase("評価対象読込失敗", runEvaluationEvaluableReadError),
		runEvaluationDriverCase("既存評価読込失敗", runEvaluationExistingReadError),
		runEvaluationDriverSuccessCase("評価一意競合後にsnapshotを再読込", runEvaluationEvaluationConflictReplay),
		runEvaluationDriverCase("評価一意競合後のsnapshot読込失敗", runEvaluationEvaluationConflictReadError),
		runEvaluationDriverCase("評価一意競合後にsnapshotがない", runEvaluationEvaluationConflictMissing),
		runEvaluationDriverCase("評価一意競合後のrollback失敗", runEvaluationEvaluationConflictRollback),
		runEvaluationDriverSuccessCase("request一意競合後にsnapshotを再読込", runEvaluationRequestConflictReplay),
		runEvaluationDriverCase("request一意競合後に別runを拒否", runEvaluationRequestConflictOther),
		runEvaluationDriverCase("request一意競合後のsnapshot読込失敗", runEvaluationRequestConflictReadError),
		runEvaluationDriverCase("request一意競合後のrollback失敗", runEvaluationRequestConflictRollback),
		runEvaluationDriverCase("操作insertのrun競合を返す", runEvaluationOperationAlreadyExists),
		runEvaluationDriverCase("操作insertの一般失敗を返す", runEvaluationOperationInsertError),
		runEvaluationDriverCase("commit失敗", runEvaluationCommitError),
		runEvaluationDriverCase("commit後snapshot欠落", runEvaluationPostCommitMissing),
		runEvaluationDriverCase("commit後snapshot読込失敗", runEvaluationPostCommitReadError),
		runEvaluationDriverCase("busy再試行上限", runEvaluationBusy),
		runEvaluationDriverCanceledCase("busy待機中の中止", runEvaluationBusy),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newRunEvaluationFailureStore(t, tt.stage)
			ctx := context.Background()
			if tt.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				time.AfterFunc(time.Microsecond, cancel)
				defer cancel()
			}
			_, _, err := store.BeginRunEvaluation(ctx, "request-1", "run-1")
			if gotSuccess := err == nil; gotSuccess != tt.wantSuccess {
				t.Errorf("BeginRunEvaluation() error = %v, want success = %v", err, tt.wantSuccess)
			}
		})
	}
}

type runEvaluationDriverTestCase struct {
	name        string
	stage       runEvaluationFailureStage
	cancel      bool
	wantSuccess bool
}

func runEvaluationDriverCase(name string, stage runEvaluationFailureStage) runEvaluationDriverTestCase {
	return runEvaluationDriverTestCase{
		name:  name,
		stage: stage,
	}
}

func runEvaluationDriverSuccessCase(name string, stage runEvaluationFailureStage) runEvaluationDriverTestCase {
	result := runEvaluationDriverCase(name, stage)
	result.wantSuccess = true
	return result
}

func runEvaluationDriverCanceledCase(name string, stage runEvaluationFailureStage) runEvaluationDriverTestCase {
	result := runEvaluationDriverCase(name, stage)
	result.cancel = true
	return result
}

type runEvaluationFailureStage string

const (
	runEvaluationInitialReadError            runEvaluationFailureStage = "initial-read-error"
	runEvaluationBeginError                  runEvaluationFailureStage = "begin-error"
	runEvaluationEvaluableReadError          runEvaluationFailureStage = "evaluable-read-error"
	runEvaluationExistingReadError           runEvaluationFailureStage = "existing-read-error"
	runEvaluationEvaluationConflictReplay    runEvaluationFailureStage = "evaluation-conflict-replay"
	runEvaluationEvaluationConflictReadError runEvaluationFailureStage = "evaluation-conflict-read-error"
	runEvaluationEvaluationConflictMissing   runEvaluationFailureStage = "evaluation-conflict-missing"
	runEvaluationEvaluationConflictRollback  runEvaluationFailureStage = "evaluation-conflict-rollback"
	runEvaluationRequestConflictReplay       runEvaluationFailureStage = "request-conflict-replay"
	runEvaluationRequestConflictOther        runEvaluationFailureStage = "request-conflict-other"
	runEvaluationRequestConflictReadError    runEvaluationFailureStage = "request-conflict-read-error"
	runEvaluationRequestConflictRollback     runEvaluationFailureStage = "request-conflict-rollback"
	runEvaluationOperationAlreadyExists      runEvaluationFailureStage = "operation-already-exists"
	runEvaluationOperationInsertError        runEvaluationFailureStage = "operation-insert-error"
	runEvaluationCommitError                 runEvaluationFailureStage = "commit-error"
	runEvaluationPostCommitMissing           runEvaluationFailureStage = "post-commit-missing"
	runEvaluationPostCommitReadError         runEvaluationFailureStage = "post-commit-read-error"
	runEvaluationBusy                        runEvaluationFailureStage = "busy"
)

const runEvaluationFailureDriverName = "context-lab-run-evaluation-failure"

var runEvaluationFailureDriverOnce sync.Once

// run評価失敗注入用storeを生成する。
func newRunEvaluationFailureStore(t *testing.T, stage runEvaluationFailureStage) *Store {
	t.Helper()
	runEvaluationFailureDriverOnce.Do(func() { sql.Register(runEvaluationFailureDriverName, runEvaluationFailureDriver{}) })
	database, err := sql.Open(runEvaluationFailureDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	return &Store{db: database}
}

type runEvaluationFailureDriver struct{}

func (runEvaluationFailureDriver) Open(stage string) (driver.Conn, error) {
	return &runEvaluationFailureConnection{stage: runEvaluationFailureStage(stage)}, nil
}

type runEvaluationFailureConnection struct {
	stage          runEvaluationFailureStage
	operationReads int
}

func (*runEvaluationFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (*runEvaluationFailureConnection) Close() error                { return nil }
func (c *runEvaluationFailureConnection) Begin() (driver.Tx, error) { return c.begin() }
func (c *runEvaluationFailureConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}
func (c *runEvaluationFailureConnection) begin() (driver.Tx, error) {
	if c.stage == runEvaluationBeginError {
		return nil, errors.New("begin failed")
	}
	return runEvaluationFailureTransaction{stage: c.stage}, nil
}

func (c *runEvaluationFailureConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "FROM experiment_evaluation_operations o JOIN"):
		c.operationReads++
		if c.stage == runEvaluationBusy {
			return nil, errors.New("database is busy")
		}
		if c.stage == runEvaluationInitialReadError || (c.stage == runEvaluationEvaluationConflictReadError && c.operationReads > 1) || (c.stage == runEvaluationRequestConflictReadError && c.operationReads > 1) || (c.stage == runEvaluationPostCommitReadError && c.operationReads > 1) {
			return nil, errors.New("operation read failed")
		}
		if c.operationReads > 1 && (c.stage == runEvaluationEvaluationConflictReplay || c.stage == runEvaluationRequestConflictReplay || c.stage == runEvaluationRequestConflictOther) {
			runID := "run-1"
			if c.stage == runEvaluationRequestConflictOther {
				runID = "other-run"
			}
			return runEvaluationOperationRows(runID), nil
		}
		return &runEvaluationRows{columns: runEvaluationOperationColumns()}, nil
	case strings.Contains(query, "FROM experiment_runs r JOIN"):
		if c.stage == runEvaluationEvaluableReadError {
			return nil, errors.New("evaluable read failed")
		}
		return &runEvaluationRows{
			columns: []string{
				"id",
				"experiment_id",
				"state",
				"summary",
				"purpose",
				"evaluation_axes",
			},
			values: [][]driver.Value{
				{
					"run-1",
					"experiment-1",
					"completed",
					"summary",
					"purpose",
					"axes",
				},
			},
		}, nil
	case strings.Contains(query, "SELECT id FROM experiment_evaluations"):
		if c.stage == runEvaluationExistingReadError {
			return nil, errors.New("existing read failed")
		}
		return &runEvaluationRows{columns: []string{"id"}}, nil
	default:
		return nil, errors.New("unexpected query")
	}
}

func (c *runEvaluationFailureConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "INSERT INTO experiment_evaluations") {
		switch c.stage {
		case runEvaluationEvaluationConflictReplay, runEvaluationEvaluationConflictReadError, runEvaluationEvaluationConflictMissing, runEvaluationEvaluationConflictRollback:
			return nil, errors.New("UNIQUE constraint failed: experiment_evaluations.run_id")
		}
	}
	if strings.Contains(query, "INSERT INTO experiment_evaluation_operations") {
		switch c.stage {
		case runEvaluationRequestConflictReplay, runEvaluationRequestConflictOther, runEvaluationRequestConflictReadError, runEvaluationRequestConflictRollback:
			return nil, errors.New("UNIQUE constraint failed: experiment_evaluation_operations.request_id")
		case runEvaluationOperationAlreadyExists:
			return nil, errors.New("UNIQUE constraint failed: experiment_evaluations.run_id")
		case runEvaluationOperationInsertError:
			return nil, errors.New("operation insert failed")
		}
	}
	return driver.RowsAffected(1), nil
}

type runEvaluationFailureTransaction struct{ stage runEvaluationFailureStage }

func (t runEvaluationFailureTransaction) Commit() error {
	if t.stage == runEvaluationCommitError {
		return errors.New("commit failed")
	}
	return nil
}
func (t runEvaluationFailureTransaction) Rollback() error {
	if t.stage == runEvaluationEvaluationConflictRollback || t.stage == runEvaluationRequestConflictRollback {
		return errors.New("rollback failed")
	}
	return nil
}

func runEvaluationOperationColumns() []string {
	return []string{
		"request_id",
		"run_id",
		"evaluation_id",
		"operation_id",
		"state",
		"failure_code",
		"summary",
		"updated_at",
	}
}
func runEvaluationOperationRows(runID string) driver.Rows {
	return &runEvaluationRows{
		columns: runEvaluationOperationColumns(),
		values: [][]driver.Value{
			{
				"request-1",
				runID,
				"evaluation-1",
				"operation-1",
				"starting",
				"",
				nil,
				"2026-01-01T00:00:00Z",
			},
		},
	}
}

type runEvaluationRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *runEvaluationRows) Columns() []string { return r.columns }
func (*runEvaluationRows) Close() error        { return nil }
func (r *runEvaluationRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
