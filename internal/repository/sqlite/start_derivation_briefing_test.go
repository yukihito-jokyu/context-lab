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
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// BeginDerivationBriefingの適格性、再生、状態更新を確認。
func TestStoreBeginDerivationBriefing(t *testing.T) {
	store, sourceID := finalizedDerivedExperimentSource(t)
	created, wasCreated, err := store.BeginDerivationBriefing(context.Background(), "request-1", sourceID)
	if err != nil || !wasCreated {
		t.Fatalf("BeginDerivationBriefing() = (%+v, %v, %v), want created", created, wasCreated, err)
	}
	if created.SourceExperimentID != sourceID || created.State != "starting" {
		t.Errorf("created = %+v, want source %q and starting", created, sourceID)
	}
	if err := store.MarkDerivationBriefingStarted(context.Background(), "request-1"); err != nil {
		t.Fatalf("MarkDerivationBriefingStarted() error = %v", err)
	}
	if err := store.MarkDerivationBriefingFailed(context.Background(), "request-1", "DERIVATION_BRIEFING_START_UNAVAILABLE"); err != nil {
		t.Fatalf("MarkDerivationBriefingFailed() error = %v", err)
	}
	replayed, wasCreated, err := store.BeginDerivationBriefing(context.Background(), "request-1", sourceID)
	if err != nil || wasCreated || replayed.State != "failed" {
		t.Errorf("replay = (%+v, %v, %v), want persisted failed result", replayed, wasCreated, err)
	}
	if _, _, err := store.BeginDerivationBriefing(context.Background(), "request-1", "other-source"); !apperr.IsCode(err, apperr.CodeDerivedExperimentRequestConflict) {
		t.Errorf("different source error = %v, want request conflict", err)
	}
	if _, _, err := store.BeginDerivationBriefing(context.Background(), "missing", "missing"); !apperr.IsCode(err, apperr.CodeDerivedExperimentSourceNotFound) {
		t.Errorf("missing source error = %v, want source not found", err)
	}
}

// BeginDerivationBriefingの未確定派生元を確認。
func TestStoreBeginDerivationBriefingRequiresEligibleSource(t *testing.T) {
	store := newExperimentPreparationDraftStore(t)
	seedExperimentPreparationDraftExperiment(t, store, "source")
	if _, _, err := store.BeginDerivationBriefing(context.Background(), "request-1", "source"); !apperr.IsCode(err, apperr.CodeDerivedExperimentSourceNotEligible) {
		t.Errorf("ineligible source error = %v, want source not eligible", err)
	}
}

// derivationBriefingUniqueError はrequest ID制約違反を表すtest double。
type derivationBriefingUniqueError struct{}

// Error はSQLiteの一意制約違反を返す。
func (derivationBriefingUniqueError) Error() string {
	return "UNIQUE constraint failed: derivation_briefing_operations.request_id"
}

// isDerivationBriefingRequestConflictの判定を確認。
func TestIsDerivationBriefingRequestConflict(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "request IDの一意制約違反",
			err:  derivationBriefingUniqueError{},
			want: true,
		},
		{
			name: "別のエラー",
			err:  errors.New("other"),
			want: false,
		},
		{
			name: "nil",
			want: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDerivationBriefingRequestConflict(tt.err); got != tt.want {
				t.Errorf("isDerivationBriefingRequestConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 派生ブリーフ開始の永続化失敗。
func TestStoreBeginDerivationBriefingFailures(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(*testing.T, *Store)
		want     string
		wantCode apperr.Code
	}{
		{
			name: "session識別子生成失敗",
			prepare: func(t *testing.T, _ *Store) {
				replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random") })
			},
			want: "generate derivation briefing session ID",
		},
		{
			name: "既存開始結果検索失敗",
			prepare: func(t *testing.T, store *Store) {
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			want: "find derivation briefing",
		},
		{
			name: "operation識別子生成失敗",
			prepare: func(t *testing.T, _ *Store) {
				calls := 0
				replaceBriefingRandom(t, func(bytes []byte) (int, error) {
					calls++
					if calls == 1 {
						return len(bytes), nil
					}
					return 0, errors.New("random")
				})
			},
			want: "generate derivation briefing operation ID",
		},
		{
			name: "transaction開始失敗",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, errors.New("begin")
				}
			},
			want: "begin derivation briefing",
		},
		{
			name: "派生元読込失敗",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{fakeBriefingRow{err: errors.New("source")}},
				})
			},
			want: "find derivation briefing source",
		},
		{
			name: "派生元不存在",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{fakeBriefingRow{err: sql.ErrNoRows}},
				})
			},
			wantCode: apperr.CodeDerivedExperimentSourceNotFound,
		},
		{
			name: "派生元不適格",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{
							"",
							"",
						}},
					},
				})
			},
			wantCode: apperr.CodeDerivedExperimentSourceNotEligible,
		},
		{
			name: "session保存失敗",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{
							"fixed",
							"conclusion",
						}},
					},
					execErrors: []error{errors.New("session")},
				})
			},
			want: "insert derivation briefing session",
		},
		{
			name: "operation保存失敗",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{
							"fixed",
							"conclusion",
						}},
					},
					execErrors: []error{
						nil,
						errors.New("operation"),
					},
				})
			},
			want: "insert derivation briefing operation",
		},
		{
			name: "commit失敗",
			prepare: func(_ *testing.T, store *Store) {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{
							"fixed",
							"conclusion",
						}},
					},
					commitError: errors.New("commit"),
				})
			},
			want: "commit derivation briefing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newDerivationBriefingTestStore(t)
			tt.prepare(t, store)
			_, _, err := store.BeginDerivationBriefing(context.Background(), "request", "source")
			if err == nil {
				t.Fatal("BeginDerivationBriefing() error = nil, want error")
			}
			if tt.wantCode != "" && !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("BeginDerivationBriefing() error = %v, want code %q", err, tt.wantCode)
			}
			if tt.want != "" && !strings.Contains(err.Error(), tt.want) {
				got := err.Error()
				t.Errorf("BeginDerivationBriefing() error = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// 派生ブリーフ状態同期の保存失敗。
func TestStoreUpdateDerivationBriefingFailures(t *testing.T) {
	tests := []struct {
		name string
		tx   *fakeBriefingTransaction
		want string
	}{
		{
			name: "transaction開始失敗",
			want: "begin derivation briefing update",
		},
		{
			name: "operation更新失敗",
			tx:   &fakeBriefingTransaction{execErrors: []error{errors.New("operation")}},
			want: "update derivation briefing operation",
		},
		{
			name: "operation更新件数取得失敗",
			tx:   &fakeBriefingTransaction{result: fakeBriefingResult{rowsAffectedError: errors.New("count")}},
			want: "count derivation briefing operation updates",
		},
		{
			name: "request不在",
			tx:   &fakeBriefingTransaction{result: fakeBriefingResult{rowsAffected: 0}},
			want: "request not found",
		},
		{
			name: "session更新失敗",
			tx: &fakeBriefingTransaction{execErrors: []error{
				nil,
				errors.New("session"),
			}},
			want: "update derivation briefing session",
		},
		{
			name: "commit失敗",
			tx:   &fakeBriefingTransaction{commitError: errors.New("commit")},
			want: "commit derivation briefing update",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newDerivationBriefingTestStore(t)
			if tt.tx == nil {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, errors.New("begin")
				}
			} else {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(tt.tx)
			}
			err := store.updateDerivationBriefing(context.Background(), "request", domain.BriefingStartStateStarted, "")
			if err == nil {
				t.Fatal("updateDerivationBriefing() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("updateDerivationBriefing() error = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// 派生ブリーフ再生と検索の補助分岐。
func TestDerivationBriefingHelpers(t *testing.T) {
	start := domain.DerivationBriefingStart{SourceExperimentID: "source"}
	if _, _, err := replayDerivationBriefing(start, "other"); !apperr.IsCode(err, apperr.CodeDerivedExperimentRequestConflict) {
		t.Errorf("replayDerivationBriefing() error = %v, want request conflict", err)
	}
	store := newDerivationBriefingTestStore(t)
	if err := store.db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, _, err := store.findDerivationBriefing(context.Background(), "request"); err == nil {
		t.Error("findDerivationBriefing() error = nil, want error")
	}
}

// 一意競合後の開始結果再読込。
func TestStoreBeginDerivationBriefingConflictReplay(t *testing.T) {
	tests := []struct {
		name     string
		stage    derivationBriefingConflictStage
		want     string
		wantCode apperr.Code
	}{
		{
			name:  "再読込失敗",
			stage: derivationBriefingConflictReadError,
			want:  "conflict read",
		},
		{
			name:  "再読込で開始結果なし",
			stage: derivationBriefingConflictMissing,
			want:  "insert derivation briefing operation",
		},
		{
			name:  "同じ派生元を再生",
			stage: derivationBriefingConflictSameSource,
		},
		{
			name:     "別の派生元を拒否",
			stage:    derivationBriefingConflictOtherSource,
			wantCode: apperr.CodeDerivedExperimentRequestConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newDerivationBriefingConflictStore(t, tt.stage)
			got, created, err := store.BeginDerivationBriefing(context.Background(), "request", "source")
			if tt.want == "" && tt.wantCode == "" {
				if err != nil {
					t.Fatalf("BeginDerivationBriefing() error = %v", err)
				}
				if created {
					t.Error("BeginDerivationBriefing() created = true, want false")
				}
				if got.SourceExperimentID != "source" {
					t.Errorf("SourceExperimentID = %q, want %q", got.SourceExperimentID, "source")
				}
				return
			}
			if err == nil {
				t.Fatal("BeginDerivationBriefing() error = nil, want error")
			}
			if tt.wantCode != "" && !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("BeginDerivationBriefing() error = %v, want code %q", err, tt.wantCode)
			}
			if tt.want != "" && !strings.Contains(err.Error(), tt.want) {
				got := err.Error()
				t.Errorf("BeginDerivationBriefing() error = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

type derivationBriefingConflictStage string

const (
	derivationBriefingConflictReadError   derivationBriefingConflictStage = "read-error"
	derivationBriefingConflictMissing     derivationBriefingConflictStage = "missing"
	derivationBriefingConflictSameSource  derivationBriefingConflictStage = "same-source"
	derivationBriefingConflictOtherSource derivationBriefingConflictStage = "other-source"
)

const derivationBriefingConflictDriverName = "context-lab-derivation-briefing-conflict"

var derivationBriefingConflictDriverOnce sync.Once

// 一意競合用SQLite driverのstore。
func newDerivationBriefingConflictStore(t *testing.T, stage derivationBriefingConflictStage) *Store {
	t.Helper()
	derivationBriefingConflictDriverOnce.Do(func() {
		sql.Register(derivationBriefingConflictDriverName, derivationBriefingConflictDriver{})
	})
	database, err := sql.Open(derivationBriefingConflictDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	return &Store{
		db: database,
		beginBriefingTransaction: func(ctx context.Context) (briefingTransaction, error) {
			tx, beginErr := database.BeginTx(ctx, nil)
			if beginErr != nil {
				return nil, beginErr
			}
			return sqliteBriefingTransaction{tx: tx}, nil
		},
	}
}

type derivationBriefingConflictDriver struct{}

func (derivationBriefingConflictDriver) Open(name string) (driver.Conn, error) {
	return &derivationBriefingConflictConnection{stage: derivationBriefingConflictStage(name)}, nil
}

type derivationBriefingConflictConnection struct {
	stage          derivationBriefingConflictStage
	operationReads int
}

func (*derivationBriefingConflictConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (*derivationBriefingConflictConnection) Close() error { return nil }
func (c *derivationBriefingConflictConnection) Begin() (driver.Tx, error) {
	return derivationBriefingConflictTransaction{}, nil
}
func (c *derivationBriefingConflictConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}
func (c *derivationBriefingConflictConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "FROM derivation_briefing_operations") {
		c.operationReads++
		if c.operationReads == 1 {
			return &derivationBriefingConflictRows{columns: derivationBriefingOperationColumns()}, nil
		}
		if c.stage == derivationBriefingConflictReadError {
			return nil, errors.New("conflict read")
		}
		if c.stage == derivationBriefingConflictMissing {
			return &derivationBriefingConflictRows{columns: derivationBriefingOperationColumns()}, nil
		}
		sourceID := "source"
		if c.stage == derivationBriefingConflictOtherSource {
			sourceID = "other"
		}
		return &derivationBriefingConflictRows{
			columns: derivationBriefingOperationColumns(),
			values: [][]driver.Value{
				{
					"request",
					sourceID,
					"session",
					"operation",
					"starting",
					"",
				},
			},
		}, nil
	}
	if strings.Contains(query, "FROM experiments e") {
		return &derivationBriefingConflictRows{
			columns: []string{
				"fixed_condition_id",
				"conclusion_id",
			},
			values: [][]driver.Value{
				{
					"fixed",
					"conclusion",
				},
			},
		}, nil
	}
	return nil, errors.New("unexpected query")
}
func (*derivationBriefingConflictConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "INSERT INTO derivation_briefing_operations") {
		return nil, derivationBriefingUniqueError{}
	}
	return driver.RowsAffected(1), nil
}

type derivationBriefingConflictTransaction struct{}

func (derivationBriefingConflictTransaction) Commit() error   { return nil }
func (derivationBriefingConflictTransaction) Rollback() error { return nil }

type derivationBriefingConflictRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *derivationBriefingConflictRows) Columns() []string { return r.columns }
func (*derivationBriefingConflictRows) Close() error        { return nil }
func (r *derivationBriefingConflictRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

// 派生ブリーフ操作の検索列。
func derivationBriefingOperationColumns() []string {
	return []string{
		"request_id",
		"source_experiment_id",
		"preparation_session_id",
		"operation_id",
		"state",
		"failure_code",
	}
}

// 派生ブリーフ用の空SQLite store。
func newDerivationBriefingTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return store
}
