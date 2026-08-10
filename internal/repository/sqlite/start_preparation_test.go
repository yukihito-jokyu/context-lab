package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 環境準備開始の作成、再送、状態更新を確認。
func TestStorePreparationStartLifecycle(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	created, wasCreated, err := store.BeginPreparation(ctx, "request-1", ".")
	if err != nil {
		t.Fatalf("BeginPreparation() error = %v", err)
	}
	if !wasCreated {
		t.Fatal("created = false, want true")
	}
	if created.State != domain.EnvironmentPreparationStateStarting {
		t.Errorf("State = %q, want %q", created.State, domain.EnvironmentPreparationStateStarting)
	}
	if err := store.MarkPreparationRunning(ctx, "request-1"); err != nil {
		t.Fatalf("MarkPreparationRunning() error = %v", err)
	}
	result := domain.EnvironmentPreparationResult{
		Candidates: []domain.EnvironmentPreparationCandidate{
			{
				EnvironmentConditions: "macOS",
				Summary:               "利用可能",
			},
		},
		Diagnostics: []domain.EnvironmentPreparationDiagnostic{
			{
				Code:        "CHECKED",
				SafeSummary: "確認済み",
			},
		},
	}
	if err := store.CompletePreparation(ctx, "request-1", result); err != nil {
		t.Fatalf("CompletePreparation() error = %v", err)
	}
	replayed, wasCreated, err := store.BeginPreparation(ctx, "request-1", ".")
	if err != nil {
		t.Fatalf("replay BeginPreparation() error = %v", err)
	}
	if wasCreated {
		t.Error("replay created = true, want false")
	}
	if replayed.State != domain.EnvironmentPreparationStateCompleted {
		t.Errorf("replay State = %q, want %q", replayed.State, domain.EnvironmentPreparationStateCompleted)
	}
	detail, found, err := store.GetPreparation(ctx, created.PreparationID)
	if err != nil {
		t.Fatalf("GetPreparation() error = %v", err)
	}
	if !found {
		t.Fatal("GetPreparation() found = false, want true")
	}
	if got := len(detail.Candidates); got != 1 {
		t.Errorf("Candidates length = %d, want 1", got)
	}
	if got := len(detail.Diagnostics); got != 1 {
		t.Errorf("Diagnostics length = %d, want 1", got)
	}
}

// 環境準備開始の競合と失敗状態を確認。
func TestStorePreparationStartConflictsAndFailure(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if _, _, err := store.BeginPreparation(ctx, "request-1", "."); err != nil {
		t.Fatalf("BeginPreparation() error = %v", err)
	}
	if _, _, err := store.BeginPreparation(ctx, "request-1", "other"); !apperr.IsCode(err, apperr.CodePreparationStartRequestConflict) {
		t.Errorf("different scope error = %v, want request conflict", err)
	}
	if _, _, err := store.BeginPreparation(ctx, "request-2", "."); !apperr.IsCode(err, apperr.CodePreparationStartPending) {
		t.Errorf("same scope error = %v, want pending", err)
	}
	if err := store.FailPreparation(ctx, "request-1", string(apperr.CodeACPNotReady)); err != nil {
		t.Fatalf("FailPreparation() error = %v", err)
	}
	start, _, err := store.BeginPreparation(ctx, "request-1", ".")
	if err != nil {
		t.Fatalf("replay BeginPreparation() error = %v", err)
	}
	if start.FailureCode != string(apperr.CodeACPNotReady) {
		t.Errorf("FailureCode = %q, want %q", start.FailureCode, apperr.CodeACPNotReady)
	}
}

// 環境準備開始の一意制約判定を確認。
func TestPreparationStartConflictPredicates(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "request ID競合",
			err:  errors.New("UNIQUE constraint failed: environment_preparation_operations.request_id"),
			want: true,
		},
		{
			name: "scope競合",
			err:  errors.New("UNIQUE constraint failed: environment_preparation_operations.scope"),
			want: true,
		},
		{
			name: "別エラー",
			err:  errors.New("other"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPreparationRequestConflict(tt.err); got != (tt.name == "request ID競合" && tt.want) {
				t.Errorf("isPreparationRequestConflict() = %v, want %v", got, tt.name == "request ID競合" && tt.want)
			}
			if got := isPreparationScopeConflict(tt.err); got != (tt.name == "scope競合" && tt.want) {
				t.Errorf("isPreparationScopeConflict() = %v, want %v", got, tt.name == "scope競合" && tt.want)
			}
		})
	}
}
