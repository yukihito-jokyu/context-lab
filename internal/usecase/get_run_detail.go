package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentRunDetailReader はrun詳細を読み出すport。
type ExperimentRunDetailReader interface {
	GetRunDetail(context.Context, string) (domain.ExperimentRunDetail, bool, error)
}

// GetRunDetail はrun詳細の再読込query。
type GetRunDetail struct {
	reader ExperimentRunDetailReader
}

// NewGetRunDetail はrun詳細再読込queryを生成。
func NewGetRunDetail(reader ExperimentRunDetailReader) *GetRunDetail {
	return &GetRunDetail{reader: reader}
}

// Execute は指定runの安全な観測結果を取得。
func (u *GetRunDetail) Execute(ctx context.Context, runID string) (domain.ExperimentRunDetail, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return domain.ExperimentRunDetail{}, apperr.New(apperr.CodeRunDetailRequestInvalid)
	}

	detail, found, err := u.reader.GetRunDetail(ctx, runID)
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.ExperimentRunDetail{}, appErr
		}

		return domain.ExperimentRunDetail{}, apperr.Wrap(apperr.CodeRunDetailUnavailable, err)
	}
	if !found {
		return domain.ExperimentRunDetail{}, apperr.New(apperr.CodeRunDetailNotFound)
	}

	return detail, nil
}
