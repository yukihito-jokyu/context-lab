package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// GetExperimentComparisonは評価がまだない実験を空配列で返す。
func TestStoreGetExperimentComparison(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	comparison, found, err := store.GetExperimentComparison(context.Background(), fixed.ExperimentID)
	if err != nil {
		t.Fatalf("GetExperimentComparison() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if comparison.Experiment.ID != fixed.ExperimentID || comparison.Experiment.Purpose == "" || comparison.Evaluations == nil {
		t.Errorf("comparison = %+v, want fixed experiment and empty evaluations", comparison)
	}
	_, found, err = store.GetExperimentComparison(context.Background(), "missing")
	if err != nil {
		t.Fatalf("missing GetExperimentComparison() error = %v", err)
	}
	if found {
		t.Error("missing found = true, want false")
	}
}

// GetExperimentComparisonは複数評価、部分結果、NULL要約を正本から返す。
func TestStoreGetExperimentComparisonEvaluations(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	start, _, err := store.BeginExperiment(context.Background(), "comparison-start", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if len(start.Runs) != 2 {
		t.Fatalf("runs = %d, want 2", len(start.Runs))
	}
	for index, run := range start.Runs {
		if err := store.CompleteExperimentRun(context.Background(), run.ID, "run summary"); err != nil {
			t.Fatalf("CompleteExperimentRun(%d) error = %v", index, err)
		}
	}
	complete, _, err := store.BeginRunEvaluation(context.Background(), "comparison-evaluation-complete", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() complete error = %v", err)
	}
	if err := store.CompleteRunEvaluation(context.Background(), complete.EvaluationID, "評価要約"); err != nil {
		t.Fatalf("CompleteRunEvaluation() error = %v", err)
	}
	partial, _, err := store.BeginRunEvaluation(context.Background(), "comparison-evaluation-partial", start.Runs[1].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() partial error = %v", err)
	}
	if err := store.FailRunEvaluation(context.Background(), partial.EvaluationID, "EVALUATION_TIMEOUT"); err != nil {
		t.Fatalf("FailRunEvaluation() error = %v", err)
	}
	const observedAt = "2026-08-10T01:02:03Z"
	execRunDetailSQL(t, store, "UPDATE experiment_runs SET summary = NULL WHERE id = ?", start.Runs[1].ID)
	execRunDetailSQL(t, store, "UPDATE experiment_evaluations SET reconciliation_state = ?, last_observed_at = ? WHERE id = ?", "reconciling", observedAt, partial.EvaluationID)

	comparison, found, err := store.GetExperimentComparison(context.Background(), fixed.ExperimentID)
	if err != nil {
		t.Fatalf("GetExperimentComparison() error = %v", err)
	}
	if !found || len(comparison.Evaluations) != 2 {
		t.Fatalf("comparison = %+v, want two evaluations", comparison)
	}
	if comparison.Evaluations[0].Result.Status != "complete" || comparison.Evaluations[0].Result.Summary == nil {
		t.Errorf("complete evaluation = %+v, want recorded complete result", comparison.Evaluations[0])
	}
	if comparison.Evaluations[1].Result.Status != "partial" || comparison.Evaluations[1].Result.ReasonCode != "EVALUATION_TIMEOUT" {
		t.Errorf("partial evaluation = %+v, want partial timeout result", comparison.Evaluations[1])
	}
	if comparison.Evaluations[1].RunSummary != nil || comparison.Evaluations[1].Result.Summary != nil {
		t.Errorf("partial nullable summaries = %+v, want nil", comparison.Evaluations[1])
	}
	if comparison.Evaluations[1].Reconciliation.State != "reconciling" || comparison.Evaluations[1].Reconciliation.LastObservedAt.Format(time.RFC3339) != observedAt {
		t.Errorf("reconciliation = %+v, want reconciling at %s", comparison.Evaluations[1].Reconciliation, observedAt)
	}
}

// 比較取得のSQLite境界異常と正本のnullable値を検証する。
func TestStoreGetExperimentComparisonDriver(t *testing.T) {
	for _, tt := range []struct {
		stage     comparisonFailureStage
		wantFound bool
		wantError bool
	}{
		{
			stage:     comparisonExperimentMissing,
			wantFound: false,
		},
		{
			stage:     comparisonExperimentReadError,
			wantError: true,
		},
		{
			stage:     comparisonExperimentTimeError,
			wantError: true,
		},
		{
			stage:     comparisonEvaluationsReadError,
			wantError: true,
		},
		{
			stage:     comparisonEvaluationScanError,
			wantError: true,
		},
		{
			stage:     comparisonEvaluationUpdatedTimeError,
			wantError: true,
		},
		{
			stage:     comparisonEvaluationObservedTimeError,
			wantError: true,
		},
		{
			stage:     comparisonEvaluationRowsError,
			wantError: true,
		},
		{
			stage:     comparisonSuccess,
			wantFound: true,
		},
		{
			stage:     comparisonNullableValues,
			wantFound: true,
		},
	} {
		t.Run(string(tt.stage), func(t *testing.T) {
			comparison, found, err := newComparisonFailureStore(t, tt.stage).GetExperimentComparison(context.Background(), "experiment-1")
			if (err != nil) != tt.wantError {
				t.Errorf("GetExperimentComparison() error = %v, want error = %v", err, tt.wantError)
			}
			if found != tt.wantFound {
				t.Errorf("GetExperimentComparison() found = %v, want %v", found, tt.wantFound)
			}
			if tt.stage == comparisonSuccess && (len(comparison.Evaluations) != 1 || comparison.Evaluations[0].RunSummary == nil || comparison.Evaluations[0].Result.Summary == nil || comparison.Evaluations[0].Reconciliation.LastObservedAt.IsZero()) {
				t.Errorf("GetExperimentComparison() = %+v, want populated comparison", comparison)
			}
		})
	}
}

type comparisonFailureStage string

const (
	comparisonExperimentMissing           comparisonFailureStage = "experiment-missing"
	comparisonExperimentReadError         comparisonFailureStage = "experiment-read-error"
	comparisonExperimentTimeError         comparisonFailureStage = "experiment-time-error"
	comparisonEvaluationsReadError        comparisonFailureStage = "evaluations-read-error"
	comparisonEvaluationScanError         comparisonFailureStage = "evaluation-scan-error"
	comparisonEvaluationUpdatedTimeError  comparisonFailureStage = "evaluation-updated-time-error"
	comparisonEvaluationObservedTimeError comparisonFailureStage = "evaluation-observed-time-error"
	comparisonEvaluationRowsError         comparisonFailureStage = "evaluation-rows-error"
	comparisonSuccess                     comparisonFailureStage = "success"
	comparisonNullableValues              comparisonFailureStage = "nullable-values"
)

const comparisonFailureDriverName = "context-lab-comparison-failure"

var comparisonFailureDriverOnce sync.Once

func newComparisonFailureStore(t *testing.T, stage comparisonFailureStage) *Store {
	t.Helper()
	comparisonFailureDriverOnce.Do(func() { sql.Register(comparisonFailureDriverName, comparisonFailureDriver{}) })
	database, err := sql.Open(comparisonFailureDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	return &Store{db: database}
}

type comparisonFailureDriver struct{}

func (comparisonFailureDriver) Open(stage string) (driver.Conn, error) {
	return &comparisonFailureConnection{stage: comparisonFailureStage(stage)}, nil
}

type comparisonFailureConnection struct{ stage comparisonFailureStage }

func (*comparisonFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}
func (*comparisonFailureConnection) Close() error { return nil }
func (*comparisonFailureConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transaction is not supported")
}

func (c *comparisonFailureConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "FROM experiments e JOIN experiment_fixed_conditions") {
		if c.stage == comparisonExperimentReadError {
			return nil, errors.New("experiment query failed")
		}
		if c.stage == comparisonExperimentMissing {
			return &comparisonFailureRows{columns: comparisonExperimentColumns()}, nil
		}
		updatedAt := "2026-08-10T00:00:00Z"
		if c.stage == comparisonExperimentTimeError {
			updatedAt = "invalid"
		}
		return &comparisonFailureRows{
			columns: comparisonExperimentColumns(),
			values: [][]driver.Value{
				{
					"experiment-1",
					"目的",
					"評価軸",
					updatedAt,
				},
			},
		}, nil
	}
	if strings.Contains(query, "FROM experiment_evaluations e JOIN experiment_runs") {
		if c.stage == comparisonEvaluationsReadError {
			return nil, errors.New("evaluations query failed")
		}
		rows := &comparisonFailureRows{columns: comparisonEvaluationColumns()}
		if c.stage == comparisonEvaluationRowsError {
			rows.nextErr = errors.New("evaluation rows failed")
			return rows, nil
		}
		updatedAt := "2026-08-10T01:00:00Z"
		observedAt := "2026-08-10T02:00:00Z"
		if c.stage == comparisonEvaluationUpdatedTimeError {
			updatedAt = "invalid"
		}
		if c.stage == comparisonEvaluationObservedTimeError {
			observedAt = "invalid"
		}
		values := []driver.Value{
			"evaluation-1",
			"run-1",
			"completed",
			"run summary",
			"complete",
			"result summary",
			"",
			"confirmed",
			observedAt,
			updatedAt,
		}
		if c.stage == comparisonEvaluationScanError {
			values[0] = nil
		}
		if c.stage == comparisonNullableValues {
			values[3] = nil
			values[5] = nil
			values[6] = nil
			values[8] = nil
		}
		rows.values = [][]driver.Value{values}
		return rows, nil
	}
	return nil, errors.New("unexpected query")
}

type comparisonFailureRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	nextErr error
}

func (r *comparisonFailureRows) Columns() []string { return r.columns }
func (*comparisonFailureRows) Close() error        { return nil }
func (r *comparisonFailureRows) Next(destination []driver.Value) error {
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

func comparisonExperimentColumns() []string {
	return []string{
		"id",
		"purpose",
		"evaluation_axes",
		"updated_at",
	}
}

func comparisonEvaluationColumns() []string {
	return []string{
		"id",
		"run_id",
		"state",
		"summary",
		"result_status",
		"result_summary",
		"result_reason_code",
		"reconciliation_state",
		"last_observed_at",
		"updated_at",
	}
}
