package usecase

import (
	"context"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// InsightWorkspaceReader は知見ワークスペース正本の読取port。
type InsightWorkspaceReader interface {
	GetInsightWorkspace(context.Context) (domain.InsightWorkspace, error)
}

// GetInsightWorkspace は知見ワークスペース再読込query。
type GetInsightWorkspace struct{ reader InsightWorkspaceReader }

// NewGetInsightWorkspace は知見ワークスペースqueryを生成する。
func NewGetInsightWorkspace(reader InsightWorkspaceReader) *GetInsightWorkspace {
	return &GetInsightWorkspace{reader: reader}
}

// Execute は知見作成画面の安全な正本を取得する。
func (u *GetInsightWorkspace) Execute(ctx context.Context) (domain.InsightWorkspace, error) {
	workspace, err := u.reader.GetInsightWorkspace(ctx)
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.InsightWorkspace{}, appErr
		}

		return domain.InsightWorkspace{}, apperr.Wrap(apperr.CodeInsightWorkspaceUnavailable, err)
	}

	return workspace, nil
}
