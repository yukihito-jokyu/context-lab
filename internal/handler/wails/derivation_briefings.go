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

// DerivationBriefingsHandler は派生実験ブリーフ開始のWails binding。
type DerivationBriefingsHandler struct {
	command *usecase.StartDerivationBriefing
	logger  logger.Logger
}

// NewDerivationBriefingsHandler は派生実験ブリーフ開始bindingを生成する。
func NewDerivationBriefingsHandler(command *usecase.StartDerivationBriefing, appLogger logger.Logger) *DerivationBriefingsHandler {
	return &DerivationBriefingsHandler{command: command, logger: appLogger}
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

// fail は内部エラーを安全な画面エラーへ変換する。
func (h *DerivationBriefingsHandler) fail(ctx context.Context, err error) StartDerivationBriefingResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "start derivation briefing failed", string(appErr.Code), slog.String("operation", "start_derivation_briefing"))

	return StartDerivationBriefingResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}
