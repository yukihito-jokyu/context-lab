package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 環境候補採用commandの入力検証と安全な候補返却。
func TestAdoptCandidateExecute(t *testing.T) {
	tests := []struct {
		name          string
		requestID     string
		preparationID string
		candidateID   string
		reader        candidateAdoptionReaderStub
		want          AdoptedCandidate
		wantCode      apperr.Code
	}{
		{
			name:          "完了済み候補を返す",
			requestID:     " request-1 ",
			preparationID: " preparation-1 ",
			candidateID:   " candidate-1 ",
			reader: candidateAdoptionReaderStub{preparation: domain.PreparationDetail{
				ID:    "preparation-1",
				State: domain.EnvironmentPreparationStateCompleted,
				Candidates: []domain.PreparationCandidate{{
					ID:                    "candidate-1",
					EnvironmentConditions: "macOS 15",
				}},
			}, found: true},
			want: AdoptedCandidate{
				PreparationID:         "preparation-1",
				CandidateID:           "candidate-1",
				EnvironmentConditions: "macOS 15",
			},
		},
		{
			name:      "要求IDが空なら拒否する",
			requestID: " ",
			wantCode:  apperr.CodeCandidateAdoptionRequestInvalid,
		},
		{
			name:          "準備IDが空なら拒否する",
			requestID:     "request-1",
			preparationID: " ",
			wantCode:      apperr.CodeCandidateAdoptionRequestInvalid,
		},
		{
			name:          "候補IDが空なら拒否する",
			requestID:     "request-1",
			preparationID: "preparation-1",
			candidateID:   " ",
			wantCode:      apperr.CodeCandidateAdoptionRequestInvalid,
		},
		{
			name:          "存在しない準備を拒否する",
			requestID:     "request-1",
			preparationID: "preparation-1",
			candidateID:   "candidate-1",
			reader:        candidateAdoptionReaderStub{},
			wantCode:      apperr.CodePreparationNotFound,
		},
		{
			name:          "未完了準備を拒否する",
			requestID:     "request-1",
			preparationID: "preparation-1",
			candidateID:   "candidate-1",
			reader: candidateAdoptionReaderStub{preparation: domain.PreparationDetail{
				State: domain.EnvironmentPreparationStateRunning,
			}, found: true},
			wantCode: apperr.CodePreparationCandidateNotReady,
		},
		{
			name:          "存在しない候補を拒否する",
			requestID:     "request-1",
			preparationID: "preparation-1",
			candidateID:   "candidate-1",
			reader: candidateAdoptionReaderStub{preparation: domain.PreparationDetail{
				State: domain.EnvironmentPreparationStateCompleted,
			}, found: true},
			wantCode: apperr.CodePreparationCandidateNotFound,
		},
		{
			name:          "読み出し失敗を安全なエラーへ変換する",
			requestID:     "request-1",
			preparationID: "preparation-1",
			candidateID:   "candidate-1",
			reader:        candidateAdoptionReaderStub{err: errors.New("private database detail")},
			wantCode:      apperr.CodeCandidateAdoptionUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAdoptCandidate(tt.reader).Execute(context.Background(), tt.requestID, tt.preparationID, tt.candidateID)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Execute() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// candidateAdoptionReaderStub は候補採用query用readerのtest double。
type candidateAdoptionReaderStub struct {
	preparation domain.PreparationDetail
	found       bool
	err         error
}

// GetPreparation は指定済み環境準備または失敗を返却。
func (s candidateAdoptionReaderStub) GetPreparation(context.Context, string) (domain.PreparationDetail, bool, error) {
	return s.preparation, s.found, s.err
}
