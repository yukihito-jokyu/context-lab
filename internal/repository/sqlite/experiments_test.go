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

// SQLite実験ブリーフ再読込の保存内容取得。
func TestStoreGetExperimentBriefing(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	statements := []struct {
		query string
		args  []any
	}{
		{
			query: "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			args: []any{
				"session-1",
				"experiment_brief",
				"started",
				"2026-08-09T00:00:00Z",
				"2026-08-09T00:01:00Z",
			},
		},
		{
			query: "INSERT INTO briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)",
			args: []any{
				"session-1",
				1,
				"user",
				"目的を確認したい",
				"2026-08-09T00:02:00Z",
			},
		},
		{
			query: "INSERT INTO briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)",
			args: []any{
				"session-1",
				2,
				"assistant",
				"比較案を提示します",
				"2026-08-09T00:03:00Z",
			},
		},
		{
			query: "INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, hypothesis, success_criteria, required_conditions, open_question, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			args: []any{
				"version-1",
				"session-1",
				1,
				"旧版",
				nil,
				"旧基準",
				"旧条件",
				nil,
				"2026-08-09T00:04:00Z",
			},
		},
		{
			query: "INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, hypothesis, success_criteria, required_conditions, open_question, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			args: []any{
				"version-2",
				"session-1",
				2,
				"新しい比較案",
				"仮説",
				"正確性",
				"固定条件",
				"追加確認",
				"2026-08-09T00:05:00Z",
			},
		},
	}
	for _, statement := range statements {
		if _, err := store.db.Exec(statement.query, statement.args...); err != nil {
			t.Fatalf("seed briefing error = %v", err)
		}
	}

	tests := []struct {
		name      string
		sessionID string
		wantFound bool
	}{
		{
			name:      "会話と最新版を返す",
			sessionID: "session-1",
			wantFound: true,
		},
		{
			name:      "未知sessionを返さない",
			sessionID: "missing",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found, err := store.GetExperimentBriefing(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("GetExperimentBriefing() error = %v", err)
			}
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}
			if !tt.wantFound {
				return
			}
			if got.State != "started" {
				t.Errorf("State = %q, want %q", got.State, "started")
			}
			if gotMessages := got.Messages; len(gotMessages) != 2 || gotMessages[0].SequenceNo != 1 || gotMessages[1].SequenceNo != 2 {
				t.Errorf("Messages = %+v, want sequence 1 and 2", gotMessages)
			}
			if got.LatestBrief == nil {
				t.Fatal("LatestBrief = nil, want latest version")
			}
			if got := got.LatestBrief.VersionID; got != "version-2" {
				t.Errorf("LatestBrief.VersionID = %q, want %q", got, "version-2")
			}
			if got := got.LastConfirmedAt; !got.Equal(time.Date(2026, time.August, 9, 0, 5, 0, 0, time.UTC)) {
				t.Errorf("LastConfirmedAt = %s, want %s", got, "2026-08-09 00:05:00 +0000 UTC")
			}
		})
	}
}

// SQLite実験ブリーフ再読込の読み出し失敗。
func TestStoreGetExperimentBriefingFailure(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	_, _, err = store.GetExperimentBriefing(context.Background(), "session-1")
	if err == nil {
		t.Fatal("GetExperimentBriefing() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "find briefing session") {
		t.Errorf("GetExperimentBriefing() error = %q, want find briefing session error", got)
	}
}

// SQLite実験ブリーフ再読込の各読み出し失敗。
func TestStoreGetExperimentBriefingReadFailures(t *testing.T) {
	tests := []struct {
		name     string
		scenario briefingReadScenario
		want     string
	}{
		{
			name:     "会話取得失敗を返す",
			scenario: briefingReadMessagesQueryError,
			want:     "query briefing messages",
		},
		{
			name:     "ブリーフ取得失敗を返す",
			scenario: briefingReadBriefQueryError,
			want:     "find latest briefing version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBriefingReadTestStore(t, tt.scenario)

			_, _, err := store.GetExperimentBriefing(context.Background(), "session-1")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("GetExperimentBriefing() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフの補助読み出し分岐。
func TestStoreExperimentBriefingReadBranches(t *testing.T) {
	tests := []struct {
		name     string
		scenario briefingReadScenario
		read     func(context.Context, *Store) error
		want     string
	}{
		{
			name:     "sessionの空updated_atはcreated_atへフォールバックする",
			scenario: briefingReadSessionEmptyUpdatedAt,
			read: func(ctx context.Context, store *Store) error {
				briefing, found, err := store.findExperimentBriefingSession(ctx, "session-1")
				if err == nil && (!found || briefing.LastConfirmedAt.IsZero()) {
					return errors.New("session fallback result is invalid")
				}

				return err
			},
		},
		{
			name:     "sessionの日時不正を返す",
			scenario: briefingReadSessionInvalidTime,
			read: func(ctx context.Context, store *Store) error {
				_, _, err := store.findExperimentBriefingSession(ctx, "session-1")

				return err
			},
			want: "parse briefing session update time",
		},
		{
			name:     "会話rowsのclose失敗を反復失敗として返す",
			scenario: briefingReadMessagesCloseError,
			read: func(ctx context.Context, store *Store) error {
				_, _, err := store.listExperimentBriefingMessages(ctx, "session-1")

				return err
			},
			want: "iterate briefing messages",
		},
		{
			name:     "会話scan失敗を返す",
			scenario: briefingReadMessagesScanError,
			read: func(ctx context.Context, store *Store) error {
				_, _, err := store.listExperimentBriefingMessages(ctx, "session-1")

				return err
			},
			want: "scan briefing message",
		},
		{
			name:     "会話日時不正を返す",
			scenario: briefingReadMessagesInvalidTime,
			read: func(ctx context.Context, store *Store) error {
				_, _, err := store.listExperimentBriefingMessages(ctx, "session-1")

				return err
			},
			want: "parse briefing message creation time",
		},
		{
			name:     "会話反復失敗を返す",
			scenario: briefingReadMessagesRowsError,
			read: func(ctx context.Context, store *Store) error {
				_, _, err := store.listExperimentBriefingMessages(ctx, "session-1")

				return err
			},
			want: "iterate briefing messages",
		},
		{
			name:     "ブリーフ日時不正を返す",
			scenario: briefingReadBriefInvalidTime,
			read: func(ctx context.Context, store *Store) error {
				_, _, err := store.findLatestExperimentBrief(ctx, "session-1")

				return err
			},
			want: "parse briefing version creation time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBriefingReadTestStore(t, tt.scenario)

			err := tt.read(context.Background(), store)
			if tt.want == "" {
				if err != nil {
					t.Errorf("read() error = %v, want nil", err)
				}

				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("read() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// briefingReadScenario は実験ブリーフ読み出し失敗を再現する種別。
type briefingReadScenario string

const (
	briefingReadMessagesQueryError    briefingReadScenario = "messages-query-error"
	briefingReadBriefQueryError       briefingReadScenario = "brief-query-error"
	briefingReadSessionEmptyUpdatedAt briefingReadScenario = "session-empty-updated-at"
	briefingReadSessionInvalidTime    briefingReadScenario = "session-invalid-time"
	briefingReadMessagesCloseError    briefingReadScenario = "messages-close-error"
	briefingReadMessagesScanError     briefingReadScenario = "messages-scan-error"
	briefingReadMessagesInvalidTime   briefingReadScenario = "messages-invalid-time"
	briefingReadMessagesRowsError     briefingReadScenario = "messages-rows-error"
	briefingReadBriefInvalidTime      briefingReadScenario = "brief-invalid-time"
)

// newBriefingReadTestStore は読み出し失敗再現用SQLiteストアを生成する。
func newBriefingReadTestStore(t *testing.T, scenario briefingReadScenario) *Store {
	t.Helper()

	database, err := sql.Open(briefingReadDriverName, string(scenario))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("database.Close() error = %v", err)
		}
	})

	return &Store{db: database}
}

const briefingReadDriverName = "context-lab-briefing-read-failure"

var briefingReadDriverOnce sync.Once

// init は読み出し失敗再現用driverを一度だけ登録する。
func init() {
	briefingReadDriverOnce.Do(func() {
		sql.Register(briefingReadDriverName, briefingReadDriver{})
	})
}

// briefingReadDriver は実験ブリーフ読込専用のdatabase driver。
type briefingReadDriver struct{}

// Open はscenarioを接続へ渡す。
func (briefingReadDriver) Open(scenario string) (driver.Conn, error) {
	return &briefingReadConnection{scenario: briefingReadScenario(scenario)}, nil
}

// briefingReadConnection はqueryごとの失敗を返す接続。
type briefingReadConnection struct {
	scenario briefingReadScenario
}

// Prepare はこのdriverで利用しない。
func (*briefingReadConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}

// Close は接続を閉じる。
func (*briefingReadConnection) Close() error {
	return nil
}

// Begin はこのdriverで利用しない。
func (*briefingReadConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported")
}

// QueryContext はscenarioに応じた実験ブリーフ読み出し結果を返す。
func (c *briefingReadConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "FROM preparation_sessions"):
		return briefingReadSessionRows(c.scenario)
	case strings.Contains(query, "FROM briefing_messages"):
		return briefingReadMessageRows(c.scenario)
	case strings.Contains(query, "FROM briefing_versions"):
		return briefingReadVersionRows(c.scenario)
	default:
		return nil, errors.New("unexpected query")
	}
}

// briefingReadSessionRows はsession queryの結果を返す。
func briefingReadSessionRows(scenario briefingReadScenario) (driver.Rows, error) {
	updatedAt := "2026-08-09T00:00:00Z"
	if scenario == briefingReadSessionEmptyUpdatedAt {
		updatedAt = ""
	}
	if scenario == briefingReadSessionInvalidTime {
		updatedAt = "invalid-time"
	}

	return &briefingReadRows{
		columns: []string{
			"state",
			"created_at",
			"updated_at",
		},
		values: [][]driver.Value{{
			"started",
			"2026-08-09T00:00:00Z",
			updatedAt,
		}},
	}, nil
}

// briefingReadMessageRows はmessage queryの結果を返す。
func briefingReadMessageRows(scenario briefingReadScenario) (driver.Rows, error) {
	if scenario == briefingReadMessagesQueryError {
		return nil, errors.New("messages query failed")
	}
	rows := &briefingReadRows{columns: []string{
		"role",
		"content",
		"sequence_no",
		"created_at",
	}}
	switch scenario {
	case briefingReadMessagesCloseError:
		rows.closeErr = errors.New("messages close failed")
	case briefingReadMessagesScanError:
		rows.values = [][]driver.Value{{
			"user",
			"content",
			"invalid-sequence",
			"2026-08-09T00:00:00Z",
		}}
	case briefingReadMessagesInvalidTime:
		rows.values = [][]driver.Value{{
			"user",
			"content",
			int64(1),
			"invalid-time",
		}}
	case briefingReadMessagesRowsError:
		rows.nextErr = errors.New("messages iteration failed")
	}

	return rows, nil
}

// briefingReadVersionRows はbrief version queryの結果を返す。
func briefingReadVersionRows(scenario briefingReadScenario) (driver.Rows, error) {
	if scenario == briefingReadBriefQueryError {
		return nil, errors.New("brief query failed")
	}
	createdAt := "2026-08-09T00:00:00Z"
	if scenario == briefingReadBriefInvalidTime {
		createdAt = "invalid-time"
	}

	return &briefingReadRows{
		columns: []string{
			"id",
			"decision",
			"hypothesis",
			"success_criteria",
			"required_conditions",
			"open_question",
			"created_at",
		},
		values: [][]driver.Value{{
			"version-1",
			"decision",
			nil,
			"criteria",
			"conditions",
			nil,
			createdAt,
		}},
	}, nil
}

// briefingReadRows は任意のrows結果または走査・close失敗を表す。
type briefingReadRows struct {
	columns  []string
	values   [][]driver.Value
	position int
	nextErr  error
	closeErr error
}

// Columns は列名を返す。
func (r *briefingReadRows) Columns() []string {
	return r.columns
}

// Close は指定済みclose失敗を返す。
func (r *briefingReadRows) Close() error {
	return r.closeErr
}

// Next は次のrowまたは指定済み反復失敗を返す。
func (r *briefingReadRows) Next(destination []driver.Value) error {
	if r.nextErr != nil {
		return r.nextErr
	}
	if r.position >= len(r.values) {
		return io.EOF
	}
	copy(destination, r.values[r.position])
	r.position++

	return nil
}

// SQLite実験ブリーフ状態更新時の確認時刻前進。
func TestStoreMarkExperimentBriefingUpdatesConfirmationTime(t *testing.T) {
	tests := []struct {
		name   string
		update func(context.Context, *Store, string) error
		state  string
	}{
		{
			name: "開始済み状態で確認時刻を更新する",
			update: func(ctx context.Context, store *Store, requestID string) error {
				return store.MarkExperimentBriefingStarted(ctx, requestID)
			},
			state: domain.BriefingStartStateStarted,
		},
		{
			name: "失敗状態で確認時刻を更新する",
			update: func(ctx context.Context, store *Store, requestID string) error {
				return store.MarkExperimentBriefingFailed(ctx, requestID, "SAFE_FAILURE")
			},
			state: domain.BriefingStartStateFailed,
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
			start, created, err := store.BeginExperimentBriefing(context.Background(), "request-1")
			if err != nil {
				t.Fatalf("BeginExperimentBriefing() error = %v", err)
			}
			if !created {
				t.Fatal("created = false, want true")
			}
			before, found, err := store.GetExperimentBriefing(context.Background(), start.BriefingSessionID)
			if err != nil {
				t.Fatalf("GetExperimentBriefing() before update error = %v", err)
			}
			if !found {
				t.Fatal("briefing found = false, want true")
			}
			if err := tt.update(context.Background(), store, "request-1"); err != nil {
				t.Fatalf("state update error = %v", err)
			}
			after, found, err := store.GetExperimentBriefing(context.Background(), start.BriefingSessionID)
			if err != nil {
				t.Fatalf("GetExperimentBriefing() after update error = %v", err)
			}
			if !found {
				t.Fatal("briefing found = false, want true")
			}
			if got := after.State; got != tt.state {
				t.Errorf("State = %q, want %q", got, tt.state)
			}
			if !after.LastConfirmedAt.After(before.LastConfirmedAt) {
				t.Errorf("LastConfirmedAt = %s, want after %s", after.LastConfirmedAt, before.LastConfirmedAt)
			}
			if after.LastConfirmedAt.Location() != time.UTC {
				t.Errorf("LastConfirmedAt location = %s, want UTC", after.LastConfirmedAt.Location())
			}
		})
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
