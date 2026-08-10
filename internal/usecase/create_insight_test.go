package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// CreateInsightの入力検証と障害変換を検証する。
func TestCreateInsightExecute(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		evidences []domain.InsightEvidence
		err       error
		wantCode  apperr.Code
	}{
		{
			name:      "成功",
			requestID: "request",
			evidences: insightEvidencePair("b", "2", "a", "1"),
		},
		{
			name:      "本文不足",
			requestID: "",
			evidences: insightEvidencePair("a", "1", "b", "2"),
			wantCode:  apperr.CodeInsightCreateInvalid,
		},
		{
			name:      "根拠不足",
			requestID: "request",
			evidences: []domain.InsightEvidence{{
				ExperimentID: "a",
				ConclusionID: "1",
			}},
			wantCode: apperr.CodeInsightCreateEvidenceInsufficient,
		},
		{
			name:      "重複実験",
			requestID: "request",
			evidences: insightEvidencePair("a", "1", "a", "2"),
			wantCode:  apperr.CodeInsightCreateInvalid,
		},
		{
			name:      "既知障害",
			requestID: "request",
			evidences: insightEvidencePair("a", "1", "b", "2"),
			err:       apperr.New(apperr.CodeInsightCreateEvidenceNotFound),
			wantCode:  apperr.CodeInsightCreateEvidenceNotFound,
		},
		{
			name:      "未知障害",
			requestID: "request",
			evidences: insightEvidencePair("a", "1", "b", "2"),
			err:       errors.New("sqlite"),
			wantCode:  apperr.CodeInsightCreateUnavailable,
		},
		{
			name:      "中止",
			requestID: "request",
			evidences: insightEvidencePair("a", "1", "b", "2"),
			err:       context.Canceled,
			wantCode:  apperr.CodeOperationCanceled,
		},
		{
			name:      "空根拠",
			requestID: "request",
			wantCode:  apperr.CodeInsightCreateEvidenceInsufficient,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCreateInsight(insightCreatorStub{err: tt.err}).Execute(context.Background(), tt.requestID, tt.evidences, " statement ", " conditions ", " gaps ")
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got.Evidences[0].ExperimentID != "a" || got.Statement != "statement" {
				t.Errorf("Execute() = %+v, want canonical result", got)
			}
		})
	}
}

// insightCreatorStub は知見永続化のtest double。
type insightCreatorStub struct{ err error }

// CreateInsight は入力を保存結果へ変換する。
func (s insightCreatorStub) CreateInsight(_ context.Context, requestID string, evidences []domain.InsightEvidence, statement, conditions, gaps string) (domain.Insight, bool, error) {
	return domain.Insight{
		RequestID:               requestID,
		Evidences:               evidences,
		Statement:               statement,
		ApplicabilityConditions: conditions,
		VerificationGaps:        gaps,
	}, true, s.err
}

// insightEvidencePair は二根拠を返す。
func insightEvidencePair(firstExperiment, firstConclusion, secondExperiment, secondConclusion string) []domain.InsightEvidence {
	return []domain.InsightEvidence{
		{
			ExperimentID: firstExperiment,
			ConclusionID: firstConclusion,
		},
		{
			ExperimentID: secondExperiment,
			ConclusionID: secondConclusion,
		},
	}
}
