package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// PreparationScopeValidator は作業root配下の安全な準備範囲を検証するport。
type PreparationScopeValidator interface {
	ValidatePreparationScope(string) (string, string, error)
}

// EnvironmentPreparer は安全な範囲をACPで照合するport。
type EnvironmentPreparer interface {
	PrepareEnvironment(context.Context, string) (domain.EnvironmentPreparationResult, error)
}

// PreparationStartStore は開始操作と安全な結果を永続化するport。
type PreparationStartStore interface {
	BeginPreparation(context.Context, string, string) (domain.EnvironmentPreparationStart, bool, error)
	MarkPreparationRunning(context.Context, string) error
	CompletePreparation(context.Context, string, domain.EnvironmentPreparationResult) error
	FailPreparation(context.Context, string, string) error
}

// StartPreparation は環境準備開始command。
type StartPreparation struct {
	store     PreparationStartStore
	validator PreparationScopeValidator
	preparer  EnvironmentPreparer
}

// NewStartPreparation は環境準備開始commandを生成。
func NewStartPreparation(store PreparationStartStore, validator PreparationScopeValidator, preparer EnvironmentPreparer) *StartPreparation {
	return &StartPreparation{store: store, validator: validator, preparer: preparer}
}

// Execute は対象範囲を検証してACP環境準備を開始する。
func (u *StartPreparation) Execute(ctx context.Context, requestID, scope string) (domain.EnvironmentPreparationStart, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return domain.EnvironmentPreparationStart{}, apperr.New(apperr.CodePreparationStartRequestInvalid)
	}

	resolvedScope, canonicalScope, err := u.validator.ValidatePreparationScope(scope)
	if err != nil {
		return domain.EnvironmentPreparationStart{}, apperr.New(apperr.CodePreparationScopeInvalid)
	}
	start, created, err := u.store.BeginPreparation(ctx, requestID, canonicalScope)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.EnvironmentPreparationStart{}, appErr
		}

		return domain.EnvironmentPreparationStart{}, apperr.Wrap(apperr.CodePreparationStartUnavailable, err)
	}
	if !created {
		return replayPreparationStart(start)
	}

	if err := u.store.MarkPreparationRunning(ctx, requestID); err != nil {
		return u.fail(ctx, requestID, apperr.Wrap(apperr.CodePreparationStartUnavailable, err))
	}
	start.State = domain.EnvironmentPreparationStateRunning

	result, err := u.preparer.PrepareEnvironment(ctx, resolvedScope)
	if err != nil {
		return u.fail(ctx, requestID, preparationStartFailure(err))
	}
	if err := u.store.CompletePreparation(ctx, requestID, result); err != nil {
		return u.fail(ctx, requestID, apperr.Wrap(apperr.CodePreparationStartUnavailable, err))
	}
	start.State = domain.EnvironmentPreparationStateCompleted

	return start, nil
}

// fail は安全な失敗をoperationへ記録して返す。
func (u *StartPreparation) fail(ctx context.Context, requestID string, failure error) (domain.EnvironmentPreparationStart, error) {
	code := preparationStartFailureCode(failure)
	if err := u.store.FailPreparation(ctx, requestID, string(code)); err != nil {
		return domain.EnvironmentPreparationStart{}, apperr.Wrap(apperr.CodePreparationStartUnavailable, err)
	}

	return domain.EnvironmentPreparationStart{}, preparationStartFailure(failure)
}

// replayPreparationStart は再送時に永続済み開始結果を復元する。
func replayPreparationStart(start domain.EnvironmentPreparationStart) (domain.EnvironmentPreparationStart, error) {
	switch start.State {
	case domain.EnvironmentPreparationStateStarting, domain.EnvironmentPreparationStateRunning:
		return domain.EnvironmentPreparationStart{}, apperr.New(apperr.CodePreparationStartPending)
	case domain.EnvironmentPreparationStateFailed:
		return domain.EnvironmentPreparationStart{}, preparationStartFailure(apperr.New(apperr.Code(start.FailureCode)))
	default:
		return start, nil
	}
}

// preparationStartFailureCode は失敗を保存可能な安全コードへ正規化する。
func preparationStartFailureCode(err error) apperr.Code {
	if appErr := apperr.As(err); appErr != nil && appErr.Code == apperr.CodeACPNotReady {
		return apperr.CodeACPNotReady
	}

	return apperr.CodePreparationStartUnavailable
}

// preparationStartFailure は開始失敗を画面安全なエラーへ正規化する。
func preparationStartFailure(err error) error {
	if preparationStartFailureCode(err) == apperr.CodeACPNotReady {
		return apperr.New(apperr.CodeACPNotReady)
	}

	return apperr.New(apperr.CodePreparationStartUnavailable)
}
