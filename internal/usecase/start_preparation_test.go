package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// StartPreparationの入力、再送、ACP失敗を確認。
func TestStartPreparationExecute(t *testing.T) {
	tests := []struct {
		name          string
		request       string
		scope         string
		store         preparationStartStoreStub
		validator     preparationScopeValidatorStub
		preparer      preparationPreparerStub
		wantCode      apperr.Code
		wantState     string
		wantCompleted bool
	}{
		{
			name:     "request ID不足",
			request:  " ",
			wantCode: apperr.CodePreparationStartRequestInvalid,
		},
		{
			name:      "範囲不正",
			request:   "request",
			scope:     "bad",
			validator: preparationScopeValidatorStub{err: errors.New("bad")},
			wantCode:  apperr.CodePreparationScopeInvalid,
		},
		{
			name:          "完了",
			request:       "request",
			scope:         ".",
			validator:     preparationScopeValidatorStub{resolved: "/root"},
			preparer:      preparationPreparerStub{result: domain.EnvironmentPreparationResult{}},
			wantState:     domain.EnvironmentPreparationStateCompleted,
			wantCompleted: true,
		},
		{
			name:      "同一範囲稼働中",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store:     preparationStartStoreStub{start: domain.EnvironmentPreparationStart{State: domain.EnvironmentPreparationStateRunning}},
			wantCode:  apperr.CodePreparationStartPending,
		},
		{
			name:      "開始済み直後の再送",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store: preparationStartStoreStub{start: domain.EnvironmentPreparationStart{
				State: domain.EnvironmentPreparationStateStarting,
			}},
			wantCode: apperr.CodePreparationStartPending,
		},
		{
			name:      "完了済みの再送",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store: preparationStartStoreStub{start: domain.EnvironmentPreparationStart{
				PreparationID: "preparation",
				State:         domain.EnvironmentPreparationStateCompleted,
			}},
			wantState: domain.EnvironmentPreparationStateCompleted,
		},
		{
			name:      "失敗済みの再送",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store: preparationStartStoreStub{start: domain.EnvironmentPreparationStart{
				State:       domain.EnvironmentPreparationStateFailed,
				FailureCode: string(apperr.CodeACPNotReady),
			}},
			wantCode: apperr.CodeACPNotReady,
		},
		{
			name:      "開始保存のアプリケーションエラー",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store:     preparationStartStoreStub{err: apperr.New(apperr.CodePreparationStartRequestConflict)},
			wantCode:  apperr.CodePreparationStartRequestConflict,
		},
		{
			name:      "開始保存の内部エラー",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store:     preparationStartStoreStub{err: errors.New("save failed")},
			wantCode:  apperr.CodePreparationStartUnavailable,
		},
		{
			name:      "実行中更新失敗",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store:     preparationStartStoreStub{markErr: errors.New("mark failed")},
			wantCode:  apperr.CodePreparationStartUnavailable,
		},
		{
			name:      "完了保存失敗",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			store:     preparationStartStoreStub{completeErr: errors.New("complete failed")},
			wantCode:  apperr.CodePreparationStartUnavailable,
		},
		{
			name:      "ACP未準備",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			preparer:  preparationPreparerStub{err: apperr.New(apperr.CodeACPNotReady)},
			wantCode:  apperr.CodeACPNotReady,
		},
		{
			name:      "ACP内部エラー",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			preparer:  preparationPreparerStub{err: errors.New("acp failed")},
			wantCode:  apperr.CodePreparationStartUnavailable,
		},
		{
			name:      "失敗保存失敗",
			request:   "request",
			scope:     ".",
			validator: preparationScopeValidatorStub{resolved: "/root"},
			preparer:  preparationPreparerStub{err: errors.New("acp failed")},
			store:     preparationStartStoreStub{failErr: errors.New("fail save failed")},
			wantCode:  apperr.CodePreparationStartUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := NewStartPreparation(&tt.store, tt.validator, tt.preparer)
			got, err := command.Execute(context.Background(), tt.request, tt.scope)
			if tt.wantCode != "" {
				appErr := apperr.As(err)
				if appErr == nil {
					t.Fatalf("Execute() error = %v, want code %q", err, tt.wantCode)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("Code = %q, want %q", appErr.Code, tt.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got.State != tt.wantState {
				t.Errorf("State = %q, want %q", got.State, tt.wantState)
			}
			if tt.store.completed != tt.wantCompleted {
				t.Errorf("CompletePreparation() called = %v, want %v", tt.store.completed, tt.wantCompleted)
			}
		})
	}
}

// preparationScopeValidatorStub は範囲検証portのtest double。
type preparationScopeValidatorStub struct {
	resolved  string
	canonical string
	err       error
}

// ValidatePreparationScope は指定済み範囲または失敗を返す。
func (s preparationScopeValidatorStub) ValidatePreparationScope(string) (string, string, error) {
	return s.resolved, s.canonical, s.err
}

// preparationPreparerStub はACP準備portのtest double。
type preparationPreparerStub struct {
	result domain.EnvironmentPreparationResult
	err    error
}

// PrepareEnvironment は指定済み結果または失敗を返す。
func (s preparationPreparerStub) PrepareEnvironment(context.Context, string) (domain.EnvironmentPreparationResult, error) {
	return s.result, s.err
}

// preparationStartStoreStub は開始保存portのtest double。
type preparationStartStoreStub struct {
	start       domain.EnvironmentPreparationStart
	created     bool
	err         error
	completed   bool
	markErr     error
	completeErr error
	failErr     error
}

// BeginPreparation は開始済み状態または初回状態を返す。
func (s *preparationStartStoreStub) BeginPreparation(context.Context, string, string) (domain.EnvironmentPreparationStart, bool, error) {
	if s.err != nil {
		return domain.EnvironmentPreparationStart{}, false, s.err
	}
	if s.start.State == "" {
		s.start = domain.EnvironmentPreparationStart{
			PreparationID: "preparation",
			State:         domain.EnvironmentPreparationStateStarting,
		}
		s.created = true
	}
	return s.start, s.created, nil
}

// MarkPreparationRunning は開始状態を更新する。
func (s *preparationStartStoreStub) MarkPreparationRunning(context.Context, string) error {
	return s.markErr
}

// CompletePreparation は完了呼出を記録する。
func (s *preparationStartStoreStub) CompletePreparation(context.Context, string, domain.EnvironmentPreparationResult) error {
	s.completed = true
	return s.completeErr
}

// FailPreparation は失敗記録を受理する。
func (s *preparationStartStoreStub) FailPreparation(context.Context, string, string) error {
	return s.failErr
}
