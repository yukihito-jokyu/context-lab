package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// DerivationBriefingStartStore は派生実験ブリーフ開始を記録するport。
type DerivationBriefingStartStore interface {
	BeginDerivationBriefing(context.Context, string, string) (domain.DerivationBriefingStart, bool, error)
	MarkDerivationBriefingStarted(context.Context, string) error
	MarkDerivationBriefingFailed(context.Context, string, string) error
}

// StartDerivationBriefing は派生実験ブリーフ開始command。
type StartDerivationBriefing struct {
	store   DerivationBriefingStartStore
	starter BriefingStarter
}

// NewStartDerivationBriefing は派生実験ブリーフ開始commandを生成。
func NewStartDerivationBriefing(store DerivationBriefingStartStore, starter BriefingStarter) *StartDerivationBriefing {
	return &StartDerivationBriefing{store: store, starter: starter}
}

// Execute は派生元を確認して派生実験ブリーフを開始する。
func (u *StartDerivationBriefing) Execute(ctx context.Context, requestID, sourceExperimentID string) (domain.DerivationBriefingStart, error) {
	requestID = strings.TrimSpace(requestID)
	sourceExperimentID = strings.TrimSpace(sourceExperimentID)
	if requestID == "" || sourceExperimentID == "" {
		return domain.DerivationBriefingStart{}, apperr.New(apperr.CodeDerivedExperimentInvalid)
	}

	start, created, err := u.store.BeginDerivationBriefing(ctx, requestID, sourceExperimentID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.DerivationBriefingStart{}, appErr
		}

		return domain.DerivationBriefingStart{}, apperr.Wrap(apperr.CodeDerivationBriefingStartFailed, err)
	}
	if !created {
		if start.State == domain.BriefingStartStateFailed {
			return domain.DerivationBriefingStart{}, derivationBriefingFailure(start.FailureCode)
		}
		if start.State == domain.BriefingStartStateStarting {
			return domain.DerivationBriefingStart{}, apperr.New(apperr.CodeDerivationBriefingPending)
		}

		return start, nil
	}

	if err := u.starter.StartExperimentBriefing(ctx, start.BriefingSessionID, start.OperationID); err != nil {
		failure := apperr.As(err)
		if failure == nil {
			failure = apperr.New(apperr.CodeDerivationBriefingStartFailed)
		}
		if markErr := u.store.MarkDerivationBriefingFailed(ctx, requestID, string(failure.Code)); markErr != nil {
			return domain.DerivationBriefingStart{}, apperr.Wrap(apperr.CodeDerivationBriefingStartFailed, markErr)
		}

		return domain.DerivationBriefingStart{}, failure
	}

	if err := u.store.MarkDerivationBriefingStarted(ctx, requestID); err != nil {
		return domain.DerivationBriefingStart{}, apperr.Wrap(apperr.CodeDerivationBriefingStartFailed, err)
	}
	start.State = domain.BriefingStartStateStarted

	return start, nil
}

// derivationBriefingFailure は永続済みの安全な開始失敗を復元。
func derivationBriefingFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeACPNotReady:
		return apperr.New(apperr.CodeACPNotReady)
	case apperr.CodeDerivedExperimentInvalid:
		return apperr.New(apperr.CodeDerivedExperimentInvalid)
	case apperr.CodeDerivedExperimentSourceNotFound:
		return apperr.New(apperr.CodeDerivedExperimentSourceNotFound)
	case apperr.CodeDerivedExperimentSourceNotEligible:
		return apperr.New(apperr.CodeDerivedExperimentSourceNotEligible)
	default:
		return apperr.New(apperr.CodeDerivationBriefingStartFailed)
	}
}
