package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// 結論確定の正本、replay、競合を検証する。
func TestStoreFinalizeExperimentConclusion(t *testing.T) {
	store, fixed := finalizedConclusionStore(t)
	first, created, err := store.FinalizeExperimentConclusion(context.Background(), "conclusion-request", fixed.ExperimentID, "結論")
	if err != nil || !created {
		t.Fatalf("FinalizeExperimentConclusion() = (%+v, %v, %v), want created result", first, created, err)
	}
	replayed, created, err := store.FinalizeExperimentConclusion(context.Background(), "conclusion-request", fixed.ExperimentID, "結論")
	if err != nil || created || replayed.ConclusionID != first.ConclusionID {
		t.Errorf("replay = (%+v, %v, %v), want persisted result", replayed, created, err)
	}
	_, _, err = store.FinalizeExperimentConclusion(context.Background(), "other-request", fixed.ExperimentID, "別結論")
	if err == nil {
		t.Error("different conclusion error = nil, want already finalized")
	}
}

// 同一requestの並行確定を検証する。
func TestStoreFinalizeExperimentConclusionConcurrently(t *testing.T) {
	store, fixed := finalizedConclusionStore(t)
	const workers = 20
	results := make(chan error, workers)
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			_, _, err := store.FinalizeExperimentConclusion(context.Background(), "concurrent-request", fixed.ExperimentID, "結論")
			results <- err
		}()
	}
	group.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Errorf("FinalizeExperimentConclusion() error = %v", err)
		}
	}
}

// 結論確定には評価が終了済みかつconfirmedであることを検証する。
func TestStoreFinalizeExperimentConclusionRequiresTerminalConfirmedEvaluations(t *testing.T) {
	tests := []struct {
		name           string
		state          string
		reconciliation string
		wantCode       apperr.Code
	}{
		{
			name:           "startingは結果記録済みでも未確定",
			state:          "starting",
			reconciliation: "confirmed",
			wantCode:       apperr.CodeExperimentConclusionNotReady,
		},
		{
			name:           "reconcilingは終了済みでも未確定",
			state:          "completed",
			reconciliation: "reconciling",
			wantCode:       apperr.CodeExperimentConclusionNotReady,
		},
		{
			name:           "failedかつconfirmedは確定可能",
			state:          "failed",
			reconciliation: "confirmed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, fixed := finalizedConclusionStore(t)
			if _, err := store.db.Exec(
				"UPDATE experiment_evaluations SET state = ?, reconciliation_state = ? WHERE experiment_id = ?",
				tt.state,
				tt.reconciliation,
				fixed.ExperimentID,
			); err != nil {
				t.Fatalf("update evaluation state error = %v", err)
			}
			_, _, err := store.FinalizeExperimentConclusion(context.Background(), "terminal-request", fixed.ExperimentID, "結論")
			if tt.wantCode == "" {
				if err != nil {
					t.Errorf("FinalizeExperimentConclusion() error = %v", err)
				}
				return
			}
			if !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("FinalizeExperimentConclusion() error = %v, want code %q", err, tt.wantCode)
			}
		})
	}
}

// 識別子生成失敗時に結論を保存しないことを検証する。
func TestStoreFinalizeExperimentConclusionIdentifierFailure(t *testing.T) {
	store, fixed := finalizedConclusionStore(t)
	previous := readBriefingRandom
	t.Cleanup(func() { readBriefingRandom = previous })
	readBriefingRandom = func([]byte) (int, error) { return 0, errors.New("random failed") }
	_, _, err := store.FinalizeExperimentConclusion(context.Background(), "identifier-request", fixed.ExperimentID, "結論")
	if err == nil {
		t.Error("FinalizeExperimentConclusion() error = nil, want identifier failure")
	}
}

// 二つのStoreとpoolで同一requestの確定がsnapshotへ収束することを検証する。
func TestStoreFinalizeExperimentConclusionAcrossStoresConcurrently(t *testing.T) {
	directory := t.TempDir()
	firstStore, err := Open(directory)
	if err != nil {
		t.Fatalf("Open(first) error = %v", err)
	}
	t.Cleanup(func() { _ = firstStore.Close() })
	secondStore, err := Open(filepath.Clean(directory))
	if err != nil {
		t.Fatalf("Open(second) error = %v", err)
	}
	t.Cleanup(func() { _ = secondStore.Close() })
	fixed := fixedConclusionConditions(t, firstStore)
	for iteration := range 20 {
		requestID := fmt.Sprintf("cross-store-request-%d", iteration)
		results := make(chan error, 2)
		var group sync.WaitGroup
		for _, store := range []*Store{
			firstStore,
			secondStore,
		} {
			group.Add(1)
			go func(store *Store) {
				defer group.Done()
				_, _, err := store.FinalizeExperimentConclusion(context.Background(), requestID, fixed.ExperimentID, "結論")
				results <- err
			}(store)
		}
		group.Wait()
		close(results)
		for result := range results {
			if result != nil {
				t.Errorf("iteration %d FinalizeExperimentConclusion() error = %v", iteration, result)
			}
		}
	}
}

// 結論確定可能な評価済みstoreを生成する。
func finalizedConclusionStore(t *testing.T) (*Store, struct{ ExperimentID string }) {
	t.Helper()
	store, fixed := fixedExperimentPreparationStore(t)
	start, _, err := store.BeginExperiment(context.Background(), "conclusion-start", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run summary"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}
	evaluation, _, err := store.BeginRunEvaluation(context.Background(), "conclusion-evaluation", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() error = %v", err)
	}
	if err := store.CompleteRunEvaluation(context.Background(), evaluation.EvaluationID, "evaluation summary"); err != nil {
		t.Fatalf("CompleteRunEvaluation() error = %v", err)
	}
	return store, struct{ ExperimentID string }{ExperimentID: fixed.ExperimentID}
}

// fixedConclusionConditions は既存Storeへ結論確定可能な正本を作る。
func fixedConclusionConditions(t *testing.T, store *Store) struct{ ExperimentID string } {
	t.Helper()
	draft := testExperimentPreparationDraft("cross-draft", "experiment-1", "目的")
	seedExperimentPreparationDraftExperiment(t, store, draft.ExperimentID)
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	fixed, err := store.FixExperimentConditions(context.Background(), fixedConditionsFromDraft(draft))
	if err != nil {
		t.Fatalf("FixExperimentConditions() error = %v", err)
	}
	start, _, err := store.BeginExperiment(context.Background(), "cross-start", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "run summary"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}
	evaluation, _, err := store.BeginRunEvaluation(context.Background(), "cross-evaluation", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() error = %v", err)
	}
	if err := store.CompleteRunEvaluation(context.Background(), evaluation.EvaluationID, "evaluation summary"); err != nil {
		t.Fatalf("CompleteRunEvaluation() error = %v", err)
	}
	return struct{ ExperimentID string }{ExperimentID: fixed.ExperimentID}
}

// SQLite query helperのread/scan/parse失敗と評価snapshotを検証する。
func TestExperimentConclusionQueryHelpers(t *testing.T) {
	tests := []struct {
		name  string
		stage conclusionQueryStage
		check func(*testing.T, *Store)
	}{
		{
			name:  "operationなし",
			stage: conclusionOperationMissing,
			check: func(t *testing.T, store *Store) {
				_, found, err := findExperimentConclusionOperation(context.Background(), store.db, "request")
				if err != nil || found {
					t.Errorf("findExperimentConclusionOperation() = (_, %v, %v), want missing", found, err)
				}
			},
		},
		{
			name:  "operation query失敗",
			stage: conclusionOperationQueryError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := findExperimentConclusionOperation(context.Background(), store.db, "request"); err == nil {
					t.Error("findExperimentConclusionOperation() error = nil, want query error")
				}
			},
		},
		{
			name:  "operation scan失敗",
			stage: conclusionOperationScanError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := findExperimentConclusionOperation(context.Background(), store.db, "request"); err == nil {
					t.Error("findExperimentConclusionOperation() error = nil, want scan error")
				}
			},
		},
		{
			name:  "operation時刻不正",
			stage: conclusionOperationTimeError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := findExperimentConclusionOperation(context.Background(), store.db, "request"); err == nil {
					t.Error("findExperimentConclusionOperation() error = nil, want parse error")
				}
			},
		},
		{
			name:  "operation成功",
			stage: conclusionOperationSuccess,
			check: func(t *testing.T, store *Store) {
				got, found, err := findExperimentConclusionOperation(context.Background(), store.db, "request")
				if err != nil || !found || got.State != "finalized" {
					t.Errorf("findExperimentConclusionOperation() = (%+v, %v, %v), want finalized", got, found, err)
				}
			},
		},
		{
			name:  "結論なし",
			stage: conclusionMissing,
			check: func(t *testing.T, store *Store) {
				_, found, err := findExperimentConclusion(context.Background(), store.db, "experiment")
				if err != nil || found {
					t.Errorf("findExperimentConclusion() = (_, %v, %v), want missing", found, err)
				}
			},
		},
		{
			name:  "結論query失敗",
			stage: conclusionQueryError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := findExperimentConclusion(context.Background(), store.db, "experiment"); err == nil {
					t.Error("findExperimentConclusion() error = nil, want query error")
				}
			},
		},
		{
			name:  "結論scan失敗",
			stage: conclusionScanError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := findExperimentConclusion(context.Background(), store.db, "experiment"); err == nil {
					t.Error("findExperimentConclusion() error = nil, want scan error")
				}
			},
		},
		{
			name:  "結論時刻不正",
			stage: conclusionTimeError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := findExperimentConclusion(context.Background(), store.db, "experiment"); err == nil {
					t.Error("findExperimentConclusion() error = nil, want parse error")
				}
			},
		},
		{
			name:  "結論成功",
			stage: conclusionSuccess,
			check: func(t *testing.T, store *Store) {
				got, found, err := findExperimentConclusion(context.Background(), store.db, "experiment")
				if err != nil || !found || got.ExperimentID != "experiment" {
					t.Errorf("findExperimentConclusion() = (%+v, %v, %v), want persisted", got, found, err)
				}
			},
		},
		{
			name:  "評価query失敗",
			stage: conclusionEvaluationQueryError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment"); err == nil {
					t.Error("experimentEvaluationSnapshotDigest() error = nil, want query error")
				}
			},
		},
		{
			name:  "評価scan失敗",
			stage: conclusionEvaluationScanError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment"); err == nil {
					t.Error("experimentEvaluationSnapshotDigest() error = nil, want scan error")
				}
			},
		},
		{
			name:  "評価rows失敗",
			stage: conclusionEvaluationRowsError,
			check: func(t *testing.T, store *Store) {
				if _, _, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment"); err == nil {
					t.Error("experimentEvaluationSnapshotDigest() error = nil, want rows error")
				}
			},
		},
		{
			name:  "評価なし",
			stage: conclusionEvaluationMissing,
			check: func(t *testing.T, store *Store) {
				_, ready, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment")
				if err != nil || ready {
					t.Errorf("experimentEvaluationSnapshotDigest() = (_, %v, %v), want not ready", ready, err)
				}
			},
		},
		{
			name:  "未記録評価",
			stage: conclusionEvaluationNotRecorded,
			check: func(t *testing.T, store *Store) {
				_, ready, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment")
				if err != nil || ready {
					t.Errorf("experimentEvaluationSnapshotDigest() = (_, %v, %v), want not ready", ready, err)
				}
			},
		},
		{
			name:  "reconciling評価",
			stage: conclusionEvaluationReconciling,
			check: func(t *testing.T, store *Store) {
				_, ready, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment")
				if err != nil || ready {
					t.Errorf("experimentEvaluationSnapshotDigest() = (_, %v, %v), want not ready", ready, err)
				}
			},
		},
		{
			name:  "非終了評価",
			stage: conclusionEvaluationNonterminal,
			check: func(t *testing.T, store *Store) {
				_, ready, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment")
				if err != nil || ready {
					t.Errorf("experimentEvaluationSnapshotDigest() = (_, %v, %v), want not ready", ready, err)
				}
			},
		},
		{
			name:  "評価成功",
			stage: conclusionEvaluationSuccess,
			check: func(t *testing.T, store *Store) {
				digest, ready, err := experimentEvaluationSnapshotDigest(context.Background(), store.db, "experiment")
				if err != nil || !ready || digest == "" {
					t.Errorf("experimentEvaluationSnapshotDigest() = (%q, %v, %v), want digest", digest, ready, err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { tt.check(t, newConclusionQueryStore(t, tt.stage)) })
	}
}

// 結論確定transaction境界の失敗を検証する。
func TestFinalizeExperimentConclusionTransactionFailures(t *testing.T) {
	for _, stage := range []conclusionQueryStage{
		conclusionFinalizeBegin,
		conclusionFinalizeOperation,
		conclusionFinalizeMissing,
		conclusionFinalizeQuery,
		conclusionFinalizeNotFixed,
		conclusionFinalizeDifferent,
		conclusionFinalizeNotReady,
		conclusionFinalizeExec,
		conclusionFinalizeCommit,
		conclusionQueryError,
		conclusionEvaluationQueryError,
	} {
		t.Run(string(stage), func(t *testing.T) {
			store := newConclusionQueryStore(t, stage)
			_, _, err := store.FinalizeExperimentConclusion(context.Background(), "request", "experiment", "text")
			if err == nil {
				t.Error("FinalizeExperimentConclusion() error = nil, want failure")
			}
		})
	}
}

// SQLite busy時に再試行待機のcontext終了を返すことを検証する。
func TestFinalizeExperimentConclusionBusyContextCancellation(t *testing.T) {
	store := newConclusionQueryStore(t, conclusionOuterBusyCancellation)
	ctx, cancel := context.WithCancel(context.Background())
	previous := conclusionQueryCancel
	t.Cleanup(func() { conclusionQueryCancel = previous })
	conclusionQueryCancel = cancel
	_, _, err := store.FinalizeExperimentConclusion(ctx, "request", "experiment", "text")
	if err == nil {
		t.Error("FinalizeExperimentConclusion() error = nil, want context cancellation")
	}
}

// 結論確定の外側再試行分岐を検証する。
func TestStoreFinalizeExperimentConclusionRetryBranches(t *testing.T) {
	tests := []struct {
		name       string
		stage      conclusionQueryStage
		requestID  string
		experiment string
		conclusion string
		limit      int
		wantCode   apperr.Code
		wantError  bool
	}{
		{
			name:       "競合後の同一operationを返す",
			stage:      conclusionOuterBusyExisting,
			requestID:  "request",
			experiment: "experiment",
			conclusion: "text",
		},
		{
			name:       "競合後の別内容operationを拒否する",
			stage:      conclusionOuterBusyConflict,
			requestID:  "request",
			experiment: "experiment",
			conclusion: "text",
			wantCode:   apperr.CodeExperimentConclusionRequestConflict,
		},
		{
			name:       "競合後のoperation読込失敗を返す",
			stage:      conclusionOuterBusyFindError,
			requestID:  "request",
			experiment: "experiment",
			conclusion: "text",
			wantError:  true,
		},
		{
			name:       "再試行上限で最後の競合を返す",
			stage:      conclusionFinalizeBusy,
			requestID:  "request",
			experiment: "experiment",
			conclusion: "text",
			limit:      1,
			wantError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := experimentConclusionRetryLimit
			if tt.limit != 0 {
				experimentConclusionRetryLimit = tt.limit
			}
			t.Cleanup(func() { experimentConclusionRetryLimit = previous })
			got, _, err := newConclusionQueryStore(t, tt.stage).FinalizeExperimentConclusion(context.Background(), tt.requestID, tt.experiment, tt.conclusion)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("FinalizeExperimentConclusion() error = %v, want code %q", err, tt.wantCode)
				}
				return
			}
			if tt.wantError {
				if err == nil {
					t.Error("FinalizeExperimentConclusion() error = nil, want error")
				}
				return
			}
			if err != nil || got.RequestID != tt.requestID {
				t.Errorf("FinalizeExperimentConclusion() = (%+v, _, %v), want replayed operation", got, err)
			}
		})
	}
}

// 結論確定transactionの未到達分岐を検証する。
func TestFinalizeExperimentConclusionRemainingTransactionBranches(t *testing.T) {
	tests := []struct {
		name  string
		stage conclusionQueryStage
	}{
		{
			name:  "既存operationの別内容を拒否する",
			stage: conclusionFinalizeOperationConflict,
		},
		{
			name:  "既存結論replayのcommit失敗を返す",
			stage: conclusionFinalizeReplayCommit,
		},
		{
			name:  "結論operation保存失敗を返す",
			stage: conclusionFinalizeOperationInsert,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := newConclusionQueryStore(t, tt.stage).finalizeExperimentConclusion(context.Background(), "request", "experiment", "text")
			if err == nil {
				t.Error("finalizeExperimentConclusion() error = nil, want error")
			}
		})
	}
}

type conclusionQueryStage string

const (
	conclusionOperationMissing          conclusionQueryStage = "operation-missing"
	conclusionOperationQueryError       conclusionQueryStage = "operation-query-error"
	conclusionOperationScanError        conclusionQueryStage = "operation-scan-error"
	conclusionOperationTimeError        conclusionQueryStage = "operation-time-error"
	conclusionOperationSuccess          conclusionQueryStage = "operation-success"
	conclusionMissing                   conclusionQueryStage = "conclusion-missing"
	conclusionQueryError                conclusionQueryStage = "conclusion-query-error"
	conclusionScanError                 conclusionQueryStage = "conclusion-scan-error"
	conclusionTimeError                 conclusionQueryStage = "conclusion-time-error"
	conclusionSuccess                   conclusionQueryStage = "conclusion-success"
	conclusionEvaluationQueryError      conclusionQueryStage = "evaluation-query-error"
	conclusionEvaluationScanError       conclusionQueryStage = "evaluation-scan-error"
	conclusionEvaluationRowsError       conclusionQueryStage = "evaluation-rows-error"
	conclusionEvaluationMissing         conclusionQueryStage = "evaluation-missing"
	conclusionEvaluationNotRecorded     conclusionQueryStage = "evaluation-not-recorded"
	conclusionEvaluationReconciling     conclusionQueryStage = "evaluation-reconciling"
	conclusionEvaluationNonterminal     conclusionQueryStage = "evaluation-nonterminal"
	conclusionEvaluationSuccess         conclusionQueryStage = "evaluation-success"
	conclusionFinalizeBegin             conclusionQueryStage = "finalize-begin"
	conclusionFinalizeOperation         conclusionQueryStage = "finalize-operation-error"
	conclusionFinalizeMissing           conclusionQueryStage = "finalize-experiment-missing"
	conclusionFinalizeQuery             conclusionQueryStage = "finalize-experiment-error"
	conclusionFinalizeNotFixed          conclusionQueryStage = "finalize-not-fixed"
	conclusionFinalizeDifferent         conclusionQueryStage = "finalize-different"
	conclusionFinalizeNotReady          conclusionQueryStage = "finalize-not-ready"
	conclusionFinalizeExec              conclusionQueryStage = "finalize-exec-error"
	conclusionFinalizeCommit            conclusionQueryStage = "finalize-commit-error"
	conclusionFinalizeBusy              conclusionQueryStage = "finalize-busy"
	conclusionOuterBusyExisting         conclusionQueryStage = "outer-busy-existing"
	conclusionOuterBusyConflict         conclusionQueryStage = "outer-busy-conflict"
	conclusionOuterBusyFindError        conclusionQueryStage = "outer-busy-find-error"
	conclusionOuterBusyCancellation     conclusionQueryStage = "outer-busy-cancellation"
	conclusionFinalizeOperationConflict conclusionQueryStage = "finalize-operation-conflict"
	conclusionFinalizeReplayCommit      conclusionQueryStage = "finalize-replay-commit-error"
	conclusionFinalizeOperationInsert   conclusionQueryStage = "finalize-operation-insert-error"
)

const conclusionQueryDriverName = "context-lab-conclusion-query"

var conclusionQueryDriverOnce sync.Once

var conclusionQueryCancel func()

// newConclusionQueryStore は結論query helperのdriver注入storeを生成する。
func newConclusionQueryStore(t *testing.T, stage conclusionQueryStage) *Store {
	t.Helper()
	conclusionQueryDriverOnce.Do(func() { sql.Register(conclusionQueryDriverName, conclusionQueryDriver{}) })
	database, err := sql.Open(conclusionQueryDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return &Store{db: database}
}

type conclusionQueryDriver struct{}

func (conclusionQueryDriver) Open(stage string) (driver.Conn, error) {
	return conclusionQueryConnection{stage: conclusionQueryStage(stage)}, nil
}

type conclusionQueryConnection struct{ stage conclusionQueryStage }

func (conclusionQueryConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (conclusionQueryConnection) Close() error { return nil }
func (c conclusionQueryConnection) Begin() (driver.Tx, error) {
	if c.stage == conclusionFinalizeBegin {
		return nil, errors.New("begin failed")
	}
	if c.stage == conclusionFinalizeBusy || c.stage == conclusionOuterBusyExisting || c.stage == conclusionOuterBusyConflict || c.stage == conclusionOuterBusyFindError || c.stage == conclusionOuterBusyCancellation {
		return nil, errors.New("database is locked")
	}
	return conclusionQueryTransaction(c), nil
}

func (c conclusionQueryConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "experiment_conclusion_operations") {
		return c.operationRows()
	}
	if strings.Contains(query, "FROM experiments WHERE") {
		return c.experimentRows()
	}
	if strings.Contains(query, "experiment_conclusions") {
		return c.conclusionRows()
	}
	if strings.Contains(query, "FROM experiment_evaluations") {
		return c.evaluationRows()
	}
	return nil, errors.New("unexpected query")
}

func (c conclusionQueryConnection) experimentRows() (driver.Rows, error) {
	if c.stage == conclusionFinalizeQuery {
		return nil, errors.New("experiment query failed")
	}
	rows := &conclusionQueryRows{columns: []string{"fixed_condition_id"}}
	if c.stage == conclusionFinalizeMissing {
		return rows, nil
	}
	if c.stage == conclusionFinalizeNotFixed {
		rows.values = [][]driver.Value{{nil}}
		return rows, nil
	}
	rows.values = [][]driver.Value{{"fixed"}}
	return rows, nil
}

func (c conclusionQueryConnection) operationRows() (driver.Rows, error) {
	if c.stage == conclusionOperationQueryError || c.stage == conclusionFinalizeOperation || c.stage == conclusionOuterBusyFindError {
		return nil, errors.New("operation query failed")
	}
	if c.stage == conclusionOuterBusyCancellation && conclusionQueryCancel != nil {
		conclusionQueryCancel()
	}
	rows := &conclusionQueryRows{
		columns: []string{
			"request_id",
			"experiment_id",
			"conclusion_id",
			"conclusion",
			"evaluation_snapshot_digest",
			"finalized_at",
		},
	}
	if c.stage != conclusionOperationSuccess && c.stage != conclusionOperationTimeError && c.stage != conclusionOperationScanError && c.stage != conclusionOuterBusyExisting && c.stage != conclusionOuterBusyConflict && c.stage != conclusionFinalizeOperationConflict {
		return rows, nil
	}
	value := "2026-08-10T00:00:00Z"
	if c.stage == conclusionOperationTimeError {
		value = "invalid"
	}
	values := []driver.Value{
		"request",
		"experiment",
		"conclusion",
		"text",
		"digest",
		value,
	}
	if c.stage == conclusionOperationScanError {
		values[0] = nil
	}
	if c.stage == conclusionOuterBusyConflict || c.stage == conclusionFinalizeOperationConflict {
		values[3] = "other"
	}
	rows.values = [][]driver.Value{
		values,
	}
	return rows, nil
}

func (c conclusionQueryConnection) conclusionRows() (driver.Rows, error) {
	if c.stage == conclusionQueryError {
		return nil, errors.New("conclusion query failed")
	}
	rows := &conclusionQueryRows{
		columns: []string{
			"id",
			"conclusion",
			"evaluation_snapshot_digest",
			"state",
			"finalized_at",
		},
	}
	if c.stage == conclusionMissing || c.stage == conclusionFinalizeNotReady || c.stage == conclusionFinalizeExec || c.stage == conclusionFinalizeCommit || c.stage == conclusionEvaluationQueryError || c.stage == conclusionFinalizeOperationInsert {
		return rows, nil
	}
	value := "2026-08-10T00:00:00Z"
	if c.stage == conclusionTimeError {
		value = "invalid"
	}
	values := []driver.Value{
		"conclusion",
		"text",
		"digest",
		"finalized",
		value,
	}
	if c.stage == conclusionFinalizeReplayCommit {
		rows.values = [][]driver.Value{values}
		return rows, nil
	}
	if c.stage == conclusionScanError {
		values[0] = nil
	}
	if c.stage == conclusionFinalizeDifferent {
		values[1] = "other"
	}
	rows.values = [][]driver.Value{
		values,
	}
	return rows, nil
}

func (c conclusionQueryConnection) evaluationRows() (driver.Rows, error) {
	if c.stage == conclusionEvaluationQueryError {
		return nil, errors.New("evaluation query failed")
	}
	rows := &conclusionQueryRows{
		columns: []string{
			"id",
			"run_id",
			"state",
			"result_status",
			"summary",
			"result_reason_code",
			"reconciliation_state",
			"last_observed_at",
		},
	}
	if c.stage == conclusionEvaluationRowsError {
		rows.nextErr = errors.New("evaluation rows failed")
		return rows, nil
	}
	if c.stage == conclusionEvaluationMissing || c.stage == conclusionFinalizeNotReady {
		return rows, nil
	}
	values := []driver.Value{
		"evaluation",
		"run",
		"completed",
		"complete",
		"summary",
		"",
		"confirmed",
		"2026-08-10T00:00:00Z",
	}
	if c.stage == conclusionEvaluationScanError {
		values[0] = nil
	}
	if c.stage == conclusionEvaluationNotRecorded {
		values[3] = "notRecorded"
	}
	if c.stage == conclusionEvaluationReconciling {
		values[6] = "reconciling"
	}
	if c.stage == conclusionEvaluationNonterminal {
		values[2] = "running"
	}
	rows.values = [][]driver.Value{
		values,
	}
	return rows, nil
}

func (c conclusionQueryConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if c.stage == conclusionFinalizeExec || (c.stage == conclusionFinalizeOperationInsert && strings.Contains(query, "experiment_conclusion_operations")) {
		return nil, errors.New("exec failed")
	}
	return driver.RowsAffected(1), nil
}

type conclusionQueryTransaction struct{ stage conclusionQueryStage }

func (t conclusionQueryTransaction) Commit() error {
	if t.stage == conclusionFinalizeCommit || t.stage == conclusionFinalizeReplayCommit {
		return errors.New("commit failed")
	}
	return nil
}

func (conclusionQueryTransaction) Rollback() error { return nil }

type conclusionQueryRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	nextErr error
}

func (r *conclusionQueryRows) Columns() []string { return r.columns }
func (*conclusionQueryRows) Close() error        { return nil }
func (r *conclusionQueryRows) Next(destination []driver.Value) error {
	if r.index >= len(r.values) {
		if r.nextErr != nil {
			return r.nextErr
		}
		return io.EOF
	}
	copy(destination, r.values[r.index])
	r.index++
	return nil
}
