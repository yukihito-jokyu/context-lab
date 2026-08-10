package wails

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

// FixExperimentConditionsRequest は固定する実験条件のフォーム値。
type FixExperimentConditionsRequest struct {
	RequestID             string   `json:"requestId"`
	ExperimentID          string   `json:"experimentId"`
	Purpose               string   `json:"purpose"`
	Hypothesis            *string  `json:"hypothesis,omitempty"`
	EnvironmentConditions string   `json:"environmentConditions"`
	InitialInput          string   `json:"initialInput"`
	Prompts               []string `json:"prompts"`
	EvaluationAxes        string   `json:"evaluationAxes"`
}

// FixExperimentConditionsResponse は条件固定の成功または失敗結果。
type FixExperimentConditionsResponse struct {
	Data  *FixExperimentConditionsData  `json:"data,omitempty"`
	Error *FixExperimentConditionsError `json:"error,omitempty"`
}

// FixExperimentConditionsData は固定済み条件の安全な識別子。
type FixExperimentConditionsData struct {
	ExperimentID     string    `json:"experimentId"`
	State            string    `json:"state"`
	FixedConditionID string    `json:"fixedConditionId"`
	OperationID      string    `json:"operationId"`
	FixedAt          time.Time `json:"fixedAt"`
}

// FixExperimentConditionsError は条件固定の安全な画面エラー。
type FixExperimentConditionsError struct {
	Code        string            `json:"code"`
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
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

// GetExperimentWorkspaceResponse は実験ワークスペースqueryの成功または失敗結果。
type GetExperimentWorkspaceResponse struct {
	Data  *GetExperimentWorkspaceData `json:"data,omitempty"`
	Error *ErrorResponse              `json:"error,omitempty"`
}

// GetExperimentWorkspaceData は画面へ返す固定条件と進行状況。
type GetExperimentWorkspaceData struct {
	ExperimentID          string                                 `json:"experimentId"`
	State                 string                                 `json:"state"`
	FixedConditions       ExperimentWorkspaceFixedConditionsData `json:"fixedConditions"`
	ConditionFixOperation ExperimentConditionFixOperationData    `json:"conditionFixOperation"`
	Runs                  []ExperimentWorkspaceRunData           `json:"runs"`
	Evaluations           []ExperimentWorkspaceEvaluationData    `json:"evaluations"`
	LastConfirmedAt       time.Time                              `json:"lastConfirmedAt"`
}

// ExperimentWorkspaceFixedConditionsData は画面表示用の不変条件。
type ExperimentWorkspaceFixedConditionsData struct {
	FixedConditionID      string                                `json:"fixedConditionId"`
	Purpose               string                                `json:"purpose"`
	Hypothesis            *string                               `json:"hypothesis,omitempty"`
	EnvironmentConditions string                                `json:"environmentConditions"`
	InitialInput          string                                `json:"initialInput"`
	Prompts               []ExperimentPreparationPromptResponse `json:"prompts"`
	EvaluationAxes        string                                `json:"evaluationAxes"`
	FixedAt               time.Time                             `json:"fixedAt"`
}

// ExperimentConditionFixOperationData は固定操作の安全な識別子。
type ExperimentConditionFixOperationData struct {
	OperationID string    `json:"operationId"`
	FixedAt     time.Time `json:"fixedAt"`
}

// ExperimentWorkspaceRunData はrunの安全な進行状況。
type ExperimentWorkspaceRunData struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	Summary   *string   `json:"summary,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ExperimentWorkspaceEvaluationData はevaluationの安全な進行状況。
type ExperimentWorkspaceEvaluationData struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	Summary   *string   `json:"summary,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StartExperimentRequest は実験開始commandの画面入力。
type StartExperimentRequest struct {
	RequestID    string `json:"requestId"`
	ExperimentID string `json:"experimentId"`
}

// StartExperimentResponse は実験開始commandの成功または失敗結果。
type StartExperimentResponse struct {
	Data  *StartExperimentData `json:"data,omitempty"`
	Error *ErrorResponse       `json:"error,omitempty"`
}

// StartExperimentData は開始済みrunの画面安全な進行状況。
type StartExperimentData struct {
	ExperimentID string                       `json:"experimentId"`
	OperationID  string                       `json:"operationId"`
	State        string                       `json:"state"`
	Runs         []ExperimentWorkspaceRunData `json:"runs"`
}

// StartRunEvaluationRequest はrun評価開始commandの画面入力。
type StartRunEvaluationRequest struct {
	RequestID string `json:"requestId"`
	RunID     string `json:"runId"`
}

// StartRunEvaluationResponse はrun評価開始commandの成功または失敗結果。
type StartRunEvaluationResponse struct {
	Data  *StartRunEvaluationData `json:"data,omitempty"`
	Error *ErrorResponse          `json:"error,omitempty"`
}

// StartRunEvaluationData は開始済み評価の画面安全な進行状況。
type StartRunEvaluationData struct {
	RunID        string `json:"runId"`
	EvaluationID string `json:"evaluationId"`
	OperationID  string `json:"operationId"`
	State        string `json:"state"`
}

// GetEvaluationDetailResponse は評価詳細queryの結果。
type GetEvaluationDetailResponse struct {
	Data  *GetEvaluationDetailData `json:"data,omitempty"`
	Error *ErrorResponse           `json:"error,omitempty"`
}

// GetEvaluationDetailData は画面へ返す安全な評価詳細。
type GetEvaluationDetailData struct {
	Evaluation      EvaluationDetailEvaluationData     `json:"evaluation"`
	Operation       EvaluationDetailOperationData      `json:"operation"`
	Evidence        EvaluationDetailEvidenceData       `json:"evidence"`
	Result          EvaluationDetailResultData         `json:"result"`
	Failure         *EvaluationDetailFailureData       `json:"failure,omitempty"`
	Reconciliation  EvaluationDetailReconciliationData `json:"reconciliation"`
	LastConfirmedAt time.Time                          `json:"lastConfirmedAt"`
}

// EvaluationDetailEvaluationData は評価実行事実。
type EvaluationDetailEvaluationData struct {
	ID           string    `json:"id"`
	ExperimentID string    `json:"experimentId"`
	RunID        string    `json:"runId"`
	State        string    `json:"state"`
	Summary      *string   `json:"summary,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// EvaluationDetailOperationData は評価操作状態。
type EvaluationDetailOperationData struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// EvaluationDetailEvidenceData は安全な評価根拠。
type EvaluationDetailEvidenceData struct {
	RunSummary     string `json:"runSummary"`
	EvaluationAxes string `json:"evaluationAxes"`
}

// EvaluationDetailResultData は評価結果または理由。
type EvaluationDetailResultData struct {
	Status     string  `json:"status"`
	Summary    *string `json:"summary,omitempty"`
	ReasonCode string  `json:"reasonCode,omitempty"`
}

// EvaluationDetailFailureData は評価不能理由。
type EvaluationDetailFailureData struct {
	Code       string    `json:"code"`
	OccurredAt time.Time `json:"occurredAt"`
}

// EvaluationDetailReconciliationData は照合状態。
type EvaluationDetailReconciliationData struct {
	State          string    `json:"state"`
	LastObservedAt time.Time `json:"lastObservedAt"`
}

// ExperimentEvaluationDetailsHandler は評価詳細queryのWails binding。
type ExperimentEvaluationDetailsHandler struct {
	getEvaluationDetail *usecase.GetEvaluationDetail
	logger              logger.Logger
}

// NewExperimentEvaluationDetailsHandler は評価詳細bindingを生成。
func NewExperimentEvaluationDetailsHandler(u *usecase.GetEvaluationDetail, l logger.Logger) *ExperimentEvaluationDetailsHandler {
	return &ExperimentEvaluationDetailsHandler{getEvaluationDetail: u, logger: l}
}

// GetEvaluationDetail は評価詳細を画面DTOで返す。
func (h *ExperimentEvaluationDetailsHandler) GetEvaluationDetail(id string) GetEvaluationDetailResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get evaluation detail called")
	if h.getEvaluationDetail == nil {
		return failGetEvaluationDetail(apperr.New(apperr.CodeEvaluationDetailUnavailable))
	}
	d, err := h.getEvaluationDetail.Execute(ctx, id)
	if err != nil {
		return failGetEvaluationDetail(err)
	}
	data := &GetEvaluationDetailData{Evaluation: EvaluationDetailEvaluationData{ID: d.Evaluation.ID, ExperimentID: d.Evaluation.ExperimentID, RunID: d.Evaluation.RunID, State: d.Evaluation.State, Summary: d.Evaluation.Summary, CreatedAt: d.Evaluation.CreatedAt.UTC(), UpdatedAt: d.Evaluation.UpdatedAt.UTC()}, Operation: EvaluationDetailOperationData{ID: d.Operation.ID, State: d.Operation.State, UpdatedAt: d.Operation.UpdatedAt.UTC()}, Evidence: EvaluationDetailEvidenceData{RunSummary: d.Evidence.RunSummary, EvaluationAxes: d.Evidence.EvaluationAxes}, Result: EvaluationDetailResultData{Status: d.Result.Status, Summary: d.Result.Summary, ReasonCode: d.Result.ReasonCode}, Reconciliation: EvaluationDetailReconciliationData{State: d.Reconciliation.State, LastObservedAt: d.Reconciliation.LastObservedAt.UTC()}, LastConfirmedAt: d.LastConfirmedAt.UTC()}
	if d.Failure != nil {
		data.Failure = &EvaluationDetailFailureData{Code: d.Failure.Code, OccurredAt: d.Failure.OccurredAt.UTC()}
	}
	return GetEvaluationDetailResponse{Data: data}
}

// failGetEvaluationDetail は内部エラーを安全なDTOへ変換。
func failGetEvaluationDetail(err error) GetEvaluationDetailResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	return GetEvaluationDetailResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// GetRunDetailResponse はrun詳細queryの成功または失敗結果。
type GetRunDetailResponse struct {
	Data  *GetRunDetailData `json:"data,omitempty"`
	Error *ErrorResponse    `json:"error,omitempty"`
}

// GetRunDetailData は画面へ返すrunの安全な実行事実と観測結果。
type GetRunDetailData struct {
	Run             RunDetailRunData                    `json:"run"`
	FixedPrompt     ExperimentPreparationPromptResponse `json:"fixedPrompt"`
	Operation       RunDetailOperationData              `json:"operation"`
	Observations    []RunDetailObservationData          `json:"observations"`
	Artifacts       RunDetailArtifactsData              `json:"artifacts"`
	Failure         *RunDetailFailureData               `json:"failure,omitempty"`
	Reconciliation  RunDetailReconciliationData         `json:"reconciliation"`
	LastConfirmedAt time.Time                           `json:"lastConfirmedAt"`
}

// RunDetailRunData はrunの安全な実行事実。
type RunDetailRunData struct {
	ID           string    `json:"id"`
	ExperimentID string    `json:"experimentId"`
	State        string    `json:"state"`
	Summary      *string   `json:"summary,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// RunDetailOperationData は開始操作の安全な状態。
type RunDetailOperationData struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RunDetailObservationData は時系列観測の安全な要約。
type RunDetailObservationData struct {
	SequenceNo int       `json:"sequenceNo"`
	Kind       string    `json:"kind"`
	OccurredAt time.Time `json:"occurredAt"`
	Summary    string    `json:"summary"`
}

// RunDetailArtifactData はartifact差分の安全な識別子。
type RunDetailArtifactData struct {
	Digest string  `json:"digest"`
	Label  *string `json:"label,omitempty"`
	Status string  `json:"status"`
}

// RunDetailArtifactsData はartifact取得の完全性。
type RunDetailArtifactsData struct {
	Status     string                  `json:"status"`
	Items      []RunDetailArtifactData `json:"items"`
	ReasonCode string                  `json:"reasonCode,omitempty"`
}

// RunDetailFailureData はrun固有の安全な失敗情報。
type RunDetailFailureData struct {
	Code           string    `json:"code"`
	OccurredAt     time.Time `json:"occurredAt"`
	PartialSummary *string   `json:"partialSummary,omitempty"`
}

// RunDetailReconciliationData は保存済み観測との照合状態。
type RunDetailReconciliationData struct {
	State          string    `json:"state"`
	LastObservedAt time.Time `json:"lastObservedAt"`
}

// ExperimentRunDetailsHandler はrun詳細queryのWails binding。
type ExperimentRunDetailsHandler struct {
	getRunDetail *usecase.GetRunDetail
	logger       logger.Logger
}

// NewExperimentRunDetailsHandler はrun詳細queryのbindingを生成。
func NewExperimentRunDetailsHandler(getRunDetail *usecase.GetRunDetail, appLogger logger.Logger) *ExperimentRunDetailsHandler {
	return &ExperimentRunDetailsHandler{getRunDetail: getRunDetail, logger: appLogger}
}

// GetRunDetail はrunの安全な実行事実と観測結果を画面DTOで返す。
func (h *ExperimentRunDetailsHandler) GetRunDetail(runID string) GetRunDetailResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get run detail called")

	if h.getRunDetail == nil {
		response := failGetRunDetail(apperr.New(apperr.CodeRunDetailUnavailable))
		h.logger.ErrorCode(ctx, "get run detail failed", response.Error.Code, slog.String("operation", "get_run_detail"))

		return response
	}
	detail, err := h.getRunDetail.Execute(ctx, runID)
	if err != nil {
		response := failGetRunDetail(err)
		h.logger.ErrorCode(ctx, "get run detail failed", response.Error.Code, slog.String("operation", "get_run_detail"))

		return response
	}
	data := toGetRunDetailData(detail)

	return GetRunDetailResponse{Data: &data}
}

// failGetRunDetail は内部エラーを安全なrun詳細エラーへ変換。
func failGetRunDetail(err error) GetRunDetailResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return GetRunDetailResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toGetRunDetailData はdomain run詳細を画面DTOへ変換する。
func toGetRunDetailData(detail domain.ExperimentRunDetail) GetRunDetailData {
	observations := make([]RunDetailObservationData, 0, len(detail.Observations))
	for _, observation := range detail.Observations {
		observations = append(observations, RunDetailObservationData{SequenceNo: observation.SequenceNo, Kind: observation.Kind, OccurredAt: observation.OccurredAt.UTC(), Summary: observation.Summary})
	}
	artifacts := make([]RunDetailArtifactData, 0, len(detail.Artifacts.Items))
	for _, artifact := range detail.Artifacts.Items {
		artifacts = append(artifacts, RunDetailArtifactData{Digest: artifact.Digest, Label: artifact.Label, Status: artifact.Status})
	}
	data := GetRunDetailData{
		Run:             RunDetailRunData{ID: detail.Run.ID, ExperimentID: detail.Run.ExperimentID, State: detail.Run.State, Summary: detail.Run.Summary, CreatedAt: detail.Run.CreatedAt.UTC(), UpdatedAt: detail.Run.UpdatedAt.UTC()},
		FixedPrompt:     ExperimentPreparationPromptResponse{SequenceNo: detail.FixedPrompt.SequenceNo, Content: detail.FixedPrompt.Content},
		Operation:       RunDetailOperationData{ID: detail.Operation.ID, State: detail.Operation.State, UpdatedAt: detail.Operation.UpdatedAt.UTC()},
		Observations:    observations,
		Artifacts:       RunDetailArtifactsData{Status: detail.Artifacts.Status, Items: artifacts, ReasonCode: detail.Artifacts.ReasonCode},
		Reconciliation:  RunDetailReconciliationData{State: detail.Reconciliation.State, LastObservedAt: detail.Reconciliation.LastObservedAt.UTC()},
		LastConfirmedAt: detail.LastConfirmedAt.UTC(),
	}
	if detail.Failure != nil {
		data.Failure = &RunDetailFailureData{Code: detail.Failure.Code, OccurredAt: detail.Failure.OccurredAt.UTC(), PartialSummary: detail.Failure.PartialSummary}
	}

	return data
}

// ExperimentEvaluationsHandler はrun評価開始commandのWails binding。
type ExperimentEvaluationsHandler struct {
	startRunEvaluation *usecase.StartRunEvaluation
	logger             logger.Logger
}

// NewExperimentEvaluationsHandler はrun評価開始bindingを生成。
func NewExperimentEvaluationsHandler(startRunEvaluation *usecase.StartRunEvaluation, appLogger logger.Logger) *ExperimentEvaluationsHandler {
	return &ExperimentEvaluationsHandler{startRunEvaluation: startRunEvaluation, logger: appLogger}
}

// StartRunEvaluation は完了済みrunの隔離評価を開始する。
func (h *ExperimentEvaluationsHandler) StartRunEvaluation(request StartRunEvaluationRequest) StartRunEvaluationResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "start run evaluation called")

	if h.startRunEvaluation == nil {
		response := failStartRunEvaluation(apperr.New(apperr.CodeRunEvaluationFailed))
		h.logger.ErrorCode(ctx, "start run evaluation failed", response.Error.Code, slog.String("operation", "start_run_evaluation"))

		return response
	}
	evaluation, err := h.startRunEvaluation.Execute(ctx, request.RequestID, request.RunID)
	if err != nil {
		response := failStartRunEvaluation(err)
		h.logger.ErrorCode(ctx, "start run evaluation failed", response.Error.Code, slog.String("operation", "start_run_evaluation"))

		return response
	}

	return StartRunEvaluationResponse{Data: &StartRunEvaluationData{RunID: evaluation.RunID, EvaluationID: evaluation.EvaluationID, OperationID: evaluation.OperationID, State: evaluation.State}}
}

// failStartRunEvaluation は内部エラーを安全な評価開始エラーへ変換。
func failStartRunEvaluation(err error) StartRunEvaluationResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return StartRunEvaluationResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// ExperimentRunsHandler は実験開始commandのWails binding。
type ExperimentRunsHandler struct {
	startExperiment *usecase.StartExperiment
	logger          logger.Logger
}

// NewExperimentRunsHandler は実験開始bindingを生成。
func NewExperimentRunsHandler(startExperiment *usecase.StartExperiment, appLogger logger.Logger) *ExperimentRunsHandler {
	return &ExperimentRunsHandler{startExperiment: startExperiment, logger: appLogger}
}

// StartExperiment は固定済み全promptの隔離runを開始する。
func (h *ExperimentRunsHandler) StartExperiment(request StartExperimentRequest) StartExperimentResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "start experiment called")

	if h.startExperiment == nil {
		response := failStartExperiment(apperr.New(apperr.CodeExperimentStartFailed))
		h.logger.ErrorCode(ctx, "start experiment failed", response.Error.Code, slog.String("operation", "start_experiment"))

		return response
	}
	start, err := h.startExperiment.Execute(ctx, request.RequestID, request.ExperimentID)
	if err != nil {
		response := failStartExperiment(err)
		h.logger.ErrorCode(ctx, "start experiment failed", response.Error.Code, slog.String("operation", "start_experiment"))

		return response
	}

	return StartExperimentResponse{Data: &StartExperimentData{ExperimentID: start.ExperimentID, OperationID: start.OperationID, State: start.State, Runs: toExperimentWorkspaceRunData(start.Runs)}}
}

// failStartExperiment は内部エラーを安全な開始エラーへ変換。
func failStartExperiment(err error) StartExperimentResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return StartExperimentResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// ExperimentWorkspacesHandler は実験ワークスペースqueryのWails binding。
type ExperimentWorkspacesHandler struct {
	getExperimentWorkspace *usecase.GetExperimentWorkspace
	logger                 logger.Logger
}

// NewExperimentWorkspacesHandler は実験ワークスペースbindingを生成。
func NewExperimentWorkspacesHandler(getExperimentWorkspace *usecase.GetExperimentWorkspace, appLogger logger.Logger) *ExperimentWorkspacesHandler {
	return &ExperimentWorkspacesHandler{getExperimentWorkspace: getExperimentWorkspace, logger: appLogger}
}

// GetExperimentWorkspace は固定条件と進行状況を画面向けDTOで返す。
func (h *ExperimentWorkspacesHandler) GetExperimentWorkspace(experimentID string) GetExperimentWorkspaceResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get experiment workspace called")

	workspace, err := h.getExperimentWorkspace.Execute(ctx, experimentID)
	if err != nil {
		response := failGetExperimentWorkspace(err)
		h.logger.ErrorCode(ctx, "get experiment workspace failed", response.Error.Code, slog.String("operation", "get_experiment_workspace"))

		return response
	}

	data := toGetExperimentWorkspaceData(workspace)

	return GetExperimentWorkspaceResponse{Data: &data}
}

// failGetExperimentWorkspace は内部エラーを安全な画面エラーへ変換。
func failGetExperimentWorkspace(err error) GetExperimentWorkspaceResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return GetExperimentWorkspaceResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toGetExperimentWorkspaceData はdomainワークスペースを画面DTOへ変換。
func toGetExperimentWorkspaceData(workspace domain.ExperimentWorkspace) GetExperimentWorkspaceData {
	return GetExperimentWorkspaceData{
		ExperimentID: workspace.ExperimentID,
		State:        workspace.State,
		FixedConditions: ExperimentWorkspaceFixedConditionsData{
			FixedConditionID:      workspace.FixedConditions.FixedConditionID,
			Purpose:               workspace.FixedConditions.Purpose,
			Hypothesis:            workspace.FixedConditions.Hypothesis,
			EnvironmentConditions: workspace.FixedConditions.EnvironmentConditions,
			InitialInput:          workspace.FixedConditions.InitialInput,
			Prompts:               toExperimentPreparationPromptResponses(workspace.FixedConditions.Prompts),
			EvaluationAxes:        workspace.FixedConditions.EvaluationAxes,
			FixedAt:               workspace.FixedConditions.FixedAt.UTC(),
		},
		ConditionFixOperation: ExperimentConditionFixOperationData{OperationID: workspace.ConditionFixOperationID, FixedAt: workspace.ConditionFixOperationAt.UTC()},
		Runs:                  toExperimentWorkspaceRunData(workspace.Runs),
		Evaluations:           toExperimentWorkspaceEvaluationData(workspace.Evaluations),
		LastConfirmedAt:       workspace.LastConfirmedAt.UTC(),
	}
}

// toExperimentWorkspaceRunData はdomain runを画面DTOへ変換する。
func toExperimentWorkspaceRunData(runs []domain.ExperimentWorkspaceRun) []ExperimentWorkspaceRunData {
	data := make([]ExperimentWorkspaceRunData, 0, len(runs))
	for _, run := range runs {
		data = append(data, ExperimentWorkspaceRunData{ID: run.ID, State: run.State, Summary: run.Summary, UpdatedAt: run.UpdatedAt.UTC()})
	}

	return data
}

// toExperimentWorkspaceEvaluationData はdomain evaluationを画面DTOへ変換する。
func toExperimentWorkspaceEvaluationData(evaluations []domain.ExperimentWorkspaceEvaluation) []ExperimentWorkspaceEvaluationData {
	data := make([]ExperimentWorkspaceEvaluationData, 0, len(evaluations))
	for _, evaluation := range evaluations {
		data = append(data, ExperimentWorkspaceEvaluationData{ID: evaluation.ID, State: evaluation.State, Summary: evaluation.Summary, UpdatedAt: evaluation.UpdatedAt.UTC()})
	}

	return data
}

// ExperimentPreparationsHandler は実験準備queryのWails binding。
type ExperimentPreparationsHandler struct {
	getExperimentPreparation       *usecase.GetExperimentPreparation
	saveExperimentPreparationDraft *usecase.SaveExperimentPreparationDraft
	fixExperimentConditions        *usecase.FixExperimentConditions
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

// NewExperimentPreparationsHandlerWithConditions は条件固定commandを含む実験準備bindingを生成。
func NewExperimentPreparationsHandlerWithConditions(getExperimentPreparation *usecase.GetExperimentPreparation, appLogger logger.Logger, saveExperimentPreparationDraft *usecase.SaveExperimentPreparationDraft, fixExperimentConditions *usecase.FixExperimentConditions) *ExperimentPreparationsHandler {
	handler := NewExperimentPreparationsHandler(getExperimentPreparation, appLogger, saveExperimentPreparationDraft)
	handler.fixExperimentConditions = fixExperimentConditions

	return handler
}

// FixExperimentConditions はフォームの条件を不変artifactとして固定する。
func (h *ExperimentPreparationsHandler) FixExperimentConditions(request FixExperimentConditionsRequest) FixExperimentConditionsResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "fix experiment conditions called")

	if h.fixExperimentConditions == nil {
		response := failFixExperimentConditions(apperr.New(apperr.CodeFixConditionsSaveFailed), request)
		h.logger.ErrorCode(ctx, "fix experiment conditions failed", response.Error.Code, slog.String("operation", "fix_experiment_conditions"))

		return response
	}
	prompts := make([]domain.ExperimentPreparationPrompt, 0, len(request.Prompts))
	for index, content := range request.Prompts {
		prompts = append(prompts, domain.ExperimentPreparationPrompt{SequenceNo: index + 1, Content: content})
	}
	fixed, err := h.fixExperimentConditions.Execute(ctx, domain.ExperimentFixedConditions{
		RequestID: request.RequestID, ExperimentID: request.ExperimentID, Purpose: request.Purpose, Hypothesis: request.Hypothesis, EnvironmentConditions: request.EnvironmentConditions, InitialInput: request.InitialInput, Prompts: prompts, EvaluationAxes: request.EvaluationAxes,
	})
	if err != nil {
		response := failFixExperimentConditions(err, request)
		h.logger.ErrorCode(ctx, "fix experiment conditions failed", response.Error.Code, slog.String("operation", "fix_experiment_conditions"))

		return response
	}

	return FixExperimentConditionsResponse{Data: &FixExperimentConditionsData{ExperimentID: fixed.ExperimentID, State: "ready", FixedConditionID: fixed.FixedConditionID, OperationID: fixed.OperationID, FixedAt: fixed.FixedAt.UTC()}}
}

// failFixExperimentConditions は内部エラーを安全な条件固定エラーへ変換。
func failFixExperimentConditions(err error, request FixExperimentConditionsRequest) FixExperimentConditionsResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	errorResponse := &FixExperimentConditionsError{Code: string(appErr.Code), Message: appErr.Error()}
	if appErr.Code == apperr.CodeConditionsInvalid {
		errorResponse.FieldErrors = fixedConditionFieldErrors(request)
	}

	return FixExperimentConditionsResponse{Error: errorResponse}
}

// fixedConditionFieldErrors は固定条件フォームの不足フィールドを返す。
func fixedConditionFieldErrors(request FixExperimentConditionsRequest) map[string]string {
	fieldErrors := make(map[string]string)
	if strings.TrimSpace(request.Purpose) == "" {
		fieldErrors["purpose"] = "目的を入力してください"
	}
	if strings.TrimSpace(request.EnvironmentConditions) == "" {
		fieldErrors["environmentConditions"] = "環境条件を入力してください"
	}
	if strings.TrimSpace(request.InitialInput) == "" {
		fieldErrors["initialInput"] = "初期入力を入力してください"
	}
	if strings.TrimSpace(request.EvaluationAxes) == "" {
		fieldErrors["evaluationAxes"] = "評価軸を入力してください"
	}
	if len(request.Prompts) == 0 {
		fieldErrors["prompts"] = "promptを1件以上入力してください"
	}
	for index, prompt := range request.Prompts {
		if strings.TrimSpace(prompt) == "" {
			fieldErrors[fmt.Sprintf("prompts.%d", index)] = "promptを入力してください"
		}
	}

	return fieldErrors
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
