package usecase

import (
	"context"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentReader は実験一覧を読み出すport。
type ExperimentReader interface {
	ListExperiments(context.Context) (domain.ExperimentCollection, error)
}

// ListExperiments は実験一覧query。
type ListExperiments struct {
	reader ExperimentReader
}

// NewListExperiments は実験一覧queryを生成。
func NewListExperiments(reader ExperimentReader) *ListExperiments {
	return &ListExperiments{reader: reader}
}

// Execute は実験一覧を取得。
func (u *ListExperiments) Execute(ctx context.Context) (domain.ExperimentCollection, error) {
	collection, err := u.reader.ListExperiments(ctx)
	if err == nil {
		return collection, nil
	}
	if appErr := apperr.FromContextError(err); appErr != nil {
		return domain.ExperimentCollection{}, appErr
	}

	return domain.ExperimentCollection{}, apperr.Wrap(apperr.CodeExperimentsLoadFailed, err)
}
