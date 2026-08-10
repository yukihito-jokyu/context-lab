package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 下書き保存commandの入力検証とエラー正規化。
func TestSaveExperimentPreparationDraftExecute(t *testing.T) {
	tests := []struct {
		name     string
		draft    domain.ExperimentPreparationDraft
		storeErr error
		wantCode apperr.Code
		wantCall bool
	}{
		{
			name: "空のrequest IDを拒否する",
			draft: domain.ExperimentPreparationDraft{
				ExperimentID: "experiment-1",
			},
			wantCode: apperr.CodeDraftRequestInvalid,
		},
		{
			name: "空のexperiment IDを拒否する",
			draft: domain.ExperimentPreparationDraft{
				RequestID: "request-1",
			},
			wantCode: apperr.CodeDraftRequestInvalid,
		},
		{
			name: "内部保存失敗を安全なコードへ変換する",
			draft: domain.ExperimentPreparationDraft{
				RequestID:    "request-1",
				ExperimentID: "experiment-1",
			},
			storeErr: errors.New("private database path"),
			wantCode: apperr.CodeDraftSaveFailed,
			wantCall: true,
		},
		{
			name: "フォーム値をそのまま保存する",
			draft: domain.ExperimentPreparationDraft{
				RequestID:    " request-1 ",
				ExperimentID: " experiment-1 ",
				Purpose:      " 目的 ",
				Prompts: []domain.ExperimentPreparationPrompt{
					{
						SequenceNo: 1,
						Content:    " prompt ",
					},
				},
			},
			wantCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeExperimentPreparationDraftStore{err: tt.storeErr}
			command := NewSaveExperimentPreparationDraft(store)

			_, err := command.Execute(context.Background(), tt.draft)
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)
			} else if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if gotCall := store.called; gotCall != tt.wantCall {
				t.Errorf("SaveExperimentPreparationDraft() called = %v, want %v", gotCall, tt.wantCall)
			}
			if !tt.wantCall || tt.wantCode != "" {
				return
			}
			if store.draft.Purpose != tt.draft.Purpose {
				t.Errorf("Purpose = %q, want %q", store.draft.Purpose, tt.draft.Purpose)
			}
			if store.draft.RequestID != "request-1" || store.draft.ExperimentID != "experiment-1" {
				t.Errorf("identifier = (%q, %q), want (request-1, experiment-1)", store.draft.RequestID, store.draft.ExperimentID)
			}
		})
	}
}

// fakeExperimentPreparationDraftStore は下書き保存portのtest double。
type fakeExperimentPreparationDraftStore struct {
	draft  domain.ExperimentPreparationDraft
	err    error
	called bool
}

// SaveExperimentPreparationDraft は指定済み下書きまたは失敗を返却。
func (f *fakeExperimentPreparationDraftStore) SaveExperimentPreparationDraft(_ context.Context, draft domain.ExperimentPreparationDraft) (domain.ExperimentPreparationDraft, error) {
	f.called = true
	f.draft = draft
	if f.err != nil {
		return domain.ExperimentPreparationDraft{}, f.err
	}

	return draft, nil
}
