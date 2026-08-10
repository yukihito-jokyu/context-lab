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

// Wails評価詳細queryは画面に必要な安全な事実だけをUTCで返す。
func TestExperimentEvaluationDetailsHandlerGetEvaluationDetail(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	summary := "評価要約"
	tests := []struct {
		name     string
		id       string
		reader   handlerEvaluationDetailReader
		wantCode apperr.Code
	}{
		{
			name: "安全な詳細を返す",
			id:   "evaluation-1",
			reader: handlerEvaluationDetailReader{
				found: true,
				detail: domain.ExperimentEvaluationDetail{
					Evaluation: domain.ExperimentEvaluationFact{
						ID:           "evaluation-1",
						ExperimentID: "experiment-1",
						RunID:        "run-1",
						State:        "failed",
						Summary:      &summary,
						CreatedAt:    confirmedAt,
						UpdatedAt:    confirmedAt,
					},
					Operation: domain.ExperimentEvaluationOperation{
						ID:        "operation-1",
						State:     "failed",
						UpdatedAt: confirmedAt,
					},
					Evidence: domain.ExperimentEvaluationEvidence{
						RunSummary:     "run要約",
						EvaluationAxes: "正確性",
					},
					Result: domain.ExperimentEvaluationResult{
						Status:     "partial",
						Summary:    &summary,
						ReasonCode: "EVALUATION_TIMEOUT",
					},
					Failure: &domain.ExperimentEvaluationFailure{
						Code:       "EVALUATION_TIMEOUT",
						OccurredAt: confirmedAt,
					},
					Reconciliation: domain.ExperimentEvaluationReconciliation{
						State:          "confirmed",
						LastObservedAt: confirmedAt,
					},
					LastConfirmedAt: confirmedAt,
				},
			},
		},
		{
			name: "内部詳細を漏らさない",
			id:   "evaluation-1",
			reader: handlerEvaluationDetailReader{
				err: errors.New("private evaluator detail"),
			},
			wantCode: apperr.CodeEvaluationDetailUnavailable,
		},
		{
			name:     "空IDを安全に返す",
			reader:   handlerEvaluationDetailReader{},
			wantCode: apperr.CodeEvaluationDetailRequestInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExperimentEvaluationDetailsHandler(usecase.NewGetEvaluationDetail(&tt.reader), newTestLogger()).GetEvaluationDetail(tt.id)
			if tt.wantCode != "" {
				if got.Error == nil || got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error = %+v, want code %q", got.Error, tt.wantCode)
				}
				if got.Error != nil && strings.Contains(got.Error.Message, "evaluator") {
					t.Errorf("Error.Message = %q, want no private detail", got.Error.Message)
				}
				return
			}
			if got.Data == nil {
				t.Fatal("Data = nil, want evaluation detail")
			}
			if got.Data.Evaluation.ID != "evaluation-1" || got.Data.Result.Status != "partial" || got.Data.Failure == nil {
				t.Errorf("Data = %+v, want safe evaluation detail", got.Data)
			}
			if got.Data.LastConfirmedAt.Location() != time.UTC || got.Data.Failure.OccurredAt.Location() != time.UTC {
				t.Errorf("returned times = %+v, want UTC", got.Data)
			}
		})
	}
}

// 評価詳細queryは依存欠落を安全なエラーへ変換する。
func TestExperimentEvaluationDetailsHandlerFailureFallback(t *testing.T) {
	got := NewExperimentEvaluationDetailsHandler(nil, newTestLogger()).GetEvaluationDetail("evaluation-1")
	if got.Error == nil || got.Error.Code != string(apperr.CodeEvaluationDetailUnavailable) {
		t.Errorf("Error = %+v, want code %q", got.Error, apperr.CodeEvaluationDetailUnavailable)
	}
}

// 評価詳細未知エラーは安全な共通エラーへ変換する。
func TestFailGetEvaluationDetail(t *testing.T) {
	got := failGetEvaluationDetail(errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want code %q", got.Error, apperr.CodeUnexpected)
	}
}

// handlerEvaluationDetailReader は評価詳細readerのtest double。
type handlerEvaluationDetailReader struct {
	detail domain.ExperimentEvaluationDetail
	found  bool
	err    error
}

// GetEvaluationDetail は設定済み詳細を返却する。
func (r *handlerEvaluationDetailReader) GetEvaluationDetail(context.Context, string) (domain.ExperimentEvaluationDetail, bool, error) {
	return r.detail, r.found, r.err
}
