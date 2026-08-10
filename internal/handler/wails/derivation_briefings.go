package wails

import (
	"context"
	"log/slog"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
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

// GetDerivationBriefingResponse は派生実験ブリーフ再読込の成功または失敗結果。
type GetDerivationBriefingResponse struct {
	Data  *GetDerivationBriefingData `json:"data,omitempty"`
	Error *ErrorResponse             `json:"error,omitempty"`
}

// GetDerivationBriefingData は画面へ返す派生実験ブリーフ再読込結果。
type GetDerivationBriefingData struct {
	State            string                                `json:"state"`
	Messages         []DerivationBriefingMessageResponse   `json:"messages"`
	LatestSuggestion *DerivationBriefingSuggestionResponse `json:"latestSuggestion"`
	LastConfirmedAt  time.Time                             `json:"lastConfirmedAt"`
}

// DerivationBriefingMessageResponse は画面表示用の派生実験ブリーフ会話。
type DerivationBriefingMessageResponse struct {
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	SequenceNo int       `json:"sequenceNo"`
	CreatedAt  time.Time `json:"createdAt"`
}

// DerivationBriefingSuggestionResponse は画面表示用の派生実験提案版。
type DerivationBriefingSuggestionResponse struct {
	ID                    string    `json:"id"`
	VersionNo             int       `json:"versionNo"`
	Purpose               string    `json:"purpose"`
	Decision              string    `json:"decision"`
	Hypothesis            *string   `json:"hypothesis,omitempty"`
	CandidatePrompts      []string  `json:"candidatePrompts"`
	EvaluationCriteria    string    `json:"evaluationCriteria"`
	EnvironmentConditions string    `json:"environmentConditions"`
	InitialInput          string    `json:"initialInput"`
	SuccessCriteria       string    `json:"successCriteria"`
	RequiredConditions    string    `json:"requiredConditions"`
	OpenQuestion          *string   `json:"openQuestion,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

// DerivationBriefingsHandler は派生実験ブリーフ開始のWails binding。
type DerivationBriefingsHandler struct {
	command                    *usecase.StartDerivationBriefing
	sendDerivationBriefMessage *usecase.SendDerivationBriefMessage
	getDerivationBriefing      *usecase.GetDerivationBriefing
	logger                     logger.Logger
}

// NewDerivationBriefingsHandler は派生実験ブリーフbindingを生成する。
func NewDerivationBriefingsHandler(command *usecase.StartDerivationBriefing, sendDerivationBriefMessage *usecase.SendDerivationBriefMessage, getDerivationBriefing *usecase.GetDerivationBriefing, appLogger logger.Logger) *DerivationBriefingsHandler {
	return &DerivationBriefingsHandler{
		command:                    command,
		sendDerivationBriefMessage: sendDerivationBriefMessage,
		getDerivationBriefing:      getDerivationBriefing,
		logger:                     appLogger,
	}
}

// GetDerivationBriefing は派生実験ブリーフを画面向けDTOで返す。
func (h *DerivationBriefingsHandler) GetDerivationBriefing(briefingSessionID string) GetDerivationBriefingResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get derivation briefing called")
	if h.getDerivationBriefing == nil {
		return h.failGet(ctx, apperr.New(apperr.CodeDerivationBriefingUnavailable))
	}

	briefing, err := h.getDerivationBriefing.Execute(ctx, briefingSessionID)
	if err != nil {
		return h.failGet(ctx, err)
	}
	data := toGetDerivationBriefingData(briefing)

	return GetDerivationBriefingResponse{Data: &data}
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

// failGet は内部エラーを安全な派生ブリーフ取得エラーへ変換する。
func (h *DerivationBriefingsHandler) failGet(ctx context.Context, err error) GetDerivationBriefingResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "get derivation briefing failed", string(appErr.Code), slog.String("operation", "get_derivation_briefing"))

	return GetDerivationBriefingResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toGetDerivationBriefingData はdomain派生実験ブリーフを画面DTOへ変換する。
func toGetDerivationBriefingData(briefing domain.DerivationBriefing) GetDerivationBriefingData {
	messages := make([]DerivationBriefingMessageResponse, 0, len(briefing.Messages))
	for _, message := range briefing.Messages {
		messages = append(messages, DerivationBriefingMessageResponse{
			Role:       message.Role,
			Content:    message.Content,
			SequenceNo: message.SequenceNo,
			CreatedAt:  message.CreatedAt.UTC(),
		})
	}

	return GetDerivationBriefingData{
		State:            briefing.State,
		Messages:         messages,
		LatestSuggestion: toDerivationBriefingSuggestionResponse(briefing.LatestSuggestion),
		LastConfirmedAt:  briefing.LastConfirmedAt.UTC(),
	}
}

// toDerivationBriefingSuggestionResponse はdomain派生実験提案を画面DTOへ変換する。
func toDerivationBriefingSuggestionResponse(suggestion *domain.DerivationBriefingSuggestion) *DerivationBriefingSuggestionResponse {
	if suggestion == nil {
		return nil
	}

	return &DerivationBriefingSuggestionResponse{
		ID:                    suggestion.ID,
		VersionNo:             suggestion.VersionNo,
		Purpose:               suggestion.Purpose,
		Decision:              suggestion.Decision,
		Hypothesis:            suggestion.Hypothesis,
		CandidatePrompts:      suggestion.CandidatePrompts,
		EvaluationCriteria:    suggestion.EvaluationCriteria,
		EnvironmentConditions: suggestion.EnvironmentConditions,
		InitialInput:          suggestion.InitialInput,
		SuccessCriteria:       suggestion.SuccessCriteria,
		RequiredConditions:    suggestion.RequiredConditions,
		OpenQuestion:          suggestion.OpenQuestion,
		CreatedAt:             suggestion.CreatedAt.UTC(),
	}
}
