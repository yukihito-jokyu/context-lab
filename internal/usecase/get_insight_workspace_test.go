package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// GetInsightWorkspaceの成功、既知障害、未知障害を検証する。
func TestGetInsightWorkspaceExecute(t *testing.T) {
	tests := []struct {
		name     string
		reader   insightWorkspaceReader
		wantCode apperr.Code
	}{
		{
			name: "成功",
			reader: insightWorkspaceReader{workspace: domain.InsightWorkspace{
				EvidenceCandidates: []domain.InsightEvidenceCandidate{
					{ExperimentID: "experiment-1"},
				},
				LastConfirmedAt: timePointer(time.Now()),
			}},
		},
		{
			name:     "既知障害",
			reader:   insightWorkspaceReader{err: apperr.New(apperr.CodeInsightWorkspaceUnavailable)},
			wantCode: apperr.CodeInsightWorkspaceUnavailable,
		},
		{
			name:     "未知障害",
			reader:   insightWorkspaceReader{err: errors.New("sqlite failed")},
			wantCode: apperr.CodeInsightWorkspaceUnavailable,
		},
		{
			name:     "中止",
			reader:   insightWorkspaceReader{err: context.Canceled},
			wantCode: apperr.CodeOperationCanceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGetInsightWorkspace(tt.reader).Execute(context.Background())
			if tt.wantCode == "" {
				if err != nil {
					t.Errorf("Execute() error = %v, want nil", err)
				}
				if len(got.EvidenceCandidates) != 1 {
					t.Errorf("EvidenceCandidates = %+v, want one", got.EvidenceCandidates)
				}
				return
			}
			if !apperr.IsCode(err, tt.wantCode) {
				t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
			}
		})
	}
}

// insightWorkspaceReader は知見ワークスペース読取のtest double。
type insightWorkspaceReader struct {
	workspace domain.InsightWorkspace
	err       error
}

// GetInsightWorkspace は固定結果を返す。
func (r insightWorkspaceReader) GetInsightWorkspace(context.Context) (domain.InsightWorkspace, error) {
	return r.workspace, r.err
}

// timePointer は時刻ポインターを返す。
func timePointer(value time.Time) *time.Time {
	return &value
}
