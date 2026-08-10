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

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SQLite派生壁打ち終了の状態遷移、再生、失敗復帰を確認。
func TestStoreStopDerivationBriefing(t *testing.T) {
	tests := []struct {
		name       string
		sessionID  string
		kind       string
		state      string
		wantCode   apperr.Code
		complete   bool
		failCode   apperr.Code
		wantState  string
		wantReplay string
	}{
		{
			name:       "停止完了を永続化して再生する",
			sessionID:  "session-complete",
			kind:       "derivation_brief",
			state:      domain.BriefingStartStateStarted,
			complete:   true,
			wantState:  domain.BriefingStartStateStopped,
			wantReplay: domain.BriefingStartStateStopped,
		},
		{
			name:       "ACP失敗後は開始済みに復帰して新規要求を受け付ける",
			sessionID:  "session-retry",
			kind:       "derivation_brief",
			state:      domain.BriefingStartStateStarted,
			failCode:   apperr.CodeACPNotReady,
			wantState:  domain.BriefingStartStateStarted,
			wantReplay: domain.BriefingStartStateFailed,
		},
		{
			name:      "別種別sessionを見つからないものとして扱う",
			sessionID: "session-other",
			kind:      "experiment_brief",
			state:     domain.BriefingStartStateStarted,
			wantCode:  apperr.CodeDerivationBriefingStopNotFound,
		},
		{
			name:      "終了済みsessionを拒否する",
			sessionID: "session-stopped",
			kind:      "derivation_brief",
			state:     domain.BriefingStartStateStopped,
			wantCode:  apperr.CodeDerivationBriefingStopNotActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			seedStopDerivationBriefingSession(t, store, tt.sessionID, tt.kind, tt.state)

			operation, created, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", tt.sessionID)
			if tt.wantCode != "" {
				assertStopDerivationBriefingError(t, err, tt.wantCode)

				return
			}
			if err != nil {
				t.Fatalf("BeginStopDerivationBriefing() error = %v", err)
			}
			if !created || operation.OperationID == "" || operation.State != domain.BriefingStartStateStarting {
				t.Fatalf("operation = (%+v, created %v), want new starting operation", operation, created)
			}
			if tt.complete {
				if err := store.CompleteStopDerivationBriefing(context.Background(), "request-1"); err != nil {
					t.Fatalf("CompleteStopDerivationBriefing() error = %v", err)
				}
			} else {
				if err := store.FailStopDerivationBriefing(context.Background(), "request-1", string(tt.failCode)); err != nil {
					t.Fatalf("FailStopDerivationBriefing() error = %v", err)
				}
			}
			assertStopDerivationBriefingSessionState(t, store, tt.sessionID, tt.wantState)

			replayed, replayCreated, replayErr := store.BeginStopDerivationBriefing(context.Background(), "request-1", tt.sessionID)
			if replayErr != nil {
				t.Fatalf("replay BeginStopDerivationBriefing() error = %v", replayErr)
			}
			if replayCreated || replayed.OperationID != operation.OperationID || replayed.State != tt.wantReplay {
				t.Errorf("replay = (%+v, created %v), want operation %q state %q", replayed, replayCreated, operation.OperationID, tt.wantReplay)
			}
			if tt.failCode != "" {
				retry, retryCreated, retryErr := store.BeginStopDerivationBriefing(context.Background(), "request-2", tt.sessionID)
				if retryErr != nil {
					t.Fatalf("retry BeginStopDerivationBriefing() error = %v", retryErr)
				}
				if !retryCreated || retry.State != domain.BriefingStartStateStarting {
					t.Errorf("retry = (%+v, created %v), want new starting operation", retry, retryCreated)
				}
			}
		})
	}
}

// SQLite派生壁打ち終了の2接続並行要求を確認。
func TestStoreBeginStopDerivationBriefingAcrossTwoPools(t *testing.T) {
	directory := t.TempDir()
	first, err := Open(directory)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := first.Close(); err != nil {
			t.Errorf("first Close() error = %v", err)
		}
	})
	second, err := Open(directory)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Errorf("second Close() error = %v", err)
		}
	})
	seedStopDerivationBriefingSession(t, first, "session-1", "derivation_brief", domain.BriefingStartStateStarted)

	type result struct {
		operation domain.DerivationBriefingStopOperation
		created   bool
		err       error
	}
	results := make(chan result, 2)
	var group sync.WaitGroup
	for index, store := range []*Store{
		first,
		second,
	} {
		group.Add(1)
		go func(index int, store *Store) {
			defer group.Done()
			operation, created, err := store.BeginStopDerivationBriefing(context.Background(), "request-"+string(rune('1'+index)), "session-1")
			results <- result{
				operation: operation,
				created:   created,
				err:       err,
			}
		}(index, store)
	}
	group.Wait()
	close(results)

	created := 0
	notActive := 0
	for got := range results {
		if got.err == nil && got.created {
			created++
			continue
		}
		if apperr.IsCode(got.err, apperr.CodeDerivationBriefingStopNotActive) {
			notActive++
			continue
		}
		t.Errorf("BeginStopDerivationBriefing() = (%+v, %v, %v), want one create or not-active", got.operation, got.created, got.err)
	}
	if created != 1 || notActive != 1 {
		t.Errorf("results = (%d created, %d not-active), want (1, 1)", created, notActive)
	}
	assertStopDerivationBriefingSessionState(t, first, "session-1", domain.BriefingStartStateStopping)
}

// SQLite派生壁打ち終了のrollbackとI/O失敗境界を確認。
func TestStoreStopDerivationBriefingFailures(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *Store) error
		want string
	}{
		{
			name: "既存停止操作検索失敗を返す",
			run: func(t *testing.T, store *Store) error {
				t.Helper()
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				return err
			},
			want: "find derivation briefing stop operation",
		},
		{
			name: "停止識別子生成失敗を返す",
			run: func(t *testing.T, store *Store) error {
				t.Helper()
				replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random unavailable") })
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				return err
			},
			want: "generate derivation briefing stop operation ID",
		},
		{
			name: "開始transaction失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) { return nil, errors.New("begin unavailable") }
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				return err
			},
			want: "begin stop derivation briefing",
		},
		{
			name: "開始中でないsessionはrollbackして拒否する",
			run: func(t *testing.T, store *Store) error {
				t.Helper()
				seedStopDerivationBriefingSession(t, store, "session-1", "derivation_brief", domain.BriefingStartStateStopped)
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				return err
			},
			want: "終了できる状態ではありません",
		},
		{
			name: "停止操作確定失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{commitError: errors.New("commit unavailable")})
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				return err
			},
			want: "commit stop derivation briefing",
		},
		{
			name: "状態変更失敗はrollbackする",
			run: func(t *testing.T, store *Store) error {
				t.Helper()
				tx := &fakeBriefingTransaction{execErrors: []error{errors.New("state update unavailable")}}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(tx)
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				if tx.rollbackCalls != 1 {
					t.Errorf("Rollback() calls = %d, want 1", tx.rollbackCalls)
				}
				return err
			},
			want: "mark derivation briefing stopping",
		},
		{
			name: "状態同期開始失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) { return nil, errors.New("begin unavailable") }
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "begin derivation briefing stop update",
		},
		{
			name: "状態同期operation更新失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{execErrors: []error{errors.New("operation unavailable")}})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "update derivation briefing stop operation",
		},
		{
			name: "状態同期operation件数取得失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{result: fakeBriefingResult{rowsAffectedError: errors.New("count unavailable")}})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "count derivation briefing stop operation updates",
		},
		{
			name: "状態同期operation不在を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{result: fakeBriefingResult{rowsAffected: 0}})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "request not found",
		},
		{
			name: "状態変更件数取得失敗はrollbackする",
			run: func(t *testing.T, store *Store) error {
				t.Helper()
				tx := &fakeBriefingTransaction{result: fakeBriefingResult{rowsAffectedError: errors.New("count unavailable")}}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(tx)
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				if tx.rollbackCalls != 1 {
					t.Errorf("Rollback() calls = %d, want 1", tx.rollbackCalls)
				}
				return err
			},
			want: "count derivation briefing stopping updates",
		},
		{
			name: "操作保存失敗はrollbackする",
			run: func(t *testing.T, store *Store) error {
				t.Helper()
				tx := &fakeBriefingTransaction{execErrors: []error{
					nil,
					errors.New("insert unavailable"),
				}}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(tx)
				_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
				if tx.rollbackCalls != 1 {
					t.Errorf("Rollback() calls = %d, want 1", tx.rollbackCalls)
				}
				return err
			},
			want: "insert derivation briefing stop operation",
		},
		{
			name: "状態同期のsession更新失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{execErrors: []error{
					nil,
					errors.New("session unavailable"),
				}})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "update derivation briefing stop session",
		},
		{
			name: "状態同期session件数取得失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{results: []sql.Result{
					fakeBriefingResult{rowsAffected: 1},
					fakeBriefingResult{rowsAffectedError: errors.New("count unavailable")},
				}})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "count derivation briefing stop session updates",
		},
		{
			name: "状態同期session不在を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{results: []sql.Result{
					fakeBriefingResult{rowsAffected: 1},
					fakeBriefingResult{rowsAffected: 0},
				}})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "session not found",
		},
		{
			name: "状態同期確定失敗を返す",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{commitError: errors.New("commit unavailable")})
				return store.updateStopDerivationBriefing(context.Background(), "request-1", domain.BriefingStartStateStopped, "")
			},
			want: "commit derivation briefing stop update",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			err := tt.run(t, store)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want containing %q", err, tt.want)
			}
		})
	}

	if !isDerivationBriefingStopRequestConflict(errors.New("UNIQUE constraint failed: derivation_briefing_stop_operations.request_id")) {
		t.Error("isDerivationBriefingStopRequestConflict() = false, want true")
	}
	if isDerivationBriefingStopRequestConflict(errors.New("other")) {
		t.Error("isDerivationBriefingStopRequestConflict() = true, want false")
	}
}

// SQLite派生壁打ち終了のrequest競合再生を確認。
func TestStoreBeginStopDerivationBriefingConflictReplay(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(*testing.T, *Store)
		wantCode  apperr.Code
		wantFound bool
	}{
		{
			name: "既存要求の別sessionを拒否する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				seedStopDerivationBriefingOperation(t, store, "operation-1", "request-1", "other-session")
			},
			wantCode: apperr.CodeDerivationBriefingStopRequestConflict,
		},
		{
			name: "状態競合後に同一session操作を再生する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					result: fakeBriefingResult{rowsAffected: 0},
					onExec: func(int) {
						seedStopDerivationBriefingOperation(t, store, "operation-1", "request-1", "session-1")
					},
				})
			},
			wantFound: true,
		},
		{
			name: "状態競合後に別session操作を拒否する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					result: fakeBriefingResult{rowsAffected: 0},
					onExec: func(int) {
						seedStopDerivationBriefingOperation(t, store, "operation-1", "request-1", "other-session")
					},
				})
			},
			wantCode: apperr.CodeDerivationBriefingStopRequestConflict,
		},
		{
			name: "停止操作の一意競合後に同一session操作を再生する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					execErrors: []error{
						nil,
						errors.New("UNIQUE constraint failed: derivation_briefing_stop_operations.request_id"),
					},
					onExec: func(call int) {
						if call == 2 {
							seedStopDerivationBriefingOperation(t, store, "operation-1", "request-1", "session-1")
						}
					},
				})
			},
			wantFound: true,
		},
		{
			name: "停止操作の一意競合後に別session操作を拒否する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					execErrors: []error{
						nil,
						errors.New("UNIQUE constraint failed: derivation_briefing_stop_operations.request_id"),
					},
					onExec: func(call int) {
						if call == 2 {
							seedStopDerivationBriefingOperation(t, store, "operation-1", "request-1", "other-session")
						}
					},
				})
			},
			wantCode: apperr.CodeDerivationBriefingStopRequestConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			tt.prepare(t, store)
			operation, created, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
			if tt.wantCode != "" {
				assertStopDerivationBriefingError(t, err, tt.wantCode)

				return
			}
			if err != nil {
				t.Fatalf("BeginStopDerivationBriefing() error = %v", err)
			}
			if !tt.wantFound || created || operation.OperationID != "operation-1" {
				t.Errorf("result = (%+v, created %v), want replayed operation", operation, created)
			}
		})
	}
}

// SQLite派生壁打ち終了の競合後読込失敗を物理driverで確認。
func TestStoreBeginStopDerivationBriefingConflictReadFailures(t *testing.T) {
	tests := []struct {
		name  string
		stage derivationBriefingStopConflictStage
		want  string
	}{
		{
			name:  "状態競合後の操作読込失敗を返す",
			stage: derivationBriefingStopConflictFindError,
			want:  "stop operation read unavailable",
		},
		{
			name:  "状態競合後のsession存在確認失敗を返す",
			stage: derivationBriefingStopConflictExistsError,
			want:  "stop session exists unavailable",
		},
		{
			name:  "一意競合後の操作読込失敗を返す",
			stage: derivationBriefingStopConflictUniqueFindError,
			want:  "stop operation read unavailable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newDerivationBriefingStopConflictStore(t, tt.stage)
			_, _, err := store.BeginStopDerivationBriefing(context.Background(), "request-1", "session-1")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("BeginStopDerivationBriefing() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

type derivationBriefingStopConflictStage string

const (
	derivationBriefingStopConflictFindError       derivationBriefingStopConflictStage = "find-error"
	derivationBriefingStopConflictExistsError     derivationBriefingStopConflictStage = "exists-error"
	derivationBriefingStopConflictUniqueFindError derivationBriefingStopConflictStage = "unique-find-error"
)

const derivationBriefingStopConflictDriverName = "context-lab-derivation-briefing-stop-conflict"

var derivationBriefingStopConflictDriverOnce sync.Once

// stopDerivationBriefingConflictStore は競合後読込失敗用driverを設定する。
func newDerivationBriefingStopConflictStore(t *testing.T, stage derivationBriefingStopConflictStage) *Store {
	t.Helper()
	derivationBriefingStopConflictDriverOnce.Do(func() {
		sql.Register(derivationBriefingStopConflictDriverName, derivationBriefingStopConflictDriver{})
	})
	database, err := sql.Open(derivationBriefingStopConflictDriverName, string(stage))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return &Store{
		db: database,
		beginBriefingTransaction: func(ctx context.Context) (briefingTransaction, error) {
			tx, err := database.BeginTx(ctx, nil)
			if err != nil {
				return nil, err
			}

			return sqliteBriefingTransaction{tx: tx}, nil
		},
	}
}

type derivationBriefingStopConflictDriver struct{}

func (derivationBriefingStopConflictDriver) Open(stage string) (driver.Conn, error) {
	return &derivationBriefingStopConflictConnection{stage: derivationBriefingStopConflictStage(stage)}, nil
}

type derivationBriefingStopConflictConnection struct {
	stage          derivationBriefingStopConflictStage
	operationReads int
}

func (*derivationBriefingStopConflictConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (*derivationBriefingStopConflictConnection) Close() error { return nil }
func (*derivationBriefingStopConflictConnection) Begin() (driver.Tx, error) {
	return derivationBriefingStopConflictTransaction{}, nil
}
func (c *derivationBriefingStopConflictConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}
func (c *derivationBriefingStopConflictConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "FROM derivation_briefing_stop_operations") {
		c.operationReads++
		if c.operationReads == 2 && (c.stage == derivationBriefingStopConflictFindError || c.stage == derivationBriefingStopConflictUniqueFindError) {
			return nil, errors.New("stop operation read unavailable")
		}

		return &derivationBriefingStopConflictRows{columns: []string{
			"preparation_session_id",
			"id",
			"state",
			"failure_code",
		}}, nil
	}
	if strings.Contains(query, "SELECT EXISTS(SELECT 1 FROM preparation_sessions") && c.stage == derivationBriefingStopConflictExistsError {
		return nil, errors.New("stop session exists unavailable")
	}

	return nil, errors.New("unexpected query")
}
func (c *derivationBriefingStopConflictConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "UPDATE preparation_sessions SET state") {
		if c.stage == derivationBriefingStopConflictFindError || c.stage == derivationBriefingStopConflictExistsError {
			return driver.RowsAffected(0), nil
		}

		return driver.RowsAffected(1), nil
	}
	if strings.Contains(query, "INSERT INTO derivation_briefing_stop_operations") {
		return nil, errors.New("UNIQUE constraint failed: derivation_briefing_stop_operations.request_id")
	}

	return nil, errors.New("unexpected exec")
}

type derivationBriefingStopConflictTransaction struct{}

func (derivationBriefingStopConflictTransaction) Commit() error   { return nil }
func (derivationBriefingStopConflictTransaction) Rollback() error { return nil }

type derivationBriefingStopConflictRows struct {
	columns []string
}

func (r *derivationBriefingStopConflictRows) Columns() []string { return r.columns }
func (*derivationBriefingStopConflictRows) Close() error        { return nil }
func (*derivationBriefingStopConflictRows) Next([]driver.Value) error {
	return io.EOF
}

// stopDerivationBriefingSession は派生壁打ち終了用sessionを保存。
func seedStopDerivationBriefingSession(t *testing.T, store *Store, sessionID, kind, state string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", sessionID, kind, state, now, now); err != nil {
		t.Fatalf("insert preparation session error = %v", err)
	}
}

// stopDerivationBriefingOperation は停止操作を保存。
func seedStopDerivationBriefingOperation(t *testing.T, store *Store, operationID, requestID, sessionID string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := store.db.Exec("INSERT INTO derivation_briefing_stop_operations (id, request_id, preparation_session_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", operationID, requestID, sessionID, domain.BriefingStartStateStarting, now, now); err != nil {
		t.Fatalf("insert stop operation error = %v", err)
	}
}

// stopDerivationBriefingSessionState はsession状態を確認。
func assertStopDerivationBriefingSessionState(t *testing.T, store *Store, sessionID, want string) {
	t.Helper()
	var got string
	if err := store.db.QueryRow("SELECT state FROM preparation_sessions WHERE id=?", sessionID).Scan(&got); err != nil {
		t.Fatalf("session state query error = %v", err)
	}
	if got != want {
		t.Errorf("State = %q, want %q", got, want)
	}
}

// stopDerivationBriefingError は安全な終了エラーを確認。
func assertStopDerivationBriefingError(t *testing.T, err error, want apperr.Code) {
	t.Helper()
	appErr := apperr.As(err)
	if appErr == nil {
		t.Fatalf("apperr.As(error) = nil, error = %v", err)
	}
	if appErr.Code != want {
		t.Errorf("Code = %q, want %q", appErr.Code, want)
	}
}
