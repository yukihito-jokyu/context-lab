package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentPreparationDraftStore は編集下書きを保存するport。
type ExperimentPreparationDraftStore interface {
	SaveExperimentPreparationDraft(context.Context, domain.ExperimentPreparationDraft) (domain.ExperimentPreparationDraft, error)
}

// SaveExperimentPreparationDraft は実験準備の下書き保存command。
type SaveExperimentPreparationDraft struct {
	store ExperimentPreparationDraftStore
}

// NewSaveExperimentPreparationDraft は下書き保存commandを生成。
func NewSaveExperimentPreparationDraft(store ExperimentPreparationDraftStore) *SaveExperimentPreparationDraft {
	return &SaveExperimentPreparationDraft{store: store}
}

// Execute は準備中実験のフォーム値を冪等に保存。
func (u *SaveExperimentPreparationDraft) Execute(ctx context.Context, draft domain.ExperimentPreparationDraft) (domain.ExperimentPreparationDraft, error) {
	draft.RequestID = strings.TrimSpace(draft.RequestID)
	draft.ExperimentID = strings.TrimSpace(draft.ExperimentID)
	if draft.RequestID == "" || draft.ExperimentID == "" {
		return domain.ExperimentPreparationDraft{}, apperr.New(apperr.CodeDraftRequestInvalid)
	}

	saved, err := u.store.SaveExperimentPreparationDraft(ctx, draft)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentPreparationDraft{}, appErr
		}

		return domain.ExperimentPreparationDraft{}, apperr.Wrap(apperr.CodeDraftSaveFailed, err)
	}

	return saved, nil
}
