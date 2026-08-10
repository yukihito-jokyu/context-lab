package sqlite

import (
	"context"
	"testing"
	"time"
)

// SQLite派生実験ブリーフ再読込の保存内容取得。
func TestStoreGetDerivationBriefing(t *testing.T) {
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
				"derivation_brief",
				"started",
				"2026-08-10T00:00:00Z",
				"2026-08-10T00:01:00Z",
			},
		},
		{
			query: "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			args: []any{
				"other-kind",
				"experiment_brief",
				"started",
				"2026-08-10T00:00:00Z",
				"2026-08-10T00:01:00Z",
			},
		},
		{
			query: "INSERT INTO derivation_briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)",
			args: []any{
				"session-1",
				1,
				"user",
				"比較対象を変えたい",
				"2026-08-10T00:02:00Z",
			},
		},
		{
			query: "INSERT INTO derivation_briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)",
			args: []any{
				"session-1",
				2,
				"assistant",
				"差分案を作成します",
				"2026-08-10T00:03:00Z",
			},
		},
		{
			query: "INSERT INTO derivation_briefing_suggestions (id, preparation_session_id, operation_id, version_no, purpose, decision, hypothesis, candidate_prompts, evaluation_criteria, environment_conditions, initial_input, success_criteria, required_conditions, open_question, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			args: []any{
				"suggestion-1",
				"session-1",
				"operation-1",
				1,
				"旧目的",
				"旧判断",
				nil,
				"[\"旧prompt\"]",
				"旧評価",
				"旧環境",
				"旧入力",
				"旧成功",
				"旧条件",
				nil,
				"2026-08-10T00:04:00Z",
			},
		},
		{
			query: "INSERT INTO derivation_briefing_suggestions (id, preparation_session_id, operation_id, version_no, purpose, decision, hypothesis, candidate_prompts, evaluation_criteria, environment_conditions, initial_input, success_criteria, required_conditions, open_question, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			args: []any{
				"suggestion-2",
				"session-1",
				"operation-2",
				2,
				"新目的",
				"新判断",
				"仮説",
				"[\"prompt A\",\"prompt B\"]",
				"新評価",
				"新環境",
				"新入力",
				"新成功",
				"新条件",
				"未解決",
				"2026-08-10T00:05:00Z",
			},
		},
	}
	for _, statement := range statements {
		if _, err := store.db.Exec(statement.query, statement.args...); err != nil {
			t.Fatalf("seed derivation briefing error = %v", err)
		}
	}

	tests := []struct {
		name      string
		sessionID string
		wantFound bool
	}{
		{
			name:      "会話と最新版提案を返す",
			sessionID: "session-1",
			wantFound: true,
		},
		{
			name:      "別kindのsessionを返さない",
			sessionID: "other-kind",
			wantFound: false,
		},
		{
			name:      "未知sessionを返さない",
			sessionID: "missing",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found, err := store.GetDerivationBriefing(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("GetDerivationBriefing() error = %v", err)
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
			if got.LatestSuggestion == nil {
				t.Fatal("LatestSuggestion = nil, want latest suggestion")
			}
			if got := got.LatestSuggestion.ID; got != "suggestion-2" {
				t.Errorf("LatestSuggestion.ID = %q, want %q", got, "suggestion-2")
			}
			if got := got.LastConfirmedAt; !got.Equal(time.Date(2026, time.August, 10, 0, 5, 0, 0, time.UTC)) {
				t.Errorf("LastConfirmedAt = %s, want %s", got, "2026-08-10 00:05:00 +0000 UTC")
			}
		})
	}
}

// SQLite派生実験ブリーフ再読込の空状態取得。
func TestStoreGetDerivationBriefingEmpty(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-empty", "derivation_brief", "starting", "2026-08-10T00:00:00Z", "2026-08-10T00:00:00Z"); err != nil {
		t.Fatalf("seed empty derivation briefing error = %v", err)
	}

	got, found, err := store.GetDerivationBriefing(context.Background(), "session-empty")
	if err != nil {
		t.Fatalf("GetDerivationBriefing() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got.Messages == nil || len(got.Messages) != 0 {
		t.Errorf("Messages = %+v, want empty slice", got.Messages)
	}
	if got.LatestSuggestion != nil {
		t.Errorf("LatestSuggestion = %+v, want nil", got.LatestSuggestion)
	}
}
