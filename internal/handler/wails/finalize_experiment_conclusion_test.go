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

// 結論確定handlerの成功と安全な失敗DTOを検証する。
func TestFinalizeExperimentConclusionsHandler(t *testing.T) {
	tests := []struct {
		name      string
		finalizer *usecase.FinalizeExperimentConclusion
		wantCode  apperr.Code
	}{
		{
			name: "成功",
			finalizer: usecase.NewFinalizeExperimentConclusion(&handlerConclusionStore{
				result: domain.ExperimentConclusion{
					ExperimentID: "experiment-1",
					ConclusionID: "conclusion-1",
					Conclusion:   "結論",
					State:        "finalized",
					FinalizedAt:  time.Now(),
				},
			}),
		},
		{
			name: "既知エラー",
			finalizer: usecase.NewFinalizeExperimentConclusion(&handlerConclusionStore{
				err: apperr.New(apperr.CodeExperimentConclusionNotReady),
			}),
			wantCode: apperr.CodeExperimentConclusionNotReady,
		},
		{
			name: "未知エラー",
			finalizer: usecase.NewFinalizeExperimentConclusion(&handlerConclusionStore{
				err: errors.New("private sqlite"),
			}),
			wantCode: apperr.CodeExperimentConclusionUnavailable,
		},
		{
			name:     "依存欠落",
			wantCode: apperr.CodeExperimentConclusionUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appLogger := &handlerConclusionLogger{}
			got := NewFinalizeExperimentConclusionsHandler(tt.finalizer, appLogger).FinalizeExperimentConclusion(FinalizeExperimentConclusionRequest{
				RequestID:    "request-1",
				ExperimentID: "experiment-1",
				Conclusion:   "結論",
			})
			if tt.wantCode == "" {
				if got.Data == nil || got.Data.ConclusionID != "conclusion-1" {
					t.Errorf("Data = %+v, want conclusion DTO", got.Data)
				}
				return
			}
			if got.Error == nil || got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error = %+v, want %q", got.Error, tt.wantCode)
			}
			if appLogger.code != string(tt.wantCode) || appLogger.operation != "finalize_experiment_conclusion" {
				t.Errorf("log = (%q, %q), want safe error log", appLogger.code, appLogger.operation)
			}
		})
	}
}

// failFinalizeExperimentConclusionの未知エラー変換を検証する。
func TestFailFinalizeExperimentConclusion(t *testing.T) {
	got := failFinalizeExperimentConclusion(errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected", got.Error)
	}
}

// handlerConclusionStore は結論確定portのtest double。
type handlerConclusionStore struct {
	result domain.ExperimentConclusion
	err    error
}

func (s *handlerConclusionStore) FinalizeExperimentConclusion(context.Context, string, string, string) (domain.ExperimentConclusion, bool, error) {
	return s.result, true, s.err
}

type handlerConclusionLogger struct{ code, operation string }

func (*handlerConclusionLogger) Debug(context.Context, string, ...slog.Attr)        {}
func (*handlerConclusionLogger) Info(context.Context, string, ...slog.Attr)         {}
func (*handlerConclusionLogger) Warn(context.Context, string, ...slog.Attr)         {}
func (*handlerConclusionLogger) Error(context.Context, string, error, ...slog.Attr) {}
func (l *handlerConclusionLogger) ErrorCode(_ context.Context, _ string, code string, attrs ...slog.Attr) {
	l.code = code
	for _, attr := range attrs {
		if attr.Key == "operation" {
			l.operation = attr.Value.String()
		}
	}
}
