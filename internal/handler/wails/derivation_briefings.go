package wails

import (
	"context"
	"log/slog"

	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/logger"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// StartDerivationBriefingResponse は派生実験ブリーフ開始の成功または失敗結果。
type StartDerivationBriefingResponse struct {
	Data  *StartDerivationBriefingData `json:"data,omitempty"`
	Error *ErrorResponse               `json:"error,omitempty"`
}

// StartDerivationBriefingData は画面へ返す派生実験ブリーフ開始結果。
type StartDerivationBriefingData struct {
	BriefingSessionID  string `json:"briefingSessionId"`
	OperationID        string `json:"operationId"`
	SourceExperimentID string `json:"sourceExperimentId"`
}

// SendDerivationBriefMessageResponse は派生実験ブリーフ会話送信の成功または失敗結果。
type SendDerivationBriefMessageResponse struct {
	Data  *SendDerivationBriefMessageData `json:"data,omitempty"`
	Error *ErrorResponse                  `json:"error,omitempty"`
}

// SendDerivationBriefMessageData は画面へ返す送信操作識別子。
type SendDerivationBriefMessageData struct {
	OperationID string `json:"operationId"`
}

// DerivationBriefingsHandler は派生実験ブリーフ開始のWails binding。
type DerivationBriefingsHandler struct {
	command                    *usecase.StartDerivationBriefing
	sendDerivationBriefMessage *usecase.SendDerivationBriefMessage
	logger                     logger.Logger
}

// NewDerivationBriefingsHandler は派生実験ブリーフ開始bindingを生成する。
func NewDerivationBriefingsHandler(command *usecase.StartDerivationBriefing, appLogger logger.Logger, sendDerivationBriefMessage ...*usecase.SendDerivationBriefMessage) *DerivationBriefingsHandler {
	handler := &DerivationBriefingsHandler{command: command, logger: appLogger}
	if len(sendDerivationBriefMessage) > 0 {
		handler.sendDerivationBriefMessage = sendDerivationBriefMessage[0]
	}

	return handler
}

// StartDerivationBriefing は派生実験ブリーフ開始を画面向けDTOで返す。
func (h *DerivationBriefingsHandler) StartDerivationBriefing(requestID, sourceExperimentID string) StartDerivationBriefingResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "start derivation briefing called")
	if h.command == nil {
		return h.fail(ctx, apperr.New(apperr.CodeDerivationBriefingStartFailed))
	}

	start, err := h.command.Execute(ctx, requestID, sourceExperimentID)
	if err != nil {
		return h.fail(ctx, err)
	}

	return StartDerivationBriefingResponse{Data: &StartDerivationBriefingData{
		BriefingSessionID:  start.BriefingSessionID,
		OperationID:        start.OperationID,
		SourceExperimentID: start.SourceExperimentID,
	}}
}

// SendDerivationBriefMessage は派生実験ブリーフ会話送信を画面向けDTOで返す。
func (h *DerivationBriefingsHandler) SendDerivationBriefMessage(requestID, briefingSessionID, message string) SendDerivationBriefMessageResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "send derivation brief message called")
	if h.sendDerivationBriefMessage == nil {
		return h.failMessage(ctx, apperr.New(apperr.CodeDerivationBriefingMessageFailed))
	}

	operation, err := h.sendDerivationBriefMessage.Execute(ctx, requestID, briefingSessionID, message)
	if err != nil {
		return h.failMessage(ctx, err)
	}

	return SendDerivationBriefMessageResponse{Data: &SendDerivationBriefMessageData{OperationID: operation.OperationID}}
}

// fail は内部エラーを安全な画面エラーへ変換する。
func (h *DerivationBriefingsHandler) fail(ctx context.Context, err error) StartDerivationBriefingResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "start derivation briefing failed", string(appErr.Code), slog.String("operation", "start_derivation_briefing"))

	return StartDerivationBriefingResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// failMessage は内部エラーを安全な派生会話送信エラーへ変換する。
func (h *DerivationBriefingsHandler) failMessage(ctx context.Context, err error) SendDerivationBriefMessageResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "send derivation brief message failed", string(appErr.Code), slog.String("operation", "send_derivation_brief_message"))

	return SendDerivationBriefMessageResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}
