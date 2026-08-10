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

// GetInsightWorkspaceResponse は知見ワークスペースqueryの成功または失敗結果。
type GetInsightWorkspaceResponse struct {
	Data  *GetInsightWorkspaceData `json:"data,omitempty"`
	Error *ErrorResponse           `json:"error,omitempty"`
}

// GetInsightWorkspaceData は知見作成画面へ返す安全な正本。
type GetInsightWorkspaceData struct {
	EvidenceCandidates  []InsightEvidenceCandidateData  `json:"evidenceCandidates"`
	SavedConsiderations []InsightSavedConsiderationData `json:"savedConsiderations"`
	Insights            []InsightSummaryData            `json:"insights"`
	LastConfirmedAt     *time.Time                      `json:"lastConfirmedAt"`
}

// InsightEvidenceCandidateData は画面選択用の確定済み根拠候補。
type InsightEvidenceCandidateData struct {
	ExperimentID   string    `json:"experimentId"`
	Purpose        string    `json:"purpose"`
	EvaluationAxes string    `json:"evaluationAxes"`
	ConclusionID   string    `json:"conclusionId"`
	Conclusion     string    `json:"conclusion"`
	FinalizedAt    time.Time `json:"finalizedAt"`
}

// InsightSavedConsiderationData は画面表示用の確定済み利用者結論。
type InsightSavedConsiderationData struct {
	ExperimentID string    `json:"experimentId"`
	ConclusionID string    `json:"conclusionId"`
	Content      string    `json:"content"`
	FinalizedAt  time.Time `json:"finalizedAt"`
}

// InsightSummaryData は画面表示用の保存済み知見要約。
type InsightSummaryData struct {
	ID                      string    `json:"id"`
	Statement               string    `json:"statement"`
	ApplicabilityConditions string    `json:"applicabilityConditions"`
	VerificationGaps        string    `json:"verificationGaps"`
	EvidenceCount           int       `json:"evidenceCount"`
	CreatedAt               time.Time `json:"createdAt"`
}

// InsightsHandler は知見ワークスペースqueryのWails binding。
type InsightsHandler struct {
	query  *usecase.GetInsightWorkspace
	create *usecase.CreateInsight
	logger logger.Logger
}

// CreateInsightEvidenceRequest は知見作成根拠の画面DTO。
type CreateInsightEvidenceRequest struct {
	ExperimentID string `json:"experimentId"`
	ConclusionID string `json:"conclusionId"`
}

// CreateInsightRequest は知見作成の画面DTO。
type CreateInsightRequest struct {
	RequestID               string                         `json:"requestId"`
	Evidences               []CreateInsightEvidenceRequest `json:"evidences"`
	Statement               string                         `json:"statement"`
	ApplicabilityConditions string                         `json:"applicabilityConditions"`
	VerificationGaps        string                         `json:"verificationGaps"`
}

// CreateInsightEvidenceData は保存済み根拠の画面DTO。
type CreateInsightEvidenceData struct {
	ExperimentID string `json:"experimentId"`
	ConclusionID string `json:"conclusionId"`
}

// CreateInsightResponse は知見作成の成功または失敗結果。
type CreateInsightResponse struct {
	Data  *CreateInsightData `json:"data,omitempty"`
	Error *ErrorResponse     `json:"error,omitempty"`
}

// CreateInsightData は保存済み知見の画面DTO。
type CreateInsightData struct {
	RequestID               string                      `json:"requestId"`
	InsightID               string                      `json:"insightId"`
	Evidences               []CreateInsightEvidenceData `json:"evidences"`
	Statement               string                      `json:"statement"`
	ApplicabilityConditions string                      `json:"applicabilityConditions"`
	VerificationGaps        string                      `json:"verificationGaps"`
	CreatedAt               time.Time                   `json:"createdAt"`
}

// CreateInsight は画面DTOを安全に知見作成usecaseへ渡す。
func (h *InsightsHandler) CreateInsight(request CreateInsightRequest) CreateInsightResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "create insight called")
	if h.create == nil {
		return h.failCreate(ctx, apperr.New(apperr.CodeInsightCreateUnavailable))
	}
	evidences := make([]domain.InsightEvidence, len(request.Evidences))
	for index, evidence := range request.Evidences {
		evidences[index] = domain.InsightEvidence{ExperimentID: evidence.ExperimentID, ConclusionID: evidence.ConclusionID}
	}
	x, err := h.create.Execute(ctx, request.RequestID, evidences, request.Statement, request.ApplicabilityConditions, request.VerificationGaps)
	if err != nil {
		return h.failCreate(ctx, err)
	}
	return CreateInsightResponse{Data: &CreateInsightData{RequestID: x.RequestID, InsightID: x.InsightID, Evidences: toCreateInsightEvidenceData(x.Evidences), Statement: x.Statement, ApplicabilityConditions: x.ApplicabilityConditions, VerificationGaps: x.VerificationGaps, CreatedAt: x.CreatedAt.UTC()}}
}

// toCreateInsightEvidenceData はdomain根拠を画面DTOへ変換する。
func toCreateInsightEvidenceData(evidences []domain.InsightEvidence) []CreateInsightEvidenceData {
	data := make([]CreateInsightEvidenceData, 0, len(evidences))
	for _, evidence := range evidences {
		data = append(data, CreateInsightEvidenceData{ExperimentID: evidence.ExperimentID, ConclusionID: evidence.ConclusionID})
	}
	return data
}

// failCreate は知見作成エラーを安全な画面DTOへ変換する。
func (h *InsightsHandler) failCreate(ctx context.Context, err error) CreateInsightResponse {
	e := apperr.As(err)
	if e == nil {
		e = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "create insight failed", string(e.Code), slog.String("operation", "create_insight"))
	return CreateInsightResponse{Error: &ErrorResponse{Code: string(e.Code), Message: e.Error()}}
}

// NewInsightsHandler は知見ワークスペースbindingを生成する。
func NewInsightsHandler(query *usecase.GetInsightWorkspace, appLogger logger.Logger) *InsightsHandler {
	return &InsightsHandler{query: query, logger: appLogger}
}

// NewInsightsHandlerWithCreate は知見queryと作成bindingを生成する。
func NewInsightsHandlerWithCreate(query *usecase.GetInsightWorkspace, create *usecase.CreateInsight, appLogger logger.Logger) *InsightsHandler {
	return &InsightsHandler{query: query, create: create, logger: appLogger}
}

// GetInsightWorkspace は知見作成画面の安全な正本を画面DTOで返す。
func (h *InsightsHandler) GetInsightWorkspace() GetInsightWorkspaceResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get insight workspace called")
	if h.query == nil {
		return h.fail(ctx, apperr.New(apperr.CodeInsightWorkspaceUnavailable))
	}

	workspace, err := h.query.Execute(ctx)
	if err != nil {
		return h.fail(ctx, err)
	}

	return GetInsightWorkspaceResponse{Data: toGetInsightWorkspaceData(workspace)}
}

// fail は内部エラーを安全な知見ワークスペースエラーへ変換する。
func (h *InsightsHandler) fail(ctx context.Context, err error) GetInsightWorkspaceResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "get insight workspace failed", string(appErr.Code), slog.String("operation", "get_insight_workspace"))

	return GetInsightWorkspaceResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toGetInsightWorkspaceData はdomain知見ワークスペースを画面DTOへ変換する。
func toGetInsightWorkspaceData(workspace domain.InsightWorkspace) *GetInsightWorkspaceData {
	data := &GetInsightWorkspaceData{
		EvidenceCandidates:  make([]InsightEvidenceCandidateData, 0, len(workspace.EvidenceCandidates)),
		SavedConsiderations: make([]InsightSavedConsiderationData, 0, len(workspace.SavedConsiderations)),
		Insights:            make([]InsightSummaryData, 0, len(workspace.Insights)),
	}
	for _, candidate := range workspace.EvidenceCandidates {
		data.EvidenceCandidates = append(data.EvidenceCandidates, InsightEvidenceCandidateData{ExperimentID: candidate.ExperimentID, Purpose: candidate.Purpose, EvaluationAxes: candidate.EvaluationAxes, ConclusionID: candidate.ConclusionID, Conclusion: candidate.Conclusion, FinalizedAt: candidate.FinalizedAt.UTC()})
	}
	for _, consideration := range workspace.SavedConsiderations {
		data.SavedConsiderations = append(data.SavedConsiderations, InsightSavedConsiderationData{ExperimentID: consideration.ExperimentID, ConclusionID: consideration.ConclusionID, Content: consideration.Content, FinalizedAt: consideration.FinalizedAt.UTC()})
	}
	for _, insight := range workspace.Insights {
		data.Insights = append(data.Insights, InsightSummaryData{ID: insight.ID, Statement: insight.Statement, ApplicabilityConditions: insight.ApplicabilityConditions, VerificationGaps: insight.VerificationGaps, EvidenceCount: insight.EvidenceCount, CreatedAt: insight.CreatedAt.UTC()})
	}
	if workspace.LastConfirmedAt != nil {
		lastConfirmedAt := workspace.LastConfirmedAt.UTC()
		data.LastConfirmedAt = &lastConfirmedAt
	}

	return data
}
