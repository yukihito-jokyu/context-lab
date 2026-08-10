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

// StartExperimentBriefingResponse は実験ブリーフ開始の成功または失敗結果。
type StartExperimentBriefingResponse struct {
	Data  *StartExperimentBriefingData `json:"data,omitempty"`
	Error *ErrorResponse               `json:"error,omitempty"`
}

// StartExperimentBriefingData は画面へ返す開始識別子。
type StartExperimentBriefingData struct {
	BriefingSessionID string `json:"briefingSessionId"`
	OperationID       string `json:"operationId"`
}

// SendExperimentBriefMessageResponse は実験ブリーフ会話送信の成功または失敗結果。
type SendExperimentBriefMessageResponse struct {
	Data  *SendExperimentBriefMessageData `json:"data,omitempty"`
	Error *ErrorResponse                  `json:"error,omitempty"`
}

// SendExperimentBriefMessageData は画面へ返す送信操作識別子。
type SendExperimentBriefMessageData struct {
	OperationID string `json:"operationId"`
}

// StopExperimentBriefingResponse は実験ブリーフ終了の成功または失敗結果。
type StopExperimentBriefingResponse struct {
	Data  *StopExperimentBriefingData `json:"data,omitempty"`
	Error *ErrorResponse              `json:"error,omitempty"`
}

// StopExperimentBriefingData は画面へ返す終了操作識別子。
type StopExperimentBriefingData struct {
	OperationID string `json:"operationId"`
}

// CreateExperimentFromBriefResponse はブリーフ採用の成功または失敗結果。
type CreateExperimentFromBriefResponse struct {
	Data  *CreateExperimentFromBriefData `json:"data,omitempty"`
	Error *ErrorResponse                 `json:"error,omitempty"`
}

// CreateExperimentFromBriefData は画面へ返す準備中実験の識別子と状態。
type CreateExperimentFromBriefData struct {
	ExperimentID string `json:"experimentId"`
	State        string `json:"state"`
}

// GetExperimentPreparationResponse は実験準備queryの成功または失敗結果。
type GetExperimentPreparationResponse struct {
	Data  *GetExperimentPreparationData `json:"data,omitempty"`
	Error *ErrorResponse                `json:"error,omitempty"`
}

// GetExperimentPreparationData は画面へ返す準備中実験の編集条件。
type GetExperimentPreparationData struct {
	ExperimentID          string                                      `json:"experimentId"`
	State                 string                                      `json:"state"`
	Purpose               string                                      `json:"purpose"`
	Hypothesis            *string                                     `json:"hypothesis,omitempty"`
	EnvironmentConditions string                                      `json:"environmentConditions"`
	InitialInput          string                                      `json:"initialInput"`
	Prompts               []ExperimentPreparationPromptResponse       `json:"prompts"`
	EvaluationAxes        string                                      `json:"evaluationAxes"`
	Source                ExperimentPreparationSourceResponse         `json:"source"`
	RequiredFields        ExperimentPreparationRequiredFieldsResponse `json:"requiredFields"`
	LastConfirmedAt       time.Time                                   `json:"lastConfirmedAt"`
}

// ExperimentPreparationPromptResponse は画面表示用prompt。
type ExperimentPreparationPromptResponse struct {
	SequenceNo int    `json:"sequenceNo"`
	Content    string `json:"content"`
}

// SaveExperimentPreparationDraftRequest は下書き保存のフォーム値。
type SaveExperimentPreparationDraftRequest struct {
	RequestID             string   `json:"requestId"`
	ExperimentID          string   `json:"experimentId"`
	Purpose               string   `json:"purpose"`
	Hypothesis            *string  `json:"hypothesis,omitempty"`
	EnvironmentConditions string   `json:"environmentConditions"`
	InitialInput          string   `json:"initialInput"`
	Prompts               []string `json:"prompts"`
	EvaluationAxes        string   `json:"evaluationAxes"`
}

// SaveExperimentPreparationDraftResponse は下書き保存の成功または失敗結果。
type SaveExperimentPreparationDraftResponse struct {
	Data  *SaveExperimentPreparationDraftData `json:"data,omitempty"`
	Error *ErrorResponse                      `json:"error,omitempty"`
}

// SaveExperimentPreparationDraftData は保存済み下書きの画面DTO。
type SaveExperimentPreparationDraftData struct {
	ExperimentID          string                                `json:"experimentId"`
	State                 string                                `json:"state"`
	Purpose               string                                `json:"purpose"`
	Hypothesis            *string                               `json:"hypothesis,omitempty"`
	EnvironmentConditions string                                `json:"environmentConditions"`
	InitialInput          string                                `json:"initialInput"`
	Prompts               []ExperimentPreparationPromptResponse `json:"prompts"`
	EvaluationAxes        string                                `json:"evaluationAxes"`
	SavedAt               time.Time                             `json:"savedAt"`
}

// ExperimentPreparationSourceResponse は採用元の安全な表示用情報。
type ExperimentPreparationSourceResponse struct {
	State     string `json:"state"`
	VersionID string `json:"versionId"`
}

// ExperimentPreparationRequiredFieldsResponse は必須入力の充足状態。
type ExperimentPreparationRequiredFieldsResponse struct {
	Purpose               bool `json:"purpose"`
	EnvironmentConditions bool `json:"environmentConditions"`
	InitialInput          bool `json:"initialInput"`
	Prompts               bool `json:"prompts"`
	EvaluationAxes        bool `json:"evaluationAxes"`
}

// ExperimentPreparationsHandler は実験準備queryのWails binding。
type ExperimentPreparationsHandler struct {
	getExperimentPreparation       *usecase.GetExperimentPreparation
	saveExperimentPreparationDraft *usecase.SaveExperimentPreparationDraft
	logger                         logger.Logger
}

// NewExperimentPreparationsHandler は実験準備bindingを生成。
func NewExperimentPreparationsHandler(getExperimentPreparation *usecase.GetExperimentPreparation, appLogger logger.Logger, saveExperimentPreparationDraft ...*usecase.SaveExperimentPreparationDraft) *ExperimentPreparationsHandler {
	handler := &ExperimentPreparationsHandler{getExperimentPreparation: getExperimentPreparation, logger: appLogger}
	if len(saveExperimentPreparationDraft) != 0 {
		handler.saveExperimentPreparationDraft = saveExperimentPreparationDraft[0]
	}

	return handler
}

// SaveExperimentPreparationDraft はフォーム下書きを画面向けDTOで保存。
func (h *ExperimentPreparationsHandler) SaveExperimentPreparationDraft(request SaveExperimentPreparationDraftRequest) SaveExperimentPreparationDraftResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "save experiment preparation draft called")

	if h.saveExperimentPreparationDraft == nil {
		response := failSaveExperimentPreparationDraft(apperr.New(apperr.CodeDraftSaveFailed))
		h.logger.ErrorCode(ctx, "save experiment preparation draft failed", response.Error.Code, slog.String("operation", "save_experiment_preparation_draft"))

		return response
	}
	prompts := make([]domain.ExperimentPreparationPrompt, 0, len(request.Prompts))
	for index, content := range request.Prompts {
		prompts = append(prompts, domain.ExperimentPreparationPrompt{SequenceNo: index + 1, Content: content})
	}
	draft, err := h.saveExperimentPreparationDraft.Execute(ctx, domain.ExperimentPreparationDraft{
		RequestID: request.RequestID, ExperimentID: request.ExperimentID, Purpose: request.Purpose, Hypothesis: request.Hypothesis, EnvironmentConditions: request.EnvironmentConditions, InitialInput: request.InitialInput, Prompts: prompts, EvaluationAxes: request.EvaluationAxes,
	})
	if err != nil {
		response := failSaveExperimentPreparationDraft(err)
		h.logger.ErrorCode(ctx, "save experiment preparation draft failed", response.Error.Code, slog.String("operation", "save_experiment_preparation_draft"))

		return response
	}

	return SaveExperimentPreparationDraftResponse{Data: &SaveExperimentPreparationDraftData{ExperimentID: draft.ExperimentID, State: "preparing", Purpose: draft.Purpose, Hypothesis: draft.Hypothesis, EnvironmentConditions: draft.EnvironmentConditions, InitialInput: draft.InitialInput, Prompts: toExperimentPreparationPromptResponses(draft.Prompts), EvaluationAxes: draft.EvaluationAxes, SavedAt: draft.SavedAt}}
}

// failSaveExperimentPreparationDraft は内部エラーを安全な画面エラーへ変換。
func failSaveExperimentPreparationDraft(err error) SaveExperimentPreparationDraftResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return SaveExperimentPreparationDraftResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toExperimentPreparationPromptResponses はdomain prompt群を画面DTOへ変換。
func toExperimentPreparationPromptResponses(prompts []domain.ExperimentPreparationPrompt) []ExperimentPreparationPromptResponse {
	responses := make([]ExperimentPreparationPromptResponse, 0, len(prompts))
	for _, prompt := range prompts {
		responses = append(responses, ExperimentPreparationPromptResponse{SequenceNo: prompt.SequenceNo, Content: prompt.Content})
	}

	return responses
}

// GetExperimentPreparation は準備中実験を画面向けDTOで返す。
func (h *ExperimentPreparationsHandler) GetExperimentPreparation(experimentID string) GetExperimentPreparationResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get experiment preparation called")

	preparation, err := h.getExperimentPreparation.Execute(ctx, experimentID)
	if err != nil {
		response := failGetExperimentPreparation(err)
		h.logger.ErrorCode(ctx, "get experiment preparation failed", response.Error.Code, slog.String("operation", "get_experiment_preparation"))

		return response
	}

	data := toGetExperimentPreparationData(preparation)

	return GetExperimentPreparationResponse{Data: &data}
}

// failGetExperimentPreparation は内部エラーを安全な画面エラーへ変換。
func failGetExperimentPreparation(err error) GetExperimentPreparationResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return GetExperimentPreparationResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toGetExperimentPreparationData はdomain実験準備を画面DTOへ変換。
func toGetExperimentPreparationData(preparation domain.ExperimentPreparation) GetExperimentPreparationData {
	prompts := make([]ExperimentPreparationPromptResponse, 0, len(preparation.Prompts))
	for _, prompt := range preparation.Prompts {
		prompts = append(prompts, ExperimentPreparationPromptResponse{SequenceNo: prompt.SequenceNo, Content: prompt.Content})
	}
	required := preparation.RequiredFields()

	return GetExperimentPreparationData{
		ExperimentID:          preparation.ExperimentID,
		State:                 preparation.State,
		Purpose:               preparation.Purpose,
		Hypothesis:            preparation.Hypothesis,
		EnvironmentConditions: preparation.EnvironmentConditions,
		InitialInput:          preparation.InitialInput,
		Prompts:               prompts,
		EvaluationAxes:        preparation.EvaluationAxes,
		Source:                ExperimentPreparationSourceResponse{State: preparation.Source.State, VersionID: preparation.Source.VersionID},
		RequiredFields:        ExperimentPreparationRequiredFieldsResponse{Purpose: required.Purpose, EnvironmentConditions: required.EnvironmentConditions, InitialInput: required.InitialInput, Prompts: required.Prompts, EvaluationAxes: required.EvaluationAxes},
		LastConfirmedAt:       preparation.LastConfirmedAt.UTC(),
	}
}

// ExperimentBriefingsHandler は実験ブリーフ開始のWails binding。
type ExperimentBriefingsHandler struct {
	startExperimentBriefing    *usecase.StartExperimentBriefing
	sendExperimentBriefMessage *usecase.SendExperimentBriefMessage
	getExperimentBriefing      *usecase.GetExperimentBriefing
	createExperimentFromBrief  *usecase.CreateExperimentFromBrief
	stopExperimentBriefing     *usecase.StopExperimentBriefing
	logger                     logger.Logger
}

// 実験ブリーフ開始binding生成。
func NewExperimentBriefingsHandler(startExperimentBriefing *usecase.StartExperimentBriefing, sendExperimentBriefMessage *usecase.SendExperimentBriefMessage, getExperimentBriefing *usecase.GetExperimentBriefing, createExperimentFromBrief *usecase.CreateExperimentFromBrief, appLogger logger.Logger, stopExperimentBriefing ...*usecase.StopExperimentBriefing) *ExperimentBriefingsHandler {
	handler := &ExperimentBriefingsHandler{
		startExperimentBriefing:    startExperimentBriefing,
		sendExperimentBriefMessage: sendExperimentBriefMessage,
		getExperimentBriefing:      getExperimentBriefing,
		createExperimentFromBrief:  createExperimentFromBrief,
		logger:                     appLogger,
	}
	if len(stopExperimentBriefing) != 0 {
		handler.stopExperimentBriefing = stopExperimentBriefing[0]
	}

	return handler
}

// StopExperimentBriefing は実験ブリーフ終了を画面向けDTOで返す。
func (h *ExperimentBriefingsHandler) StopExperimentBriefing(requestID, briefingSessionID string) StopExperimentBriefingResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "stop experiment briefing called")

	operation, err := h.stopExperimentBriefing.Execute(ctx, requestID, briefingSessionID)
	if err != nil {
		response := failStopExperimentBriefing(err)
		h.logger.ErrorCode(ctx, "stop experiment briefing failed", response.Error.Code, slog.String("operation", "stop_experiment_briefing"))

		return response
	}

	return StopExperimentBriefingResponse{Data: &StopExperimentBriefingData{OperationID: operation.OperationID}}
}

// failStopExperimentBriefing は内部エラーを安全な画面エラーへ変換。
func failStopExperimentBriefing(err error) StopExperimentBriefingResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return StopExperimentBriefingResponse{Error: &ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Error(),
	}}
}

// 採用済みブリーフからの準備中実験画面DTO返却。
func (h *ExperimentBriefingsHandler) CreateExperimentFromBrief(requestID, briefingSessionID, briefVersionID string) CreateExperimentFromBriefResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "create experiment from brief called")

	creation, err := h.createExperimentFromBrief.Execute(ctx, requestID, briefingSessionID, briefVersionID)
	if err != nil {
		response := failCreateExperimentFromBrief(err)
		h.logger.ErrorCode(ctx, "create experiment from brief failed", response.Error.Code, slog.String("operation", "create_experiment_from_brief"))

		return response
	}

	return CreateExperimentFromBriefResponse{Data: &CreateExperimentFromBriefData{
		ExperimentID: creation.ExperimentID,
		State:        creation.State,
	}}
}

// 内部エラーから安全な採用画面エラー変換。
func failCreateExperimentFromBrief(err error) CreateExperimentFromBriefResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return CreateExperimentFromBriefResponse{Error: &ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Error(),
	}}
}

// SendExperimentBriefMessage は実験ブリーフ会話送信を画面向けDTOで返す。
func (h *ExperimentBriefingsHandler) SendExperimentBriefMessage(requestID, briefingSessionID, message string) SendExperimentBriefMessageResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "send experiment brief message called")

	operation, err := h.sendExperimentBriefMessage.Execute(ctx, requestID, briefingSessionID, message)
	if err != nil {
		response := failSendExperimentBriefMessage(err)
		h.logger.ErrorCode(ctx, "send experiment brief message failed", response.Error.Code, slog.String("operation", "send_experiment_brief_message"))

		return response
	}

	return SendExperimentBriefMessageResponse{Data: &SendExperimentBriefMessageData{OperationID: operation.OperationID}}
}

// failSendExperimentBriefMessage は内部エラーを安全な画面エラーへ変換。
func failSendExperimentBriefMessage(err error) SendExperimentBriefMessageResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return SendExperimentBriefMessageResponse{Error: &ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Error(),
	}}
}

// GetExperimentBriefingResponse は実験ブリーフ再読込の成功または失敗結果。
type GetExperimentBriefingResponse struct {
	Data  *GetExperimentBriefingData `json:"data,omitempty"`
	Error *ErrorResponse             `json:"error,omitempty"`
}

// GetExperimentBriefingData は画面へ返す実験ブリーフ再読込結果。
type GetExperimentBriefingData struct {
	State           string                      `json:"state"`
	Messages        []ExperimentMessageResponse `json:"messages"`
	LatestBrief     *ExperimentBriefResponse    `json:"latestBrief"`
	LastConfirmedAt time.Time                   `json:"lastConfirmedAt"`
}

// ExperimentMessageResponse は画面表示用会話メッセージ。
type ExperimentMessageResponse struct {
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	SequenceNo int       `json:"sequenceNo"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ExperimentBriefResponse は画面表示用実験ブリーフ版。
type ExperimentBriefResponse struct {
	VersionID             string   `json:"versionId"`
	Purpose               string   `json:"purpose"`
	Decision              string   `json:"decision"`
	Hypothesis            *string  `json:"hypothesis,omitempty"`
	CandidatePrompts      []string `json:"candidatePrompts"`
	EvaluationCriteria    string   `json:"evaluationAxes"`
	EnvironmentConditions string   `json:"environmentConditions"`
	InitialInput          string   `json:"initialInput"`
	SuccessCriteria       string   `json:"successCriteria"`
	RequiredConditions    string   `json:"requiredConditions"`
	OpenQuestion          *string  `json:"openQuestion,omitempty"`
}

// GetExperimentBriefing は実験ブリーフを画面向けDTOで返す。
func (h *ExperimentBriefingsHandler) GetExperimentBriefing(briefingSessionID string) GetExperimentBriefingResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get experiment briefing called")

	briefing, err := h.getExperimentBriefing.Execute(ctx, briefingSessionID)
	if err != nil {
		response := failGetExperimentBriefing(err)
		h.logger.ErrorCode(ctx, "get experiment briefing failed", response.Error.Code, slog.String("operation", "get_experiment_briefing"))

		return response
	}

	data := toGetExperimentBriefingData(briefing)

	return GetExperimentBriefingResponse{Data: &data}
}

// failGetExperimentBriefing は内部エラーを安全な画面エラーへ変換。
func failGetExperimentBriefing(err error) GetExperimentBriefingResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return GetExperimentBriefingResponse{Error: &ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Error(),
	}}
}

// toGetExperimentBriefingData はdomain実験ブリーフを画面DTOへ変換。
func toGetExperimentBriefingData(briefing domain.ExperimentBriefing) GetExperimentBriefingData {
	messages := make([]ExperimentMessageResponse, 0, len(briefing.Messages))
	for _, message := range briefing.Messages {
		messages = append(messages, ExperimentMessageResponse{
			Role:       message.Role,
			Content:    message.Content,
			SequenceNo: message.SequenceNo,
			CreatedAt:  message.CreatedAt.UTC(),
		})
	}

	return GetExperimentBriefingData{
		State:           briefing.State,
		Messages:        messages,
		LatestBrief:     toExperimentBriefResponse(briefing.LatestBrief),
		LastConfirmedAt: briefing.LastConfirmedAt.UTC(),
	}
}

// toExperimentBriefResponse はdomain実験ブリーフ版を画面DTOへ変換。
func toExperimentBriefResponse(brief *domain.ExperimentBrief) *ExperimentBriefResponse {
	if brief == nil {
		return nil
	}

	return &ExperimentBriefResponse{
		VersionID:             brief.VersionID,
		Purpose:               brief.Purpose,
		Decision:              brief.Decision,
		Hypothesis:            brief.Hypothesis,
		CandidatePrompts:      brief.CandidatePrompts,
		EvaluationCriteria:    brief.EvaluationCriteria,
		EnvironmentConditions: brief.EnvironmentConditions,
		InitialInput:          brief.InitialInput,
		SuccessCriteria:       brief.SuccessCriteria,
		RequiredConditions:    brief.RequiredConditions,
		OpenQuestion:          brief.OpenQuestion,
	}
}

// StartExperimentBriefing は実験ブリーフ開始を画面向けDTOで返す。
func (h *ExperimentBriefingsHandler) StartExperimentBriefing(requestID string) StartExperimentBriefingResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "start experiment briefing called")

	start, err := h.startExperimentBriefing.Execute(ctx, requestID)
	if err != nil {
		response := failStartExperimentBriefing(err)
		h.logger.ErrorCode(ctx, "start experiment briefing failed", response.Error.Code, slog.String("operation", "start_experiment_briefing"))

		return response
	}

	data := toStartExperimentBriefingData(start)

	return StartExperimentBriefingResponse{Data: &data}
}

// failStartExperimentBriefing は内部エラーを安全な画面エラーへ変換。
func failStartExperimentBriefing(err error) StartExperimentBriefingResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return StartExperimentBriefingResponse{Error: &ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Error(),
	}}
}

// toStartExperimentBriefingData はdomain開始結果を画面DTOへ変換。
func toStartExperimentBriefingData(start domain.ExperimentBriefingStart) StartExperimentBriefingData {
	return StartExperimentBriefingData{
		BriefingSessionID: start.BriefingSessionID,
		OperationID:       start.OperationID,
	}
}

// ListExperimentsResponse はWails bindingの成功または失敗結果。
type ListExperimentsResponse struct {
	Data  *ListExperimentsData `json:"data,omitempty"`
	Error *ErrorResponse       `json:"error,omitempty"`
}

// ErrorResponse は画面へ返す安全なエラー情報。
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ListExperimentsData は実験一覧queryの返却データ。
type ListExperimentsData struct {
	Experiments          []ExperimentResponse  `json:"experiments"`
	CancelledExperiments []ExperimentResponse  `json:"cancelledExperiments"`
	ResumeSummary        ResumeSummaryResponse `json:"resumeSummary"`
	LastConfirmedAt      *time.Time            `json:"lastConfirmedAt,omitempty"`
}

// ExperimentResponse は画面表示用の実験行。
type ExperimentResponse struct {
	ID                      string    `json:"id"`
	Purpose                 string    `json:"purpose"`
	State                   string    `json:"state"`
	ProgressSummary         string    `json:"progressSummary"`
	DerivedFromExperimentID *string   `json:"derivedFromExperimentId,omitempty"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

// ResumeSummaryResponse は再開候補と状態別件数。
type ResumeSummaryResponse struct {
	RecommendedExperimentID *string        `json:"recommendedExperimentId,omitempty"`
	StatusCounts            map[string]int `json:"statusCounts"`
}

// ExperimentsHandler は実験一覧のWails binding。
type ExperimentsHandler struct {
	listExperiments *usecase.ListExperiments
	logger          logger.Logger
}

// NewExperimentsHandler は実験一覧bindingを生成。
func NewExperimentsHandler(listExperiments *usecase.ListExperiments, appLogger logger.Logger) *ExperimentsHandler {
	return &ExperimentsHandler{listExperiments: listExperiments, logger: appLogger}
}

// ListExperiments は実験一覧を画面向けDTOで返す。
func (h *ExperimentsHandler) ListExperiments() ListExperimentsResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "list experiments called")

	collection, err := h.listExperiments.Execute(ctx)
	if err != nil {
		response := failListExperiments(err)
		h.logger.ErrorCode(ctx, "list experiments failed", response.Error.Code, slog.String("operation", "list_experiments"))

		return response
	}

	data := toListExperimentsResponse(collection)

	return ListExperimentsResponse{Data: &data}
}

// failListExperiments は内部エラーを安全な画面エラーへ変換。
func failListExperiments(err error) ListExperimentsResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return ListExperimentsResponse{Error: &ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Error(),
	}}
}

// toListExperimentsResponse はdomain実験一覧を画面DTOへ変換。
func toListExperimentsResponse(collection domain.ExperimentCollection) ListExperimentsData {
	experiments := toExperimentResponses(collection.Experiments)
	cancelledExperiments := toExperimentResponses(collection.CancelledExperiments)
	statusCounts := make(map[string]int)
	for _, experiment := range collection.Experiments {
		statusCounts[experiment.State]++
	}

	var recommendedExperimentID *string
	if len(collection.Experiments) > 0 {
		recommendedID := collection.Experiments[0].ID
		recommendedExperimentID = &recommendedID
	}

	return ListExperimentsData{
		Experiments:          experiments,
		CancelledExperiments: cancelledExperiments,
		ResumeSummary: ResumeSummaryResponse{
			RecommendedExperimentID: recommendedExperimentID,
			StatusCounts:            statusCounts,
		},
		LastConfirmedAt: collection.LastConfirmedAt,
	}
}

// toExperimentResponses はdomain実験のsliceを画面DTOへ変換。
func toExperimentResponses(experiments []domain.Experiment) []ExperimentResponse {
	responses := make([]ExperimentResponse, 0, len(experiments))
	for _, experiment := range experiments {
		responses = append(responses, ExperimentResponse{
			ID:                      experiment.ID,
			Purpose:                 experiment.Purpose,
			State:                   experiment.State,
			ProgressSummary:         experiment.ProgressSummary,
			DerivedFromExperimentID: experiment.DerivedFromExperimentID,
			UpdatedAt:               experiment.UpdatedAt,
		})
	}

	return responses
}
