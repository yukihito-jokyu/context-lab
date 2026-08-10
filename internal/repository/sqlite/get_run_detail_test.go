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

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// SQLite run詳細の正本読込。
func TestStoreGetRunDetail(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T) (*Store, string)
		want  string
	}{
		{
			name: "保存済みの観測と成果物を返す",
			setup: func(t *testing.T) (*Store, string) {
				return newRunDetailStore(t)
			},
			want: "detail",
		},
		{
			name: "存在しないrunは未検出を返す",
			setup: func(t *testing.T) (*Store, string) {
				store, _ := newRunDetailStore(t)

				return store, "missing-run"
			},
			want: "missing",
		},
		{
			name: "観測表の読込失敗を返す",
			setup: func(t *testing.T) (*Store, string) {
				store, runID := newRunDetailStore(t)
				execRunDetailSQL(t, store, "DROP TABLE experiment_run_observations")

				return store, runID
			},
			want: "find run observations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, runID := tt.setup(t)
			detail, found, err := store.GetRunDetail(context.Background(), runID)
			if tt.want == "detail" {
				if err != nil {
					t.Fatalf("GetRunDetail() error = %v", err)
				}
				if !found {
					t.Fatal("found = false, want true")
				}
				assertRunDetail(t, detail)

				return
			}
			if tt.want == "missing" {
				if err != nil {
					t.Errorf("GetRunDetail() error = %v, want nil", err)
				}
				if found {
					t.Error("found = true, want false")
				}

				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("GetRunDetail() error = %v, want containing %q", err, tt.want)
			}
			if found {
				t.Error("found = true, want false")
			}
		})
	}
}

// run詳細本体の入力値と時刻異常。
func TestStoreFindRunDetail(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store, string)
		wantErr string
		want    string
	}{
		{
			name: "未記録の照合状態と最終観測時刻を補う",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET reconciliation_state = NULL, last_observed_at = NULL WHERE id = ?", runID)
			},
			want: "fallback",
		},
		{
			name: "操作時刻が不正なら失敗する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_start_operations SET updated_at = 'invalid' WHERE operation_id = (SELECT operation_id FROM experiment_runs WHERE id = ?)", runID)
			},
			wantErr: "parse run operation update time",
		},
		{
			name: "作成時刻が不正なら失敗する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET created_at = 'invalid' WHERE id = ?", runID)
			},
			wantErr: "parse run creation time",
		},
		{
			name: "更新時刻が不正なら失敗する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET updated_at = 'invalid' WHERE id = ?", runID)
			},
			wantErr: "parse run update time",
		},
		{
			name: "最終観測時刻が不正なら失敗する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET last_observed_at = 'invalid' WHERE id = ?", runID)
			},
			wantErr: "parse run last observed time",
		},
		{
			name: "run本体の読込失敗を返す",
			prepare: func(t *testing.T, store *Store, _ string) {
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			wantErr: "find run detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, runID := newRunDetailStore(t)
			tt.prepare(t, store, runID)
			detail, found, err := store.findRunDetail(context.Background(), runID)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("findRunDetail() error = %v, want containing %q", err, tt.wantErr)
				}
				if found {
					t.Error("found = true, want false")
				}

				return
			}
			if err != nil {
				t.Fatalf("findRunDetail() error = %v", err)
			}
			if !found {
				t.Fatal("found = false, want true")
			}
			if detail.Reconciliation.State != "confirmed" {
				t.Errorf("Reconciliation.State = %q, want %q", detail.Reconciliation.State, "confirmed")
			}
			if !detail.Reconciliation.LastObservedAt.Equal(detail.Run.UpdatedAt) {
				t.Errorf("LastObservedAt = %s, want %s", detail.Reconciliation.LastObservedAt, detail.Run.UpdatedAt)
			}
		})
	}
}

// run詳細の反復読込失敗。
func TestStoreRunDetailIterationFailures(t *testing.T) {
	tests := []struct {
		name string
		kind string
	}{
		{
			name: "観測の反復失敗を返す",
			kind: "observations",
		},
		{
			name: "成果物の反復失敗を返す",
			kind: "artifacts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newRunDetailRowsErrorStore(t, tt.kind)
			if tt.kind == "observations" {
				_, err := store.findRunObservations(context.Background(), "run-1")
				if err == nil || !strings.Contains(err.Error(), "iterate run observations") {
					t.Errorf("findRunObservations() error = %v, want iteration error", err)
				}

				return
			}
			_, err := store.findRunArtifacts(context.Background(), "run-1")
			if err == nil || !strings.Contains(err.Error(), "iterate run artifacts") {
				t.Errorf("findRunArtifacts() error = %v, want iteration error", err)
			}
		})
	}
}

// run詳細の関連表読込失敗。
func TestStorePopulateRunDetail(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		wantErr string
	}{
		{
			name:    "全関連情報を設定する",
			prepare: func(*testing.T, *Store) {},
		},
		{
			name: "成果物表の読込失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				execRunDetailSQL(t, store, "DROP TABLE experiment_run_artifacts")
			},
			wantErr: "find run artifacts",
		},
		{
			name: "失敗表の読込失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				execRunDetailSQL(t, store, "DROP TABLE experiment_run_failures")
			},
			wantErr: "find run failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, runID := newRunDetailStore(t)
			detail, found, err := store.findRunDetail(context.Background(), runID)
			if err != nil {
				t.Fatalf("findRunDetail() error = %v", err)
			}
			if !found {
				t.Fatal("found = false, want true")
			}
			tt.prepare(t, store)
			err = store.populateRunDetail(context.Background(), &detail)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("populateRunDetail() error = %v, want containing %q", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("populateRunDetail() error = %v", err)
			}
			if got := len(detail.Observations); got != 2 {
				t.Errorf("Observations length = %d, want 2", got)
			}
			if got := len(detail.Artifacts.Items); got != 2 {
				t.Errorf("Artifacts.Items length = %d, want 2", got)
			}
			if detail.Failure == nil {
				t.Error("Failure = nil, want persisted failure")
			}
		})
	}
}

// run観測の読込境界。
func TestStoreFindRunObservations(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store, string)
		wantErr string
		wantLen int
	}{
		{
			name:    "時系列順に返す",
			prepare: func(*testing.T, *Store, string) {},
			wantLen: 2,
		},
		{
			name: "観測時刻が不正なら失敗する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_run_observations SET occurred_at = 'invalid' WHERE run_id = ? AND sequence_no = 1", runID)
			},
			wantErr: "parse run observation time",
		},
		{
			name: "観測値のscan失敗を返す",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_run_observations SET sequence_no = 'invalid' WHERE run_id = ? AND sequence_no = 1", runID)
			},
			wantErr: "scan run observation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, runID := newRunDetailStore(t)
			tt.prepare(t, store, runID)
			observations, err := store.findRunObservations(context.Background(), runID)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("findRunObservations() error = %v, want containing %q", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("findRunObservations() error = %v", err)
			}
			if got := len(observations); got != tt.wantLen {
				t.Errorf("observations length = %d, want %d", got, tt.wantLen)
			}
			if observations[0].SequenceNo != 1 || observations[1].SequenceNo != 2 {
				t.Errorf("observations = %+v, want sequence 1 then 2", observations)
			}
		})
	}
}

// run成果物の読込境界。
func TestStoreFindRunArtifacts(t *testing.T) {
	tests := []struct {
		name       string
		prepare    func(*testing.T, *Store, string)
		wantErr    string
		wantStatus string
		wantReason string
	}{
		{
			name:       "部分取得の理由と成果物を返す",
			prepare:    func(*testing.T, *Store, string) {},
			wantStatus: domain.ExperimentRunArtifactStatusPartial,
			wantReason: "RUNNER_FAILED",
		},
		{
			name: "完了済みの空成果物を完全集計する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "DELETE FROM experiment_run_artifacts WHERE run_id = ?", runID)
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET artifact_status = ?, artifact_reason_code = NULL WHERE id = ?", domain.ExperimentRunArtifactStatusComplete, runID)
			},
			wantStatus: domain.ExperimentRunArtifactStatusComplete,
		},
		{
			name: "未記録状態を正本から返す",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "DELETE FROM experiment_run_artifacts WHERE run_id = ?", runID)
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET artifact_status = ?, artifact_reason_code = NULL WHERE id = ?", domain.ExperimentRunArtifactStatusNotRecorded, runID)
			},
			wantStatus: domain.ExperimentRunArtifactStatusNotRecorded,
		},
		{
			name: "理由のない部分取得を安全な理由で補う",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_runs SET artifact_reason_code = NULL WHERE id = ?", runID)
			},
			wantStatus: domain.ExperimentRunArtifactStatusPartial,
			wantReason: "ARTIFACT_PARTIAL",
		},
		{
			name: "成果物値のscan失敗を返す",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "DROP TABLE experiment_run_artifacts")
				execRunDetailSQL(t, store, "CREATE TABLE experiment_run_artifacts (run_id TEXT NOT NULL, digest TEXT, label TEXT, status TEXT, PRIMARY KEY (run_id, digest))")
				execRunDetailSQL(t, store, "INSERT INTO experiment_run_artifacts (run_id, digest, status) VALUES (?, NULL, 'available')", runID)
			},
			wantErr: "scan run artifact",
		},
		{
			name: "状態読込失敗を返す",
			prepare: func(t *testing.T, store *Store, _ string) {
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			wantErr: "find run artifact status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, runID := newRunDetailStore(t)
			tt.prepare(t, store, runID)
			artifacts, err := store.findRunArtifacts(context.Background(), runID)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("findRunArtifacts() error = %v, want containing %q", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("findRunArtifacts() error = %v", err)
			}
			if artifacts.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", artifacts.Status, tt.wantStatus)
			}
			if artifacts.ReasonCode != tt.wantReason {
				t.Errorf("ReasonCode = %q, want %q", artifacts.ReasonCode, tt.wantReason)
			}
		})
	}
}

// run失敗記録の読込境界。
func TestStoreFindRunFailure(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store, string)
		wantErr string
		wantNil bool
	}{
		{
			name:    "run固有の失敗記録を返す",
			prepare: func(*testing.T, *Store, string) {},
		},
		{
			name: "失敗記録がなければnilを返す",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "DELETE FROM experiment_run_failures WHERE run_id = ?", runID)
			},
			wantNil: true,
		},
		{
			name: "失敗時刻が不正なら失敗する",
			prepare: func(t *testing.T, store *Store, runID string) {
				execRunDetailSQL(t, store, "UPDATE experiment_run_failures SET occurred_at = 'invalid' WHERE run_id = ?", runID)
			},
			wantErr: "parse run failure time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, runID := newRunDetailStore(t)
			tt.prepare(t, store, runID)
			failure, err := store.findRunFailure(context.Background(), runID)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("findRunFailure() error = %v, want containing %q", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("findRunFailure() error = %v", err)
			}
			if tt.wantNil {
				if failure != nil {
					t.Errorf("failure = %+v, want nil", failure)
				}

				return
			}
			if failure == nil {
				t.Fatal("failure = nil, want persisted failure")
			}
			if failure.Code != "RUNNER_FAILED" {
				t.Errorf("Code = %q, want %q", failure.Code, "RUNNER_FAILED")
			}
		})
	}
}

// run詳細用の完全なSQLite正本。
func newRunDetailStore(t *testing.T) (*Store, string) {
	t.Helper()
	store, fixed := fixedExperimentPreparationStore(t)
	start, _, err := store.BeginExperiment(context.Background(), "run-detail-request", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	runID := start.Runs[0].ID
	const observedAt = "2026-08-10T01:02:03Z"
	execRunDetailSQL(t, store, "UPDATE experiment_runs SET state = ?, summary = ?, artifact_status = ?, artifact_reason_code = ?, created_at = ?, updated_at = ?, last_observed_at = ?, reconciliation_state = ? WHERE id = ?", domain.ExperimentRunStateFailed, "安全な実行要約", domain.ExperimentRunArtifactStatusPartial, "RUNNER_FAILED", observedAt, observedAt, observedAt, "reconciling", runID)
	execRunDetailSQL(t, store, "UPDATE experiment_start_operations SET updated_at = ? WHERE operation_id = ?", observedAt, start.OperationID)
	execRunDetailSQL(t, store, "INSERT INTO experiment_run_observations (run_id, sequence_no, kind, occurred_at, summary) VALUES (?, ?, ?, ?, ?)", runID, 2, "completed", observedAt, "完了を観測")
	execRunDetailSQL(t, store, "INSERT INTO experiment_run_observations (run_id, sequence_no, kind, occurred_at, summary) VALUES (?, ?, ?, ?, ?)", runID, 1, "started", observedAt, "開始を観測")
	execRunDetailSQL(t, store, "INSERT INTO experiment_run_artifacts (run_id, digest, label, status) VALUES (?, ?, ?, ?)", runID, "sha256:bbb", nil, "available")
	execRunDetailSQL(t, store, "INSERT INTO experiment_run_artifacts (run_id, digest, label, status) VALUES (?, ?, ?, ?)", runID, "sha256:aaa", "安全な成果物", "available")
	execRunDetailSQL(t, store, "INSERT INTO experiment_run_failures (run_id, code, occurred_at, partial_summary) VALUES (?, ?, ?, ?)", runID, "RUNNER_FAILED", observedAt, "途中成果を保存済み")

	return store, runID
}

// run詳細SQLの実行。
func execRunDetailSQL(t *testing.T, store *Store, query string, arguments ...any) {
	t.Helper()
	if _, err := store.db.Exec(query, arguments...); err != nil {
		t.Fatalf("Exec(%q) error = %v", query, err)
	}
}

// run詳細の成功値検証。
func assertRunDetail(t *testing.T, detail domain.ExperimentRunDetail) {
	t.Helper()
	if detail.Run.ID == "" || detail.Run.ExperimentID != "experiment-1" {
		t.Errorf("Run = %+v, want persisted run", detail.Run)
	}
	if detail.FixedPrompt.SequenceNo == 0 || detail.FixedPrompt.Content == "" {
		t.Errorf("FixedPrompt = %+v, want persisted fixed prompt", detail.FixedPrompt)
	}
	if detail.Operation.ID == "" || detail.Operation.State == "" {
		t.Errorf("Operation = %+v, want persisted operation", detail.Operation)
	}
	if len(detail.Observations) != 2 || detail.Observations[0].SequenceNo != 1 {
		t.Errorf("Observations = %+v, want ordered observations", detail.Observations)
	}
	if detail.Artifacts.Status != domain.ExperimentRunArtifactStatusPartial || detail.Artifacts.ReasonCode != "RUNNER_FAILED" || len(detail.Artifacts.Items) != 2 {
		t.Errorf("Artifacts = %+v, want partial persisted artifacts", detail.Artifacts)
	}
	if detail.Failure == nil || detail.Failure.Code != "RUNNER_FAILED" {
		t.Errorf("Failure = %+v, want run-specific failure", detail.Failure)
	}
	if detail.Reconciliation.State != "reconciling" {
		t.Errorf("Reconciliation.State = %q, want %q", detail.Reconciliation.State, "reconciling")
	}
	if detail.LastConfirmedAt.IsZero() {
		t.Error("LastConfirmedAt = zero, want run update time")
	}
}

const runDetailRowsErrorDriverName = "context-lab-run-detail-rows-error"

var runDetailRowsErrorDriverOnce sync.Once

// run詳細反復失敗用storeを生成する。
func newRunDetailRowsErrorStore(t *testing.T, kind string) *Store {
	t.Helper()
	runDetailRowsErrorDriverOnce.Do(func() {
		sql.Register(runDetailRowsErrorDriverName, runDetailRowsErrorDriver{})
	})
	database, err := sql.Open(runDetailRowsErrorDriverName, kind)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	return &Store{db: database}
}

type runDetailRowsErrorDriver struct{}

func (runDetailRowsErrorDriver) Open(kind string) (driver.Conn, error) {
	return runDetailRowsErrorConnection{kind: kind}, nil
}

type runDetailRowsErrorConnection struct {
	kind string
}

func (runDetailRowsErrorConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}

func (runDetailRowsErrorConnection) Close() error {
	return nil
}

func (runDetailRowsErrorConnection) Begin() (driver.Tx, error) {
	return runDetailRowsErrorTransaction{}, nil
}

func (c runDetailRowsErrorConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.kind == "artifacts" && strings.Contains(query, "SELECT artifact_status, artifact_reason_code") {
		return &runDetailArtifactStatusRows{}, nil
	}
	if c.kind == "observations" && strings.Contains(query, "experiment_run_observations") {
		return runDetailRowsError{columns: []string{
			"sequence_no",
			"kind",
			"occurred_at",
			"summary",
		}}, nil
	}
	if c.kind == "artifacts" && strings.Contains(query, "experiment_run_artifacts") {
		return runDetailRowsError{columns: []string{
			"digest",
			"label",
			"status",
		}}, nil
	}

	return nil, errors.New("unexpected query")
}

type runDetailRowsErrorTransaction struct{}

func (runDetailRowsErrorTransaction) Commit() error {
	return nil
}

func (runDetailRowsErrorTransaction) Rollback() error {
	return nil
}

type runDetailRowsError struct {
	columns []string
}

type runDetailArtifactStatusRows struct {
	returned bool
}

// artifact状態用列定義。
func (*runDetailArtifactStatusRows) Columns() []string {
	return []string{
		"artifact_status",
		"artifact_reason_code",
	}
}

// artifact状態用行クローズ。
func (*runDetailArtifactStatusRows) Close() error {
	return nil
}

// artifact状態用行返却。
func (r *runDetailArtifactStatusRows) Next(destination []driver.Value) error {
	if r.returned {
		return io.EOF
	}
	r.returned = true
	destination[0] = domain.ExperimentRunArtifactStatusComplete
	destination[1] = nil

	return nil
}

func (r runDetailRowsError) Columns() []string {
	return r.columns
}

func (runDetailRowsError) Close() error {
	return nil
}

func (runDetailRowsError) Next([]driver.Value) error {
	return io.ErrUnexpectedEOF
}
