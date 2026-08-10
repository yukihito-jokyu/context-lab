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

type derivedFailureStage string

const (
	derivedBusy               derivedFailureStage = "busy"
	derivedBusyLimit          derivedFailureStage = "busy-limit"
	derivedBegin              derivedFailureStage = "begin"
	derivedOperationRead      derivedFailureStage = "operation-read"
	derivedSourceRead         derivedFailureStage = "source-read"
	derivedSourceMissing      derivedFailureStage = "source-missing"
	derivedSourceNotEligible  derivedFailureStage = "source-not-eligible"
	derivedPromptRead         derivedFailureStage = "prompt-read"
	derivedPromptScan         derivedFailureStage = "prompt-scan"
	derivedPromptRows         derivedFailureStage = "prompt-rows"
	derivedInsertExperiment   derivedFailureStage = "insert-experiment"
	derivedInsertPreparation  derivedFailureStage = "insert-preparation"
	derivedApplyChange        derivedFailureStage = "apply-change"
	derivedInsertPrompt       derivedFailureStage = "insert-prompt"
	derivedInsertOperation    derivedFailureStage = "insert-operation"
	derivedCommit             derivedFailureStage = "commit"
	derivedConflictRead       derivedFailureStage = "conflict-read"
	derivedConflictRollback   derivedFailureStage = "conflict-rollback"
	derivedConflictReplay     derivedFailureStage = "conflict-replay"
	derivedConflictOther      derivedFailureStage = "conflict-other"
	derivedFindRead           derivedFailureStage = "find-read"
	derivedFindTime           derivedFailureStage = "find-time"
	derivedFindMissing        derivedFailureStage = "find-missing"
	derivedInitialReplay      derivedFailureStage = "initial-replay"
	derivedInitialReplayOther derivedFailureStage = "initial-replay-other"
	derivedTransactionReplay  derivedFailureStage = "transaction-replay"
)

const derivedFailureDriverName = "context-lab-derived-failure"

var (
	derivedFailureDriverOnce sync.Once
	derivedFailureCancel     func()
)

// newDerivedFailureStore は派生作成のSQLite境界へ決定的な異常を注入する。
func newDerivedFailureStore(t *testing.T, stage derivedFailureStage) *Store {
	t.Helper()
	derivedFailureDriverOnce.Do(func() { sql.Register(derivedFailureDriverName, derivedFailureDriver{}) })
	database, err := sql.Open(derivedFailureDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return &Store{db: database}
}

type derivedFailureDriver struct{}

func (derivedFailureDriver) Open(stage string) (driver.Conn, error) {
	return &derivedFailureConnection{stage: derivedFailureStage(stage)}, nil
}

type derivedFailureConnection struct {
	stage         derivedFailureStage
	operationRead int
}

func (*derivedFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}
func (*derivedFailureConnection) Close() error                { return nil }
func (c *derivedFailureConnection) Begin() (driver.Tx, error) { return c.begin() }
func (c *derivedFailureConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}
func (c *derivedFailureConnection) begin() (driver.Tx, error) {
	if c.stage == derivedBegin {
		return nil, errors.New("begin failed")
	}
	return derivedFailureTransaction{stage: c.stage}, nil
}

func (c *derivedFailureConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "experiment_derived_operations") {
		c.operationRead++
		if c.stage == derivedBusy || c.stage == derivedBusyLimit {
			if derivedFailureCancel != nil {
				derivedFailureCancel()
			}
			return nil, errors.New("database is busy")
		}
		if c.stage == derivedOperationRead || c.stage == derivedFindRead || (c.stage == derivedConflictRead && c.operationRead > 1) {
			return nil, errors.New("operation read failed")
		}
		if c.stage == derivedFindMissing {
			return &derivedFailureRows{columns: derivedOperationColumns()}, nil
		}
		if c.stage == derivedInitialReplay || c.stage == derivedInitialReplayOther || c.stage == derivedTransactionReplay || ((c.stage == derivedConflictReplay || c.stage == derivedConflictOther) && c.operationRead > 1) {
			payload, _ := canonicalDerivedExperimentPayload(domain.DerivedExperimentChanges{Purpose: stringPointer("changed")}, "reason")
			if c.stage == derivedInitialReplayOther || c.stage == derivedConflictOther {
				payload = "other-payload"
			}
			createdAt := "2026-08-10T00:00:00Z"
			if c.stage == derivedFindTime {
				createdAt = "invalid"
			}
			return &derivedFailureRows{
				columns: derivedOperationColumns(),
				values: [][]driver.Value{
					{
						"request",
						"derived",
						"source",
						payload,
						createdAt,
					},
				},
			}, nil
		}
		if c.stage == derivedFindTime {
			return &derivedFailureRows{
				columns: derivedOperationColumns(),
				values: [][]driver.Value{
					{
						"request",
						"derived",
						"source",
						"payload",
						"invalid",
					},
				},
			}, nil
		}
		return &derivedFailureRows{columns: derivedOperationColumns()}, nil
	}
	if strings.Contains(query, "FROM experiments e") {
		switch c.stage {
		case derivedSourceRead:
			return nil, errors.New("source read failed")
		case derivedSourceMissing:
			return &derivedFailureRows{columns: derivedSourceColumns()}, nil
		case derivedSourceNotEligible:
			return &derivedFailureRows{
				columns: derivedSourceColumns(),
				values: [][]driver.Value{
					{
						"",
						"",
						"source",
						nil,
						"environment",
						"input",
						"axes",
						"session",
						"version",
					},
				},
			}, nil
		}
		return &derivedFailureRows{
			columns: derivedSourceColumns(),
			values: [][]driver.Value{
				{
					"fixed",
					"conclusion",
					"source",
					nil,
					"environment",
					"input",
					"axes",
					"session",
					"version",
				},
			},
		}, nil
	}
	if strings.Contains(query, "experiment_fixed_condition_prompts") {
		if c.stage == derivedPromptRead {
			return nil, errors.New("prompt read failed")
		}
		if c.stage == derivedPromptScan {
			return &derivedFailureRows{
				columns: []string{
					"sequence_no",
					"content",
				},
				values: [][]driver.Value{
					{
						"not-number",
						"prompt",
					},
				},
			}, nil
		}
		if c.stage == derivedPromptRows {
			return &derivedFailureRows{
				columns: []string{
					"sequence_no",
					"content",
				},
				err: errors.New("prompt rows failed"),
			}, nil
		}
		return &derivedFailureRows{
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
		}, nil
	}
	return nil, errors.New("unexpected query")
}

func (c *derivedFailureConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.Contains(query, "INSERT INTO experiments") && c.stage == derivedInsertExperiment:
		return nil, errors.New("experiment insert failed")
	case strings.Contains(query, "INSERT INTO experiment_preparations") && c.stage == derivedInsertPreparation:
		return nil, errors.New("preparation insert failed")
	case strings.Contains(query, "UPDATE ") && c.stage == derivedApplyChange:
		return nil, errors.New("change apply failed")
	case strings.Contains(query, "INSERT INTO experiment_preparation_prompts") && c.stage == derivedInsertPrompt:
		return nil, errors.New("prompt insert failed")
	case strings.Contains(query, "INSERT INTO experiment_derived_operations"):
		switch c.stage {
		case derivedInsertOperation:
			return nil, errors.New("operation insert failed")
		case derivedConflictRead, derivedConflictRollback, derivedConflictReplay, derivedConflictOther:
			return nil, errors.New("UNIQUE constraint failed: experiment_derived_operations.request_id")
		}
	}
	return driver.RowsAffected(1), nil
}

type derivedFailureTransaction struct{ stage derivedFailureStage }

func (t derivedFailureTransaction) Commit() error {
	if t.stage == derivedCommit {
		return errors.New("commit failed")
	}
	return nil
}
func (t derivedFailureTransaction) Rollback() error {
	if t.stage == derivedConflictRollback {
		return errors.New("rollback failed")
	}
	return nil
}

type derivedFailureRows struct {
	columns []string
	values  [][]driver.Value
	err     error
	index   int
}

func (r *derivedFailureRows) Columns() []string { return r.columns }
func (*derivedFailureRows) Close() error        { return nil }
func (r *derivedFailureRows) Next(dest []driver.Value) error {
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

func derivedOperationColumns() []string {
	return []string{
		"request_id",
		"experiment_id",
		"source_experiment_id",
		"canonical_payload",
		"created_at",
	}
}
func derivedSourceColumns() []string {
	return []string{
		"fixed_id",
		"conclusion_id",
		"purpose",
		"hypothesis",
		"environment_conditions",
		"initial_input",
		"evaluation_axes",
		"briefing_session_id",
		"briefing_version_id",
	}
}

// 派生作成のtransaction境界異常。
func TestStoreCreateDerivedExperimentDriverFailures(t *testing.T) {
	changed := "changed"
	tests := []struct {
		name    string
		stage   derivedFailureStage
		changes domain.DerivedExperimentChanges
	}{
		derivedFailureCase("transaction開始失敗", derivedBegin, changed),
		derivedFailureCase("操作読込失敗", derivedOperationRead, changed),
		derivedFailureCase("派生元読込失敗", derivedSourceRead, changed),
		derivedFailureCase("派生元不存在", derivedSourceMissing, changed),
		derivedFailureCase("派生元不適格", derivedSourceNotEligible, changed),
		derivedFailureCase("固定prompt読込失敗", derivedPromptRead, changed),
		derivedFailureCase("固定prompt scan失敗", derivedPromptScan, changed),
		derivedFailureCase("固定prompt反復失敗", derivedPromptRows, changed),
		derivedFailureCase("実験insert失敗", derivedInsertExperiment, changed),
		derivedFailureCase("準備insert失敗", derivedInsertPreparation, changed),
		derivedFailureCase("変更反映失敗", derivedApplyChange, changed),
		derivedFailureCase("prompt insert失敗", derivedInsertPrompt, changed),
		derivedFailureCase("操作insert失敗", derivedInsertOperation, changed),
		derivedFailureCase("commit失敗", derivedCommit, changed),
		derivedFailureCase("transaction内の既存snapshot", derivedTransactionReplay, changed),
		derivedFailureCase("一意競合後の読込失敗", derivedConflictRead, changed),
		derivedFailureCase("一意競合後のrollback失敗", derivedConflictRollback, changed),
		derivedFailureCase("一意競合後の同一snapshot", derivedConflictReplay, changed),
		derivedFailureCase("一意競合後の別snapshot", derivedConflictOther, changed),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, payloadErr := canonicalDerivedExperimentPayload(tt.changes, "reason")
			if payloadErr != nil {
				t.Fatalf("canonicalDerivedExperimentPayload() error = %v", payloadErr)
			}
			_, _, err := newDerivedFailureStore(t, tt.stage).createDerivedExperiment(context.Background(), "request", "source", tt.changes, payload)
			if tt.stage == derivedConflictReplay || tt.stage == derivedTransactionReplay {
				if err != nil {
					t.Errorf("createDerivedExperiment() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Error("createDerivedExperiment() error = nil, want error")
			}
		})
	}
}

// derivedFailureCase は派生作成失敗テストの入力を組み立てる。
func derivedFailureCase(name string, stage derivedFailureStage, changed string) struct {
	name    string
	stage   derivedFailureStage
	changes domain.DerivedExperimentChanges
} {
	return struct {
		name    string
		stage   derivedFailureStage
		changes domain.DerivedExperimentChanges
	}{
		name:  name,
		stage: stage,
		changes: domain.DerivedExperimentChanges{
			Purpose: &changed,
		},
	}
}

// 派生作成のreplay、競合待機、補助関数の未到達分岐。
func TestStoreCreateDerivedExperimentRetryAndHelperFailures(t *testing.T) {
	changed := "changed"
	payload, err := canonicalDerivedExperimentPayload(domain.DerivedExperimentChanges{Purpose: &changed}, "reason")
	if err != nil {
		t.Fatalf("canonicalDerivedExperimentPayload() error = %v", err)
	}
	for _, tt := range []struct {
		name  string
		stage derivedFailureStage
		want  bool
	}{
		{
			name:  "初期replay",
			stage: derivedInitialReplay,
			want:  true,
		},
		{
			name:  "初期別payload",
			stage: derivedInitialReplayOther,
			want:  false,
		},
		{
			name:  "時刻不正",
			stage: derivedFindTime,
			want:  false,
		},
		{
			name:  "操作未検出",
			stage: derivedFindMissing,
			want:  true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := newDerivedFailureStore(t, tt.stage)
			if tt.stage == derivedFindTime || tt.stage == derivedFindMissing {
				_, _, findErr := store.findDerivedExperiment(context.Background(), "request")
				if (findErr == nil) != tt.want {
					t.Errorf("findDerivedExperiment() error = %v", findErr)
				}
				return
			}
			_, _, retryErr := store.createDerivedExperimentWithRetry(context.Background(), "request", "source", domain.DerivedExperimentChanges{Purpose: &changed}, "reason")
			if (retryErr == nil) != tt.want {
				t.Errorf("createDerivedExperimentWithRetry() error = %v", retryErr)
			}
		})
	}
	canceled, cancel := context.WithCancel(context.Background())
	derivedFailureCancel = cancel
	t.Cleanup(func() { derivedFailureCancel = nil })
	if _, _, err := newDerivedFailureStore(t, derivedBusy).createDerivedExperimentWithRetry(canceled, "request", "source", domain.DerivedExperimentChanges{Purpose: &changed}, "reason"); err == nil {
		t.Error("busy retry error = nil, want error")
	}
	previous := marshalDerivedExperimentPayload
	t.Cleanup(func() { marshalDerivedExperimentPayload = previous })
	marshalDerivedExperimentPayload = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }
	if _, err := canonicalDerivedExperimentPayload(domain.DerivedExperimentChanges{}, "reason"); err == nil {
		t.Error("canonicalDerivedExperimentPayload() error = nil, want error")
	}
	if payload == "" {
		t.Error("payload = empty, want canonical payload")
	}
}

// 派生作成の再試行入口と識別子生成の失敗。
func TestStoreCreateDerivedExperimentRemainingFailures(t *testing.T) {
	changed := "changed"
	if _, _, err := newDerivedFailureStore(t, derivedFindRead).createDerivedExperimentWithRetry(context.Background(), "request", "source", domain.DerivedExperimentChanges{Purpose: &changed}, "reason"); err == nil {
		t.Error("find failure error = nil, want error")
	}
	previousMarshal := marshalDerivedExperimentPayload
	t.Cleanup(func() { marshalDerivedExperimentPayload = previousMarshal })
	marshalDerivedExperimentPayload = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }
	if _, _, err := newDerivedFailureStore(t, derivedInitialReplay).createDerivedExperimentWithRetry(context.Background(), "request", "source", domain.DerivedExperimentChanges{}, "reason"); err == nil {
		t.Error("payload marshal error = nil, want error")
	}
	marshalDerivedExperimentPayload = previousMarshal
	replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random failed") })
	payload, err := canonicalDerivedExperimentPayload(domain.DerivedExperimentChanges{Purpose: &changed}, "reason")
	if err != nil {
		t.Fatalf("canonicalDerivedExperimentPayload() error = %v", err)
	}
	if _, _, err = newDerivedFailureStore(t, derivedCommit).createDerivedExperiment(context.Background(), "request", "source", domain.DerivedExperimentChanges{Purpose: &changed}, payload); err == nil {
		t.Error("identifier error = nil, want error")
	}
	if _, _, err = newDerivedFailureStore(t, derivedBusyLimit).createDerivedExperimentWithRetry(context.Background(), "request", "source", domain.DerivedExperimentChanges{Purpose: &changed}, "reason"); err == nil {
		t.Error("retry limit error = nil, want error")
	}
}
