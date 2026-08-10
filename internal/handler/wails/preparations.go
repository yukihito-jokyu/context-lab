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

// ListPreparationsResponse は環境準備session一覧queryの成功または失敗結果。
type ListPreparationsResponse struct {
	Data  *ListPreparationsData `json:"data,omitempty"`
	Error *ErrorResponse        `json:"error,omitempty"`
}

// ListPreparationsData は画面へ返す環境準備session一覧。
type ListPreparationsData struct {
	Preparations []PreparationResponse `json:"preparations"`
}

// PreparationResponse は画面表示用の環境準備session。
type PreparationResponse struct {
	PreparationID  string    `json:"preparationId"`
	State          string    `json:"state"`
	StartedAt      time.Time `json:"startedAt"`
	LastObservedAt time.Time `json:"lastObservedAt"`
}

// GetPreparationResponse は環境準備session詳細queryの成功または失敗結果。
type GetPreparationResponse struct {
	Data  *GetPreparationData `json:"data,omitempty"`
	Error *ErrorResponse      `json:"error,omitempty"`
}

// GetPreparationData は画面へ返す環境準備session詳細。
type GetPreparationData struct {
	PreparationID  string                            `json:"preparationId"`
	State          string                            `json:"state"`
	StartedAt      time.Time                         `json:"startedAt"`
	LastObservedAt time.Time                         `json:"lastObservedAt"`
	Candidates     []PreparationCandidateResponse    `json:"candidates"`
	Diagnostics    []PreparationDiagnosticResponse   `json:"diagnostics"`
	Failure        *PreparationFailureResponse       `json:"failure,omitempty"`
	Reconciliation PreparationReconciliationResponse `json:"reconciliation"`
}

// PreparationCandidateResponse は画面表示用の環境候補。
type PreparationCandidateResponse struct {
	ID                    string    `json:"id"`
	EnvironmentConditions string    `json:"environmentConditions"`
	Summary               string    `json:"summary"`
	CreatedAt             time.Time `json:"createdAt"`
}

// PreparationDiagnosticResponse は画面表示用の安全な診断情報。
type PreparationDiagnosticResponse struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Summary    string    `json:"summary"`
	OccurredAt time.Time `json:"occurredAt"`
}

// PreparationFailureResponse は画面表示用の安全な失敗情報。
type PreparationFailureResponse struct {
	Code       string    `json:"code"`
	OccurredAt time.Time `json:"occurredAt"`
}

// PreparationReconciliationResponse は画面表示用の再照合状態。
type PreparationReconciliationResponse struct {
	State          string    `json:"state"`
	LastObservedAt time.Time `json:"lastObservedAt"`
}

// PreparationsHandler は環境準備session一覧queryのWails binding。
type PreparationsHandler struct {
	listPreparations *usecase.ListPreparations
	getPreparation   *usecase.GetPreparation
	logger           logger.Logger
}

// NewPreparationsHandler は環境準備queryのWails bindingを生成。
func NewPreparationsHandler(listPreparations *usecase.ListPreparations, getPreparation *usecase.GetPreparation, appLogger logger.Logger) *PreparationsHandler {
	return &PreparationsHandler{listPreparations: listPreparations, getPreparation: getPreparation, logger: appLogger}
}

// ListPreparations は環境準備session一覧を画面向けDTOで返す。
func (h *PreparationsHandler) ListPreparations() ListPreparationsResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "list preparations called")

	preparations, err := h.listPreparations.Execute(ctx)
	if err != nil {
		response := failListPreparations(err)
		h.logger.ErrorCode(ctx, "list preparations failed", response.Error.Code, slog.String("operation", "list_preparations"))

		return response
	}

	data := toListPreparationsData(preparations)

	return ListPreparationsResponse{Data: &data}
}

// GetPreparation は環境準備session詳細を画面向けDTOで返す。
func (h *PreparationsHandler) GetPreparation(preparationID string) GetPreparationResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get preparation called")
	preparation, err := h.getPreparation.Execute(ctx, preparationID)
	if err != nil {
		response := failGetPreparation(err)
		h.logger.ErrorCode(ctx, "get preparation failed", response.Error.Code, slog.String("operation", "get_preparation"))

		return response
	}

	data := toGetPreparationData(preparation)

	return GetPreparationResponse{Data: &data}
}

// failListPreparations は内部エラーを安全な画面エラーへ変換。
func failListPreparations(err error) ListPreparationsResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return ListPreparationsResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// failGetPreparation は内部エラーを安全な画面エラーへ変換。
func failGetPreparation(err error) GetPreparationResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return GetPreparationResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}

// toListPreparationsData はdomain環境準備sessionを画面DTOへ変換。
func toListPreparationsData(preparations []domain.Preparation) ListPreparationsData {
	responses := make([]PreparationResponse, 0, len(preparations))
	for _, preparation := range preparations {
		responses = append(responses, PreparationResponse{
			PreparationID:  preparation.ID,
			State:          preparation.State,
			StartedAt:      preparation.StartedAt.UTC(),
			LastObservedAt: preparation.LastObservedAt.UTC(),
		})
	}

	return ListPreparationsData{Preparations: responses}
}

// toGetPreparationData はdomain環境準備詳細を画面DTOへ変換。
func toGetPreparationData(preparation domain.PreparationDetail) GetPreparationData {
	candidates := make([]PreparationCandidateResponse, 0, len(preparation.Candidates))
	for _, candidate := range preparation.Candidates {
		candidates = append(candidates, PreparationCandidateResponse{ID: candidate.ID, EnvironmentConditions: candidate.EnvironmentConditions, Summary: candidate.Summary, CreatedAt: candidate.CreatedAt.UTC()})
	}
	diagnostics := make([]PreparationDiagnosticResponse, 0, len(preparation.Diagnostics))
	for _, diagnostic := range preparation.Diagnostics {
		diagnostics = append(diagnostics, PreparationDiagnosticResponse{ID: diagnostic.ID, Code: diagnostic.Code, Summary: diagnostic.SafeSummary, OccurredAt: diagnostic.OccurredAt.UTC()})
	}
	var failure *PreparationFailureResponse
	if preparation.Failure != nil {
		failure = &PreparationFailureResponse{Code: preparation.Failure.Code, OccurredAt: preparation.Failure.OccurredAt.UTC()}
	}

	return GetPreparationData{PreparationID: preparation.ID, State: preparation.State, StartedAt: preparation.StartedAt.UTC(), LastObservedAt: preparation.LastObservedAt.UTC(), Candidates: candidates, Diagnostics: diagnostics, Failure: failure, Reconciliation: PreparationReconciliationResponse{State: preparation.Reconciliation.State, LastObservedAt: preparation.Reconciliation.LastObservedAt.UTC()}}
}
