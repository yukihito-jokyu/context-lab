package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

const insightFailureDriverName = "context-lab-insight-create-failure"

var insightFailureDriverOnce sync.Once
var insightReplayErrorCalls int
var insightFailureCancel func()

// newInsightFailureStore は知見作成のSQLite異常を注入する。
func newInsightFailureStore(t *testing.T, stage string) *Store {
	t.Helper()
	if stage == "replay-error" {
		insightReplayErrorCalls = 0
	}
	insightFailureDriverOnce.Do(func() { sql.Register(insightFailureDriverName, insightFailureDriver{}) })
	database, err := sql.Open(insightFailureDriverName, stage)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return &Store{db: database}
}

type insightFailureDriver struct{}

func (insightFailureDriver) Open(stage string) (driver.Conn, error) {
	return &insightFailureConnection{stage: stage}, nil
}

type insightFailureConnection struct {
	stage            string
	operationQueries int
}

func (*insightFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unsupported")
}
func (*insightFailureConnection) Close() error { return nil }
func (c *insightFailureConnection) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *insightFailureConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	if c.stage == "begin" {
		return nil, errors.New("begin")
	}
	return insightFailureTransaction{stage: c.stage}, nil
}
func (c *insightFailureConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.stage == "busy" && strings.Contains(query, "insight_create_operations") {
		return nil, errors.New("database is locked")
	}
	if c.stage == "read" {
		return nil, errors.New("read")
	}
	if strings.Contains(query, "insight_create_operations") {
		c.operationQueries++
		if c.stage == "busy-cancel" {
			if c.operationQueries == 2 && insightFailureCancel != nil {
				insightFailureCancel()
			}
			return nil, errors.New("database is locked")
		}
		if c.stage == "outer-read" {
			if c.operationQueries == 1 {
				return nil, errors.New("database is locked")
			}
			return nil, errors.New("outer read")
		}
		if c.stage == "replay" || c.stage == "conflict" || c.stage == "replay-error" {
			if c.stage == "replay-error" {
				insightReplayErrorCalls++
				if insightReplayErrorCalls == 1 {
					return nil, errors.New("database is locked")
				}
				return nil, errors.New("read")
			}
			if c.operationQueries == 1 {
				return nil, errors.New("database is locked")
			}
			if c.stage == "replay-error" {
				return nil, errors.New("read")
			}
			fingerprint := insightFingerprint([]domain.InsightEvidence{
				{
					ExperimentID: "a",
					ConclusionID: "a",
				},
				{
					ExperimentID: "b",
					ConclusionID: "b",
				},
			}, "s", "c", "g")
			if c.stage == "conflict" {
				fingerprint = "other"
			}
			return &insightFailureRows{
				columns: []string{
					"a",
					"b",
					"c",
					"d",
					"e",
					"f",
				},
				values: [][]driver.Value{
					{
						fingerprint,
						"id",
						"s",
						"c",
						"g",
						"2026-01-01T00:00:00Z",
					},
				},
			}, nil
		}
		if c.stage == "operation-parse" {
			return &insightFailureRows{
				columns: []string{
					"a",
					"b",
					"c",
					"d",
					"e",
					"f",
				},
				values: [][]driver.Value{
					{
						"fp",
						"id",
						"s",
						"c",
						"g",
						"invalid",
					},
				},
			}, nil
		}
		if c.stage == "operation" || c.stage == "evidence-query" || c.stage == "evidence-scan" || c.stage == "evidence-rows" {
			return &insightFailureRows{
				columns: []string{
					"a",
					"b",
					"c",
					"d",
					"e",
					"f",
				},
				values: [][]driver.Value{
					{
						"fp",
						"id",
						"s",
						"c",
						"g",
						"2026-01-01T00:00:00Z",
					},
				},
			}, nil
		}
		return &insightFailureRows{}, nil
	}
	if strings.Contains(query, "FROM insight_evidences") {
		if c.stage == "evidence-query" {
			return nil, errors.New("evidence query")
		}
		if c.stage == "evidence-scan" {
			return &insightFailureRows{
				columns: []string{
					"a",
					"b",
				},
				values: [][]driver.Value{
					{
						nil,
						"conclusion",
					},
				},
			}, nil
		}
		if c.stage == "evidence-rows" {
			return &insightFailureRows{err: errors.New("rows")}, nil
		}
		return &insightFailureRows{
			columns: []string{
				"a",
				"b",
			},
			values: [][]driver.Value{
				{
					"experiment",
					"conclusion",
				},
			},
		}, nil
	}
	if strings.Contains(query, "experiment_conclusions") {
		if c.stage == "evidence" {
			return nil, errors.New("evidence")
		}
		return &insightFailureRows{values: [][]driver.Value{{true}}}, nil
	}
	return &insightFailureRows{}, nil
}
func (c *insightFailureConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if (c.stage == "insight" && strings.Contains(query, "INSERT INTO insights")) || (c.stage == "evidences" && strings.Contains(query, "INSERT INTO insight_evidences")) || (c.stage == "insert-operation" && strings.Contains(query, "insight_create_operations")) {
		return nil, errors.New("insert")
	}
	return driver.RowsAffected(1), nil
}

type insightFailureTransaction struct{ stage string }

func (t insightFailureTransaction) Commit() error {
	if t.stage == "commit" {
		return errors.New("commit")
	}
	return nil
}
func (insightFailureTransaction) Rollback() error { return nil }

type insightFailureRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	err     error
}

func (r *insightFailureRows) Columns() []string {
	if len(r.columns) > 0 {
		return r.columns
	}
	return []string{"value"}
}
func (*insightFailureRows) Close() error { return nil }
func (r *insightFailureRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		if r.err != nil {
			return r.err
		}
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func insightEvidencePair(firstExperiment, firstConclusion, secondExperiment, secondConclusion string) []domain.InsightEvidence {
	return []domain.InsightEvidence{
		{
			ExperimentID: firstExperiment,
			ConclusionID: firstConclusion,
		},
		{
			ExperimentID: secondExperiment,
			ConclusionID: secondConclusion,
		},
	}
}

// CreateInsightの保存、replay、異常根拠を検証する。
func TestStoreCreateInsight(t *testing.T) {
	store := newTestStore(t)
	seedInsightEvidence(t, store, "experiment-a", "conclusion-a")
	seedInsightEvidence(t, store, "experiment-b", "conclusion-b")
	evidences := []domain.InsightEvidence{
		{
			ExperimentID: "experiment-a",
			ConclusionID: "conclusion-a",
		},
		{
			ExperimentID: "experiment-b",
			ConclusionID: "conclusion-b",
		},
	}
	created, wasCreated, err := store.CreateInsight(context.Background(), "request", evidences, "statement", "conditions", "gaps")
	if err != nil || !wasCreated {
		t.Fatalf("CreateInsight() = (%+v, %v, %v), want created", created, wasCreated, err)
	}
	replayed, wasCreated, err := store.CreateInsight(context.Background(), "request", evidences, "statement", "conditions", "gaps")
	if err != nil || wasCreated || replayed.InsightID != created.InsightID || len(replayed.Evidences) != 2 {
		t.Errorf("replay = (%+v, %v, %v), want persisted result", replayed, wasCreated, err)
	}
	_, _, err = store.CreateInsight(context.Background(), "request", evidences, "other", "conditions", "gaps")
	if !apperr.IsCode(err, apperr.CodeInsightCreateRequestConflict) {
		t.Errorf("conflict error = %v, want conflict", err)
	}
	missing := []domain.InsightEvidence{
		{
			ExperimentID: "experiment-a",
			ConclusionID: "missing",
		},
		{
			ExperimentID: "experiment-b",
			ConclusionID: "conclusion-b",
		},
	}
	_, _, err = store.CreateInsight(context.Background(), "other", missing, "statement", "conditions", "gaps")
	if !apperr.IsCode(err, apperr.CodeInsightCreateEvidenceNotFound) {
		t.Errorf("missing error = %v, want not found", err)
	}
}

// createInsightOnceの根拠不足、根拠不存在、識別子失敗を検証する。
func TestStoreCreateInsightOnceFailures(t *testing.T) {
	tests := []struct {
		name       string
		evidences  []domain.InsightEvidence
		identifier bool
		wantCode   apperr.Code
	}{
		{
			name: "根拠不足",
			evidences: []domain.InsightEvidence{{
				ExperimentID: "a",
				ConclusionID: "a",
			}},
			wantCode: apperr.CodeInsightCreateEvidenceInsufficient,
		},
		{
			name:      "根拠不存在",
			evidences: insightEvidencePair("a", "a", "b", "b"),
			wantCode:  apperr.CodeInsightCreateEvidenceNotFound,
		},
		{
			name:       "識別子失敗",
			evidences:  insightEvidencePair("a", "a", "b", "b"),
			identifier: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			if tt.identifier {
				seedInsightEvidence(t, store, "a", "a")
				seedInsightEvidence(t, store, "b", "b")
				previous := readBriefingRandom
				t.Cleanup(func() { readBriefingRandom = previous })
				readBriefingRandom = func([]byte) (int, error) { return 0, errors.New("random") }
			}
			_, _, err := store.createInsightOnce(context.Background(), "request", tt.evidences, "s", "c", "g", "fp")
			if tt.wantCode != "" && !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("createInsightOnce() error = %v, want %q", err, tt.wantCode)
			}
			if tt.identifier && err == nil {
				t.Error("createInsightOnce() error = nil, want identifier error")
			}
		})
	}
}

// createInsightOnceのSQLite transaction異常を検証する。
func TestStoreCreateInsightOnceDriverFailures(t *testing.T) {
	evidences := insightEvidencePair("a", "a", "b", "b")
	for _, stage := range []string{
		"begin",
		"read",
		"evidence",
		"insight",
		"evidences",
		"insert-operation",
		"commit",
	} {
		t.Run(stage, func(t *testing.T) {
			_, _, err := newInsightFailureStore(t, stage).createInsightOnce(context.Background(), "request", evidences, "s", "c", "g", "fp")
			if err == nil {
				t.Errorf("createInsightOnce(%s) error = nil", stage)
			}
		})
	}
}

// CreateInsightのbusy retry終了と中止を検証する。
func TestStoreCreateInsightBusyFailures(t *testing.T) {
	evidences := insightEvidencePair("a", "a", "b", "b")
	_, _, err := newInsightFailureStore(t, "busy").CreateInsight(context.Background(), "request", evidences, "s", "c", "g")
	if err == nil {
		t.Error("CreateInsight() error = nil, want retry exhausted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = newInsightFailureStore(t, "busy").CreateInsight(ctx, "request", evidences, "s", "c", "g")
	if err == nil {
		t.Error("CreateInsight() cancel error = nil")
	}
	cancelCtx, cancelBusy := context.WithCancel(context.Background())
	insightFailureCancel = cancelBusy
	t.Cleanup(func() { insightFailureCancel = nil })
	_, _, err = newInsightFailureStore(t, "busy-cancel").CreateInsight(cancelCtx, "request", evidences, "s", "c", "g")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("CreateInsight() busy cancel error = %v, want canceled", err)
	}
	for _, stage := range []string{
		"read",
		"replay",
		"conflict",
		"replay-error",
		"outer-read",
	} {
		t.Run(stage, func(t *testing.T) {
			_, _, callErr := newInsightFailureStore(t, stage).CreateInsight(context.Background(), "request", evidences, "s", "c", "g")
			if stage == "replay" {
				if callErr != nil {
					t.Errorf("CreateInsight(replay) error = %v", callErr)
				}
				return
			}
			if callErr == nil {
				t.Errorf("CreateInsight(%s) error = nil", stage)
			}
			if stage == "outer-read" && !strings.Contains(callErr.Error(), "outer read") {
				t.Errorf("CreateInsight(outer-read) error = %v, want outer read", callErr)
			}
		})
	}
}

// findInsightOperationのSQLite読取異常を検証する。
func TestFindInsightOperationDriverFailures(t *testing.T) {
	for _, stage := range []string{
		"read",
		"operation-parse",
		"evidence-query",
		"evidence-scan",
		"evidence-rows",
		"operation",
	} {
		t.Run(stage, func(t *testing.T) {
			_, _, err := findInsightOperation(context.Background(), newInsightFailureStore(t, stage).db, "request")
			if stage == "operation" {
				if err != nil {
					t.Errorf("findInsightOperation() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("findInsightOperation(%s) error = nil", stage)
			}
		})
	}
}

// CreateInsightの二pool同一request収束を検証する。
func TestStoreCreateInsightAcrossStores(t *testing.T) {
	directory := t.TempDir()
	first, err := Open(directory)
	if err != nil {
		t.Fatalf("Open(first) error = %v", err)
	}
	t.Cleanup(func() { _ = first.Close() })
	second, err := Open(filepath.Clean(directory))
	if err != nil {
		t.Fatalf("Open(second) error = %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })
	seedInsightEvidence(t, first, "experiment-a", "conclusion-a")
	seedInsightEvidence(t, first, "experiment-b", "conclusion-b")
	evidences := insightEvidencePair("experiment-a", "conclusion-a", "experiment-b", "conclusion-b")
	results := make(chan domain.Insight, 2)
	errors := make(chan error, 2)
	var group sync.WaitGroup
	for _, store := range []*Store{
		first,
		second,
	} {
		group.Add(1)
		go func(store *Store) {
			defer group.Done()
			insight, _, createErr := store.CreateInsight(context.Background(), "same-request", evidences, "statement", "conditions", "gaps")
			results <- insight
			errors <- createErr
		}(store)
	}
	group.Wait()
	close(results)
	close(errors)
	var id string
	for createErr := range errors {
		if createErr != nil {
			t.Errorf("CreateInsight() error = %v", createErr)
		}
	}
	for insight := range results {
		if id == "" {
			id = insight.InsightID
		} else if insight.InsightID != id {
			t.Errorf("InsightID = %q, want %q", insight.InsightID, id)
		}
	}
}

// seedInsightEvidence は確定済み根拠を直接投入する。
func seedInsightEvidence(t *testing.T, store *Store, experimentID, conclusionID string) {
	t.Helper()
	if _, err := store.db.Exec("INSERT INTO experiments(id,purpose,state,updated_at) VALUES(?,?,?,?)", experimentID, "purpose", "completed", "2026-01-01T00:00:00Z"); err != nil {
		t.Fatalf("insert experiment error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO experiment_conclusions(id,experiment_id,conclusion,evaluation_snapshot_digest,state,finalized_at) VALUES(?,?,?,?,?,?)", conclusionID, experimentID, "conclusion", "digest", "finalized", "2026-01-01T00:00:00Z"); err != nil {
		t.Fatalf("insert conclusion error = %v", err)
	}
}
