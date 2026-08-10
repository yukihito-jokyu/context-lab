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

// SQLite実験開始の全run生成、冪等性、失敗記録。
func TestStoreBeginExperiment(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	start, created, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if !created {
		t.Fatal("created = false, want true")
	}
	if start.State != domain.ExperimentStartStateStarting || len(start.Runs) != len(fixed.Prompts) {
		t.Errorf("start = %+v, want starting and %d runs", start, len(fixed.Prompts))
	}
	for _, run := range start.Runs {
		if run.ID == "" || run.State != domain.ExperimentRunStateQueued {
			t.Errorf("run = %+v, want queued persisted run", run)
		}
		assertExperimentRunArtifactState(t, store, run.ID, domain.ExperimentRunArtifactStatusNotRecorded, "")
	}
	if err := store.MarkExperimentRunRunning(context.Background(), start.Runs[0].ID); err != nil {
		t.Fatalf("MarkExperimentRunRunning() error = %v", err)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "安全な要約"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}
	if err := store.FailExperimentRun(context.Background(), start.Runs[1].ID, string(apperr.CodeExperimentStartFailed)); err != nil {
		t.Fatalf("FailExperimentRun() error = %v", err)
	}
	assertExperimentRunArtifactState(t, store, start.Runs[0].ID, domain.ExperimentRunArtifactStatusComplete, "")
	assertExperimentRunArtifactState(t, store, start.Runs[1].ID, domain.ExperimentRunArtifactStatusPartial, string(apperr.CodeExperimentStartFailed))

	replayed, created, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("replayed BeginExperiment() error = %v", err)
	}
	if created || replayed.State != domain.ExperimentStartStateFailed || len(replayed.Runs) != len(fixed.Prompts) {
		t.Errorf("replayed = %+v, created = %v, want failed snapshot", replayed, created)
	}
	if _, _, err := store.BeginExperiment(context.Background(), "start-request", "another-experiment"); !apperr.IsCode(err, apperr.CodeExperimentStartRequestInvalid) {
		t.Errorf("different experiment BeginExperiment() error = %v, want code %q", err, apperr.CodeExperimentStartRequestInvalid)
	}
}

// run artifact状態の永続値検証。
func assertExperimentRunArtifactState(t *testing.T, store *Store, runID, wantStatus, wantReason string) {
	t.Helper()

	var status string
	var reason sql.NullString
	if err := store.db.QueryRow("SELECT artifact_status, artifact_reason_code FROM experiment_runs WHERE id = ?", runID).Scan(&status, &reason); err != nil {
		t.Fatalf("artifact state query error = %v", err)
	}
	if status != wantStatus {
		t.Errorf("artifact status = %q, want %q", status, wantStatus)
	}
	if gotReason := reason.String; gotReason != wantReason {
		t.Errorf("artifact reason = %q, want %q", gotReason, wantReason)
	}
}

// 同一request IDを別々のdatabase/sql poolから同時に送っても一つの開始snapshotへ収束する。
func TestStoreBeginExperimentConvergesConcurrentRequestsAcrossStores(t *testing.T) {
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

			type result struct {
				start domain.ExperimentStart
				err   error
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
					start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
					results <- result{
						start: start,
						err:   err,
					}
				}(store)
			}
			<-entered
			<-entered
			close(ready)
			waitGroup.Wait()
			close(results)

			var starts []domain.ExperimentStart
			for result := range results {
				if result.err != nil {
					t.Fatalf("BeginExperiment() error = %v", result.err)
				}
				starts = append(starts, result.start)
			}
			if !reflect.DeepEqual(starts[0], starts[1]) {
				t.Errorf("concurrent snapshots = %+v, want identical", starts)
			}
			var operations int
			if err := firstStore.db.QueryRow("SELECT COUNT(*) FROM experiment_start_operations WHERE experiment_id = ?", fixed.ExperimentID).Scan(&operations); err != nil {
				t.Fatalf("count operations error = %v", err)
			}
			if operations != 1 {
				t.Errorf("start operations = %d, want 1", operations)
			}
		})
	}
}

// SQLite実験開始のready状態検証。
func TestStoreBeginExperimentRequiresReadyFixedConditions(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	_, _, err := store.BeginExperiment(context.Background(), "start-request", "experiment-1")
	if !apperr.IsCode(err, apperr.CodeExperimentStartNotReady) {
		t.Errorf("BeginExperiment() error = %v, want code %q", err, apperr.CodeExperimentStartNotReady)
	}
}

// SQLite run更新と開始完了の成功・失敗。
func TestStoreExperimentRunStateUpdates(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	t.Cleanup(func() { _ = store.Close() })
	start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err := store.CompleteExperimentStart(context.Background(), start.RequestID); err != nil {
		t.Errorf("CompleteExperimentStart() error = %v", err)
	}
	if err := store.FailExperimentRun(context.Background(), start.Runs[0].ID, string(apperr.CodeOperationTimeout)); err != nil {
		t.Errorf("FailExperimentRun() error = %v", err)
	}
	if _, err := store.db.Exec("CREATE TRIGGER fail_complete_run BEFORE UPDATE ON experiment_runs WHEN NEW.state = 'completed' BEGIN SELECT RAISE(ABORT, 'complete run failed'); END"); err != nil {
		t.Fatalf("create complete trigger error = %v", err)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[1].ID, "安全な要約"); err == nil {
		t.Error("CompleteExperimentRun() error = nil, want run update failure")
	}
	if _, err := store.db.Exec("DROP TRIGGER fail_complete_run"); err != nil {
		t.Fatalf("drop complete trigger error = %v", err)
	}
	if _, err := store.db.Exec("CREATE TRIGGER fail_run_failure_insert BEFORE INSERT ON experiment_run_failures BEGIN SELECT RAISE(ABORT, 'failure record failed'); END"); err != nil {
		t.Fatalf("create failure insert trigger error = %v", err)
	}
	if err := store.FailExperimentRun(context.Background(), start.Runs[1].ID, string(apperr.CodeOperationTimeout)); err == nil {
		t.Error("FailExperimentRun() error = nil, want failure record error")
	}
	if _, err := store.db.Exec("DROP TRIGGER fail_run_failure_insert"); err != nil {
		t.Fatalf("drop failure insert trigger error = %v", err)
	}
	if _, err := store.db.Exec("CREATE TRIGGER fail_start_operation_update BEFORE UPDATE ON experiment_start_operations BEGIN SELECT RAISE(ABORT, 'operation update failed'); END"); err != nil {
		t.Fatalf("create trigger error = %v", err)
	}
	if err := store.FailExperimentRun(context.Background(), start.Runs[1].ID, string(apperr.CodeOperationTimeout)); err == nil {
		t.Error("FailExperimentRun() error = nil, want operation update failure")
	}
	if err := store.db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := store.FailExperimentRun(context.Background(), "run-1", "failed"); err == nil {
		t.Error("FailExperimentRun() error = nil, want update failure")
	}
	if err := store.CompleteExperimentStart(context.Background(), "request-1"); err == nil {
		t.Error("CompleteExperimentStart() error = nil, want update failure")
	}
}

// SQLite実験開始の識別子、途中保存、固定条件読込の失敗。
func TestStoreBeginExperimentFailurePaths(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store, domain.ExperimentFixedConditions)
		want    apperr.Code
	}{
		{
			name: "実験が見つからない場合はワークスペース不存在を返す",
			prepare: func(t *testing.T, store *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
				if _, err := store.db.Exec("DELETE FROM experiments WHERE id = ?", "experiment-1"); err != nil {
					t.Fatalf("delete experiment error = %v", err)
				}
			},
			want: apperr.CodeExperimentWorkspaceNotFound,
		},
		{
			name: "固定promptがない場合は開始不可を返す",
			prepare: func(t *testing.T, store *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
				if _, err := store.db.Exec("DELETE FROM experiment_fixed_condition_prompts WHERE fixed_condition_id = (SELECT fixed_condition_id FROM experiments WHERE id = ?)", "experiment-1"); err != nil {
					t.Fatalf("delete fixed prompts error = %v", err)
				}
			},
			want: apperr.CodeExperimentStartNotReady,
		},
		{
			name: "開始操作IDの生成失敗を返す",
			prepare: func(t *testing.T, _ *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
				replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random failed") })
			},
			want: "",
		},
		{
			name: "run IDの生成失敗を返す",
			prepare: func(t *testing.T, _ *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
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
			},
			want: "",
		},
		{
			name: "開始操作insert失敗を返す",
			prepare: func(t *testing.T, store *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
				if _, err := store.db.Exec("CREATE TRIGGER fail_start_operation BEFORE INSERT ON experiment_start_operations BEGIN SELECT RAISE(ABORT, 'operation failure'); END"); err != nil {
					t.Fatalf("create trigger error = %v", err)
				}
			},
			want: "",
		},
		{
			name: "run insert失敗を返す",
			prepare: func(t *testing.T, store *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
				if _, err := store.db.Exec("CREATE TRIGGER fail_start_run BEFORE INSERT ON experiment_runs BEGIN SELECT RAISE(ABORT, 'run failure'); END"); err != nil {
					t.Fatalf("create trigger error = %v", err)
				}
			},
			want: "",
		},
		{
			name: "実験状態更新失敗を返す",
			prepare: func(t *testing.T, store *Store, _ domain.ExperimentFixedConditions) {
				t.Helper()
				if _, err := store.db.Exec("CREATE TRIGGER fail_start_transition BEFORE UPDATE OF state ON experiments BEGIN SELECT RAISE(ABORT, 'transition failure'); END"); err != nil {
					t.Fatalf("create trigger error = %v", err)
				}
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, fixed := fixedExperimentPreparationStore(t)
			t.Cleanup(func() { _ = store.Close() })
			tt.prepare(t, store, fixed)

			_, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
			if err == nil {
				t.Fatal("BeginExperiment() error = nil, want error")
			}
			if tt.want != "" && !apperr.IsCode(err, tt.want) {
				t.Errorf("BeginExperiment() error = %v, want code %q", err, tt.want)
			}
		})
	}
}

// SQLite driver境界で開始前後の読込とtransaction失敗を注入する。
func TestStoreBeginExperimentDriverFailures(t *testing.T) {
	tests := []struct {
		name        string
		stage       experimentStartFailureStage
		call        func(*Store) error
		wantSuccess bool
	}{
		{
			name:  "commit後の開始snapshot欠落を返す",
			stage: experimentStartPostCommitMissing,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "commit後の開始snapshot読込失敗を返す",
			stage: experimentStartPostCommitReadError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "run一覧読込失敗を返す",
			stage: experimentStartFindRunsError,
			call: func(store *Store) error {
				_, _, err := store.findExperimentStart(context.Background(), "request-1")

				return err
			},
		},
		{
			name:  "ワークスペース読込失敗を返す",
			stage: experimentStartFindWorkspaceError,
			call: func(store *Store) error {
				_, _, err := store.findExperimentStart(context.Background(), "request-1")

				return err
			},
		},
		{
			name:  "ワークスペース欠落を返す",
			stage: experimentStartFindWorkspaceMissing,
			call: func(store *Store) error {
				_, _, err := store.findExperimentStart(context.Background(), "request-1")

				return err
			},
		},
		{
			name:  "SQLite競合の再試行上限を返す",
			stage: experimentStartBusy,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始transactionのSQLite競合の再試行上限を返す",
			stage: experimentStartBeginBusy,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始transactionのSQLite競合待機中の中止を返す",
			stage: experimentStartBeginBusy,
			call: func(store *Store) error {
				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(500*time.Microsecond, cancel)
				defer cancel()
				_, _, err := store.BeginExperiment(ctx, "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "SQLite競合待機中の中止を返す",
			stage: experimentStartBusy,
			call: func(store *Store) error {
				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(500*time.Microsecond, cancel)
				defer cancel()
				_, _, err := store.BeginExperiment(ctx, "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "既存開始操作の読込失敗を返す",
			stage: experimentStartInitialQueryError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:        "初期開始snapshot読込のSQLite競合後に既存snapshotを返す",
			stage:       experimentStartInitialBusyThenReplay,
			wantSuccess: true,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "transaction開始失敗を返す",
			stage: experimentStartBeginError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "固定条件主record読込失敗を返す",
			stage: experimentStartConditionQueryError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始可能でない競合後のsnapshot読込失敗を返す",
			stage: experimentStartNotReadySnapshotReadError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:        "開始可能でない競合後に同じsnapshotを返す",
			stage:       experimentStartNotReadySnapshotReplay,
			wantSuccess: true,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始可能でない競合後に別実験snapshotを拒否する",
			stage: experimentStartNotReadySnapshotOther,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:        "開始可能でない競合後のsnapshot読込競合を再試行する",
			stage:       experimentStartNotReadySnapshotBusyThenReplay,
			wantSuccess: true,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "固定prompt読込失敗を返す",
			stage: experimentStartPromptQueryError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "固定prompt scan失敗を返す",
			stage: experimentStartPromptScanError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "固定prompt反復失敗を返す",
			stage: experimentStartPromptRowsError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始操作一意競合後の読込失敗を返す",
			stage: experimentStartConflictReadError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始操作一意競合後のrollback失敗を返す",
			stage: experimentStartConflictRollbackError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:        "開始操作一意競合後に同じsnapshotを返す",
			stage:       experimentStartConflictReplay,
			wantSuccess: true,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "開始操作一意競合後に別実験を拒否する",
			stage: experimentStartConflictOther,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
		{
			name:  "commit失敗を返す",
			stage: experimentStartCommitError,
			call: func(store *Store) error {
				_, _, err := store.BeginExperiment(context.Background(), "request-1", "experiment-1")

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentStartFailureStore(t, tt.stage)
			if err := tt.call(store); (err == nil) != tt.wantSuccess {
				t.Fatalf("operation error = %v, want success = %v", err, tt.wantSuccess)
			}
		})
	}
}

type experimentStartFailureStage string

const (
	experimentStartInitialQueryError              experimentStartFailureStage = "initial-query-error"
	experimentStartInitialBusyThenReplay          experimentStartFailureStage = "initial-busy-then-replay"
	experimentStartBeginError                     experimentStartFailureStage = "begin-error"
	experimentStartConditionQueryError            experimentStartFailureStage = "condition-query-error"
	experimentStartPromptQueryError               experimentStartFailureStage = "prompt-query-error"
	experimentStartPromptScanError                experimentStartFailureStage = "prompt-scan-error"
	experimentStartPromptRowsError                experimentStartFailureStage = "prompt-rows-error"
	experimentStartConflictReadError              experimentStartFailureStage = "conflict-read-error"
	experimentStartConflictRollbackError          experimentStartFailureStage = "conflict-rollback-error"
	experimentStartConflictReplay                 experimentStartFailureStage = "conflict-replay"
	experimentStartConflictOther                  experimentStartFailureStage = "conflict-other"
	experimentStartCommitError                    experimentStartFailureStage = "commit-error"
	experimentStartBusy                           experimentStartFailureStage = "busy"
	experimentStartBeginBusy                      experimentStartFailureStage = "begin-busy"
	experimentStartNotReadySnapshotReadError      experimentStartFailureStage = "not-ready-snapshot-read-error"
	experimentStartNotReadySnapshotReplay         experimentStartFailureStage = "not-ready-snapshot-replay"
	experimentStartNotReadySnapshotOther          experimentStartFailureStage = "not-ready-snapshot-other"
	experimentStartNotReadySnapshotBusyThenReplay experimentStartFailureStage = "not-ready-snapshot-busy-then-replay"
	experimentStartPostCommitMissing              experimentStartFailureStage = "post-commit-missing"
	experimentStartPostCommitReadError            experimentStartFailureStage = "post-commit-read-error"
	experimentStartFindRunsError                  experimentStartFailureStage = "find-runs-error"
	experimentStartFindRunsScanError              experimentStartFailureStage = "find-runs-scan-error"
	experimentStartFindRunsTimeError              experimentStartFailureStage = "find-runs-time-error"
	experimentStartFindRunsRowsError              experimentStartFailureStage = "find-runs-rows-error"
	experimentStartFindWorkspaceError             experimentStartFailureStage = "find-workspace-error"
	experimentStartFindWorkspaceMissing           experimentStartFailureStage = "find-workspace-missing"
)

// 開始run一覧のSQLite読込失敗を返す。
func TestFindExperimentStartRunsDriverFailures(t *testing.T) {
	tests := []struct {
		name  string
		stage experimentStartFailureStage
	}{
		{
			name:  "query失敗",
			stage: experimentStartFindRunsError,
		},
		{
			name:  "scan失敗",
			stage: experimentStartFindRunsScanError,
		},
		{
			name:  "時刻変換失敗",
			stage: experimentStartFindRunsTimeError,
		},
		{
			name:  "行走査失敗",
			stage: experimentStartFindRunsRowsError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newExperimentStartFailureStore(t, tt.stage).findExperimentStartRuns(context.Background(), "operation-1")
			if err == nil {
				t.Error("findExperimentStartRuns() error = nil, want SQLite read error")
			}
		})
	}
}

const experimentStartFailureDriverName = "context-lab-experiment-start-failure"

var experimentStartFailureDriverOnce sync.Once

// newExperimentStartFailureStore は開始処理のdatabase/sql失敗注入用storeを返す。
func newExperimentStartFailureStore(t *testing.T, stage experimentStartFailureStage) *Store {
	t.Helper()
	experimentStartFailureDriverOnce.Do(func() { sql.Register(experimentStartFailureDriverName, experimentStartFailureDriver{}) })
	database, err := sql.Open(experimentStartFailureDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	return &Store{db: database}
}

// experimentStartFailureDriver は開始処理専用の失敗注入driver。
type experimentStartFailureDriver struct{}

// Open は指定stageを接続へ渡す。
func (experimentStartFailureDriver) Open(stage string) (driver.Conn, error) {
	return &experimentStartFailureConnection{stage: experimentStartFailureStage(stage)}, nil
}

// experimentStartFailureConnection は開始処理のSQL境界を模擬する。
type experimentStartFailureConnection struct {
	stage          experimentStartFailureStage
	operationReads int
}

// Prepare は未使用statementを拒否する。
func (*experimentStartFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}

// Close は接続を閉じる。
func (*experimentStartFailureConnection) Close() error { return nil }

// Begin は互換transactionを開始する。
func (c *experimentStartFailureConnection) Begin() (driver.Tx, error) { return c.begin() }

// BeginTx はcontext付きtransactionを開始する。
func (c *experimentStartFailureConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}

// begin は指定stageの開始失敗を返す。
func (c *experimentStartFailureConnection) begin() (driver.Tx, error) {
	if c.stage == experimentStartBeginError {
		return nil, errors.New("begin failed")
	}
	if c.stage == experimentStartBeginBusy {
		return nil, errors.New("database is busy")
	}

	return experimentStartFailureTransaction{stage: c.stage}, nil
}

// QueryContext は開始操作、固定条件、promptの結果を返す。
func (c *experimentStartFailureConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "FROM experiment_start_operations"):
		c.operationReads++
		if c.stage == experimentStartBusy {
			return nil, errors.New("database is busy")
		}
		if c.stage == experimentStartInitialBusyThenReplay && c.operationReads == 1 {
			return nil, errors.New("database is locked")
		}
		if (c.stage == experimentStartNotReadySnapshotReadError || c.stage == experimentStartNotReadySnapshotReplay || c.stage == experimentStartNotReadySnapshotOther || c.stage == experimentStartNotReadySnapshotBusyThenReplay) && c.operationReads > 1 {
			if c.stage == experimentStartNotReadySnapshotReadError {
				return nil, errors.New("operation query failed")
			}
			if c.stage == experimentStartNotReadySnapshotBusyThenReplay && c.operationReads == 2 {
				return nil, errors.New("database is locked")
			}
			experimentID := "experiment-1"
			if c.stage == experimentStartNotReadySnapshotOther {
				experimentID = "other-experiment"
			}

			return &experimentStartFailureRows{
				columns: []string{
					"request_id",
					"experiment_id",
					"operation_id",
					"state",
					"failure_code",
				},
				values: [][]driver.Value{
					{
						"request-1",
						experimentID,
						"operation-1",
						"starting",
						"",
					},
				},
			}, nil
		}
		if c.stage == experimentStartInitialQueryError ||
			(c.stage == experimentStartConflictReadError && c.operationReads > 1) ||
			(c.stage == experimentStartPostCommitReadError && c.operationReads > 1) {
			return nil, errors.New("operation query failed")
		}
		if c.stage == experimentStartInitialBusyThenReplay || c.stage == experimentStartFindRunsError || c.stage == experimentStartFindWorkspaceError || c.stage == experimentStartFindWorkspaceMissing || ((c.stage == experimentStartConflictReplay || c.stage == experimentStartConflictOther) && c.operationReads > 1) {
			experimentID := "experiment-1"
			if c.stage == experimentStartConflictOther {
				experimentID = "other-experiment"
			}
			return &experimentStartFailureRows{
				columns: []string{
					"request_id",
					"experiment_id",
					"operation_id",
					"state",
					"failure_code",
				},
				values: [][]driver.Value{
					{
						"request-1",
						experimentID,
						"operation-1",
						"starting",
						"",
					},
				},
			}, nil
		}
		return &experimentStartFailureRows{
			columns: []string{
				"request_id",
				"experiment_id",
				"operation_id",
				"state",
				"failure_code",
			},
		}, nil
	case strings.Contains(query, "FROM experiments e LEFT JOIN experiment_fixed_conditions"):
		if c.stage == experimentStartConditionQueryError {
			return nil, errors.New("condition query failed")
		}
		state := "ready"
		if c.stage == experimentStartNotReadySnapshotReadError || c.stage == experimentStartNotReadySnapshotReplay || c.stage == experimentStartNotReadySnapshotOther || c.stage == experimentStartNotReadySnapshotBusyThenReplay {
			state = "preparing"
		}
		return &experimentStartFailureRows{
			columns: []string{
				"state",
				"fixed_condition_id",
				"purpose",
				"hypothesis",
				"environment_conditions",
				"initial_input",
				"evaluation_axes",
			},
			values: [][]driver.Value{
				{
					state,
					"condition-1",
					"purpose",
					nil,
					"environment",
					"input",
					"axes",
				},
			},
		}, nil
	case strings.Contains(query, "FROM experiment_fixed_condition_prompts"):
		if c.stage == experimentStartPromptQueryError {
			return nil, errors.New("prompt query failed")
		}
		rows := &experimentStartFailureRows{
			columns: []string{
				"sequence_no",
				"content",
			},
			values: [][]driver.Value{
				{
					1,
					"prompt",
				},
			},
		}
		if c.stage == experimentStartPromptScanError {
			rows.values = [][]driver.Value{
				{
					"invalid",
					"prompt",
				},
			}
		}
		if c.stage == experimentStartPromptRowsError {
			rows.values = nil
			rows.nextErr = errors.New("prompt rows failed")
		}
		return rows, nil
	case strings.Contains(query, "FROM experiment_runs"):
		if c.stage == experimentStartFindRunsError {
			return nil, errors.New("runs query failed")
		}
		rows := &experimentStartFailureRows{
			columns: []string{
				"id",
				"state",
				"summary",
				"updated_at",
			},
		}
		if c.stage == experimentStartFindRunsScanError {
			rows.values = [][]driver.Value{
				{
					nil,
					"failed",
					nil,
					"2026-08-10T00:00:00Z",
				},
			}
		}
		if c.stage == experimentStartFindRunsTimeError {
			rows.values = [][]driver.Value{
				{
					"run-1",
					"failed",
					nil,
					"invalid",
				},
			}
		}
		if c.stage == experimentStartFindRunsRowsError {
			rows.nextErr = errors.New("run rows failed")
		}
		return rows, nil
	case strings.Contains(query, "FROM experiments e JOIN experiment_fixed_conditions"):
		if c.stage == experimentStartFindWorkspaceError {
			return nil, errors.New("workspace query failed")
		}
		if c.stage == experimentStartInitialBusyThenReplay || c.stage == experimentStartNotReadySnapshotReplay || c.stage == experimentStartNotReadySnapshotOther || c.stage == experimentStartNotReadySnapshotBusyThenReplay || c.stage == experimentStartConflictReplay || c.stage == experimentStartConflictOther {
			return &experimentStartFailureRows{
				columns: []string{
					"state",
					"updated_at",
					"id",
					"purpose",
					"hypothesis",
					"environment_conditions",
					"initial_input",
					"evaluation_axes",
					"fixed_at",
					"operation_id",
					"operation_fixed_at",
				},
				values: [][]driver.Value{
					{
						"ready",
						"2026-08-10T00:00:00Z",
						"condition-1",
						"purpose",
						nil,
						"environment",
						"input",
						"axes",
						"2026-08-10T00:00:00Z",
						"fix-operation-1",
						"2026-08-10T00:00:00Z",
					},
				},
			}, nil
		}
		return &experimentStartFailureRows{columns: []string{"state"}}, nil
	case strings.Contains(query, "FROM experiment_evaluations"):
		return &experimentStartFailureRows{
			columns: []string{
				"id",
				"state",
				"summary",
				"updated_at",
			},
		}, nil
	default:
		return nil, errors.New("unexpected query")
	}
}

// ExecContext は開始operationの一意競合または成功を返す。
func (c *experimentStartFailureConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if (c.stage == experimentStartConflictReadError || c.stage == experimentStartConflictRollbackError || c.stage == experimentStartConflictReplay || c.stage == experimentStartConflictOther) && strings.Contains(query, "INSERT INTO experiment_start_operations") {
		return nil, errors.New("UNIQUE constraint failed: experiment_start_operations.request_id")
	}

	return driver.RowsAffected(1), nil
}

// experimentStartFailureTransaction はcommit失敗を注入する。
type experimentStartFailureTransaction struct{ stage experimentStartFailureStage }

// Commit は指定stageで失敗する。
func (t experimentStartFailureTransaction) Commit() error {
	if t.stage == experimentStartCommitError {
		return errors.New("commit failed")
	}

	return nil
}

// Rollback はtransaction取消を受理する。
func (t experimentStartFailureTransaction) Rollback() error {
	if t.stage == experimentStartConflictRollbackError {
		return errors.New("rollback failed")
	}

	return nil
}

// experimentStartFailureRows はdriver.Rowsの最小実装。
type experimentStartFailureRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	nextErr error
}

// Columns は列名を返す。
func (r *experimentStartFailureRows) Columns() []string { return r.columns }

// Close はrowsを閉じる。
func (*experimentStartFailureRows) Close() error { return nil }

// Next は次の行または指定済み反復失敗を返す。
func (r *experimentStartFailureRows) Next(destination []driver.Value) error {
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
