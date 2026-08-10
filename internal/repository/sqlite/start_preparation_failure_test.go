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

const preparationFailureDriverName = "context-lab-preparation-start-failure"

var preparationFailureDriverOnce sync.Once

// StartPreparationのdatabase/sql異常分岐を確認。
func TestStorePreparationStartDatabaseFailures(t *testing.T) {
	tests := []struct {
		name  string
		stage string
		call  func(*Store) error
	}{
		{
			name:  "開始結果検索失敗",
			stage: "find",
			call: func(store *Store) error {
				_, _, err := store.BeginPreparation(context.Background(), "request", ".")
				return err
			},
		},
		{
			name:  "開始transaction失敗",
			stage: "begin",
			call: func(store *Store) error {
				_, _, err := store.BeginPreparation(context.Background(), "request", ".")
				return err
			},
		},
		{
			name:  "session保存失敗",
			stage: "session",
			call: func(store *Store) error {
				_, _, err := store.BeginPreparation(context.Background(), "request", ".")
				return err
			},
		},
		{
			name:  "operation保存失敗",
			stage: "operation",
			call: func(store *Store) error {
				_, _, err := store.BeginPreparation(context.Background(), "request", ".")
				return err
			},
		},
		{
			name:  "開始commit失敗",
			stage: "commit",
			call: func(store *Store) error {
				_, _, err := store.BeginPreparation(context.Background(), "request", ".")
				return err
			},
		},
		{
			name:  "更新transaction失敗",
			stage: "begin",
			call: func(store *Store) error {
				return store.MarkPreparationRunning(context.Background(), "request")
			},
		},
		{
			name:  "更新operation検索失敗",
			stage: "update-find",
			call: func(store *Store) error {
				return store.MarkPreparationRunning(context.Background(), "request")
			},
		},
		{
			name:  "更新operation不存在",
			stage: "update-notfound",
			call: func(store *Store) error {
				return store.MarkPreparationRunning(context.Background(), "request")
			},
		},
		{
			name:  "更新operation保存失敗",
			stage: "update-operation",
			call: func(store *Store) error {
				return store.MarkPreparationRunning(context.Background(), "request")
			},
		},
		{
			name:  "更新session保存失敗",
			stage: "update-session",
			call: func(store *Store) error {
				return store.MarkPreparationRunning(context.Background(), "request")
			},
		},
		{
			name:  "更新commit失敗",
			stage: "commit",
			call: func(store *Store) error {
				return store.MarkPreparationRunning(context.Background(), "request")
			},
		},
		{
			name:  "候補保存失敗",
			stage: "candidate",
			call: func(store *Store) error {
				return store.CompletePreparation(context.Background(), "request", domain.EnvironmentPreparationResult{Candidates: []domain.EnvironmentPreparationCandidate{
					{
						EnvironmentConditions: "macOS",
						Summary:               "safe",
					},
				}})
			},
		},
		{
			name:  "診断保存失敗",
			stage: "diagnostic",
			call: func(store *Store) error {
				return store.CompletePreparation(context.Background(), "request", domain.EnvironmentPreparationResult{Diagnostics: []domain.EnvironmentPreparationDiagnostic{
					{
						Code:        "CHECKED",
						SafeSummary: "safe",
					},
				}})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(newPreparationFailureStore(t, tt.stage)); err == nil {
				t.Error("operation error = nil, want database error")
			}
		})
	}
}

// 環境準備開始の識別子生成失敗を確認。
func TestStorePreparationStartIdentifierFailures(t *testing.T) {
	replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random") })
	if _, _, err := newTestStore(t).BeginPreparation(context.Background(), "request", "."); err == nil {
		t.Error("BeginPreparation() error = nil, want identifier error")
	}
}

// 開始記録transaction境界の異常を確認。
func TestInsertPreparationStartDatabaseFailures(t *testing.T) {
	tests := []struct {
		name  string
		stage string
	}{
		{
			name:  "transaction開始",
			stage: "begin",
		},
		{
			name:  "session保存",
			stage: "session",
		},
		{
			name:  "operation保存",
			stage: "operation",
		},
		{
			name:  "commit",
			stage: "commit",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newPreparationFailureStore(t, tt.stage)
			err := store.insertPreparationStart(context.Background(), domain.EnvironmentPreparationStart{
				RequestID:     "request",
				PreparationID: "preparation",
				Scope:         ".",
				State:         domain.EnvironmentPreparationStateStarting,
			})
			if err == nil {
				t.Error("insertPreparationStart() error = nil, want database error")
			}
		})
	}
}

// BeginPreparationの競合再読込異常を確認。
func TestStoreBeginPreparationConflictRecovery(t *testing.T) {
	tests := []struct {
		name      string
		find      func() func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error)
		insertErr error
		wantCode  bool
		wantOK    bool
	}{
		{
			name: "競合後の検索失敗",
			find: func() func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
				calls := 0
				return func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
					calls++
					if calls == 2 {
						return domain.EnvironmentPreparationStart{}, false, errors.New("find failed")
					}
					return domain.EnvironmentPreparationStart{}, false, nil
				}
			},
			insertErr: errors.New("UNIQUE constraint failed: environment_preparation_operations.request_id"),
		},
		{
			name: "競合後の別scope",
			find: func() func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
				calls := 0
				return func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
					calls++
					if calls == 2 {
						return domain.EnvironmentPreparationStart{Scope: "other"}, true, nil
					}
					return domain.EnvironmentPreparationStart{}, false, nil
				}
			},
			insertErr: errors.New("UNIQUE constraint failed: environment_preparation_operations.request_id"),
			wantCode:  true,
		},
		{
			name: "競合後の同一scope",
			find: func() func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
				calls := 0
				return func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
					calls++
					if calls == 2 {
						return domain.EnvironmentPreparationStart{Scope: "."}, true, nil
					}
					return domain.EnvironmentPreparationStart{}, false, nil
				}
			},
			insertErr: errors.New("UNIQUE constraint failed: environment_preparation_operations.request_id"),
			wantOK:    true,
		},
		{
			name: "競合後に見つからない",
			find: func() func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
				return func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
					return domain.EnvironmentPreparationStart{}, false, nil
				}
			},
			insertErr: errors.New("UNIQUE constraint failed: environment_preparation_operations.request_id"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				findPreparationStartOverride:   tt.find(),
				insertPreparationStartOverride: func(context.Context, domain.EnvironmentPreparationStart) error { return tt.insertErr },
			}
			_, _, err := store.BeginPreparation(context.Background(), "request", ".")
			if (err == nil) != tt.wantOK {
				t.Fatal("BeginPreparation() error = nil, want conflict recovery error")
			}
		})
	}
}

// BeginPreparationの既存開始結果復元を確認。
func TestStoreBeginPreparationExistingStart(t *testing.T) {
	store := &Store{
		findPreparationStartOverride: func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
			return domain.EnvironmentPreparationStart{
				PreparationID: "preparation",
				Scope:         ".",
				State:         domain.EnvironmentPreparationStateRunning,
			}, true, nil
		},
	}
	got, created, err := store.BeginPreparation(context.Background(), "request", ".")
	if err != nil {
		t.Fatalf("BeginPreparation() error = %v", err)
	}
	if created {
		t.Error("created = true, want false")
	}
	if got.PreparationID != "preparation" {
		t.Errorf("PreparationID = %q, want preparation", got.PreparationID)
	}
}

// BeginPreparationの識別子生成失敗を確認。
func TestStoreBeginPreparationIdentifierFailure(t *testing.T) {
	replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random") })
	store := &Store{
		findPreparationStartOverride: func(context.Context, string) (domain.EnvironmentPreparationStart, bool, error) {
			return domain.EnvironmentPreparationStart{}, false, nil
		},
	}
	if _, _, err := store.BeginPreparation(context.Background(), "request", "."); err == nil {
		t.Error("BeginPreparation() error = nil, want identifier error")
	}
}

// 環境準備結果の識別子生成失敗を確認。
func TestInsertPreparationResultIdentifierFailures(t *testing.T) {
	tests := []struct {
		name   string
		result domain.EnvironmentPreparationResult
	}{
		{
			name: "候補識別子",
			result: domain.EnvironmentPreparationResult{Candidates: []domain.EnvironmentPreparationCandidate{
				{
					EnvironmentConditions: "macOS",
					Summary:               "safe",
				},
			}},
		},
		{
			name: "診断識別子",
			result: domain.EnvironmentPreparationResult{Diagnostics: []domain.EnvironmentPreparationDiagnostic{
				{
					Code:        "CHECKED",
					SafeSummary: "safe",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random") })
			if err := newPreparationFailureStore(t, "").CompletePreparation(context.Background(), "request", tt.result); err == nil {
				t.Error("CompletePreparation() error = nil, want identifier error")
			}
		})
	}
}

// preparation failure storeはdatabase/sql失敗注入用storeを返す。
func newPreparationFailureStore(t *testing.T, stage string) *Store {
	t.Helper()
	preparationFailureDriverOnce.Do(func() { sql.Register(preparationFailureDriverName, preparationFailureDriver{}) })
	database, err := sql.Open(preparationFailureDriverName, stage)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return &Store{db: database}
}

type preparationFailureDriver struct{}

func (preparationFailureDriver) Open(stage string) (driver.Conn, error) {
	return &preparationFailureConn{stage: stage}, nil
}

type preparationFailureConn struct {
	stage string
}

func (*preparationFailureConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (*preparationFailureConn) Close() error { return nil }
func (c *preparationFailureConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *preparationFailureConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	if c.stage == "begin" {
		return nil, errors.New("begin failed")
	}
	return preparationFailureTx{stage: c.stage}, nil
}
func (c *preparationFailureConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.stage == "find" || (c.stage == "update-find" && strings.Contains(query, "preparation_session_id")) {
		return nil, errors.New("query failed")
	}
	if c.stage == "update-notfound" && strings.Contains(query, "preparation_session_id") {
		return &preparationFailureRows{columns: []string{"preparation_session_id"}}, nil
	}
	if strings.Contains(query, "preparation_session_id") {
		return &preparationFailureRows{
			columns: []string{"preparation_session_id"},
			values:  [][]driver.Value{{"preparation-1"}},
		}, nil
	}
	return &preparationFailureRows{
		columns: []string{
			"request_id",
			"preparation_session_id",
			"scope",
			"state",
			"failure_code",
		},
	}, nil
}
func (c *preparationFailureConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if (c.stage == "session" && strings.Contains(query, "INSERT INTO preparation_sessions")) ||
		(c.stage == "operation" && strings.Contains(query, "INSERT INTO environment_preparation_operations")) ||
		(c.stage == "candidate" && strings.Contains(query, "INSERT INTO environment_preparation_candidates")) ||
		(c.stage == "diagnostic" && strings.Contains(query, "INSERT INTO environment_preparation_diagnostics")) ||
		(c.stage == "update-operation" && strings.Contains(query, "UPDATE environment_preparation_operations")) ||
		(c.stage == "update-session" && strings.Contains(query, "UPDATE preparation_sessions")) {
		return nil, errors.New("exec failed")
	}
	return driver.RowsAffected(1), nil
}

type preparationFailureTx struct{ stage string }

func (t preparationFailureTx) Commit() error {
	if t.stage == "commit" {
		return errors.New("commit failed")
	}
	return nil
}
func (preparationFailureTx) Rollback() error { return nil }

type preparationFailureRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *preparationFailureRows) Columns() []string { return r.columns }
func (*preparationFailureRows) Close() error        { return nil }
func (r *preparationFailureRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
