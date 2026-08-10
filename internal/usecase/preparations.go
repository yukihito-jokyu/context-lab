package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// PreparationReader は環境準備session一覧を読み出すport。
type PreparationReader interface {
	ListPreparations(context.Context) ([]domain.Preparation, error)
	GetPreparation(context.Context, string) (domain.PreparationDetail, bool, error)
}

// GetPreparation は環境準備session詳細query。
type GetPreparation struct {
	reader PreparationReader
}

// NewGetPreparation は環境準備session詳細queryを生成。
func NewGetPreparation(reader PreparationReader) *GetPreparation {
	return &GetPreparation{reader: reader}
}

// Execute は環境準備session詳細を取得。
func (u *GetPreparation) Execute(ctx context.Context, preparationID string) (domain.PreparationDetail, error) {
	preparationID = strings.TrimSpace(preparationID)
	if preparationID == "" {
		return domain.PreparationDetail{}, apperr.New(apperr.CodePreparationRequestInvalid)
	}

	preparation, found, err := u.reader.GetPreparation(ctx, preparationID)
	if err != nil {
		return domain.PreparationDetail{}, apperr.Wrap(apperr.CodePreparationUnavailable, err)
	}
	if !found {
		return domain.PreparationDetail{}, apperr.New(apperr.CodePreparationNotFound)
	}

	return preparation, nil
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
