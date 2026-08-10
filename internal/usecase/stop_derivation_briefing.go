package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// DerivationBriefingStopStore は派生実験ブリーフ終了を記録するport。
type DerivationBriefingStopStore interface {
	BeginStopDerivationBriefing(context.Context, string, string) (domain.DerivationBriefingStopOperation, bool, error)
	CompleteStopDerivationBriefing(context.Context, string) error
	FailStopDerivationBriefing(context.Context, string, string) error
}

// DerivationBriefingStopper は派生実験ブリーフ終了をACPへ委譲するport。
type DerivationBriefingStopper interface {
	StopDerivationBriefing(context.Context, string, string) error
}

// StopDerivationBriefing は派生実験ブリーフ終了command。
type StopDerivationBriefing struct {
	store   DerivationBriefingStopStore
	stopper DerivationBriefingStopper
}

// NewStopDerivationBriefing は派生実験ブリーフ終了commandを生成する。
func NewStopDerivationBriefing(store DerivationBriefingStopStore, stopper DerivationBriefingStopper) *StopDerivationBriefing {
	return &StopDerivationBriefing{store: store, stopper: stopper}
}

// Execute は派生実験ブリーフ終了を記録してACP停止を委譲する。
func (u *StopDerivationBriefing) Execute(ctx context.Context, requestID, briefingSessionID string) (domain.DerivationBriefingStopOperation, error) {
	requestID = strings.TrimSpace(requestID)
	briefingSessionID = strings.TrimSpace(briefingSessionID)
	if requestID == "" || briefingSessionID == "" {
		return domain.DerivationBriefingStopOperation{}, apperr.New(apperr.CodeDerivationBriefingStopInvalid)
	}

	operation, created, err := u.store.BeginStopDerivationBriefing(ctx, requestID, briefingSessionID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.DerivationBriefingStopOperation{}, appErr
		}

		return domain.DerivationBriefingStopOperation{}, apperr.Wrap(apperr.CodeDerivationBriefingStopFailed, err)
	}
	if !created {
		if operation.State == domain.BriefingStartStateFailed {
			return domain.DerivationBriefingStopOperation{}, derivationBriefingStopFailure(operation.FailureCode)
		}
		if operation.State == domain.BriefingStartStateStarting {
			return domain.DerivationBriefingStopOperation{}, apperr.New(apperr.CodeDerivationBriefingStopPending)
		}

		return operation, nil
	}

	if err := u.stopper.StopDerivationBriefing(ctx, operation.BriefingSessionID, operation.OperationID); err != nil {
		failure := apperr.As(err)
		if failure == nil {
			failure = apperr.New(apperr.CodeDerivationBriefingStopFailed)
		}
		if markErr := u.store.FailStopDerivationBriefing(ctx, operation.RequestID, string(failure.Code)); markErr != nil {
			return domain.DerivationBriefingStopOperation{}, apperr.Wrap(apperr.CodeDerivationBriefingStopFailed, markErr)
		}

		return domain.DerivationBriefingStopOperation{}, failure
	}
	if err := u.store.CompleteStopDerivationBriefing(ctx, operation.RequestID); err != nil {
		return domain.DerivationBriefingStopOperation{}, apperr.Wrap(apperr.CodeDerivationBriefingStopFailed, err)
	}
	operation.State = domain.BriefingStartStateStopped

	return operation, nil
}

// derivationBriefingStopFailure は永続済みの安全な終了失敗を復元する。
func derivationBriefingStopFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeACPNotReady:
		return apperr.New(apperr.CodeACPNotReady)
	case apperr.CodeDerivationBriefingStopInvalid:
		return apperr.New(apperr.CodeDerivationBriefingStopInvalid)
	case apperr.CodeDerivationBriefingStopNotFound:
		return apperr.New(apperr.CodeDerivationBriefingStopNotFound)
	case apperr.CodeDerivationBriefingStopNotActive:
		return apperr.New(apperr.CodeDerivationBriefingStopNotActive)
	default:
		return apperr.New(apperr.CodeDerivationBriefingStopFailed)
	}
}
