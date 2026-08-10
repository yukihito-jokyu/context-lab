package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentConditionsStore は固定条件を永続化するport。
type ExperimentConditionsStore interface {
	FixExperimentConditions(context.Context, domain.ExperimentFixedConditions) (domain.ExperimentFixedConditions, error)
}

// FixExperimentConditions は実験条件固定command。
type FixExperimentConditions struct {
	store ExperimentConditionsStore
}

// NewFixExperimentConditions は条件固定commandを生成。
func NewFixExperimentConditions(store ExperimentConditionsStore) *FixExperimentConditions {
	return &FixExperimentConditions{store: store}
}

// Execute は準備中の条件を冪等に固定する。
func (u *FixExperimentConditions) Execute(ctx context.Context, conditions domain.ExperimentFixedConditions) (domain.ExperimentFixedConditions, error) {
	conditions.RequestID = strings.TrimSpace(conditions.RequestID)
	conditions.ExperimentID = strings.TrimSpace(conditions.ExperimentID)
	if conditions.RequestID == "" || conditions.ExperimentID == "" {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeFixConditionsRequestInvalid)
	}
	if !conditions.Valid() {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeConditionsInvalid)
	}

	fixed, err := u.store.FixExperimentConditions(ctx, conditions)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentFixedConditions{}, appErr
		}

		return domain.ExperimentFixedConditions{}, apperr.Wrap(apperr.CodeFixConditionsSaveFailed, err)
	}

	return fixed, nil
}
