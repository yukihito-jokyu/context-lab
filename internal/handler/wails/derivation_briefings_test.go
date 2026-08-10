package wails

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// DerivationBriefingsHandlerの成功と安全な失敗DTOを確認。
func TestDerivationBriefingsHandlerStartDerivationBriefing(t *testing.T) {
	tests := []struct {
		name     string
		command  *usecase.StartDerivationBriefing
		wantCode apperr.Code
	}{
		{
			name:    "成功",
			command: usecase.NewStartDerivationBriefing(derivationBriefingHandlerStore{}, derivationBriefingHandlerStarter{}),
		},
		{
			name:     "既知エラー",
			command:  usecase.NewStartDerivationBriefing(derivationBriefingHandlerStore{beginErr: apperr.New(apperr.CodeDerivedExperimentSourceNotEligible)}, derivationBriefingHandlerStarter{}),
			wantCode: apperr.CodeDerivedExperimentSourceNotEligible,
		},
		{
			name:     "未知エラー",
			command:  usecase.NewStartDerivationBriefing(derivationBriefingHandlerStore{beginErr: errors.New("private sqlite")}, derivationBriefingHandlerStarter{}),
			wantCode: apperr.CodeDerivationBriefingStartFailed,
		},
		{
			name:     "依存欠落",
			wantCode: apperr.CodeDerivationBriefingStartFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appLogger := &derivationBriefingHandlerLogger{}
			got := NewDerivationBriefingsHandler(tt.command, appLogger).StartDerivationBriefing("request-1", "source-1")
			if tt.wantCode == "" {
				if got.Data == nil || got.Data.SourceExperimentID != "source-1" || got.Data.BriefingSessionID != "session-1" {
					t.Errorf("Data = %+v, want derivation briefing identifiers", got.Data)
				}

				return
			}
			if got.Error == nil || got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error = %+v, want %q", got.Error, tt.wantCode)
			}
			if appLogger.code != string(tt.wantCode) || appLogger.operation != "start_derivation_briefing" {
				t.Errorf("safe error log = (%q, %q)", appLogger.code, appLogger.operation)
			}
		})
	}
}

// DerivationBriefingsHandlerの未知エラー変換を確認。
func TestDerivationBriefingsHandlerFailureFallback(t *testing.T) {
	appLogger := &derivationBriefingHandlerLogger{}
	got := NewDerivationBriefingsHandler(nil, appLogger).fail(context.Background(), errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected safe error", got.Error)
	}
	if appLogger.code != string(apperr.CodeUnexpected) {
		t.Errorf("ErrorCode = %q, want %q", appLogger.code, apperr.CodeUnexpected)
	}
}

// derivationBriefingHandlerStore はhandler用開始記録portのtest double。
type derivationBriefingHandlerStore struct{ beginErr error }

// BeginDerivationBriefing は指定済み開始結果を返す。
func (s derivationBriefingHandlerStore) BeginDerivationBriefing(context.Context, string, string) (domain.DerivationBriefingStart, bool, error) {
	return domain.DerivationBriefingStart{
		RequestID:          "request-1",
		SourceExperimentID: "source-1",
		BriefingSessionID:  "session-1",
		OperationID:        "operation-1",
		State:              domain.BriefingStartStateStarting,
	}, true, s.beginErr
}

// MarkDerivationBriefingStarted は開始済み状態を受理する。
func (derivationBriefingHandlerStore) MarkDerivationBriefingStarted(context.Context, string) error {
	return nil
}

// MarkDerivationBriefingFailed は失敗状態を受理する。
func (derivationBriefingHandlerStore) MarkDerivationBriefingFailed(context.Context, string, string) error {
	return nil
}

// derivationBriefingHandlerStarter はhandler用ACP開始portのtest double。
type derivationBriefingHandlerStarter struct{}

// StartExperimentBriefing は開始を受理する。
func (derivationBriefingHandlerStarter) StartExperimentBriefing(context.Context, string, string) error {
	return nil
}

// derivationBriefingHandlerLogger は安全なログを記録するtest double。
type derivationBriefingHandlerLogger struct {
	code      string
	operation string
}

// Debug はdebugログを受理する。
func (*derivationBriefingHandlerLogger) Debug(context.Context, string, ...slog.Attr) {}

// Info はinfoログを受理する。
func (*derivationBriefingHandlerLogger) Info(context.Context, string, ...slog.Attr) {}

// Warn はwarnログを受理する。
func (*derivationBriefingHandlerLogger) Warn(context.Context, string, ...slog.Attr) {}

// Error はerrorログを受理する。
func (*derivationBriefingHandlerLogger) Error(context.Context, string, error, ...slog.Attr) {}

// ErrorCode は安全なエラー情報を記録する。
func (l *derivationBriefingHandlerLogger) ErrorCode(_ context.Context, _ string, code string, attributes ...slog.Attr) {
	l.code = code
	for _, attribute := range attributes {
		if attribute.Key == "operation" {
			l.operation = attribute.Value.String()
		}
	}
}
