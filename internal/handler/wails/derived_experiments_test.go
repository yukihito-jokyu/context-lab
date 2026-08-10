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

func TestCreateDerivedExperimentsHandler(t *testing.T) {
	purpose := "派生目的"
	createdAt := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	tests := []struct {
		name      string
		command   *usecase.CreateDerivedExperiment
		wantCode  apperr.Code
		wantDebug bool
	}{
		{
			name: "成功",
			command: usecase.NewCreateDerivedExperiment(derivedExperimentHandlerCreator{
				result: domain.DerivedExperiment{
					RequestID:          "request-1",
					ExperimentID:       "derived-1",
					SourceExperimentID: "source-1",
					State:              "preparing",
					CreatedAt:          createdAt,
				},
			}),
			wantDebug: true,
		},
		{
			name: "既知エラー",
			command: usecase.NewCreateDerivedExperiment(derivedExperimentHandlerCreator{
				err: apperr.New(apperr.CodeDerivedExperimentSourceNotEligible),
			}),
			wantCode: apperr.CodeDerivedExperimentSourceNotEligible,
		},
		{
			name: "未知エラー",
			command: usecase.NewCreateDerivedExperiment(derivedExperimentHandlerCreator{
				err: errors.New("private sqlite"),
			}),
			wantCode: apperr.CodeDerivedExperimentUnavailable,
		},
		{
			name:     "依存欠落",
			wantCode: apperr.CodeDerivedExperimentUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appLogger := &derivedExperimentHandlerLogger{}
			got := NewCreateDerivedExperimentsHandler(tt.command, appLogger).CreateDerivedExperiment(CreateDerivedExperimentRequest{
				RequestID:          "request-1",
				SourceExperimentID: "source-1",
				Changes: domain.DerivedExperimentChanges{
					Purpose: &purpose,
				},
				Reason: "比較のため",
			})
			if tt.wantCode == "" {
				if got.Data == nil || got.Data.ExperimentID != "derived-1" || got.Data.CreatedAt != "2026-08-10T03:00:00Z" {
					t.Errorf("Data = %+v, want UTC derived experiment DTO", got.Data)
				}
				if tt.wantDebug && !appLogger.debugged {
					t.Error("Debug was not called")
				}
				return
			}
			if got.Error == nil || got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error = %+v, want %q", got.Error, tt.wantCode)
			}
			if appLogger.code != string(tt.wantCode) || appLogger.operation != "create_derived_experiment" {
				t.Errorf("safe error log = (%q, %q)", appLogger.code, appLogger.operation)
			}
		})
	}
}

func TestCreateDerivedExperimentHandlerFailureFallback(t *testing.T) {
	appLogger := &derivedExperimentHandlerLogger{}
	got := NewCreateDerivedExperimentsHandler(nil, appLogger).fail(context.Background(), errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected safe error", got.Error)
	}
	if appLogger.code != string(apperr.CodeUnexpected) {
		t.Errorf("ErrorCode = %q, want %q", appLogger.code, apperr.CodeUnexpected)
	}
}

type derivedExperimentHandlerCreator struct {
	result domain.DerivedExperiment
	err    error
}

func (c derivedExperimentHandlerCreator) CreateDerivedExperiment(context.Context, string, string, domain.DerivedExperimentChanges, string) (domain.DerivedExperiment, bool, error) {
	return c.result, true, c.err
}

type derivedExperimentHandlerLogger struct {
	debugged  bool
	code      string
	operation string
}

func (l *derivedExperimentHandlerLogger) Debug(context.Context, string, ...slog.Attr) {
	l.debugged = true
}
func (*derivedExperimentHandlerLogger) Info(context.Context, string, ...slog.Attr) {}
func (*derivedExperimentHandlerLogger) Warn(context.Context, string, ...slog.Attr) {}
func (*derivedExperimentHandlerLogger) Error(context.Context, string, error, ...slog.Attr) {
}
func (l *derivedExperimentHandlerLogger) ErrorCode(_ context.Context, _ string, code string, attributes ...slog.Attr) {
	l.code = code
	for _, attribute := range attributes {
		if attribute.Key == "operation" {
			l.operation = attribute.Value.String()
		}
	}
}
