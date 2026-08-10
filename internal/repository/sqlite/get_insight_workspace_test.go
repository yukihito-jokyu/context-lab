package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"
)

// GetInsightWorkspaceの確定済み候補と空状態を検証する。
func TestStoreGetInsightWorkspace(t *testing.T) {
	store, first := finalizedConclusionStore(t)
	firstConclusion, _, err := store.FinalizeExperimentConclusion(context.Background(), "first-conclusion", first.ExperimentID, "最初の結論")
	if err != nil {
		t.Fatalf("FinalizeExperimentConclusion() error = %v", err)
	}
	got, err := store.GetInsightWorkspace(context.Background())
	if err != nil {
		t.Fatalf("GetInsightWorkspace() error = %v", err)
	}
	if len(got.EvidenceCandidates) != 1 || len(got.SavedConsiderations) != 1 || len(got.Insights) != 0 {
		t.Errorf("workspace = %+v, want one candidate and consideration with no insights", got)
	}
	if got.LastConfirmedAt == nil {
		t.Fatal("LastConfirmedAt = nil, want latest conclusion time")
	}
	if got.EvidenceCandidates[0].ConclusionID != firstConclusion.ConclusionID {
		t.Errorf("EvidenceCandidates = %+v, want finalized conclusions", got.EvidenceCandidates)
	}

	empty, err := newTestStore(t).GetInsightWorkspace(context.Background())
	if err != nil {
		t.Fatalf("empty GetInsightWorkspace() error = %v", err)
	}
	if len(empty.EvidenceCandidates) != 0 || len(empty.SavedConsiderations) != 0 || len(empty.Insights) != 0 || empty.LastConfirmedAt != nil {
		t.Errorf("empty workspace = %+v, want empty arrays and nil time", empty)
	}
}

// GetInsightWorkspaceのSQLite読取異常を検証する。
func TestStoreGetInsightWorkspaceDriverFailures(t *testing.T) {
	tests := []struct {
		name      string
		stage     insightWorkspaceFailureStage
		wantError bool
	}{
		{
			name:      "query失敗",
			stage:     insightWorkspaceQueryFailure,
			wantError: true,
		},
		{
			name:      "scan失敗",
			stage:     insightWorkspaceScanFailure,
			wantError: true,
		},
		{
			name:      "時刻不正",
			stage:     insightWorkspaceTimeFailure,
			wantError: true,
		},
		{
			name:      "rows失敗",
			stage:     insightWorkspaceRowsFailure,
			wantError: true,
		},
		{
			name:  "成功",
			stage: insightWorkspaceSuccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, err := newInsightWorkspaceFailureStore(t, tt.stage).GetInsightWorkspace(context.Background())
			if gotError := err != nil; gotError != tt.wantError {
				t.Errorf("GetInsightWorkspace() error = %v, want error = %v", err, tt.wantError)
			}
			if !tt.wantError && (len(workspace.EvidenceCandidates) != 1 || workspace.LastConfirmedAt == nil) {
				t.Errorf("GetInsightWorkspace() = %+v, want populated workspace", workspace)
			}
		})
	}
}

// insightWorkspaceFailureStage はSQLite読取異常の段階。
type insightWorkspaceFailureStage string

const (
	insightWorkspaceQueryFailure insightWorkspaceFailureStage = "query"
	insightWorkspaceScanFailure  insightWorkspaceFailureStage = "scan"
	insightWorkspaceTimeFailure  insightWorkspaceFailureStage = "time"
	insightWorkspaceRowsFailure  insightWorkspaceFailureStage = "rows"
	insightWorkspaceSuccess      insightWorkspaceFailureStage = "success"
)

const insightWorkspaceFailureDriverName = "context-lab-insight-workspace-failure"

var insightWorkspaceFailureDriverOnce sync.Once

// newInsightWorkspaceFailureStore は読取異常を返すSQLite storeを生成する。
func newInsightWorkspaceFailureStore(t *testing.T, stage insightWorkspaceFailureStage) *Store {
	t.Helper()
	insightWorkspaceFailureDriverOnce.Do(func() { sql.Register(insightWorkspaceFailureDriverName, insightWorkspaceFailureDriver{}) })
	database, err := sql.Open(insightWorkspaceFailureDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	return &Store{db: database}
}

// insightWorkspaceFailureDriver は読取異常用のSQL driver。
type insightWorkspaceFailureDriver struct{}

// Open は異常段階を持つ接続を返す。
func (insightWorkspaceFailureDriver) Open(stage string) (driver.Conn, error) {
	return insightWorkspaceFailureConnection{stage: insightWorkspaceFailureStage(stage)}, nil
}

// insightWorkspaceFailureConnection は読取異常を返すSQL接続。
type insightWorkspaceFailureConnection struct{ stage insightWorkspaceFailureStage }

// Prepare は未使用のstatementを返す。
func (insightWorkspaceFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is unsupported")
}

// Close は接続を閉じる。
func (insightWorkspaceFailureConnection) Close() error { return nil }

// Begin は未使用transactionを拒否する。
func (insightWorkspaceFailureConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is unsupported")
}

// QueryContext は段階別の候補行を返す。
func (c insightWorkspaceFailureConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if c.stage == insightWorkspaceQueryFailure {
		return nil, errors.New("query failed")
	}
	rows := &insightWorkspaceFailureRows{columns: []string{
		"experiment_id",
		"purpose",
		"evaluation_axes",
		"conclusion_id",
		"conclusion",
		"finalized_at",
	}}
	if c.stage == insightWorkspaceRowsFailure {
		rows.nextErr = errors.New("rows failed")
		return rows, nil
	}
	finalizedAt := "2026-08-10T00:00:00Z"
	if c.stage == insightWorkspaceTimeFailure {
		finalizedAt = "invalid"
	}
	values := []driver.Value{
		"experiment-1",
		"目的",
		"軸",
		"conclusion-1",
		"結論",
		finalizedAt,
	}
	if c.stage == insightWorkspaceScanFailure {
		values[0] = nil
	}
	rows.values = [][]driver.Value{values}
	return rows, nil
}

// insightWorkspaceFailureRows は候補行を返すdriver rows。
type insightWorkspaceFailureRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	nextErr error
}

// Columns は列名を返す。
func (r *insightWorkspaceFailureRows) Columns() []string { return r.columns }

// Close はrowsを閉じる。
func (*insightWorkspaceFailureRows) Close() error { return nil }

// Next は次の候補行または異常を返す。
func (r *insightWorkspaceFailureRows) Next(destination []driver.Value) error {
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
