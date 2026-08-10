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

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 同一request IDを別々のdatabase/sql poolから同時に送っても一つのretry snapshotへ収束する。
func TestStoreRetryEndedRunConvergesConcurrentRequestsAcrossStores(t *testing.T) {
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
			_, sourceRunID := newRunDetailStoreForRetry(t, firstStore)

			type result struct {
				retry domain.ExperimentRunRetry
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
					retry, _, err := store.RetryEndedRun(context.Background(), "retry-request", sourceRunID)
					operationResult := result{
						retry: retry,
						err:   err,
					}
					results <- operationResult
				}(store)
			}
			<-entered
			<-entered
			close(ready)
			waitGroup.Wait()
			close(results)

			var retries []domain.ExperimentRunRetry
			for result := range results {
				if result.err != nil {
					t.Fatalf("RetryEndedRun() error = %v", result.err)
				}
				retries = append(retries, result.retry)
			}
			if !reflect.DeepEqual(retries[0], retries[1]) {
				t.Errorf("concurrent snapshots = %+v, want identical", retries)
			}
			var operations int
			if err := firstStore.db.QueryRow(
				"SELECT COUNT(*) FROM experiment_run_retry_operations WHERE request_id = ?",
				"retry-request",
			).Scan(&operations); err != nil {
				t.Fatalf("count retry operations error = %v", err)
			}
			if operations != 1 {
				t.Errorf("retry operations = %d, want 1", operations)
			}
		})
	}
}

// newRunDetailStoreForRetry は既存storeに失敗済みsource runを作成する。
func newRunDetailStoreForRetry(t *testing.T, store *Store) (domain.ExperimentStart, string) {
	t.Helper()
	seedExperimentPreparationDraftExperiment(t, store, "experiment-1")
	draft := testExperimentPreparationDraft("draft-request", "experiment-1", "目的")
	if _, err := store.SaveExperimentPreparationDraft(context.Background(), draft); err != nil {
		t.Fatalf("SaveExperimentPreparationDraft() error = %v", err)
	}
	fixed := fixedConditionsFromDraft(draft)
	if _, err := store.FixExperimentConditions(context.Background(), fixed); err != nil {
		t.Fatalf("FixExperimentConditions() error = %v", err)
	}
	start, _, err := store.BeginExperiment(context.Background(), "start-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err := store.FailExperimentRun(context.Background(), start.Runs[0].ID, "RUNNER_FAILED"); err != nil {
		t.Fatalf("FailExperimentRun() error = %v", err)
	}

	return start, start.Runs[0].ID
}

// RetryEndedRunは失敗済みrunから同じ固定promptのqueued runだけを作成する。
func TestStoreRetryEndedRun(t *testing.T) {
	store, sourceRunID := newRunDetailStore(t)
	ctx := context.Background()

	retry, created, err := store.RetryEndedRun(ctx, "retry-request-1", sourceRunID)
	if err != nil {
		t.Fatalf("RetryEndedRun() error = %v", err)
	}
	if !created {
		t.Fatal("created = false, want true")
	}
	if retry.SourceRunID != sourceRunID || retry.RetryRunID == "" || retry.OperationID == "" || retry.State != domain.ExperimentRunStateQueued {
		t.Errorf("RetryEndedRun() = %+v, want persisted queued retry", retry)
	}

	var retryOfRunID string
	var promptSequenceNo int
	var state string
	var operationID string
	var sourcePromptSequenceNo int
	if err := store.db.QueryRow("SELECT prompt_sequence_no FROM experiment_runs WHERE id = ?", sourceRunID).Scan(&sourcePromptSequenceNo); err != nil {
		t.Fatalf("query source run error = %v", err)
	}
	if err := store.db.QueryRow("SELECT retry_of_run_id, prompt_sequence_no, state, operation_id FROM experiment_runs WHERE id = ?", retry.RetryRunID).Scan(&retryOfRunID, &promptSequenceNo, &state, &operationID); err != nil {
		t.Fatalf("query retry run error = %v", err)
	}
	if retryOfRunID != sourceRunID || promptSequenceNo != sourcePromptSequenceNo || state != domain.ExperimentRunStateQueued || operationID != retry.OperationID {
		t.Errorf("retry run = (%q, %d, %q, %q), want source/prompt/queued/operation", retryOfRunID, promptSequenceNo, state, operationID)
	}

	replayed, created, err := store.RetryEndedRun(ctx, "retry-request-1", sourceRunID)
	if err != nil {
		t.Fatalf("RetryEndedRun() replay error = %v", err)
	}
	if created || replayed.RetryRunID != retry.RetryRunID {
		t.Errorf("replay = (%+v, %v), want existing retry", replayed, created)
	}

	if _, _, err := store.RetryEndedRun(ctx, "retry-request-1", "other-run"); !apperr.IsCode(err, apperr.CodeRunRetryRequestConflict) {
		t.Errorf("RetryEndedRun() conflict error = %v, want code %q", err, apperr.CodeRunRetryRequestConflict)
	}
	if _, _, err := store.RetryEndedRun(ctx, "retry-request-2", retry.RetryRunID); !apperr.IsCode(err, apperr.CodeRunRetryNotAllowed) {
		t.Errorf("RetryEndedRun() queued error = %v, want code %q", err, apperr.CodeRunRetryNotAllowed)
	}
	if _, _, err := store.RetryEndedRun(ctx, "retry-request-3", "missing-run"); !apperr.IsCode(err, apperr.CodeRunRetryNotFound) {
		t.Errorf("RetryEndedRun() missing error = %v, want code %q", err, apperr.CodeRunRetryNotFound)
	}
}

// 開始requestのreplayは後から作成されたretry runを含めない。
func TestStoreBeginExperimentReplayExcludesRetryRun(t *testing.T) {
	store, sourceRunID := newRunDetailStore(t)
	retry, _, err := store.RetryEndedRun(context.Background(), "retry-request", sourceRunID)
	if err != nil {
		t.Fatalf("RetryEndedRun() error = %v", err)
	}

	start, created, err := store.BeginExperiment(context.Background(), "run-detail-request", "experiment-1")
	if err != nil {
		t.Fatalf("BeginExperiment() replay error = %v", err)
	}
	if created {
		t.Fatal("created = true, want replay")
	}
	for _, run := range start.Runs {
		if run.ID == retry.RetryRunID {
			t.Errorf("replayed start runs = %+v, want no retry run", start.Runs)
		}
	}
}

// SQLite境界の異常を、実DBのロック状態に依存せずに検証する。
func TestStoreRetryEndedRunDriverFailures(t *testing.T) {
	for _, stage := range []retryFailureStage{
		retryInitialReadError,
		retryBeginError,
		retrySourceReadError,
		retrySourceMissing,
		retrySourceNotAllowed,
		retryInsertRunError,
		retryInsertOperationError,
		retryCommitError,
		retryPostCommitReadError,
		retryPostCommitMissing,
		retryConflictReadError,
		retryConflictRollbackError,
	} {
		t.Run(string(stage), func(t *testing.T) {
			store := newRetryFailureStore(t, stage)
			var err error
			if stage == retryInitialReadError || stage == retryConflictReadError || stage == retryConflictRollbackError {
				_, _, err = store.RetryEndedRun(context.Background(), "request-1", "source-1")
			} else {
				_, _, err = store.createRunRetry(context.Background(), "request-1", "source-1")
			}
			if err == nil {
				t.Error("operation error = nil, want error")
			}
		})
	}
}

func TestStoreFindRunRetryDriverFailures(t *testing.T) {
	for _, stage := range []retryFailureStage{
		retryFindReadError,
		retryFindTimeError,
	} {
		t.Run(string(stage), func(t *testing.T) {
			_, _, err := newRetryFailureStore(t, stage).findRunRetry(context.Background(), "request-1")
			if err == nil {
				t.Error("findRunRetry() error = nil, want error")
			}
		})
	}
	_, found, err := newRetryFailureStore(t, retryFindMissing).findRunRetry(context.Background(), "request-1")
	if err != nil || found {
		t.Errorf("findRunRetry() = (_, %v, %v), want (_, false, nil)", found, err)
	}
}

func TestIsRunRetryRequestConflict(t *testing.T) {
	if !isRunRetryRequestConflict(errors.New("UNIQUE constraint failed: experiment_run_retry_operations.request_id")) {
		t.Error("isRunRetryRequestConflict() = false, want true")
	}
	if isRunRetryRequestConflict(nil) || isRunRetryRequestConflict(errors.New("other error")) {
		t.Error("isRunRetryRequestConflict() = true, want false")
	}
}

func TestStoreRetryEndedRunAdditionalDriverBranches(t *testing.T) {
	for _, tt := range []struct {
		stage retryFailureStage
		want  bool
	}{
		{
			stage: retryInitialFound,
			want:  true,
		},
		{
			stage: retryInitialFoundOther,
			want:  false,
		},
		{
			stage: retryConflictReplay,
			want:  true,
		},
		{
			stage: retryConflictOther,
			want:  false,
		},
		{
			stage: retrySuccess,
			want:  true,
		},
	} {
		t.Run(string(tt.stage), func(t *testing.T) {
			store := newRetryFailureStore(t, tt.stage)
			if tt.stage == retryInitialFound || tt.stage == retryInitialFoundOther {
				_, _, err := store.RetryEndedRun(context.Background(), "request-1", "source-1")
				if (err == nil) != tt.want {
					t.Errorf("RetryEndedRun() error = %v", err)
				}
				return
			}
			_, _, err := store.createRunRetry(context.Background(), "request-1", "source-1")
			if (err == nil) != tt.want {
				t.Errorf("createRunRetry() error = %v", err)
			}
		})
	}
	canceled, cancel := context.WithCancel(context.Background())
	retryFailureCancel = cancel
	t.Cleanup(func() { retryFailureCancel = nil })
	if _, _, err := newRetryFailureStore(t, retryBusy).RetryEndedRun(canceled, "request-1", "source-1"); err == nil {
		t.Error("RetryEndedRun() busy error = nil, want error")
	}
}

func TestStoreCreateRunRetryIdentifierFailure(t *testing.T) {
	previous := readBriefingRandom
	t.Cleanup(func() { readBriefingRandom = previous })
	readBriefingRandom = func([]byte) (int, error) { return 0, errors.New("random failed") }
	if _, _, err := newRetryFailureStore(t, retrySuccess).createRunRetry(context.Background(), "request-1", "source-1"); err == nil || !strings.Contains(err.Error(), "generate retry run ID") {
		t.Errorf("createRunRetry() first identifier error = %v, want generate retry run ID", err)
	}
	reads := 0
	readBriefingRandom = func(bytes []byte) (int, error) {
		reads++
		if reads == 2 {
			return 0, errors.New("random failed")
		}
		for index := range bytes {
			bytes[index] = byte(index)
		}
		return len(bytes), nil
	}
	if _, _, err := newRetryFailureStore(t, retrySuccess).createRunRetry(context.Background(), "request-1", "source-1"); err == nil || !strings.Contains(err.Error(), "generate run retry operation ID") {
		t.Errorf("createRunRetry() second identifier error = %v, want generate run retry operation ID", err)
	}
}

func TestStoreRetryEndedRunBusyLimit(t *testing.T) {
	_, _, err := newRetryFailureStore(t, retryBusyLimit).RetryEndedRun(context.Background(), "request-1", "source-1")
	if err == nil {
		t.Error("RetryEndedRun() busy limit error = nil, want error")
	}
}

type retryFailureStage string

const (
	retryInitialReadError      retryFailureStage = "initial-read-error"
	retryBeginError            retryFailureStage = "begin-error"
	retrySourceReadError       retryFailureStage = "source-read-error"
	retrySourceMissing         retryFailureStage = "source-missing"
	retrySourceNotAllowed      retryFailureStage = "source-not-allowed"
	retryInsertRunError        retryFailureStage = "insert-run-error"
	retryInsertOperationError  retryFailureStage = "insert-operation-error"
	retryCommitError           retryFailureStage = "commit-error"
	retryPostCommitReadError   retryFailureStage = "post-commit-read-error"
	retryPostCommitMissing     retryFailureStage = "post-commit-missing"
	retryConflictReadError     retryFailureStage = "conflict-read-error"
	retryConflictRollbackError retryFailureStage = "conflict-rollback-error"
	retryFindReadError         retryFailureStage = "find-read-error"
	retryFindTimeError         retryFailureStage = "find-time-error"
	retryFindMissing           retryFailureStage = "find-missing"
	retryInitialFound          retryFailureStage = "initial-found"
	retryInitialFoundOther     retryFailureStage = "initial-found-other"
	retryConflictReplay        retryFailureStage = "conflict-replay"
	retryConflictOther         retryFailureStage = "conflict-other"
	retrySuccess               retryFailureStage = "success"
	retryBusy                  retryFailureStage = "busy"
	retryBusyLimit             retryFailureStage = "busy-limit"
)

const retryFailureDriverName = "context-lab-retry-failure"

var retryFailureDriverOnce sync.Once
var retryFailureCancel func()

func newRetryFailureStore(t *testing.T, stage retryFailureStage) *Store {
	t.Helper()
	retryFailureDriverOnce.Do(func() { sql.Register(retryFailureDriverName, retryFailureDriver{}) })
	database, err := sql.Open(retryFailureDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	return &Store{db: database}
}

type retryFailureDriver struct{}

func (retryFailureDriver) Open(stage string) (driver.Conn, error) {
	return &retryFailureConnection{stage: retryFailureStage(stage)}, nil
}

type retryFailureConnection struct {
	stage retryFailureStage
	reads int
}

func (*retryFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}
func (*retryFailureConnection) Close() error                { return nil }
func (c *retryFailureConnection) Begin() (driver.Tx, error) { return c.begin() }
func (c *retryFailureConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}
func (c *retryFailureConnection) begin() (driver.Tx, error) {
	if c.stage == retryBeginError {
		return nil, errors.New("begin failed")
	}
	return retryFailureTransaction{stage: c.stage}, nil
}

func (c *retryFailureConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "experiment_run_retry_operations") {
		c.reads++
		if c.stage == retryBusy {
			if retryFailureCancel != nil {
				retryFailureCancel()
			}
			return nil, errors.New("database is busy")
		}
		if c.stage == retryBusyLimit {
			return nil, errors.New("database is busy")
		}
		if c.stage == retryInitialReadError || c.stage == retryFindReadError || (c.stage == retryConflictReadError && c.reads > 1) || (c.stage == retryPostCommitReadError && c.reads > 0) {
			return nil, errors.New("retry query failed")
		}
		if c.stage == retryInitialFound || c.stage == retryInitialFoundOther || ((c.stage == retryConflictReplay || c.stage == retryConflictOther) && c.reads > 0) {
			rows := retryFailureResultRows("2026-08-10T00:00:00Z").(*retryFailureRows)
			if c.stage == retryInitialFoundOther || c.stage == retryConflictOther {
				rows.values[0][1] = "other-source"
			}
			return rows, nil
		}
		if c.stage == retryFindMissing || (c.stage == retryPostCommitMissing && c.reads > 0) || (c.reads == 1 && c.stage != retryFindTimeError && c.stage != retrySuccess) {
			return &retryFailureRows{columns: retryColumns()}, nil
		}
		createdAt := "2026-08-10T00:00:00Z"
		if c.stage == retryFindTimeError {
			createdAt = "invalid"
		}
		return retryFailureResultRows(createdAt), nil
	}
	if strings.Contains(query, "FROM experiment_runs WHERE id") {
		if c.stage == retrySourceReadError {
			return nil, errors.New("source query failed")
		}
		if c.stage == retrySourceMissing {
			return &retryFailureRows{
				columns: []string{
					"experiment_id",
					"state",
					"prompt_sequence_no",
					"isolation_kind",
				},
			}, nil
		}
		state := driver.Value(string(domain.ExperimentRunStateFailed))
		if c.stage == retrySourceNotAllowed {
			state = string(domain.ExperimentRunStateQueued)
		}
		return &retryFailureRows{
			columns: []string{
				"experiment_id",
				"state",
				"prompt_sequence_no",
				"isolation_kind",
			},
			values: [][]driver.Value{
				{
					"experiment-1",
					state,
					1,
					"docker",
				},
			},
		}, nil
	}
	return nil, errors.New("unexpected query")
}

func (c *retryFailureConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "INSERT INTO experiment_runs") && c.stage == retryInsertRunError {
		return nil, errors.New("insert run failed")
	}
	if strings.Contains(query, "INSERT INTO experiment_run_retry_operations") {
		if c.stage == retryInsertOperationError {
			return nil, errors.New("insert operation failed")
		}
		if c.stage == retryConflictReadError || c.stage == retryConflictRollbackError || c.stage == retryConflictReplay || c.stage == retryConflictOther {
			return nil, errors.New("UNIQUE constraint failed: experiment_run_retry_operations.request_id")
		}
	}
	return driver.RowsAffected(1), nil
}

type retryFailureTransaction struct{ stage retryFailureStage }

func (t retryFailureTransaction) Commit() error {
	if t.stage == retryCommitError {
		return errors.New("commit failed")
	}
	return nil
}
func (t retryFailureTransaction) Rollback() error {
	if t.stage == retryConflictRollbackError {
		return errors.New("rollback failed")
	}
	return nil
}

type retryFailureRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *retryFailureRows) Columns() []string { return r.columns }
func (*retryFailureRows) Close() error        { return nil }
func (r *retryFailureRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
func retryColumns() []string {
	return []string{
		"request_id",
		"source_run_id",
		"experiment_id",
		"run_id",
		"operation_id",
		"state",
		"created_at",
	}
}
func retryFailureResultRows(createdAt string) driver.Rows {
	return &retryFailureRows{
		columns: retryColumns(),
		values: [][]driver.Value{
			{
				"request-1",
				"source-1",
				"experiment-1",
				"retry-1",
				"operation-1",
				"queued",
				createdAt,
			},
		},
	}
}
