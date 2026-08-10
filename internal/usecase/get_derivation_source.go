package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentDerivationSourceReader は派生元の正本を読むport。
type ExperimentDerivationSourceReader interface {
	GetDerivationSource(context.Context, string) (domain.ExperimentDerivationSource, bool, error)
}

// GetDerivationSource は派生元確認query。
type GetDerivationSource struct {
	reader ExperimentDerivationSourceReader
}

// NewGetDerivationSource は派生元確認queryを生成する。
func NewGetDerivationSource(reader ExperimentDerivationSourceReader) *GetDerivationSource {
	return &GetDerivationSource{reader: reader}
}

// Execute は派生元確認の安全な結果を返す。
func (u *GetDerivationSource) Execute(ctx context.Context, experimentID string) (domain.ExperimentDerivationSource, error) {
	experimentID = strings.TrimSpace(experimentID)
	if experimentID == "" {
		return domain.ExperimentDerivationSource{}, apperr.New(apperr.CodeExperimentDerivationSourceInvalid)
	}
	source, found, err := u.reader.GetDerivationSource(ctx, experimentID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentDerivationSource{}, appErr
		}
		return domain.ExperimentDerivationSource{}, apperr.Wrap(apperr.CodeExperimentDerivationSourceUnavailable, err)
	}
	if !found {
		return domain.ExperimentDerivationSource{}, apperr.New(apperr.CodeExperimentDerivationSourceNotFound)
	}
	return source, nil
}
