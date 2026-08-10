package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

type DerivedExperimentCreator interface {
	CreateDerivedExperiment(context.Context, string, string, domain.DerivedExperimentChanges, string) (domain.DerivedExperiment, bool, error)
}
type CreateDerivedExperiment struct{ creator DerivedExperimentCreator }

func NewCreateDerivedExperiment(creator DerivedExperimentCreator) *CreateDerivedExperiment {
	return &CreateDerivedExperiment{creator: creator}
}
func (u *CreateDerivedExperiment) Execute(ctx context.Context, requestID, sourceExperimentID string, changes domain.DerivedExperimentChanges, reason string) (domain.DerivedExperiment, error) {
	requestID = strings.TrimSpace(requestID)
	sourceExperimentID = strings.TrimSpace(sourceExperimentID)
	reason = strings.TrimSpace(reason)
	if requestID == "" || sourceExperimentID == "" || reason == "" || !hasDerivedExperimentChanges(changes) {
		return domain.DerivedExperiment{}, apperr.New(apperr.CodeDerivedExperimentInvalid)
	}
	result, _, err := u.creator.CreateDerivedExperiment(ctx, requestID, sourceExperimentID, changes, reason)
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.DerivedExperiment{}, appErr
		}
		if appErr := apperr.As(err); appErr != nil {
			return domain.DerivedExperiment{}, appErr
		}
		return domain.DerivedExperiment{}, apperr.Wrap(apperr.CodeDerivedExperimentUnavailable, err)
	}
	return result, nil
}
func hasDerivedExperimentChanges(changes domain.DerivedExperimentChanges) bool {
	return changes.Purpose != nil || changes.Hypothesis != nil || changes.EnvironmentConditions != nil || changes.InitialInput != nil || changes.Prompts != nil || changes.EvaluationAxes != nil
}
