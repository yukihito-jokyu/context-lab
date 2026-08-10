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
			got := NewDerivationBriefingsHandler(tt.command, nil, nil, nil, appLogger).StartDerivationBriefing("request-1", "source-1")
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
	got := NewDerivationBriefingsHandler(nil, nil, nil, nil, appLogger).fail(context.Background(), errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected safe error", got.Error)
	}
	if appLogger.code != string(apperr.CodeUnexpected) {
		t.Errorf("ErrorCode = %q, want %q", appLogger.code, apperr.CodeUnexpected)
	}
}

// DerivationBriefingsHandlerの派生会話送信を確認。
func TestDerivationBriefingsHandlerSendDerivationBriefMessage(t *testing.T) {
	tests := []struct {
		name     string
		command  *usecase.SendDerivationBriefMessage
		wantCode apperr.Code
	}{
		{
			name: "成功",
			command: usecase.NewSendDerivationBriefMessage(
				&derivationBriefingMessageHandlerStore{},
				derivationBriefingMessageHandlerSender{},
			),
		},
		{
			name: "安全な失敗",
			command: usecase.NewSendDerivationBriefMessage(
				&derivationBriefingMessageHandlerStore{beginErr: apperr.New(apperr.CodeDerivationBriefingMessageNotActive)},
				derivationBriefingMessageHandlerSender{},
			),
			wantCode: apperr.CodeDerivationBriefingMessageNotActive,
		},
		{
			name:     "依存欠落",
			wantCode: apperr.CodeDerivationBriefingMessageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appLogger := &derivationBriefingHandlerLogger{}
			handler := NewDerivationBriefingsHandler(nil, tt.command, nil, nil, appLogger)
			got := handler.SendDerivationBriefMessage("request-1", "session-1", "比較したい")
			if tt.wantCode == "" {
				if got.Data == nil || got.Data.OperationID == "" {
					t.Errorf("Data = %+v, want operation ID", got.Data)
				}

				return
			}
			if got.Error == nil || got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error = %+v, want %q", got.Error, tt.wantCode)
			}
			if appLogger.code != string(tt.wantCode) || appLogger.operation != "send_derivation_brief_message" {
				t.Errorf("safe error log = (%q, %q)", appLogger.code, appLogger.operation)
			}
		})
	}
}

// DerivationBriefingsHandlerの派生会話送信未知エラー変換を確認。
func TestDerivationBriefingsHandlerMessageFailureFallback(t *testing.T) {
	appLogger := &derivationBriefingHandlerLogger{}
	got := NewDerivationBriefingsHandler(nil, nil, nil, nil, appLogger).failMessage(context.Background(), errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected safe error", got.Error)
	}
	if appLogger.code != string(apperr.CodeUnexpected) {
		t.Errorf("ErrorCode = %q, want %q", appLogger.code, apperr.CodeUnexpected)
	}
}

// DerivationBriefingsHandlerの派生会話再読込を確認。
func TestDerivationBriefingsHandlerGetDerivationBriefing(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	tests := []struct {
		name     string
		reader   derivationBriefingHandlerReader
		session  string
		wantCode apperr.Code
		wantData bool
	}{
		{
			name:    "会話と最新版提案を画面DTOへ変換する",
			session: "session-1",
			reader: derivationBriefingHandlerReader{
				found: true,
				briefing: domain.DerivationBriefing{
					State: "started",
					Messages: []domain.DerivationBriefingMessage{{
						Role:       "user",
						Content:    "比較する",
						SequenceNo: 1,
						CreatedAt:  confirmedAt,
					}},
					LatestSuggestion: &domain.DerivationBriefingSuggestion{
						ID:        "suggestion-1",
						VersionNo: 2,
						CreatedAt: confirmedAt,
					},
					LastConfirmedAt: confirmedAt,
				},
			},
			wantData: true,
		},
		{
			name:     "未知sessionを安全なコードへ変換する",
			session:  "missing",
			wantCode: apperr.CodeDerivationBriefingNotFound,
		},
		{
			name: "内部エラーを安全なコードへ変換する",
			reader: derivationBriefingHandlerReader{
				err: errors.New("private sqlite"),
			},
			session:  "session-2",
			wantCode: apperr.CodeDerivationBriefingUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appLogger := &derivationBriefingHandlerLogger{}
			handler := NewDerivationBriefingsHandler(nil, nil, usecase.NewGetDerivationBriefing(tt.reader), nil, appLogger)
			got := handler.GetDerivationBriefing(tt.session)
			if gotData := got.Data != nil; gotData != tt.wantData {
				t.Fatalf("Data available = %v, want %v", gotData, tt.wantData)
			}
			if tt.wantData {
				if got := got.Data.Messages; len(got) != 1 || got[0].Content != "比較する" {
					t.Errorf("Messages = %+v, want one message", got)
				}
				if got := got.Data.LatestSuggestion; got == nil || got.ID != "suggestion-1" || got.VersionNo != 2 {
					t.Errorf("LatestSuggestion = %+v, want suggestion-1 version 2", got)
				}
				if got := got.Data.LastConfirmedAt; !got.Equal(confirmedAt.UTC()) || got.Location() != time.UTC {
					t.Errorf("LastConfirmedAt = %s, want UTC %s", got, confirmedAt.UTC())
				}

				return
			}
			if got.Error == nil || got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error = %+v, want %q", got.Error, tt.wantCode)
			}
			if appLogger.code != string(tt.wantCode) || appLogger.operation != "get_derivation_briefing" {
				t.Errorf("safe error log = (%q, %q)", appLogger.code, appLogger.operation)
			}
		})
	}
}

// DerivationBriefingsHandlerの派生会話再読込依存欠落を確認。
func TestDerivationBriefingsHandlerGetDerivationBriefingMissingDependency(t *testing.T) {
	appLogger := &derivationBriefingHandlerLogger{}
	got := NewDerivationBriefingsHandler(nil, nil, nil, nil, appLogger).GetDerivationBriefing("session-1")
	if got.Error == nil || got.Error.Code != string(apperr.CodeDerivationBriefingUnavailable) {
		t.Errorf("Error = %+v, want %q", got.Error, apperr.CodeDerivationBriefingUnavailable)
	}
}

// DerivationBriefingsHandlerの派生壁打ち終了DTOを確認。
func TestDerivationBriefingsHandlerStopDerivationBriefing(t *testing.T) {
	tests := []struct {
		name       string
		stopperErr error
		wantCode   apperr.Code
	}{
		{
			name: "終了操作識別子だけを返す",
		},
		{
			name:       "内部エラーを安全なコードへ変換する",
			stopperErr: errors.New("private ACP credential"),
			wantCode:   apperr.CodeDerivationBriefingStopFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appLogger := &derivationBriefingHandlerLogger{}
			handler := NewDerivationBriefingsHandler(nil, nil, nil, usecase.NewStopDerivationBriefing(&derivationBriefingStopHandlerStore{}, derivationBriefingStopHandler{err: tt.stopperErr}), appLogger)
			got := handler.StopDerivationBriefing("request-1", "session-1")
			if tt.wantCode != "" {
				if got.Error == nil || got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error = %+v, want %q", got.Error, tt.wantCode)
				}
				if got.Data != nil {
					t.Errorf("Data = %+v, want nil", got.Data)
				}
				if appLogger.code != string(tt.wantCode) || appLogger.operation != "stop_derivation_briefing" {
					t.Errorf("safe error log = (%q, %q)", appLogger.code, appLogger.operation)
				}

				return
			}
			if got.Error != nil {
				t.Fatalf("Error = %+v, want nil", got.Error)
			}
			if got.Data == nil || got.Data.OperationID != "stop-operation-1" {
				t.Errorf("Data = %+v, want stop operation identifier", got.Data)
			}
		})
	}
}

// DerivationBriefingsHandlerの終了依存欠落と未知エラー変換を確認。
func TestDerivationBriefingsHandlerStopDerivationBriefingFailureFallback(t *testing.T) {
	appLogger := &derivationBriefingHandlerLogger{}
	handler := NewDerivationBriefingsHandler(nil, nil, nil, nil, appLogger)
	got := handler.StopDerivationBriefing("request-1", "session-1")
	if got.Error == nil || got.Error.Code != string(apperr.CodeDerivationBriefingStopFailed) {
		t.Errorf("missing dependency Error = %+v, want %q", got.Error, apperr.CodeDerivationBriefingStopFailed)
	}
	got = handler.failStop(context.Background(), errors.New("private ACP credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("fallback Error = %+v, want %q", got.Error, apperr.CodeUnexpected)
	}
}

// 派生実験提案DTOの空状態変換を確認。
func TestToDerivationBriefingSuggestionResponseNil(t *testing.T) {
	if got := toDerivationBriefingSuggestionResponse(nil); got != nil {
		t.Errorf("toDerivationBriefingSuggestionResponse(nil) = %+v, want nil", got)
	}
}

// derivationBriefingHandlerStore はhandler用開始記録portのtest double。
type derivationBriefingHandlerStore struct{ beginErr error }

// derivationBriefingMessageHandlerStore はhandler用会話送信記録portのtest double。
type derivationBriefingMessageHandlerStore struct{ beginErr error }

// derivationBriefingStopHandlerStore はhandler用終了記録portのtest double。
type derivationBriefingStopHandlerStore struct{}

// BeginStopDerivationBriefing は開始中の終了操作を返す。
func (*derivationBriefingStopHandlerStore) BeginStopDerivationBriefing(_ context.Context, requestID, briefingSessionID string) (domain.DerivationBriefingStopOperation, bool, error) {
	return domain.DerivationBriefingStopOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "stop-operation-1",
		State:             domain.BriefingStartStateStarting,
	}, true, nil
}

// CompleteStopDerivationBriefing は終了完了を受理する。
func (*derivationBriefingStopHandlerStore) CompleteStopDerivationBriefing(context.Context, string) error {
	return nil
}

// FailStopDerivationBriefing は終了失敗を受理する。
func (*derivationBriefingStopHandlerStore) FailStopDerivationBriefing(context.Context, string, string) error {
	return nil
}

// derivationBriefingStopHandler はhandler用ACP停止portのtest double。
type derivationBriefingStopHandler struct{ err error }

// StopDerivationBriefing は指定済み停止結果を返す。
func (s derivationBriefingStopHandler) StopDerivationBriefing(context.Context, string, string) error {
	return s.err
}

// derivationBriefingHandlerReader はhandler用派生実験ブリーフ読み出しportのtest double。
type derivationBriefingHandlerReader struct {
	briefing domain.DerivationBriefing
	found    bool
	err      error
}

// GetDerivationBriefing は指定済み派生実験ブリーフを返す。
func (r derivationBriefingHandlerReader) GetDerivationBriefing(context.Context, string) (domain.DerivationBriefing, bool, error) {
	return r.briefing, r.found, r.err
}

// BeginDerivationBriefMessage は指定済み送信開始結果を返す。
func (s *derivationBriefingMessageHandlerStore) BeginDerivationBriefMessage(context.Context, string, string) (domain.DerivationBriefingMessageOperation, bool, error) {
	return domain.DerivationBriefingMessageOperation{
		OperationID: "message-operation",
		State:       domain.BriefingStartStateStarting,
	}, true, s.beginErr
}

// CompleteDerivationBriefMessage は送信完了を受理する。
func (*derivationBriefingMessageHandlerStore) CompleteDerivationBriefMessage(context.Context, string, string, domain.DerivationBriefingMessageResult) error {
	return nil
}

// FailDerivationBriefMessage は送信失敗を受理する。
func (*derivationBriefingMessageHandlerStore) FailDerivationBriefMessage(context.Context, string, string) error {
	return nil
}

// derivationBriefingMessageHandlerSender はhandler用ACP送信portのtest double。
type derivationBriefingMessageHandlerSender struct{}

// SendDerivationBriefMessage は送信成功を返す。
func (derivationBriefingMessageHandlerSender) SendDerivationBriefMessage(context.Context, string, string, string) (domain.DerivationBriefingMessageResult, error) {
	return domain.DerivationBriefingMessageResult{}, nil
}

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
