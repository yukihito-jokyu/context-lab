package wails

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// Wails run詳細queryの成功と安全な失敗返却。
func TestExperimentRunDetailsHandlerGetRunDetail(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	tests := []struct {
		name     string
		runID    string
		reader   handlerRunDetailReader
		wantCode apperr.Code
	}{
		{
			name:  "安全な詳細を返す",
			runID: "run-1",
			reader: handlerRunDetailReader{
				found: true,
				detail: domain.ExperimentRunDetail{
					Run: domain.ExperimentRunFact{
						ID:           "run-1",
						ExperimentID: "experiment-1",
						State:        "completed",
						CreatedAt:    confirmedAt,
						UpdatedAt:    confirmedAt,
					},
					FixedPrompt: domain.ExperimentPreparationPrompt{
						SequenceNo: 1,
						Content:    "prompt",
					},
					Artifacts: domain.ExperimentRunArtifacts{
						Status:     "partial",
						ReasonCode: "ARTIFACTS_PARTIAL",
						Items: []domain.ExperimentRunArtifact{
							{
								Digest: "digest-1",
								Label:  handlerRunDetailString("差分"),
								Status: "available",
							},
						},
					},
					Observations: []domain.ExperimentRunObservation{
						{
							SequenceNo: 1,
							Kind:       "output",
							OccurredAt: confirmedAt,
							Summary:    "観測",
						},
					},
					Failure: &domain.ExperimentRunFailure{
						Code:           "OPERATION_TIMEOUT",
						OccurredAt:     confirmedAt,
						PartialSummary: handlerRunDetailString("部分結果"),
					},
					Reconciliation: domain.ExperimentRunReconciliation{
						State:          "confirmed",
						LastObservedAt: confirmedAt,
					},
					LastConfirmedAt: confirmedAt,
				},
			},
		},
		{
			name:  "内部詳細を漏らさない",
			runID: "run-1",
			reader: handlerRunDetailReader{
				err: errors.New("private docker ID"),
			},
			wantCode: apperr.CodeRunDetailUnavailable,
		},
		{
			name:     "空IDを安全に返す",
			wantCode: apperr.CodeRunDetailRequestInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExperimentRunDetailsHandler(usecase.NewGetRunDetail(tt.reader), newTestLogger()).GetRunDetail(tt.runID)
			if tt.wantCode != "" {
				if got.Error == nil {
					t.Fatal("Error = nil, want safe error")
				}
				if got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
				}
				if strings.Contains(got.Error.Message, "docker") {
					t.Errorf("Error.Message = %q, want no private detail", got.Error.Message)
				}

				return
			}
			if got.Data == nil {
				t.Fatal("Data = nil, want run detail")
			}
			if got.Data.Run.ID != "run-1" || got.Data.Artifacts.Status != "partial" || got.Data.Failure == nil || len(got.Data.Observations) != 1 {
				t.Errorf("Data = %+v, want safe run detail", got.Data)
			}
			if got.Data.LastConfirmedAt.Location() != time.UTC {
				t.Errorf("LastConfirmedAt location = %s, want UTC", got.Data.LastConfirmedAt.Location())
			}
		})
	}
}

// Wails run詳細queryの依存欠落。
func TestExperimentRunDetailsHandlerFailureFallback(t *testing.T) {
	got := NewExperimentRunDetailsHandler(nil, newTestLogger()).GetRunDetail("run-1")
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeRunDetailUnavailable) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeRunDetailUnavailable)
	}
}

// run詳細未知エラーの安全な変換。
func TestFailGetRunDetail(t *testing.T) {
	got := failGetRunDetail(errors.New("private credential"))
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeUnexpected)
	}
}

// handlerRunDetailString は文字列pointerを返却。
func handlerRunDetailString(value string) *string {
	return &value
}

// handlerRunDetailReader はrun詳細readerのtest double。
type handlerRunDetailReader struct {
	detail domain.ExperimentRunDetail
	found  bool
	err    error
}

// GetRunDetail は設定済み詳細を返却。
func (r handlerRunDetailReader) GetRunDetail(context.Context, string) (domain.ExperimentRunDetail, bool, error) {
	return r.detail, r.found, r.err
}
