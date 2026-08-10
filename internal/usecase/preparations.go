package usecase

import (
	"context"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// PreparationReader は環境準備session一覧を読み出すport。
type PreparationReader interface {
	ListPreparations(context.Context) ([]domain.Preparation, error)
}

// ListPreparations は環境準備session一覧query。
type ListPreparations struct {
	reader PreparationReader
}

// NewListPreparations は環境準備session一覧queryを生成。
func NewListPreparations(reader PreparationReader) *ListPreparations {
	return &ListPreparations{reader: reader}
}

// Execute は環境準備session一覧を取得。
func (u *ListPreparations) Execute(ctx context.Context) ([]domain.Preparation, error) {
	preparations, err := u.reader.ListPreparations(ctx)
	if err == nil {
		return preparations, nil
	}
	return nil, apperr.Wrap(apperr.CodePreparationsLoadFailed, err)
}
