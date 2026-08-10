package sqlite

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SQLite派生壁打ちメッセージの保存、再生、session検証を確認。
func TestStoreDerivationBriefMessage(t *testing.T) {
	tests := []struct {
		name     string
		session  string
		kind     string
		state    string
		wantCode apperr.Code
	}{
		{
			name:    "開始済み派生壁打ちへ保存する",
			session: "session-started",
			kind:    "derivation_brief",
			state:   domain.BriefingStartStateStarted,
		},
		{
			name:     "別種別sessionを拒否する",
			session:  "session-other",
			kind:     "experiment_brief",
			state:    domain.BriefingStartStateStarted,
			wantCode: apperr.CodeDerivationBriefingMessageNotFound,
		},
		{
			name:     "未開始sessionを拒否する",
			session:  "session-starting",
			kind:     "derivation_brief",
			state:    domain.BriefingStartStateStarting,
			wantCode: apperr.CodeDerivationBriefingMessageNotActive,
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
			if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", tt.session, tt.kind, tt.state, "2026-08-10T00:00:00Z", "2026-08-10T00:00:00Z"); err != nil {
				t.Fatalf("seed preparation session error = %v", err)
			}

			operation, created, err := store.BeginDerivationBriefMessage(context.Background(), "request-1", tt.session)
			if tt.wantCode != "" {
				assertDerivationBriefingMessageError(t, err, tt.wantCode)

				return
			}
			if err != nil {
				t.Fatalf("BeginDerivationBriefMessage() error = %v", err)
			}
			if !created {
				t.Fatal("created = false, want true")
			}
			result := domain.DerivationBriefingMessageResult{
				AssistantMessage: "提案を整理しました",
				Suggestion: &domain.ExperimentBrief{
					Decision:           "比較する",
					SuccessCriteria:    "正確性",
					RequiredConditions: "固定条件",
				},
			}
			if err := store.CompleteDerivationBriefMessage(context.Background(), "request-1", "派生案を考えたい", result); err != nil {
				t.Fatalf("CompleteDerivationBriefMessage() error = %v", err)
			}
			second, secondCreated, secondErr := store.BeginDerivationBriefMessage(context.Background(), "request-1", tt.session)
			if secondErr != nil {
				t.Fatalf("second BeginDerivationBriefMessage() error = %v", secondErr)
			}
			if secondCreated {
				t.Error("second created = true, want false")
			}
			if second.OperationID != operation.OperationID || second.State != domain.BriefingStartStateStarted {
				t.Errorf("second = %+v, want completed operation", second)
			}
			var createdAt, updatedAt string
			if err := store.db.QueryRow("SELECT created_at, updated_at FROM derivation_briefing_message_operations WHERE request_id=?", "request-1").Scan(&createdAt, &updatedAt); err != nil {
				t.Fatalf("operation timestamps error = %v", err)
			}
			if createdAt == "" || updatedAt == "" {
				t.Errorf("timestamps = (%q, %q), want both populated", createdAt, updatedAt)
			}
			var messages, suggestions int
			if err := store.db.QueryRow("SELECT COUNT(*) FROM derivation_briefing_messages").Scan(&messages); err != nil {
				t.Fatalf("message count error = %v", err)
			}
			if err := store.db.QueryRow("SELECT COUNT(*) FROM derivation_briefing_suggestions").Scan(&suggestions); err != nil {
				t.Fatalf("suggestion count error = %v", err)
			}
			if messages != 2 || suggestions != 1 {
				t.Errorf("stored = (%d messages, %d suggestions), want (2, 1)", messages, suggestions)
			}
		})
	}
}

// SQLite派生壁打ち会話送信完了の全I/O失敗を確認。
func TestStoreCompleteDerivationBriefMessageFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		result  domain.DerivationBriefingMessageResult
		want    string
	}{
		{
			name: "連番取得失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{err: errors.New("sequence unavailable")}}})
			},
			want: "find next derivation briefing message sequence",
		},
		{
			name: "利用者会話保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows:       []briefingRow{fakeBriefingRow{values: []any{1}}},
					execErrors: []error{errors.New("user insert unavailable")},
				})
			},
			want: "insert derivation user briefing message",
		},
		{
			name: "AI会話保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{fakeBriefingRow{values: []any{1}}},
					execErrors: []error{
						nil,
						errors.New("assistant insert unavailable"),
					},
				})
			},
			result: domain.DerivationBriefingMessageResult{AssistantMessage: "AI応答"},
			want:   "insert derivation assistant briefing message",
		},
		{
			name: "提案保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{1}},
						fakeBriefingRow{values: []any{1}},
					},
					execErrors: []error{
						nil,
						errors.New("suggestion insert unavailable"),
					},
				})
			},
			result: domain.DerivationBriefingMessageResult{Suggestion: &domain.ExperimentBrief{
				Decision:           "判断",
				SuccessCriteria:    "基準",
				RequiredConditions: "条件",
			}},
			want: "insert derivation briefing suggestion",
		},
		{
			name: "operation更新失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{fakeBriefingRow{values: []any{1}}},
					execErrors: []error{
						nil,
						errors.New("operation update unavailable"),
					},
				})
			},
			want: "complete derivation briefing message operation",
		},
		{
			name: "session更新失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{fakeBriefingRow{values: []any{1}}},
					execErrors: []error{
						nil,
						nil,
						errors.New("session update unavailable"),
					},
				})
			},
			want: "update derivation briefing session",
		},
		{
			name: "確定失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows:        []briefingRow{fakeBriefingRow{values: []any{1}}},
					commitError: errors.New("commit unavailable"),
				})
			},
			want: "commit derivation briefing message completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newCompleteDerivationBriefingMessageTestStore(t)
			tt.prepare(t, store)

			err := store.CompleteDerivationBriefMessage(context.Background(), "request-1", "利用者メッセージ", tt.result)
			if err == nil {
				t.Fatal("CompleteDerivationBriefMessage() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("CompleteDerivationBriefMessage() error = %q, want containing %q", got, tt.want)
			}
		})
	}
}

// SQLite派生壁打ち会話送信完了の境界失敗を確認。
func TestStoreCompleteDerivationBriefMessageBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		want    string
	}{
		{
			name: "operation検索失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			want: "database is closed",
		},
		{
			name: "operation不在を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DELETE FROM derivation_briefing_message_operations"); err != nil {
					t.Fatalf("delete derivation briefing message operation error = %v", err)
				}
			},
			want: "request not found",
		},
		{
			name: "トランザクション開始失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, errors.New("begin unavailable")
				}
			},
			want: "begin derivation briefing message completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newCompleteDerivationBriefingMessageTestStore(t)
			tt.prepare(t, store)

			err := store.CompleteDerivationBriefMessage(context.Background(), "request-1", "利用者メッセージ", domain.DerivationBriefingMessageResult{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("CompleteDerivationBriefMessage() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// newCompleteDerivationBriefingMessageTestStore は会話送信完了用の保存済みoperationを生成。
func newCompleteDerivationBriefingMessageTestStore(t *testing.T) *Store {
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
	createdAt := "2026-08-10T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "derivation_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO derivation_briefing_message_operations (id, request_id, preparation_session_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", "operation-1", "request-1", "session-1", domain.BriefingStartStateStarting, createdAt, createdAt); err != nil {
		t.Fatalf("seed derivation briefing message operation error = %v", err)
	}

	return store
}

// SQLite派生壁打ちメッセージのrequest ID競合を確認。
func TestStoreDerivationBriefMessageRequestConflict(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	for _, sessionID := range []string{
		"session-1",
		"session-2",
	} {
		if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", sessionID, "derivation_brief", domain.BriefingStartStateStarted, "2026-08-10T00:00:00Z", "2026-08-10T00:00:00Z"); err != nil {
			t.Fatalf("seed preparation session error = %v", err)
		}
	}
	if _, _, err := store.BeginDerivationBriefMessage(context.Background(), "request-1", "session-1"); err != nil {
		t.Fatalf("BeginDerivationBriefMessage() error = %v", err)
	}
	_, _, err = store.BeginDerivationBriefMessage(context.Background(), "request-1", "session-2")
	assertDerivationBriefingMessageError(t, err, apperr.CodeDerivationBriefingMessageRequestConflict)
}

// SQLite派生壁打ちメッセージの同一session並行完了を確認。
func TestStoreCompleteDerivationBriefMessageConcurrently(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "derivation_brief", domain.BriefingStartStateStarted, "2026-08-10T00:00:00Z", "2026-08-10T00:00:00Z"); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}
	for _, requestID := range []string{
		"request-1",
		"request-2",
	} {
		if _, _, err := store.BeginDerivationBriefMessage(context.Background(), requestID, "session-1"); err != nil {
			t.Fatalf("BeginDerivationBriefMessage() error = %v", err)
		}
	}

	var group sync.WaitGroup
	errors := make(chan error, 2)
	for _, requestID := range []string{
		"request-1",
		"request-2",
	} {
		group.Add(1)
		go func(requestID string) {
			defer group.Done()
			errors <- store.CompleteDerivationBriefMessage(
				context.Background(),
				requestID,
				"派生案を考えたい",
				domain.DerivationBriefingMessageResult{
					AssistantMessage: "提案を整理しました",
					Suggestion: &domain.ExperimentBrief{
						Decision:           "比較する",
						SuccessCriteria:    "正確性",
						RequiredConditions: "固定条件",
					},
				},
			)
		}(requestID)
	}
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("CompleteDerivationBriefMessage() error = %v", err)
		}
	}
	var messages, suggestions int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM derivation_briefing_messages").Scan(&messages); err != nil {
		t.Fatalf("message count error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM derivation_briefing_suggestions").Scan(&suggestions); err != nil {
		t.Fatalf("suggestion count error = %v", err)
	}
	if messages != 4 || suggestions != 2 {
		t.Errorf("stored = (%d messages, %d suggestions), want (4, 2)", messages, suggestions)
	}
}

// assertDerivationBriefingMessageError は安全なエラーコードを確認する。
func assertDerivationBriefingMessageError(t *testing.T, err error, want apperr.Code) {
	t.Helper()
	appErr := apperr.As(err)
	if appErr == nil {
		t.Fatalf("apperr.As(error) = nil, error = %v", err)
	}
	if appErr.Code != want {
		t.Errorf("Code = %q, want %q", appErr.Code, want)
	}
}
