package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentConclusionFinalizer は実験結論を永続的に確定するport。
type ExperimentConclusionFinalizer interface {
	FinalizeExperimentConclusion(context.Context, string, string, string) (domain.ExperimentConclusion, bool, error)
}

// FinalizeExperimentConclusion は実験結論確定command。
type FinalizeExperimentConclusion struct{ store ExperimentConclusionFinalizer }

// NewFinalizeExperimentConclusion は実験結論確定commandを生成する。
func NewFinalizeExperimentConclusion(store ExperimentConclusionFinalizer) *FinalizeExperimentConclusion {
	return &FinalizeExperimentConclusion{store: store}
}

// Execute は評価済み実験の結論をrequest単位で確定する。
func (u *FinalizeExperimentConclusion) Execute(ctx context.Context, requestID, experimentID, conclusion string) (domain.ExperimentConclusion, error) {
	requestID = strings.TrimSpace(requestID)
	experimentID = strings.TrimSpace(experimentID)
	conclusion = strings.TrimSpace(conclusion)
	if requestID == "" || experimentID == "" || conclusion == "" || utf8.RuneCountInString(conclusion) > 8000 {
		return domain.ExperimentConclusion{}, apperr.New(apperr.CodeExperimentConclusionInvalid)
	}
	finalized, _, err := u.store.FinalizeExperimentConclusion(ctx, requestID, experimentID, conclusion)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentConclusion{}, appErr
		}
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.ExperimentConclusion{}, appErr
		}
		return domain.ExperimentConclusion{}, apperr.Wrap(apperr.CodeExperimentConclusionUnavailable, err)
	}
	return finalized, nil
}
