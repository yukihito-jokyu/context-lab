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
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
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

// SQLite実験ブリーフ会話送信の保存と冪等記録。
func TestStoreExperimentBriefMessage(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	createdAt := "2026-08-09T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}

	operation, created, err := store.BeginExperimentBriefMessage(context.Background(), "request-1", "session-1")
	if err != nil {
		t.Fatalf("BeginExperimentBriefMessage() error = %v", err)
	}
	if !created {
		t.Fatal("created = false, want true")
	}
	result := domain.ExperimentBriefingMessageResult{
		AssistantMessage: "評価基準を確認します",
		Brief: &domain.ExperimentBrief{
			Decision:           "比較する",
			SuccessCriteria:    "正確性",
			RequiredConditions: "固定条件",
		},
	}
	if err := store.CompleteExperimentBriefMessage(context.Background(), "request-1", "目的を確認したい", result); err != nil {
		t.Fatalf("CompleteExperimentBriefMessage() error = %v", err)
	}

	second, created, err := store.BeginExperimentBriefMessage(context.Background(), "request-1", "session-1")
	if err != nil {
		t.Fatalf("second BeginExperimentBriefMessage() error = %v", err)
	}
	if created {
		t.Error("second created = true, want false")
	}
	if second.OperationID != operation.OperationID || second.State != domain.BriefingStartStateStarted {
		t.Errorf("second = %+v, want completed operation", second)
	}
	briefing, found, err := store.GetExperimentBriefing(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetExperimentBriefing() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got := len(briefing.Messages); got != 2 {
		t.Errorf("Messages length = %d, want 2", got)
	}
	if briefing.LatestBrief == nil || briefing.LatestBrief.Decision != "比較する" {
		t.Errorf("LatestBrief = %+v, want saved brief", briefing.LatestBrief)
	}
}

// SQLiteブリーフ採用の原子的保存と冪等性。
func TestStoreCreateExperimentFromBrief(t *testing.T) {
	tests := []struct {
		name           string
		requestID      string
		briefVersionID string
		wantCode       apperr.Code
		wantCreated    bool
	}{
		{
			name:           "採用値を準備中実験へ保存する",
			requestID:      "request-1",
			briefVersionID: "version-1",
			wantCreated:    true,
		},
		{
			name:           "他sessionの版を拒否する",
			requestID:      "request-2",
			briefVersionID: "other-version",
			wantCode:       apperr.CodeExperimentBriefVersionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentCreationTestStore(t)

			got, created, err := store.CreateExperimentFromBrief(context.Background(), tt.requestID, "session-1", tt.briefVersionID)
			if tt.wantCode != "" {
				if gotCode := apperr.As(err); gotCode == nil || gotCode.Code != tt.wantCode {
					t.Errorf("CreateExperimentFromBrief() error = %v, want code %q", err, tt.wantCode)
				}
				var count int
				if countErr := store.db.QueryRow("SELECT COUNT(*) FROM experiments").Scan(&count); countErr != nil {
					t.Fatalf("experiment count error = %v", countErr)
				}
				if count != 0 {
					t.Errorf("experiments = %d, want 0", count)
				}

				return
			}
			if err != nil {
				t.Fatalf("CreateExperimentFromBrief() error = %v", err)
			}
			if created != tt.wantCreated {
				t.Errorf("created = %v, want %v", created, tt.wantCreated)
			}
			if got.State != "preparing" || got.ExperimentID == "" {
				t.Errorf("creation = %+v, want preparing experiment", got)
			}
			var purpose, environment, evaluation, initialInput string
			if err := store.db.QueryRow("SELECT e.purpose, p.environment_conditions, p.evaluation_criteria, p.initial_input FROM experiments e JOIN experiment_preparations p ON p.experiment_id = e.id WHERE e.id = ?", got.ExperimentID).Scan(&purpose, &environment, &evaluation, &initialInput); err != nil {
				t.Fatalf("saved preparation error = %v", err)
			}
			if purpose != "目的" || environment != "隔離環境" || evaluation != "正確性" || initialInput != "初期入力" {
				t.Errorf("saved preparation = %q, %q, %q, %q, want adopted values", purpose, environment, evaluation, initialInput)
			}
			var promptCount int
			if err := store.db.QueryRow("SELECT COUNT(*) FROM experiment_preparation_prompts WHERE experiment_id = ?", got.ExperimentID).Scan(&promptCount); err != nil {
				t.Fatalf("prompt count error = %v", err)
			}
			if promptCount != 2 {
				t.Errorf("prompt count = %d, want 2", promptCount)
			}
			preparation, found, preparationErr := store.GetExperimentPreparation(context.Background(), got.ExperimentID)
			if preparationErr != nil {
				t.Fatalf("GetExperimentPreparation() error = %v", preparationErr)
			}
			if !found {
				t.Fatal("GetExperimentPreparation() found = false, want true")
			}
			if preparation.Purpose != "目的" || preparation.EnvironmentConditions != "隔離環境" || preparation.EvaluationAxes != "正確性" || preparation.InitialInput != "初期入力" {
				t.Errorf("GetExperimentPreparation() = %+v, want adopted values", preparation)
			}
			if gotPrompts := len(preparation.Prompts); gotPrompts != 2 {
				t.Errorf("GetExperimentPreparation() prompts = %d, want 2", gotPrompts)
			}
			second, secondCreated, secondErr := store.CreateExperimentFromBrief(context.Background(), tt.requestID, "session-1", tt.briefVersionID)
			if secondErr != nil {
				t.Fatalf("second CreateExperimentFromBrief() error = %v", secondErr)
			}
			if secondCreated {
				t.Error("second created = true, want false")
			}
			if second != got {
				t.Errorf("second creation = %+v, want %+v", second, got)
			}
			if _, _, conflictErr := store.CreateExperimentFromBrief(context.Background(), tt.requestID, "session-2", "other-version"); !apperr.IsCode(conflictErr, apperr.CodeExperimentCreateRequestConflict) {
				t.Errorf("conflicting request error = %v, want %q", conflictErr, apperr.CodeExperimentCreateRequestConflict)
			}
		})
	}
}

// SQLite実験準備queryの読込と時系列prompt返却。
func TestStoreGetExperimentPreparation(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	createdAt := "2026-08-09T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, success_criteria, required_conditions, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "version-1", "session-1", 1, "採用", "基準", "条件", createdAt); err != nil {
		t.Fatalf("seed briefing version error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO experiments (id, purpose, state, updated_at) VALUES (?, ?, ?, ?)", "experiment-1", "目的", "preparing", createdAt); err != nil {
		t.Fatalf("seed experiment error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO experiment_preparations (experiment_id, briefing_session_id, briefing_version_id, hypothesis, environment_conditions, initial_input, evaluation_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", "experiment-1", "session-1", "version-1", "仮説", "隔離環境", "初期入力", "正確性", createdAt, createdAt); err != nil {
		t.Fatalf("seed experiment preparation error = %v", err)
	}
	prompts := []struct {
		sequence int
		content  string
	}{
		{
			sequence: 2,
			content:  "prompt B",
		},
		{
			sequence: 1,
			content:  "prompt A",
		},
	}
	for _, prompt := range prompts {
		if _, err := store.db.Exec("INSERT INTO experiment_preparation_prompts (experiment_id, sequence_no, content) VALUES (?, ?, ?)", "experiment-1", prompt.sequence, prompt.content); err != nil {
			t.Fatalf("seed experiment preparation prompt error = %v", err)
		}
	}

	got, found, err := store.GetExperimentPreparation(context.Background(), "experiment-1")
	if err != nil {
		t.Fatalf("GetExperimentPreparation() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got.ExperimentID != "experiment-1" || got.Purpose != "目的" || got.Source.VersionID != "version-1" || got.Source.State != "採用" {
		t.Errorf("preparation = %+v, want safe persisted values", got)
	}
	if got.Hypothesis == nil || *got.Hypothesis != "仮説" {
		t.Errorf("Hypothesis = %v, want 仮説", got.Hypothesis)
	}
	if len(got.Prompts) != 2 || got.Prompts[0].SequenceNo != 1 || got.Prompts[1].Content != "prompt B" {
		t.Errorf("Prompts = %+v, want sequence order", got.Prompts)
	}
	if !got.LastConfirmedAt.Equal(time.Date(2026, time.August, 9, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("LastConfirmedAt = %s, want %s", got.LastConfirmedAt, createdAt)
	}

	_, found, err = store.GetExperimentPreparation(context.Background(), "missing")
	if err != nil {
		t.Fatalf("missing GetExperimentPreparation() error = %v", err)
	}
	if found {
		t.Error("missing found = true, want false")
	}
}

// SQLite実験ワークスペースqueryの固定条件と操作読込。
func TestStoreGetExperimentWorkspace(t *testing.T) {
	store, fixed := fixedExperimentPreparationStore(t)
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	got, found, err := store.GetExperimentWorkspace(context.Background(), fixed.ExperimentID)
	if err != nil {
		t.Fatalf("GetExperimentWorkspace() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got.ExperimentID != fixed.ExperimentID || got.State != "ready" {
		t.Errorf("workspace = %+v, want ready experiment", got)
	}
	if got.FixedConditions.FixedConditionID == "" || got.FixedConditions.Purpose != fixed.Purpose {
		t.Errorf("FixedConditions = %+v, want persisted conditions", got.FixedConditions)
	}
	if got.ConditionFixOperationID == "" || got.ConditionFixOperationAt.IsZero() {
		t.Errorf("condition fix operation = (%q, %s), want persisted operation", got.ConditionFixOperationID, got.ConditionFixOperationAt)
	}
	if len(got.FixedConditions.Prompts) != len(fixed.Prompts) || got.FixedConditions.Prompts[0].Content != fixed.Prompts[0].Content {
		t.Errorf("Prompts = %+v, want persisted prompt order", got.FixedConditions.Prompts)
	}
	updatedAt := "2026-08-10T10:00:00Z"
	if _, err := store.db.Exec("INSERT INTO experiment_runs (id, experiment_id, state, summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", "run-1", fixed.ExperimentID, "completed", "要約", updatedAt, updatedAt); err != nil {
		t.Fatalf("insert run error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO experiment_evaluations (id, experiment_id, state, summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", "evaluation-1", fixed.ExperimentID, "completed", "評価要約", updatedAt, updatedAt); err != nil {
		t.Fatalf("insert evaluation error = %v", err)
	}
	got, found, err = store.GetExperimentWorkspace(context.Background(), fixed.ExperimentID)
	if err != nil {
		t.Fatalf("GetExperimentWorkspace() after progress error = %v", err)
	}
	if !found {
		t.Fatal("found after progress = false, want true")
	}
	if len(got.Runs) != 1 || got.Runs[0].ID != "run-1" || got.Runs[0].State != "completed" || got.Runs[0].Summary == nil || *got.Runs[0].Summary != "要約" {
		t.Errorf("Runs = %+v, want persisted run", got.Runs)
	}
	if len(got.Evaluations) != 1 || got.Evaluations[0].ID != "evaluation-1" || got.Evaluations[0].State != "completed" || got.Evaluations[0].Summary == nil || *got.Evaluations[0].Summary != "評価要約" {
		t.Errorf("Evaluations = %+v, want persisted evaluation", got.Evaluations)
	}

	_, found, err = store.GetExperimentWorkspace(context.Background(), "missing")
	if err != nil {
		t.Fatalf("missing GetExperimentWorkspace() error = %v", err)
	}
	if found {
		t.Error("missing found = true, want false")
	}
}

// SQLite実験ワークスペースqueryの代表的な読込失敗。
func TestStoreGetExperimentWorkspaceFailures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *Store)
		want  string
	}{
		{
			name: "主record読込失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE experiments"); err != nil {
					t.Fatalf("drop experiments error = %v", err)
				}
			},
			want: "find experiment workspace",
		},
		{
			name: "更新日時の形式不正を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("UPDATE experiments SET updated_at = ? WHERE id = ?", "invalid", "experiment-1"); err != nil {
					t.Fatalf("update experiment error = %v", err)
				}
			},
			want: "parse experiment workspace update time",
		},
		{
			name: "固定日時の形式不正を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("UPDATE experiment_fixed_conditions SET fixed_at = ?", "invalid"); err != nil {
					t.Fatalf("update fixed conditions error = %v", err)
				}
			},
			want: "parse experiment condition fixed time",
		},
		{
			name: "操作日時の形式不正を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("UPDATE experiment_condition_fix_operations SET fixed_at = ?", "invalid"); err != nil {
					t.Fatalf("update condition operation error = %v", err)
				}
			},
			want: "parse experiment condition operation fixed time",
		},
		{
			name: "prompt読込失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE experiment_fixed_condition_prompts"); err != nil {
					t.Fatalf("drop prompts error = %v", err)
				}
			},
			want: "query experiment workspace prompts",
		},
		{
			name: "prompt走査失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("UPDATE experiment_fixed_condition_prompts SET sequence_no = ? WHERE sequence_no = ?", "invalid", 1); err != nil {
					t.Fatalf("update fixed prompt error = %v", err)
				}
			},
			want: "scan experiment workspace prompt",
		},
		{
			name: "run読込失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE experiment_runs"); err != nil {
					t.Fatalf("drop runs error = %v", err)
				}
			},
			want: "query experiment workspace runs",
		},
		{
			name: "run走査失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				recreateWorkspaceProgressTable(t, store, "experiment_runs")
				if _, err := store.db.Exec("INSERT INTO experiment_runs (id, experiment_id, state, summary, created_at, updated_at) VALUES (?, ?, NULL, NULL, ?, ?)", "run-1", "experiment-1", "2026-08-10T10:00:00Z", "2026-08-10T10:00:00Z"); err != nil {
					t.Fatalf("insert invalid run error = %v", err)
				}
			},
			want: "scan experiment workspace run",
		},
		{
			name: "run更新日時の形式不正を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("INSERT INTO experiment_runs (id, experiment_id, state, summary, created_at, updated_at) VALUES (?, ?, ?, NULL, ?, ?)", "run-1", "experiment-1", "completed", "2026-08-10T10:00:00Z", "invalid"); err != nil {
					t.Fatalf("insert invalid run time error = %v", err)
				}
			},
			want: "parse experiment workspace run update time",
		},
		{
			name: "evaluation読込失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE experiment_evaluations"); err != nil {
					t.Fatalf("drop evaluations error = %v", err)
				}
			},
			want: "query experiment workspace evaluations",
		},
		{
			name: "evaluation走査失敗を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				recreateWorkspaceProgressTable(t, store, "experiment_evaluations")
				if _, err := store.db.Exec("INSERT INTO experiment_evaluations (id, experiment_id, state, summary, created_at, updated_at) VALUES (?, ?, NULL, NULL, ?, ?)", "evaluation-1", "experiment-1", "2026-08-10T10:00:00Z", "2026-08-10T10:00:00Z"); err != nil {
					t.Fatalf("insert invalid evaluation error = %v", err)
				}
			},
			want: "scan experiment workspace evaluation",
		},
		{
			name: "evaluation更新日時の形式不正を返す",
			setup: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("INSERT INTO experiment_evaluations (id, experiment_id, state, summary, created_at, updated_at) VALUES (?, ?, ?, NULL, ?, ?)", "evaluation-1", "experiment-1", "completed", "2026-08-10T10:00:00Z", "invalid"); err != nil {
					t.Fatalf("insert invalid evaluation time error = %v", err)
				}
			},
			want: "parse experiment workspace evaluation update time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := fixedExperimentPreparationStore(t)
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})
			tt.setup(t, store)

			_, _, err := store.GetExperimentWorkspace(context.Background(), "experiment-1")
			if err == nil {
				t.Fatal("GetExperimentWorkspace() error = nil, want read failure")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("GetExperimentWorkspace() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// SQLite実験ワークスペースの補助読取失敗。
func TestStoreGetExperimentWorkspaceProgressReadFailures(t *testing.T) {
	tests := []struct {
		name     string
		scenario briefingReadScenario
		want     string
	}{
		{
			name:     "固定promptのclose失敗を返す",
			scenario: workspaceReadPromptsCloseError,
			want:     "iterate experiment workspace prompts",
		},
		{
			name:     "固定promptの反復失敗を返す",
			scenario: workspaceReadPromptsRowsError,
			want:     "iterate experiment workspace prompts",
		},
		{
			name:     "run読取失敗を返す",
			scenario: workspaceReadRunsQueryError,
			want:     "query experiment workspace runs",
		},
		{
			name:     "evaluation読取失敗を返す",
			scenario: workspaceReadEvaluationsQueryError,
			want:     "query experiment workspace evaluations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBriefingReadTestStore(t, tt.scenario)

			_, _, err := store.GetExperimentWorkspace(context.Background(), "experiment-1")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("GetExperimentWorkspace() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// SQLite実験ワークスペース進行状況helperの失敗正規化。
func TestStoreExperimentWorkspaceProgressReadFailures(t *testing.T) {
	tests := []struct {
		name     string
		scenario briefingReadScenario
		read     func(context.Context, *Store) error
		want     string
	}{
		{
			name:     "run query失敗を返す",
			scenario: workspaceReadRunsQueryError,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceRuns(ctx, "experiment-1")

				return err
			},
			want: "query experiment workspace runs",
		},
		{
			name:     "run scan失敗を返す",
			scenario: workspaceReadRunsScanError,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceRuns(ctx, "experiment-1")

				return err
			},
			want: "scan experiment workspace run",
		},
		{
			name:     "run日時不正を返す",
			scenario: workspaceReadRunsInvalidTime,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceRuns(ctx, "experiment-1")

				return err
			},
			want: "parse experiment workspace run update time",
		},
		{
			name:     "run反復失敗を返す",
			scenario: workspaceReadRunsRowsError,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceRuns(ctx, "experiment-1")

				return err
			},
			want: "iterate experiment workspace runs",
		},
		{
			name:     "evaluation query失敗を返す",
			scenario: workspaceReadEvaluationsQueryError,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceEvaluations(ctx, "experiment-1")

				return err
			},
			want: "query experiment workspace evaluations",
		},
		{
			name:     "evaluation scan失敗を返す",
			scenario: workspaceReadEvaluationsScanError,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceEvaluations(ctx, "experiment-1")

				return err
			},
			want: "scan experiment workspace evaluation",
		},
		{
			name:     "evaluation日時不正を返す",
			scenario: workspaceReadEvaluationsInvalidTime,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceEvaluations(ctx, "experiment-1")

				return err
			},
			want: "parse experiment workspace evaluation update time",
		},
		{
			name:     "evaluation反復失敗を返す",
			scenario: workspaceReadEvaluationsRowsError,
			read: func(ctx context.Context, store *Store) error {
				_, err := store.findExperimentWorkspaceEvaluations(ctx, "experiment-1")

				return err
			},
			want: "iterate experiment workspace evaluations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBriefingReadTestStore(t, tt.scenario)

			err := tt.read(context.Background(), store)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("read() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// recreateWorkspaceProgressTable は走査失敗検証用のnullable進行テーブルを再作成する。
func recreateWorkspaceProgressTable(t *testing.T, store *Store, table string) {
	t.Helper()

	if _, err := store.db.Exec("DROP TABLE " + table); err != nil {
		t.Fatalf("drop %s error = %v", table, err)
	}
	if _, err := store.db.Exec("CREATE TABLE " + table + " (id TEXT, experiment_id TEXT, retry_of_run_id TEXT, state TEXT, summary TEXT, created_at TEXT, updated_at TEXT)"); err != nil {
		t.Fatalf("create %s error = %v", table, err)
	}
}

// SQLite実験準備queryの失敗正規化。
func TestStoreGetExperimentPreparationFailures(t *testing.T) {
	tests := []struct {
		name     string
		scenario briefingReadScenario
		want     string
	}{
		{
			name:     "実験準備query失敗を返す",
			scenario: preparationReadQueryError,
			want:     "find experiment preparation",
		},
		{
			name:     "確認時刻不正を返す",
			scenario: preparationReadInvalidTime,
			want:     "parse experiment preparation update time",
		},
		{
			name:     "prompt query失敗を返す",
			scenario: preparationReadPromptsQueryError,
			want:     "query experiment preparation prompts",
		},
		{
			name:     "prompt scan失敗を返す",
			scenario: preparationReadPromptsScanError,
			want:     "scan experiment preparation prompt",
		},
		{
			name:     "prompt反復失敗を返す",
			scenario: preparationReadPromptsRowsError,
			want:     "iterate experiment preparation prompts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBriefingReadTestStore(t, tt.scenario)

			_, _, err := store.GetExperimentPreparation(context.Background(), "experiment-1")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("GetExperimentPreparation() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// ブリーフ採用用保存済み版生成。
func newExperimentCreationTestStore(t *testing.T) *Store {
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
	createdAt := "2026-08-09T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-2", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed other preparation session error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, success_criteria, required_conditions, created_at, purpose, candidate_prompts, evaluation_criteria, environment_conditions, initial_input) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "version-1", "session-1", 1, "採用", "基準", "条件", createdAt, "目的", `["prompt A","prompt B"]`, "正確性", "隔離環境", "初期入力"); err != nil {
		t.Fatalf("seed briefing version error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, success_criteria, required_conditions, created_at, purpose, candidate_prompts, evaluation_criteria, environment_conditions, initial_input) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "other-version", "session-2", 1, "採用", "基準", "条件", createdAt, "目的", `["prompt A","prompt B"]`, "正確性", "隔離環境", "初期入力"); err != nil {
		t.Fatalf("seed other briefing version error = %v", err)
	}

	return store
}

// ブリーフ採用補助関数の境界。
func TestExperimentBriefAdoptionHelpers(t *testing.T) {
	tests := []struct {
		name     string
		brief    domain.ExperimentBrief
		wantCode apperr.Code
	}{
		{
			name: "必須項目不足を拒否する",
			brief: domain.ExperimentBrief{
				CandidatePrompts: []string{
					"prompt A",
					"prompt B",
				},
			},
			wantCode: apperr.CodeExperimentBriefIncomplete,
		},
		{
			name: "空promptを拒否する",
			brief: domain.ExperimentBrief{
				Purpose:               "目的",
				EvaluationCriteria:    "基準",
				EnvironmentConditions: "環境",
				CandidatePrompts: []string{
					"prompt A",
					" ",
				},
			},
			wantCode: apperr.CodeExperimentBriefIncomplete,
		},
		{
			name: "必要条件が揃う",
			brief: domain.ExperimentBrief{
				Purpose:               "目的",
				EvaluationCriteria:    "基準",
				EnvironmentConditions: "環境",
				CandidatePrompts: []string{
					"prompt A",
					"prompt B",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExperimentBriefForAdoption(tt.brief)
			if tt.wantCode == "" {
				if err != nil {
					t.Errorf("validateExperimentBriefForAdoption() error = %v, want nil", err)
				}

				return
			}
			if !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("validateExperimentBriefForAdoption() error = %v, want %q", err, tt.wantCode)
			}
		})
	}
	if !isExperimentCreationRequestConflict(errors.New("UNIQUE constraint failed: experiment_creation_operations.request_id")) {
		t.Error("isExperimentCreationRequestConflict() = false, want true")
	}
	if isExperimentCreationRequestConflict(errors.New("database locked")) {
		t.Error("isExperimentCreationRequestConflict() = true, want false")
	}
}

// SQLiteブリーフ採用保存の失敗時ロールバック。
func TestStoreCreateExperimentFromBriefFailures(t *testing.T) {
	tests := []struct {
		name         string
		beginErr     error
		rows         []briefingRow
		execErrors   []error
		commitErr    error
		prepare      func(*testing.T, *Store, *fakeBriefingTransaction)
		want         string
		wantRollback bool
	}{
		{
			name:     "トランザクション開始失敗",
			beginErr: errors.New("begin unavailable"),
			want:     "begin experiment creation",
		},
		{
			name: "版検索失敗",
			rows: []briefingRow{
				fakeBriefingRow{err: errors.New("version unavailable")},
			},
			want:         "find experiment brief version",
			wantRollback: true,
		},
		{
			name: "未完全ブリーフ",
			rows: []briefingRow{
				fakeBriefingRow{values: []any{
					"",
					nil,
					`["prompt A","prompt B"]`,
					"基準",
					"環境",
					"",
				}},
			},
			want:         "条件が不足",
			wantRollback: true,
		},
		{
			name: "識別子生成失敗",
			rows: []briefingRow{
				validExperimentBriefingRow(),
			},
			prepare: func(t *testing.T, _ *Store, _ *fakeBriefingTransaction) {
				t.Helper()
				replaceBriefingRandom(t, func([]byte) (int, error) {
					return 0, errors.New("random unavailable")
				})
			},
			want:         "generate experiment identifier",
			wantRollback: true,
		},
		{
			name: "実験保存失敗",
			rows: []briefingRow{
				validExperimentBriefingRow(),
			},
			execErrors:   []error{errors.New("experiment unavailable")},
			want:         "insert experiment",
			wantRollback: true,
		},
		{
			name: "準備値保存失敗",
			rows: []briefingRow{
				validExperimentBriefingRow(),
			},
			execErrors: []error{
				nil,
				errors.New("preparation unavailable"),
			},
			want:         "insert experiment preparation",
			wantRollback: true,
		},
		{
			name: "prompt保存失敗",
			rows: []briefingRow{
				validExperimentBriefingRow(),
			},
			execErrors: []error{
				nil,
				nil,
				errors.New("prompt unavailable"),
			},
			want:         "insert experiment preparation prompt",
			wantRollback: true,
		},
		{
			name: "operation保存失敗",
			rows: []briefingRow{
				validExperimentBriefingRow(),
			},
			execErrors: []error{
				nil,
				nil,
				nil,
				nil,
				errors.New("operation unavailable"),
			},
			want:         "insert experiment creation operation",
			wantRollback: true,
		},
		{
			name: "確定失敗",
			rows: []briefingRow{
				validExperimentBriefingRow(),
			},
			commitErr: errors.New("commit unavailable"),
			want:      "commit experiment creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentCreationTestStore(t)
			transaction := &fakeBriefingTransaction{
				rows:        tt.rows,
				execErrors:  tt.execErrors,
				commitError: tt.commitErr,
			}
			if tt.beginErr != nil {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, tt.beginErr
				}
			} else {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)
			}
			if tt.prepare != nil {
				tt.prepare(t, store, transaction)
			}

			_, _, err := store.CreateExperimentFromBrief(context.Background(), "request-failure", "session-1", "version-1")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("CreateExperimentFromBrief() error = %v, want containing %q", err, tt.want)
			}
			if gotRollback := transaction.rollbackCalls > 0; gotRollback != tt.wantRollback {
				t.Errorf("rollback = %v, want %v", gotRollback, tt.wantRollback)
			}
		})
	}
}

// SQLiteブリーフ採用操作の競合再読込。
func TestStoreCreateExperimentFromBriefConflictRecovery(t *testing.T) {
	store := newExperimentCreationTestStore(t)
	transaction := &fakeBriefingTransaction{
		rows: []briefingRow{
			validExperimentBriefingRow(),
		},
		execErrors: []error{
			nil,
			nil,
			nil,
			nil,
			errors.New("UNIQUE constraint failed: experiment_creation_operations.request_id"),
		},
	}
	transaction.onExec = func(call int) {
		if call != 5 {
			return
		}
		if _, err := store.db.Exec("INSERT INTO experiments (id, purpose, state, updated_at) VALUES (?, ?, ?, ?)", "experiment-race", "目的", "preparing", "2026-08-09T00:00:00Z"); err != nil {
			t.Fatalf("seed raced experiment error = %v", err)
		}
		if _, err := store.db.Exec("INSERT INTO experiment_creation_operations (request_id, briefing_session_id, briefing_version_id, experiment_id) VALUES (?, ?, ?, ?)", "request-race", "session-1", "version-1", "experiment-race"); err != nil {
			t.Fatalf("seed raced operation error = %v", err)
		}
	}
	store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)

	got, created, err := store.CreateExperimentFromBrief(context.Background(), "request-race", "session-1", "version-1")
	if err != nil {
		t.Fatalf("CreateExperimentFromBrief() error = %v", err)
	}
	if created {
		t.Error("created = true, want false")
	}
	if got.ExperimentID != "experiment-race" || got.State != "preparing" {
		t.Errorf("creation = %+v, want recovered operation", got)
	}
}

// SQLiteブリーフ採用操作の検索失敗境界。
func TestStoreCreateExperimentFromBriefLookupFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		want    string
	}{
		{
			name: "既存操作検索失敗",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			want: "find experiment creation operation",
		},
		{
			name: "競合後操作検索失敗",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				transaction := &fakeBriefingTransaction{
					rows: []briefingRow{
						validExperimentBriefingRow(),
					},
					execErrors: []error{
						nil,
						nil,
						nil,
						nil,
						errors.New("UNIQUE constraint failed: experiment_creation_operations.request_id"),
					},
				}
				transaction.onExec = func(call int) {
					if call != 5 {
						return
					}
					if err := store.db.Close(); err != nil {
						t.Fatalf("Close() error = %v", err)
					}
				}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)
			},
			want: "find experiment creation operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newExperimentCreationTestStore(t)
			tt.prepare(t, store)

			_, _, err := store.CreateExperimentFromBrief(context.Background(), "request-lookup", "session-1", "version-1")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("CreateExperimentFromBrief() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// ブリーフ採用読込補助関数の失敗境界。
func TestExperimentBriefAdoptionReadHelpers(t *testing.T) {
	store := newExperimentCreationTestStore(t)
	if err := store.db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, _, err := store.findExperimentCreation(context.Background(), "request-1"); err == nil || !strings.Contains(err.Error(), "find experiment creation operation") {
		t.Errorf("findExperimentCreation() error = %v, want query failure", err)
	}

	tests := []struct {
		name string
		row  briefingRow
		want string
	}{
		{
			name: "仮説を復元する",
			row: fakeBriefingRow{values: []any{
				"目的",
				"仮説",
				`["prompt A","prompt B"]`,
				"基準",
				"環境",
				"初期入力",
			}},
		},
		{
			name: "prompt形式不正を返す",
			row: fakeBriefingRow{values: []any{
				"目的",
				nil,
				"invalid-json",
				"基準",
				"環境",
				"初期入力",
			}},
			want: "unmarshal experiment brief prompts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brief, err := findExperimentBriefForAdoption(context.Background(), &fakeBriefingTransaction{rows: []briefingRow{tt.row}}, "session-1", "version-1")
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Errorf("findExperimentBriefForAdoption() error = %v, want containing %q", err, tt.want)
				}

				return
			}
			if err != nil {
				t.Fatalf("findExperimentBriefForAdoption() error = %v", err)
			}
			if brief.Hypothesis == nil || *brief.Hypothesis != "仮説" {
				t.Errorf("Hypothesis = %v, want 仮説", brief.Hypothesis)
			}
		})
	}
}

// 採用可能ブリーフ版単一行。
func validExperimentBriefingRow() briefingRow {
	return fakeBriefingRow{values: []any{
		"目的",
		nil,
		`["prompt A","prompt B"]`,
		"基準",
		"環境",
		"初期入力",
	}}
}

// SQLite実験ブリーフ会話の並行完了保存。
func TestStoreCompleteExperimentBriefMessageConcurrently(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	createdAt := "2026-08-09T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}
	for _, requestID := range []string{
		"request-1",
		"request-2",
	} {
		if _, created, err := store.BeginExperimentBriefMessage(context.Background(), requestID, "session-1"); err != nil {
			t.Fatalf("BeginExperimentBriefMessage() error = %v", err)
		} else if !created {
			t.Fatalf("created = false, want true")
		}
	}

	errorsCh := make(chan error, 2)
	for _, requestID := range []string{
		"request-1",
		"request-2",
	} {
		go func(requestID string) {
			errorsCh <- store.CompleteExperimentBriefMessage(context.Background(), requestID, "利用者メッセージ", domain.ExperimentBriefingMessageResult{
				AssistantMessage: "AI応答",
				Brief: &domain.ExperimentBrief{
					Decision:           "判断",
					SuccessCriteria:    "基準",
					RequiredConditions: "条件",
				},
			})
		}(requestID)
	}
	for range 2 {
		if err := <-errorsCh; err != nil {
			t.Errorf("CompleteExperimentBriefMessage() error = %v", err)
		}
	}

	briefing, found, err := store.GetExperimentBriefing(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetExperimentBriefing() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got := len(briefing.Messages); got != 4 {
		t.Errorf("Messages length = %d, want 4", got)
	}
	for index, message := range briefing.Messages {
		if got, want := message.SequenceNo, index+1; got != want {
			t.Errorf("Messages[%d].SequenceNo = %d, want %d", index, got, want)
		}
	}
	var operationCount, completedCount, versionCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM briefing_message_operations").Scan(&operationCount); err != nil {
		t.Fatalf("message operation count error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM briefing_message_operations WHERE state = ?", domain.BriefingStartStateStarted).Scan(&completedCount); err != nil {
		t.Fatalf("completed message operation count error = %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM briefing_versions").Scan(&versionCount); err != nil {
		t.Fatalf("briefing version count error = %v", err)
	}
	if operationCount != 2 {
		t.Errorf("message operations = %d, want 2", operationCount)
	}
	if completedCount != 2 {
		t.Errorf("completed message operations = %d, want 2", completedCount)
	}
	if versionCount != 2 {
		t.Errorf("briefing versions = %d, want 2", versionCount)
	}
}

// SQLite実験ブリーフ会話送信開始の失敗分岐。
func TestStoreBeginExperimentBriefMessageFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		want    string
	}{
		{
			name: "既存operation検索失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				if err := store.db.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			want: "database is closed",
		},
		{
			name: "未知sessionを返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{err: sql.ErrNoRows}}})
			},
			want: "見つかりません",
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
			name: "トランザクション開始失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, errors.New("begin unavailable")
				}
			},
			want: "begin briefing message",
		},
		{
			name: "session読込失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{err: errors.New("session unavailable")}}})
			},
			want: "find briefing message session",
		},
		{
			name: "非開始sessionを拒否する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{values: []any{domain.BriefingStartStateFailed}}}})
			},
			want: "送信できる状態ではありません",
		},
		{
			name: "operation保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows:       []briefingRow{fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}}},
					execErrors: []error{errors.New("insert unavailable")},
				})
			},
			want: "insert briefing message operation",
		},
		{
			name: "競合後のoperation検索失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				transaction := &fakeBriefingTransaction{
					rows:       []briefingRow{fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}}},
					execErrors: []error{errors.New("UNIQUE constraint failed: briefing_message_operations.request_id")},
				}
				transaction.onExec = func(int) {
					if err := store.db.Close(); err != nil {
						t.Fatalf("Close() error = %v", err)
					}
				}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)
			},
			want: "database is closed",
		},
		{
			name: "確定失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows:        []briefingRow{fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}}},
					commitError: errors.New("commit unavailable"),
				})
			},
			want: "commit briefing message",
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

			_, _, err = store.BeginExperimentBriefMessage(context.Background(), "request-1", "session-1")
			if err == nil {
				t.Fatal("BeginExperimentBriefMessage() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("BeginExperimentBriefMessage() error = %q, want containing %q", got, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ会話送信開始の既存operation境界。
func TestStoreBeginExperimentBriefMessageExistingOperations(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	createdAt := "2026-08-09T00:00:00Z"
	for _, sessionID := range []string{
		"session-1",
		"session-2",
	} {
		if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", sessionID, "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
			t.Fatalf("seed preparation session error = %v", err)
		}
	}
	if _, created, err := store.BeginExperimentBriefMessage(context.Background(), "request-1", "session-1"); err != nil || !created {
		t.Fatalf("BeginExperimentBriefMessage() created = %v, error = %v", created, err)
	}
	if _, _, err := store.BeginExperimentBriefMessage(context.Background(), "request-1", "session-2"); err == nil {
		t.Fatal("BeginExperimentBriefMessage() error = nil, want target mismatch")
	}

	store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
		transaction := &fakeBriefingTransaction{
			rows:       []briefingRow{fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}}},
			execErrors: []error{errors.New("UNIQUE constraint failed: briefing_message_operations.request_id")},
		}
		transaction.onExec = func(int) {
			if _, err := store.db.Exec("INSERT INTO briefing_message_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", "operation-race", "request-race", "session-1", domain.BriefingStartStateStarted); err != nil {
				t.Fatalf("seed raced operation error = %v", err)
			}
		}

		return transaction, nil
	}
	got, created, err := store.BeginExperimentBriefMessage(context.Background(), "request-race", "session-1")
	if err != nil {
		t.Fatalf("raced BeginExperimentBriefMessage() error = %v", err)
	}
	if created {
		t.Error("created = true, want false")
	}
	if got.OperationID != "operation-race" {
		t.Errorf("OperationID = %q, want %q", got.OperationID, "operation-race")
	}
}

// SQLite実験ブリーフ会話送信完了の失敗分岐。
func TestStoreCompleteExperimentBriefMessageFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Store)
		result  domain.ExperimentBriefingMessageResult
		want    string
	}{
		{
			name: "連番取得失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{err: errors.New("sequence unavailable")}}})
			},
			want: "find next briefing message sequence",
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
			want: "insert user briefing message",
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
			result: domain.ExperimentBriefingMessageResult{AssistantMessage: "AI応答"},
			want:   "insert assistant briefing message",
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
			want: "complete briefing message operation",
		},
		{
			name: "ブリーフ候補保存失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{1}},
						fakeBriefingRow{values: []any{1}},
					},
					execErrors: []error{
						nil,
						errors.New("brief insert unavailable"),
					},
				})
			},
			result: domain.ExperimentBriefingMessageResult{Brief: &domain.ExperimentBrief{
				Decision:           "判断",
				SuccessCriteria:    "基準",
				RequiredConditions: "条件",
			}},
			want: "insert briefing version",
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
			want: "update briefing message session",
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
			want: "commit briefing message completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newCompleteBriefingMessageTestStore(t)
			tt.prepare(t, store)

			err := store.CompleteExperimentBriefMessage(context.Background(), "request-1", "利用者メッセージ", tt.result)
			if err == nil {
				t.Fatal("CompleteExperimentBriefMessage() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("CompleteExperimentBriefMessage() error = %q, want containing %q", got, tt.want)
			}
		})
	}
}

// newCompleteBriefingMessageTestStore は会話送信完了用の保存済みoperationを生成。
func newCompleteBriefingMessageTestStore(t *testing.T) *Store {
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
	createdAt := "2026-08-09T00:00:00Z"
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("seed preparation session error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO briefing_message_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", "operation-1", "request-1", "session-1", domain.BriefingStartStateStarting); err != nil {
		t.Fatalf("seed briefing message operation error = %v", err)
	}

	return store
}

// SQLite実験ブリーフ会話送信完了の境界失敗。
func TestStoreCompleteExperimentBriefMessageBoundaries(t *testing.T) {
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
				if _, err := store.db.Exec("DELETE FROM briefing_message_operations"); err != nil {
					t.Fatalf("delete briefing message operation error = %v", err)
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
			want: "begin briefing message completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newCompleteBriefingMessageTestStore(t)
			tt.prepare(t, store)

			err := store.CompleteExperimentBriefMessage(context.Background(), "request-1", "利用者メッセージ", domain.ExperimentBriefingMessageResult{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("CompleteExperimentBriefMessage() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ版保存の失敗分岐。
func TestStoreInsertExperimentBriefVersionFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T)
		tx      *fakeBriefingTransaction
		want    string
	}{
		{
			name: "識別子生成失敗を返す",
			prepare: func(t *testing.T) {
				t.Helper()
				replaceBriefingRandom(t, func([]byte) (int, error) {
					return 0, errors.New("random unavailable")
				})
			},
			tx:   &fakeBriefingTransaction{},
			want: "generate briefing version identifier",
		},
		{
			name:    "版番号取得失敗を返す",
			prepare: func(*testing.T) {},
			tx:      &fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{err: errors.New("version unavailable")}}},
			want:    "find next briefing version",
		},
		{
			name:    "版保存失敗を返す",
			prepare: func(*testing.T) {},
			tx: &fakeBriefingTransaction{
				rows:       []briefingRow{fakeBriefingRow{values: []any{1}}},
				execErrors: []error{errors.New("insert unavailable")},
			},
			want: "insert briefing version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(t)

			err := (&Store{}).insertExperimentBriefVersion(context.Background(), tt.tx, "session-1", domain.ExperimentBrief{Decision: "判断"}, "2026-08-09T00:00:00Z")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("insertExperimentBriefVersion() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ会話送信失敗保存の分岐。
func TestStoreFailExperimentBriefMessage(t *testing.T) {
	tests := []struct {
		name   string
		result sql.Result
		err    error
		want   string
	}{
		{
			name: "更新失敗を返す",
			err:  errors.New("update unavailable"),
			want: "fail briefing message operation",
		},
		{
			name: "更新件数取得失敗を返す",
			result: fakeBriefingResult{
				rowsAffectedError: errors.New("count unavailable"),
			},
			want: "count failed briefing message operations",
		},
		{
			name: "operation不在を返す",
			result: fakeBriefingResult{
				rowsAffected: 0,
			},
			want: "request not found",
		},
		{
			name: "失敗状態を保存する",
			result: fakeBriefingResult{
				rowsAffected: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{failBriefingMessageOperation: func(context.Context, string, string) (sql.Result, error) {
				return tt.result, tt.err
			}}

			err := store.FailExperimentBriefMessage(context.Background(), "request-1", "ACP_NOT_READY")
			if tt.want == "" {
				if err != nil {
					t.Errorf("FailExperimentBriefMessage() error = %v, want nil", err)
				}

				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("FailExperimentBriefMessage() error = %v, want containing %q", err, tt.want)
			}
		})
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
	briefingReadMessagesQueryError      briefingReadScenario = "messages-query-error"
	briefingReadBriefQueryError         briefingReadScenario = "brief-query-error"
	briefingReadSessionEmptyUpdatedAt   briefingReadScenario = "session-empty-updated-at"
	briefingReadSessionInvalidTime      briefingReadScenario = "session-invalid-time"
	briefingReadMessagesCloseError      briefingReadScenario = "messages-close-error"
	briefingReadMessagesScanError       briefingReadScenario = "messages-scan-error"
	briefingReadMessagesInvalidTime     briefingReadScenario = "messages-invalid-time"
	briefingReadMessagesRowsError       briefingReadScenario = "messages-rows-error"
	briefingReadBriefInvalidTime        briefingReadScenario = "brief-invalid-time"
	preparationReadQueryError           briefingReadScenario = "preparation-query-error"
	preparationReadInvalidTime          briefingReadScenario = "preparation-invalid-time"
	preparationReadPromptsQueryError    briefingReadScenario = "preparation-prompts-query-error"
	preparationReadPromptsScanError     briefingReadScenario = "preparation-prompts-scan-error"
	preparationReadPromptsRowsError     briefingReadScenario = "preparation-prompts-rows-error"
	workspaceReadPromptsCloseError      briefingReadScenario = "workspace-prompts-close-error"
	workspaceReadPromptsRowsError       briefingReadScenario = "workspace-prompts-rows-error"
	workspaceReadRunsQueryError         briefingReadScenario = "workspace-runs-query-error"
	workspaceReadRunsScanError          briefingReadScenario = "workspace-runs-scan-error"
	workspaceReadRunsInvalidTime        briefingReadScenario = "workspace-runs-invalid-time"
	workspaceReadRunsRowsError          briefingReadScenario = "workspace-runs-rows-error"
	workspaceReadEvaluationsQueryError  briefingReadScenario = "workspace-evaluations-query-error"
	workspaceReadEvaluationsScanError   briefingReadScenario = "workspace-evaluations-scan-error"
	workspaceReadEvaluationsInvalidTime briefingReadScenario = "workspace-evaluations-invalid-time"
	workspaceReadEvaluationsRowsError   briefingReadScenario = "workspace-evaluations-rows-error"
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
	case strings.Contains(query, "FROM experiments e JOIN experiment_fixed_conditions"):
		return workspaceReadRows(c.scenario)
	case strings.Contains(query, "FROM experiments e"):
		return preparationReadRows(c.scenario)
	case strings.Contains(query, "FROM experiment_fixed_condition_prompts"):
		return workspacePromptReadRows(c.scenario)
	case strings.Contains(query, "FROM experiment_runs"):
		return workspaceRunReadRows(c.scenario)
	case strings.Contains(query, "FROM experiment_evaluations"):
		return workspaceEvaluationReadRows(c.scenario)
	case strings.Contains(query, "FROM experiment_preparation_prompts"):
		return preparationPromptReadRows(c.scenario)
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

// workspaceReadRows は実験ワークスペースqueryの結果を返す。
func workspaceReadRows(briefingReadScenario) (driver.Rows, error) {
	return &briefingReadRows{
		columns: []string{
			"state",
			"updated_at",
			"fixed_condition_id",
			"purpose",
			"hypothesis",
			"environment_conditions",
			"initial_input",
			"evaluation_axes",
			"fixed_at",
			"operation_id",
			"operation_fixed_at",
		},
		values: [][]driver.Value{{
			"ready",
			"2026-08-10T00:00:00Z",
			"condition-1",
			"purpose",
			nil,
			"environment",
			"input",
			"criteria",
			"2026-08-10T00:00:00Z",
			"operation-1",
			"2026-08-10T00:00:00Z",
		}},
	}, nil
}

// workspacePromptReadRows は固定prompt queryの結果または失敗を返す。
func workspacePromptReadRows(scenario briefingReadScenario) (driver.Rows, error) {
	rows := &briefingReadRows{columns: []string{
		"sequence_no",
		"content",
	}}
	switch scenario {
	case workspaceReadPromptsCloseError:
		rows.closeErr = errors.New("workspace prompts close failed")
	case workspaceReadPromptsRowsError:
		rows.nextErr = errors.New("workspace prompts iteration failed")
	}

	return rows, nil
}

// workspaceRunReadRows はrun queryの結果または失敗を返す。
func workspaceRunReadRows(scenario briefingReadScenario) (driver.Rows, error) {
	if scenario == workspaceReadRunsQueryError {
		return nil, errors.New("workspace runs query failed")
	}
	rows := &briefingReadRows{columns: []string{
		"id",
		"retry_of_run_id",
		"state",
		"summary",
		"updated_at",
	}}
	switch scenario {
	case workspaceReadRunsScanError:
		rows.values = [][]driver.Value{{
			"run-1",
			nil,
			"completed",
			"summary",
			nil,
		}}
	case workspaceReadRunsInvalidTime:
		rows.values = [][]driver.Value{{
			"run-1",
			nil,
			"completed",
			nil,
			"invalid-time",
		}}
	case workspaceReadRunsRowsError:
		rows.nextErr = errors.New("workspace runs iteration failed")
	}

	return rows, nil
}

// workspaceEvaluationReadRows はevaluation queryの結果または失敗を返す。
func workspaceEvaluationReadRows(scenario briefingReadScenario) (driver.Rows, error) {
	if scenario == workspaceReadEvaluationsQueryError {
		return nil, errors.New("workspace evaluations query failed")
	}
	rows := &briefingReadRows{columns: []string{
		"id",
		"state",
		"summary",
		"updated_at",
	}}
	switch scenario {
	case workspaceReadEvaluationsScanError:
		rows.values = [][]driver.Value{{
			"evaluation-1",
			"completed",
			"summary",
			nil,
		}}
	case workspaceReadEvaluationsInvalidTime:
		rows.values = [][]driver.Value{{
			"evaluation-1",
			"completed",
			nil,
			"invalid-time",
		}}
	case workspaceReadEvaluationsRowsError:
		rows.nextErr = errors.New("workspace evaluations iteration failed")
	}

	return rows, nil
}

// preparationReadRows は実験準備queryの結果または失敗を返す。
func preparationReadRows(scenario briefingReadScenario) (driver.Rows, error) {
	if scenario == preparationReadQueryError {
		return nil, errors.New("preparation query failed")
	}
	updatedAt := "2026-08-09T00:00:00Z"
	if scenario == preparationReadInvalidTime {
		updatedAt = "invalid-time"
	}

	return &briefingReadRows{
		columns: []string{
			"state",
			"purpose",
			"hypothesis",
			"environment_conditions",
			"initial_input",
			"evaluation_criteria",
			"briefing_version_id",
			"decision",
			"updated_at",
		},
		values: [][]driver.Value{{
			"preparing",
			"purpose",
			nil,
			"environment",
			"input",
			"criteria",
			"version-1",
			"adopted",
			updatedAt,
		}},
	}, nil
}

// preparationPromptReadRows は実験準備prompt queryの結果または失敗を返す。
func preparationPromptReadRows(scenario briefingReadScenario) (driver.Rows, error) {
	if scenario == preparationReadPromptsQueryError {
		return nil, errors.New("preparation prompts query failed")
	}
	rows := &briefingReadRows{columns: []string{
		"sequence_no",
		"content",
	}}
	switch scenario {
	case preparationReadPromptsScanError:
		rows.values = [][]driver.Value{{
			"invalid-sequence",
			"prompt",
		}}
	case preparationReadPromptsRowsError:
		rows.nextErr = errors.New("preparation prompts iteration failed")
	}

	return rows, nil
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
			"purpose",
			"decision",
			"hypothesis",
			"candidate_prompts",
			"evaluation_criteria",
			"environment_conditions",
			"initial_input",
			"success_criteria",
			"required_conditions",
			"open_question",
			"created_at",
		},
		values: [][]driver.Value{{
			"version-1",
			"purpose",
			"decision",
			nil,
			"[]",
			"criteria",
			"conditions",
			"",
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
	execErrors    []error
	execCalls     int
	result        sql.Result
	results       []sql.Result
	commitError   error
	onExec        func(int)
	rows          []briefingRow
	rowCalls      int
	rollbackCalls int
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

// QueryRowContext はこのtest doubleで未使用の行を返却。
func (f *fakeBriefingTransaction) QueryRowContext(_ context.Context, query string, _ ...any) briefingRow {
	if strings.Contains(query, "FROM preparation_sessions WHERE id = ? AND kind = ?") && f.usesExperimentBriefingRow() {
		return fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}}
	}

	f.rowCalls++
	if f.rowCalls <= len(f.rows) {
		return f.rows[f.rowCalls-1]
	}

	return fakeBriefingRow{err: errors.New("unexpected query row")}
}

// usesExperimentBriefingRow は採用用のブリーフ行を識別する。
func (f *fakeBriefingTransaction) usesExperimentBriefingRow() bool {
	if len(f.rows) == 0 {
		return false
	}
	row, ok := f.rows[0].(fakeBriefingRow)

	return ok && (len(row.values) > 1 || (row.err != nil && strings.Contains(row.err.Error(), "version")))
}

// fakeBriefingRow は単一行読込のtest double。
type fakeBriefingRow struct {
	values []any
	err    error
}

// Scan は指定済みの単一行値をコピー。
func (f fakeBriefingRow) Scan(destinations ...any) error {
	if f.err != nil {
		return f.err
	}
	for index, destination := range destinations {
		if index >= len(f.values) {
			return errors.New("missing fake row value")
		}
		switch target := destination.(type) {
		case *string:
			value, ok := f.values[index].(string)
			if !ok {
				return errors.New("fake row string conversion failed")
			}
			*target = value
		case *int:
			value, ok := f.values[index].(int)
			if !ok {
				return errors.New("fake row int conversion failed")
			}
			*target = value
		case *sql.NullString:
			if f.values[index] == nil {
				target.Valid = false

				continue
			}
			value, ok := f.values[index].(string)
			if !ok {
				return errors.New("fake row null string conversion failed")
			}
			target.String = value
			target.Valid = true
		default:
			return errors.New("unsupported fake row destination")
		}
	}

	return nil
}

// Commit は指定済みの確定結果を返却。
func (f *fakeBriefingTransaction) Commit() error {
	return f.commitError
}

// Rollback はロールバックを受理。
func (f *fakeBriefingTransaction) Rollback() error {
	f.rollbackCalls++

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

// SQLite実験ブリーフ停止の永続化と冪等性。
func TestStoreStopExperimentBriefing(t *testing.T) {
	tests := []struct {
		name         string
		requestID    string
		sessionID    string
		sessionState string
		wantCode     apperr.Code
	}{
		{
			name:         "停止意図を保存して停止済みを返す",
			requestID:    "request-1",
			sessionID:    "session-1",
			sessionState: domain.BriefingStartStateStarted,
		},
		{
			name:         "非active sessionを拒否する",
			requestID:    "request-2",
			sessionID:    "session-1",
			sessionState: domain.BriefingStartStateStopped,
			wantCode:     apperr.CodeBriefingNotActive,
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
			createdAt := time.Now().UTC().Format(time.RFC3339Nano)
			if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", tt.sessionID, "experiment_brief", tt.sessionState, createdAt, createdAt); err != nil {
				t.Fatalf("insert preparation session error = %v", err)
			}

			operation, created, err := store.BeginStopExperimentBriefing(context.Background(), tt.requestID, tt.sessionID)
			if tt.wantCode != "" {
				appErr := apperr.As(err)
				if appErr == nil {
					t.Fatal("apperr.As(error) = nil, want app error")
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("Code = %q, want %q", appErr.Code, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("BeginStopExperimentBriefing() error = %v", err)
			}
			if !created {
				t.Error("created = false, want true")
			}
			if operation.OperationID == "" {
				t.Error("OperationID = empty, want identifier")
			}
			if err := store.CompleteStopExperimentBriefing(context.Background(), tt.requestID); err != nil {
				t.Fatalf("CompleteStopExperimentBriefing() error = %v", err)
			}
			briefing, found, err := store.GetExperimentBriefing(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("GetExperimentBriefing() error = %v", err)
			}
			if !found {
				t.Fatal("found = false, want true")
			}
			if briefing.State != domain.BriefingStartStateStopped {
				t.Errorf("State = %q, want %q", briefing.State, domain.BriefingStartStateStopped)
			}
			second, secondCreated, err := store.BeginStopExperimentBriefing(context.Background(), tt.requestID, tt.sessionID)
			if err != nil {
				t.Fatalf("second BeginStopExperimentBriefing() error = %v", err)
			}
			if secondCreated {
				t.Error("second created = true, want false")
			}
			if second.OperationID != operation.OperationID {
				t.Errorf("second OperationID = %q, want %q", second.OperationID, operation.OperationID)
			}
		})
	}
}

// SQLite実験ブリーフ停止のrepository失敗境界。
func TestStoreStopExperimentBriefingFailures(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *Store) error
	}{
		{
			name: "停止操作検索失敗",
			run: func(_ *testing.T, store *Store) error {
				if err := store.db.Close(); err != nil {
					return err
				}
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
		{
			name: "停止識別子生成失敗",
			run: func(t *testing.T, store *Store) error {
				replaceBriefingRandom(t, func([]byte) (int, error) { return 0, errors.New("random") })
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
		{
			name: "停止transaction開始失敗",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) { return nil, errors.New("begin") }
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
		{
			name: "停止session検索失敗",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{rows: []briefingRow{fakeBriefingRow{err: errors.New("session")}}})
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
		{
			name: "停止非active session",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{values: []any{domain.BriefingStartStateFailed}},
					},
				})
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
		{
			name: "停止operation保存失敗",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					execErrors: []error{
						errors.New("insert"),
					},
					rows: []briefingRow{
						fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}},
					},
				})
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
		{
			name: "停止確定失敗",
			run: func(_ *testing.T, store *Store) error {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					commitError: errors.New("commit"),
					rows: []briefingRow{
						fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}},
					},
				})
				_, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			if err := tt.run(t, store); err == nil {
				t.Fatal("run() error = nil, want error")
			}
		})
	}
	if !isBriefingStopRequestConflict(errors.New("UNIQUE constraint failed: briefing_stop_operations.request_id")) {
		t.Error("isBriefingStopRequestConflict() = false, want true")
	}
	if isBriefingStopRequestConflict(errors.New("other")) {
		t.Error("isBriefingStopRequestConflict() = true, want false")
	}
}

// SQLite実験ブリーフ停止の冪等競合境界。
func TestStoreBeginStopExperimentBriefingConflictBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(*testing.T, *Store)
		wantCode  apperr.Code
		wantFound bool
	}{
		{
			name: "既存requestの別sessionを拒否する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("INSERT INTO briefing_stop_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", "operation-1", "request-1", "other-session", domain.BriefingStartStateStarting); err != nil {
					t.Fatalf("insert stop operation error = %v", err)
				}
			},
			wantCode: apperr.CodeBriefingRequestInvalid,
		},
		{
			name: "session不在を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					rows: []briefingRow{
						fakeBriefingRow{err: sql.ErrNoRows},
					},
				})
			},
			wantCode: apperr.CodeBriefingNotFound,
		},
		{
			name: "競合後の同一session操作を再利用する",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				transaction := &fakeBriefingTransaction{
					execErrors: []error{
						errors.New("UNIQUE constraint failed: briefing_stop_operations.request_id"),
					},
					rows: []briefingRow{
						fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}},
					},
					onExec: func(int) {
						if _, err := store.db.Exec("INSERT INTO briefing_stop_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", "operation-1", "request-1", "session-1", domain.BriefingStartStateStarting); err != nil {
							t.Fatalf("insert conflicting stop operation error = %v", err)
						}
					},
				}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)
			},
			wantFound: true,
		},
		{
			name: "競合後に操作が見つからない場合を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(&fakeBriefingTransaction{
					execErrors: []error{
						errors.New("UNIQUE constraint failed: briefing_stop_operations.request_id"),
					},
					rows: []briefingRow{
						fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}},
					},
				})
			},
		},
		{
			name: "競合後の操作検索失敗を返す",
			prepare: func(t *testing.T, store *Store) {
				t.Helper()
				transaction := &fakeBriefingTransaction{
					execErrors: []error{
						errors.New("UNIQUE constraint failed: briefing_stop_operations.request_id"),
					},
					rows: []briefingRow{
						fakeBriefingRow{values: []any{domain.BriefingStartStateStarted}},
					},
					onExec: func(int) {
						if err := store.db.Close(); err != nil {
							t.Fatalf("Close() error = %v", err)
						}
					},
				}
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(transaction)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			tt.prepare(t, store)

			operation, created, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1")
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("BeginStopExperimentBriefing() error = %v, want %q", err, tt.wantCode)
				}

				return
			}
			if tt.wantFound {
				if err != nil {
					t.Fatalf("BeginStopExperimentBriefing() error = %v", err)
				}
				if created {
					t.Error("created = true, want false")
				}
				if operation.OperationID != "operation-1" {
					t.Errorf("OperationID = %q, want %q", operation.OperationID, "operation-1")
				}

				return
			}
			if err == nil {
				t.Error("BeginStopExperimentBriefing() error = nil, want error")
			}
		})
	}
}

// SQLite実験ブリーフ停止状態同期の保存失敗。
func TestStoreUpdateStopExperimentBriefingFailures(t *testing.T) {
	tests := []struct {
		name  string
		state string
		tx    *fakeBriefingTransaction
		want  string
	}{
		{
			name: "transaction開始失敗を返す",
			want: "begin briefing stop update",
		},
		{
			name: "停止operation更新失敗を返す",
			tx: &fakeBriefingTransaction{
				execErrors: []error{
					errors.New("operation update failed"),
				},
			},
			want: "update briefing stop operation",
		},
		{
			name: "停止operation更新件数取得失敗を返す",
			tx: &fakeBriefingTransaction{
				result: fakeBriefingResult{rowsAffectedError: errors.New("count failed")},
			},
			want: "count briefing stop operation updates",
		},
		{
			name: "停止operation不在を返す",
			tx: &fakeBriefingTransaction{
				result: fakeBriefingResult{rowsAffected: 0},
			},
			want: "request not found",
		},
		{
			name:  "停止session更新失敗を返す",
			state: domain.BriefingStartStateStopped,
			tx: &fakeBriefingTransaction{
				execErrors: []error{
					nil,
					errors.New("session update failed"),
				},
			},
			want: "update stopped briefing session",
		},
		{
			name:  "停止session更新件数取得失敗を返す",
			state: domain.BriefingStartStateStopped,
			tx: &fakeBriefingTransaction{
				results: []sql.Result{
					fakeBriefingResult{rowsAffected: 1},
					fakeBriefingResult{rowsAffectedError: errors.New("session count failed")},
				},
			},
			want: "count stopped briefing session updates",
		},
		{
			name:  "停止session不在を返す",
			state: domain.BriefingStartStateStopped,
			tx: &fakeBriefingTransaction{
				results: []sql.Result{
					fakeBriefingResult{rowsAffected: 1},
					fakeBriefingResult{rowsAffected: 0},
				},
			},
			want: "session not found",
		},
		{
			name: "停止状態同期確定失敗を返す",
			tx: &fakeBriefingTransaction{
				commitError: errors.New("commit failed"),
			},
			want: "commit briefing stop update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			if tt.tx == nil {
				store.beginBriefingTransaction = func(context.Context) (briefingTransaction, error) {
					return nil, errors.New("begin failed")
				}
			} else {
				store.beginBriefingTransaction = fakeBriefingTransactionFactory(tt.tx)
			}

			state := tt.state
			if state == "" {
				state = domain.BriefingStartStateStopped
			}
			err := store.updateStopExperimentBriefing(context.Background(), "request-1", state, "failure")
			if err == nil {
				t.Fatal("updateStopExperimentBriefing() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("updateStopExperimentBriefing() error = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// SQLite実験ブリーフ停止失敗の保存。
func TestStoreFailStopExperimentBriefing(t *testing.T) {
	store := newTestStore(t)
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "session-1", "experiment_brief", domain.BriefingStartStateStarted, createdAt, createdAt); err != nil {
		t.Fatalf("insert preparation session error = %v", err)
	}
	if _, _, err := store.BeginStopExperimentBriefing(context.Background(), "request-1", "session-1"); err != nil {
		t.Fatalf("BeginStopExperimentBriefing() error = %v", err)
	}
	if err := store.FailStopExperimentBriefing(context.Background(), "request-1", string(apperr.CodeACPNotReady)); err != nil {
		t.Fatalf("FailStopExperimentBriefing() error = %v", err)
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

// 同時の一覧確認は確認時刻の書込みを直列化する。
func TestStoreListExperimentsSerializesConcurrentConfirmation(t *testing.T) {
	store := newTestStore(t)
	store.listMu.Lock()
	result := make(chan error, 1)
	go func() {
		_, err := store.ListExperiments(context.Background())
		result <- err
	}()

	select {
	case err := <-result:
		t.Fatalf("ListExperiments() completed while locked: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	store.listMu.Unlock()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("ListExperiments() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ListExperiments() did not complete after unlock")
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
