package wails

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// GetInsightWorkspace handlerの安全DTOと失敗契約を検証する。
func TestInsightsHandlerGetInsightWorkspace(t *testing.T) {
	tests := []struct {
		name     string
		command  *usecase.GetInsightWorkspace
		wantCode apperr.Code
	}{
		{
			name: "成功",
			command: usecase.NewGetInsightWorkspace(insightHandlerReader{workspace: domain.InsightWorkspace{
				EvidenceCandidates: []domain.InsightEvidenceCandidate{
					{
						ExperimentID: "experiment-1",
						FinalizedAt:  time.Now(),
					},
				},
				SavedConsiderations: []domain.InsightSavedConsideration{
					{
						ExperimentID: "experiment-1",
						FinalizedAt:  time.Now(),
					},
				},
				Insights: []domain.InsightSummary{
					{
						ID:            "insight-1",
						Statement:     "知見",
						EvidenceCount: 2,
						CreatedAt:     time.Now(),
					},
				},
				LastConfirmedAt: insightTimePointer(time.Now()),
			}}),
		},
		{
			name: "未知障害",
			command: usecase.NewGetInsightWorkspace(insightHandlerReader{
				err: errors.New("private sqlite"),
			}),
			wantCode: apperr.CodeInsightWorkspaceUnavailable,
		},
		{
			name:     "依存欠落",
			wantCode: apperr.CodeInsightWorkspaceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInsightsHandler(tt.command, newTestLogger()).GetInsightWorkspace()
			if tt.wantCode == "" {
				if got.Data == nil || len(got.Data.EvidenceCandidates) != 1 || got.Data.EvidenceCandidates[0].FinalizedAt.Location() != time.UTC {
					t.Errorf("Data = %+v, want UTC workspace DTO", got.Data)
				}
				return
			}
			if got.Error == nil || got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error = %+v, want code %q", got.Error, tt.wantCode)
			}
		})
	}
}

// InsightsHandlerの未知エラー変換を検証する。
func TestInsightsHandlerFailureFallback(t *testing.T) {
	got := NewInsightsHandler(nil, newTestLogger()).fail(context.Background(), errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected error", got.Error)
	}
}

// CreateInsight handlerのDTO変換と失敗契約を検証する。
func TestInsightsHandlerCreateInsight(t *testing.T) {
	tests := []struct {
		name     string
		command  *usecase.CreateInsight
		wantCode apperr.Code
	}{
		{
			name:    "成功",
			command: usecase.NewCreateInsight(insightCreateHandlerStub{}),
		},
		{
			name:     "障害",
			command:  usecase.NewCreateInsight(insightCreateHandlerStub{err: errors.New("sqlite")}),
			wantCode: apperr.CodeInsightCreateUnavailable,
		},
		{
			name:     "依存欠落",
			wantCode: apperr.CodeInsightCreateUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInsightsHandlerWithCreate(nil, tt.command, newTestLogger()).CreateInsight(CreateInsightRequest{
				RequestID: "request",
				Evidences: []CreateInsightEvidenceRequest{
					{
						ExperimentID: "experiment-a",
						ConclusionID: "conclusion-a",
					},
					{
						ExperimentID: "experiment-b",
						ConclusionID: "conclusion-b",
					},
				},
				Statement:               "statement",
				ApplicabilityConditions: "conditions",
				VerificationGaps:        "gaps",
			})
			if tt.wantCode != "" {
				if got.Error == nil || got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error = %+v, want code %q", got.Error, tt.wantCode)
				}
				return
			}
			if got.Data == nil || len(got.Data.Evidences) != 2 || got.Data.Evidences[0].ExperimentID != "experiment-a" {
				t.Errorf("Data = %+v, want safe insight DTO", got.Data)
			}
		})
	}
}

// CreateInsight handlerの未知エラー変換を検証する。
func TestInsightsHandlerCreateInsightFallback(t *testing.T) {
	got := NewInsightsHandlerWithCreate(nil, nil, newTestLogger()).failCreate(context.Background(), errors.New("private"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected", got.Error)
	}
}

// insightCreateHandlerStub は知見作成handler用のtest double。
type insightCreateHandlerStub struct{ err error }

// CreateInsight は固定の知見結果を返す。
func (s insightCreateHandlerStub) CreateInsight(_ context.Context, requestID string, evidences []domain.InsightEvidence, statement, conditions, gaps string) (domain.Insight, bool, error) {
	return domain.Insight{
		RequestID:               requestID,
		InsightID:               "insight",
		Evidences:               evidences,
		Statement:               statement,
		ApplicabilityConditions: conditions,
		VerificationGaps:        gaps,
		CreatedAt:               time.Now(),
	}, true, s.err
}

// insightHandlerReader は知見handler用のtest double。
type insightHandlerReader struct {
	workspace domain.InsightWorkspace
	err       error
}

// GetInsightWorkspace は固定結果を返す。
func (r insightHandlerReader) GetInsightWorkspace(context.Context) (domain.InsightWorkspace, error) {
	return r.workspace, r.err
}

// insightTimePointer は時刻ポインターを返す。
func insightTimePointer(value time.Time) *time.Time {
	return &value
}
