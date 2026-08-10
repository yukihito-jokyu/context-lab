package wails

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// GetExperimentComparison handlerは安全な比較DTOを返す。
func TestExperimentComparisonsHandlerGetExperimentComparison(t *testing.T) {
	reader := comparisonHandlerReader{
		comparison: domain.ExperimentComparison{
			Experiment: domain.ExperimentComparisonExperiment{
				ID:             "experiment-1",
				Purpose:        "目的",
				EvaluationAxes: "軸",
			},
			Evaluations: []domain.ExperimentComparisonEvaluation{
				{
					EvaluationID: "evaluation-1",
					RunID:        "run-1",
					State:        "completed",
					RunSummary:   stringPointer("run要約"),
					Result: domain.ExperimentEvaluationResult{
						Status:     "complete",
						Summary:    stringPointer("評価要約"),
						ReasonCode: "",
					},
					Reconciliation: domain.ExperimentEvaluationReconciliation{
						State:          "confirmed",
						LastObservedAt: time.Now(),
					},
					UpdatedAt: time.Now(),
				},
			},
			LastConfirmedAt: time.Now(),
		},
	}
	got := NewExperimentComparisonsHandler(usecase.NewGetExperimentComparison(reader), newTestLogger()).GetExperimentComparison("experiment-1")
	if got.Data == nil || got.Data.Experiment.ID != "experiment-1" || len(got.Data.Evaluations) != 1 {
		t.Errorf("Data = %+v, want comparison DTO", got.Data)
	}
	if got.Data != nil && (got.Data.Evaluations[0].Result.Status != "complete" || got.Data.Evaluations[0].Reconciliation.LastObservedAt.Location() != time.UTC) {
		t.Errorf("Evaluations = %+v, want safe result and UTC times", got.Data.Evaluations)
	}
}

// GetExperimentComparison handlerは未知エラーを安全な共通エラーへ変換する。
func TestFailGetExperimentComparison(t *testing.T) {
	got := failGetExperimentComparison(errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want code %q", got.Error, apperr.CodeUnexpected)
	}
}

// GetExperimentComparison handlerは依存欠落を安全に返す。
func TestExperimentComparisonsHandlerFailureFallback(t *testing.T) {
	appLogger := &comparisonHandlerLogger{}
	got := NewExperimentComparisonsHandler(nil, appLogger).GetExperimentComparison("experiment-1")
	if got.Error == nil || got.Error.Code != string(apperr.CodeExperimentComparisonUnavailable) {
		t.Errorf("Error = %+v, want unavailable", got.Error)
	}
	if appLogger.errorCode != string(apperr.CodeExperimentComparisonUnavailable) || appLogger.operation != "get_experiment_comparison" {
		t.Errorf("ErrorCode log = code %q operation %q, want unavailable comparison operation", appLogger.errorCode, appLogger.operation)
	}
}

// usecase失敗は画面へ安全な比較取得失敗DTOとして返す。
func TestExperimentComparisonsHandlerUsecaseFailure(t *testing.T) {
	reader := comparisonHandlerReader{err: errors.New("private sqlite")}
	got := NewExperimentComparisonsHandler(usecase.NewGetExperimentComparison(reader), newTestLogger()).GetExperimentComparison("experiment-1")
	if got.Error == nil || got.Error.Code != string(apperr.CodeExperimentComparisonUnavailable) {
		t.Errorf("Error = %+v, want unavailable", got.Error)
	}
}

type comparisonHandlerReader struct {
	comparison domain.ExperimentComparison
	err        error
}

func (r comparisonHandlerReader) GetExperimentComparison(context.Context, string) (domain.ExperimentComparison, bool, error) {
	return r.comparison, true, r.err
}

func stringPointer(value string) *string { return &value }

type comparisonHandlerLogger struct {
	errorCode string
	operation string
}

func (*comparisonHandlerLogger) Debug(context.Context, string, ...slog.Attr) {}
func (*comparisonHandlerLogger) Info(context.Context, string, ...slog.Attr)  {}
func (*comparisonHandlerLogger) Warn(context.Context, string, ...slog.Attr)  {}
func (*comparisonHandlerLogger) Error(context.Context, string, error, ...slog.Attr) {
}
func (l *comparisonHandlerLogger) ErrorCode(_ context.Context, _ string, code string, attributes ...slog.Attr) {
	l.errorCode = code
	for _, attribute := range attributes {
		if attribute.Key == "operation" {
			l.operation = attribute.Value.String()
		}
	}
}
