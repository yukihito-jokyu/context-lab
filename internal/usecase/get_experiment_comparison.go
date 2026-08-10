package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentComparisonReader は実験比較の正本を読み出すport。
type ExperimentComparisonReader interface {
	GetExperimentComparison(context.Context, string) (domain.ExperimentComparison, bool, error)
}

// GetExperimentComparison は実験比較の再読込query。
type GetExperimentComparison struct{ reader ExperimentComparisonReader }

// NewGetExperimentComparison は実験比較queryを生成する。
func NewGetExperimentComparison(reader ExperimentComparisonReader) *GetExperimentComparison {
	return &GetExperimentComparison{reader: reader}
}

// Execute は指定実験の安全な比較結果を取得する。
func (u *GetExperimentComparison) Execute(ctx context.Context, experimentID string) (domain.ExperimentComparison, error) {
	experimentID = strings.TrimSpace(experimentID)
	if experimentID == "" {
		return domain.ExperimentComparison{}, apperr.New(apperr.CodeExperimentComparisonInvalid)
	}
	comparison, found, err := u.reader.GetExperimentComparison(ctx, experimentID)
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.ExperimentComparison{}, appErr
		}
		return domain.ExperimentComparison{}, apperr.Wrap(apperr.CodeExperimentComparisonUnavailable, err)
	}
	if !found {
		return domain.ExperimentComparison{}, apperr.New(apperr.CodeExperimentComparisonNotFound)
	}
	return comparison, nil
}
