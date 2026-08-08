package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	"reflect"
	"time"
)

// SQLite実験ブリーフ開始の原子的記録と再利用。
func TestStoreBeginExperimentBriefing(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	first, created, err := store.BeginExperimentBriefing(context.Background(), "request-1")
	if err != nil {
		t.Fatalf("BeginExperimentBriefing() error = %v", err)
	}
	if !created {
		t.Error("created = false, want true")
	}
	if first.BriefingSessionID == "" || first.OperationID == "" {
		t.Errorf("start = %+v, want generated identifiers", first)
	}

	second, created, err := store.BeginExperimentBriefing(context.Background(), "request-1")
	if err != nil {
		t.Fatalf("second BeginExperimentBriefing() error = %v", err)
	}
	if created {
		t.Error("second created = true, want false")
	}
	if second != first {
		t.Errorf("second start = %+v, want %+v", second, first)
	}

	var sessionCount, operationCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM preparation_sessions").Scan(&sessionCount); err != nil {
		t.Fatalf("preparation session count error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM briefing_operations").Scan(&operationCount); err != nil {
		t.Fatalf("briefing operation count error = %v", err)
	}
	if sessionCount != 1 {
		t.Errorf("preparation sessions = %d, want 1", sessionCount)
	}
	if operationCount != 1 {
		t.Errorf("briefing operations = %d, want 1", operationCount)
	}
}

// SQLite実験ブリーフ開始の並行request ID再利用。
func TestStoreBeginExperimentBriefingConcurrently(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	store.db.SetMaxOpenConns(1)

	type result struct {
		start   domain.ExperimentBriefingStart
		created bool
		err     error
	}
	ready := make(chan struct{})
	results := make(chan result, 2)
	for range 2 {
		go func() {
			<-ready
			start, created, beginErr := store.BeginExperimentBriefing(context.Background(), "request-1")
			results <- result{
				start:   start,
				created: created,
				err:     beginErr,
			}
		}()
	}
	close(ready)

	first := <-results
	second := <-results
	if first.err != nil {
		t.Fatalf("first BeginExperimentBriefing() error = %v", first.err)
	}
	if second.err != nil {
		t.Fatalf("second BeginExperimentBriefing() error = %v", second.err)
	}
	if first.start.BriefingSessionID != second.start.BriefingSessionID {
		t.Errorf("BriefingSessionID = %q and %q, want same identifier", first.start.BriefingSessionID, second.start.BriefingSessionID)
	}
	if first.start.OperationID != second.start.OperationID {
		t.Errorf("OperationID = %q and %q, want same identifier", first.start.OperationID, second.start.OperationID)
	}
	if first.created == second.created {
		t.Errorf("created = %v and %v, want exactly one creation", first.created, second.created)
	}
}

// SQLite実験ブリーフrequest ID競合判定。
func TestIsBriefingRequestConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "request IDの一意制約競合を検出する",
			err:  errors.New("UNIQUE constraint failed: briefing_operations.request_id"),
			want: true,
		},
		{
			name: "別の保存失敗を競合と扱わない",
			err:  errors.New("database locked"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBriefingRequestConflict(tt.err); got != tt.want {
				t.Errorf("isBriefingRequestConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ開始の保存失敗。
func TestStoreBeginExperimentBriefingFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		want    string
	}{
		{
			name: "開始結果の検索失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			want: "find briefing operation",
		},
		{
			name: "識別子生成失敗を返す",
			prepare: func(t *testing.T, _ *Store) {
				t.Helper()
				replaceBriefingRandom(t, func([]byte) (int, error) {
					return 0, errors.New("random unavailable")
				})
			},
			want: "read random identifier",
		},
		{
			name: "操作識別子生成失敗を返す",
			prepare: func(t *testing.T, _ *Store) {
				t.Helper()
				calls := 0
				replaceBriefingRandom(t, func(bytes []byte) (int, error) {
					calls++
					if calls == 1 {
						return len(bytes), nil
					}

					return 0, errors.New("random unavailable")
				})
			},
			want: "read random identifier",
		},
		{
			name: "トランザクション開始失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, errors.New("begin unavailable")
				}
			},
			want: "begin experiment briefing",
		},
		{
			name: "準備セッション保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{execErrors: []error{errors.New("session insert failed")}})
			},
			want: "insert preparation session",
		},
		{
			name: "操作開始意図保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					execErrors: []error{
						nil,
						errors.New("operation insert failed"),
					},
				})
			},
			want: "insert briefing operation",
		},
		{
			name: "競合後の既存開始結果検索失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				transaction := &fakeBriefingTransaction{
					execErrors: []error{
						nil,
						errors.New("UNIQUE constraint failed: briefing_operations.request_id"),
					},
				}
				transaction.onExec = func(calls int) {
					if calls != 2 {
						return
					}
					if err := store.db.Close(); err != nil {
						t.Fatalf("Close() error = %v", err)
					}
				}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)
			},
			want: "database is closed",
		},
		{
			name: "トランザクション確定失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{commitError: errors.New("commit failed")})
			},
			want: "commit experiment briefing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := Open(t.TempDir())
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})
			tt.prepare(t, store)

			_, _, err = store.BeginExperimentBriefing(context.Background(), "request-1")
			if err == nil {
				t.Fatal("BeginExperimentBriefing() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("BeginExperimentBriefing() error = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ開始状態同期の保存失敗。
func TestStoreUpdateExperimentBriefingFailures(t *testing.T) {
	tests := []struct {
		name string
		tx   *fakeBriefingTransaction
		want string
	}{
		{
			name: "操作状態更新失敗を返す",
			tx: &fakeBriefingTransaction{
				execErrors: []error{errors.New("operation update failed")},
			},
			want: "update briefing operation",
		},
		{
			name: "操作更新件数取得失敗を返す",
			tx: &fakeBriefingTransaction{
				result: fakeBriefingResult{rowsAffectedError: errors.New("count failed")},
			},
			want: "count briefing operation updates",
		},
		{
			name: "開始要求不在を返す",
			tx: &fakeBriefingTransaction{
				result: fakeBriefingResult{rowsAffected: 0},
			},
			want: "request not found",
		},
		{
			name: "準備セッション状態更新失敗を返す",
			tx: &fakeBriefingTransaction{
				execErrors: []error{
					nil,
					errors.New("session update failed"),
				},
			},
			want: "update preparation session",
		},
		{
			name: "状態同期確定失敗を返す",
			tx: &fakeBriefingTransaction{
				commitError: errors.New("commit failed"),
			},
			want: "commit briefing update",
		},
		{
			name: "準備セッション更新件数取得失敗を返す",
			tx: &fakeBriefingTransaction{
				results: []sql.Result{
					fakeBriefingResult{rowsAffected: 1},
					fakeBriefingResult{rowsAffectedError: errors.New("session count failed")},
				},
			},
			want: "count preparation session updates",
		},
		{
			name: "準備セッション不在を返す",
			tx: &fakeBriefingTransaction{
				results: []sql.Result{
					fakeBriefingResult{rowsAffected: 1},
					fakeBriefingResult{rowsAffected: 0},
				},
			},
			want: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := Open(t.TempDir())
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})
			store.beginBriefingTransaction = fakeBriefingTransactionFactory(tt.tx)

			err = store.updateExperimentBriefing(context.Background(), "request-1", domain.BriefingStartStateStarted, "")
			if err == nil {
				t.Fatal("updateExperimentBriefing() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("updateExperimentBriefing() error = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ状態同期の開始失敗。
func TestStoreUpdateExperimentBriefingBeginFailure(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
		return nil, errors.New("begin unavailable")
	}

	err = store.updateExperimentBriefing(context.Background(), "request-1", domain.BriefingStartStateStarted, "")
	if err == nil {
		t.Fatal("updateExperimentBriefing() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "begin briefing update") {
		t.Errorf("updateExperimentBriefing() error = %q, want begin briefing update error", got)
	}
}

// 乱数読み出し差し替え。
func replaceBriefingRandom(t *testing.T, replacement func([]byte) (int, error)) {
	t.Helper()

	previous := readBriefingRandom
	readBriefingRandom = replacement
	t.Cleanup(func() {
		readBriefingRandom = previous
	})
}

// fakeBriefingTransactionFactory はトランザクション開始test doubleを生成。
func fakeBriefingTransactionFactory(transaction briefingTransaction) func(context.Context) (briefingTransaction, error) {
	return func(context.Context) (briefingTransaction, error) {
		return transaction, nil
	}
}

// fakeBriefingTransaction はトランザクション境界のtest double。
type fakeBriefingTransaction struct {
	execErrors  []error
	execCalls   int
	result      sql.Result
	results     []sql.Result
	commitError error
	onExec      func(int)
}

// ExecContext は指定済みの実行結果を返却。
func (f *fakeBriefingTransaction) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	f.execCalls++
	if f.onExec != nil {
		f.onExec(f.execCalls)
	}
	if f.execCalls <= len(f.execErrors) {
		err := f.execErrors[f.execCalls-1]
		if err != nil {
			return nil, err
		}
	}
	if f.execCalls <= len(f.results) {
		return f.results[f.execCalls-1], nil
	}
	if f.result != nil {
		return f.result, nil
	}

	return fakeBriefingResult{rowsAffected: 1}, nil
}

// Commit は指定済みの確定結果を返却。
func (f *fakeBriefingTransaction) Commit() error {
	return f.commitError
}

// Rollback はロールバックを受理。
func (*fakeBriefingTransaction) Rollback() error {
	return nil
}

// fakeBriefingResult はSQL実行結果のtest double。
type fakeBriefingResult struct {
	rowsAffected      int64
	rowsAffectedError error
}

// LastInsertId は未使用の挿入識別子を返却。
func (fakeBriefingResult) LastInsertId() (int64, error) {
	return 0, nil
}

// RowsAffected は指定済みの更新件数を返却。
func (f fakeBriefingResult) RowsAffected() (int64, error) {
	return f.rowsAffected, f.rowsAffectedError
}

// SQLite実験ブリーフ開始状態の同期。
func TestStoreMarkExperimentBriefing(t *testing.T) {
	tests := []struct {
		name        string
		mark        func(*Store, context.Context, string) error
		wantState   string
		wantFailure string
	}{
		{
			name: "開始済み状態を同期する",
			mark: func(store *Store, ctx context.Context, requestID string) error {
				return store.MarkExperimentBriefingStarted(ctx, requestID)
			},
			wantState: domain.BriefingStartStateStarted,
		},
		{
			name: "安全な失敗コードを同期する",
			mark: func(store *Store, ctx context.Context, requestID string) error {
				return store.MarkExperimentBriefingFailed(ctx, requestID, "ACP_NOT_READY")
			},
			wantState:   domain.BriefingStartStateFailed,
			wantFailure: "ACP_NOT_READY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := Open(t.TempDir())
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})
			if _, _, err := store.BeginExperimentBriefing(context.Background(), "request-1"); err != nil {
				t.Fatalf("BeginExperimentBriefing() error = %v", err)
			}

			if err := tt.mark(store, context.Background(), "request-1"); err != nil {
				t.Fatalf("mark() error = %v", err)
			}
			start, found, err := store.findExperimentBriefing(context.Background(), "request-1")
			if err != nil {
				t.Fatalf("findExperimentBriefing() error = %v", err)
			}
			if !found {
				t.Fatal("found = false, want true")
			}
			if start.State != tt.wantState {
				t.Errorf("State = %q, want %q", start.State, tt.wantState)
			}
			if start.FailureCode != tt.wantFailure {
				t.Errorf("FailureCode = %q, want %q", start.FailureCode, tt.wantFailure)
			}
		})
	}
}

// SQLite初期化と実験一覧読み出し。
func TestStoreListExperiments(t *testing.T) {
	tests := []struct {
		name                string
		seed                func(*testing.T, *Store)
		wantExperiments     []string
		wantCancelled       []string
		wantLastConfirmedAt bool
		wantDriverAvailable bool
	}{
		{
			name: "空のデータベースは空配列を返す",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
			},
			wantExperiments:     []string{},
			wantCancelled:       []string{},
			wantLastConfirmedAt: true,
			wantDriverAvailable: true,
		},
		{
			name: "取消済みを別配列へ分離する",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				seedExperiments(t, store)
			},
			wantExperiments: []string{
				"experiment-running",
				"experiment-planned",
			},
			wantCancelled: []string{
				"experiment-cancelled",
			},
			wantLastConfirmedAt: true,
			wantDriverAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			tt.seed(t, store)

			var version string
			if err := store.db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
				t.Fatalf("sqlite_version() error = %v", err)
			}
			if gotDriverAvailable := version != ""; gotDriverAvailable != tt.wantDriverAvailable {
				t.Errorf("driver available = %v, want %v", gotDriverAvailable, tt.wantDriverAvailable)
			}

			got, err := store.ListExperiments(context.Background())
			if err != nil {
				t.Fatalf("ListExperiments() error = %v", err)
			}
			if gotIDs := experimentIDs(got.Experiments); !reflect.DeepEqual(gotIDs, tt.wantExperiments) {
				t.Errorf("Experiments IDs = %v, want %v", gotIDs, tt.wantExperiments)
			}
			if gotIDs := experimentIDs(got.CancelledExperiments); !reflect.DeepEqual(gotIDs, tt.wantCancelled) {
				t.Errorf("CancelledExperiments IDs = %v, want %v", gotIDs, tt.wantCancelled)
			}
			if gotLastConfirmedAt := got.LastConfirmedAt != nil; gotLastConfirmedAt != tt.wantLastConfirmedAt {
				t.Errorf("LastConfirmedAt available = %v, want %v", gotLastConfirmedAt, tt.wantLastConfirmedAt)
			}
			if got.LastConfirmedAt != nil {
				var persisted string
				if err := store.db.QueryRow("SELECT value FROM application_metadata WHERE key = ?", "last_confirmed_at").Scan(&persisted); err != nil {
					t.Fatalf("last_confirmed_at query error = %v", err)
				}
				if got := got.LastConfirmedAt.Format(time.RFC3339Nano); got != persisted {
					t.Errorf("LastConfirmedAt = %q, want persisted %q", got, persisted)
				}
			}
		})
	}
}

// SQLite読み出し失敗の安全なrepositoryエラー化。
func TestStoreListExperimentsErrors(t *testing.T) {
	tests := []struct {
		name string
		seed func(*testing.T, *Store)
	}{
		{
			name: "通常実験のquery失敗",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE experiments"); err != nil {
					t.Fatalf("DROP TABLE error = %v", err)
				}
			},
		},
		{
			name: "取消済み実験の変換失敗",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("INSERT INTO experiments (id, purpose, state, progress_summary, updated_at) VALUES (?, ?, ?, ?, ?)", "cancelled", "中止", "cancelled", "中止", "invalid-time"); err != nil {
					t.Fatalf("INSERT cancelled experiment error = %v", err)
				}
			},
		},
		{
			name: "最終確認時刻の記録失敗",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE application_metadata"); err != nil {
					t.Fatalf("DROP TABLE metadata error = %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			tt.seed(t, store)

			_, err := store.ListExperiments(context.Background())
			if err == nil {
				t.Error("ListExperiments() error = nil, want repository error")
			}
		})
	}
}

// 一時SQLiteストア生成。
func newTestStore(t *testing.T) *Store {
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

// 実験fixture投入。
func seedExperiments(t *testing.T, store *Store) {
	t.Helper()

	statements := []struct {
		query string
		args  []any
	}{
		{
			query: "INSERT INTO experiments (id, purpose, state, progress_summary, derived_from_experiment_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			args: []any{
				"experiment-planned",
				"計画",
				"planned",
				"未開始",
				nil,
				"2026-08-08T01:00:00Z",
			},
		},
		{
			query: "INSERT INTO experiments (id, purpose, state, progress_summary, derived_from_experiment_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			args: []any{
				"experiment-running",
				"実行中",
				"running",
				"評価中",
				"experiment-planned",
				"2026-08-08T02:00:00Z",
			},
		},
		{
			query: "INSERT INTO experiments (id, purpose, state, progress_summary, derived_from_experiment_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			args: []any{
				"experiment-cancelled",
				"中止済み",
				"cancelled",
				"中止",
				nil,
				"2026-08-08T03:00:00Z",
			},
		},
		{
			query: "INSERT INTO application_metadata (key, value) VALUES (?, ?)",
			args: []any{
				"last_confirmed_at",
				"2026-08-08T01:02:03Z",
			},
		},
	}

	for _, statement := range statements {
		if _, err := store.db.Exec(statement.query, statement.args...); err != nil {
			t.Fatalf("seed database error = %v", err)
		}
	}
}

// 実験ID抽出。
func experimentIDs(experiments []domain.Experiment) []string {
	ids := make([]string, 0, len(experiments))
	for _, experiment := range experiments {
		ids = append(ids, experiment.ID)
	}

	return ids
}
