package usecase

import (
	"context"
	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"strings"
)

// ExperimentEvaluationDetailReader は評価詳細を読み出すport。
type ExperimentEvaluationDetailReader interface {
	GetEvaluationDetail(context.Context, string) (domain.ExperimentEvaluationDetail, bool, error)
}

// GetEvaluationDetail は評価詳細の再読込query。
type GetEvaluationDetail struct {
	reader ExperimentEvaluationDetailReader
}

// NewGetEvaluationDetail は評価詳細queryを生成。
func NewGetEvaluationDetail(reader ExperimentEvaluationDetailReader) *GetEvaluationDetail {
	return &GetEvaluationDetail{reader: reader}
}

// Execute は指定評価の安全な結果を取得。
func (u *GetEvaluationDetail) Execute(ctx context.Context, evaluationID string) (domain.ExperimentEvaluationDetail, error) {
	if strings.TrimSpace(evaluationID) == "" {
		return domain.ExperimentEvaluationDetail{}, apperr.New(apperr.CodeEvaluationDetailRequestInvalid)
	}
	detail, found, err := u.reader.GetEvaluationDetail(ctx, strings.TrimSpace(evaluationID))
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.ExperimentEvaluationDetail{}, appErr
		}
		return domain.ExperimentEvaluationDetail{}, apperr.Wrap(apperr.CodeEvaluationDetailUnavailable, err)
	}
	if !found {
		return domain.ExperimentEvaluationDetail{}, apperr.New(apperr.CodeEvaluationDetailNotFound)
	}
	return detail, nil
}
