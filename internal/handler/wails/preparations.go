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

// StartPreparationResponse は環境準備開始commandの成功または失敗結果。
type StartPreparationResponse struct {
	Data  *StartPreparationData `json:"data,omitempty"`
	Error *ErrorResponse        `json:"error,omitempty"`
}

// StartPreparationData は画面へ返す環境準備開始結果。
type StartPreparationData struct {
	PreparationID string `json:"preparationId"`
	State         string `json:"state"`
}

// AdoptCandidateResponse は環境候補採用commandの成功または失敗結果。
type AdoptCandidateResponse struct {
	Data  *AdoptCandidateData `json:"data,omitempty"`
	Error *ErrorResponse      `json:"error,omitempty"`
}

// AdoptCandidateData は画面へ返す採用確認済み環境候補。
type AdoptCandidateData struct {
	PreparationID         string `json:"preparationId"`
	CandidateID           string `json:"candidateId"`
	EnvironmentConditions string `json:"environmentConditions"`
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
	startPreparation *usecase.StartPreparation
	adoptCandidate   *usecase.AdoptCandidate
	logger           logger.Logger
}

// NewPreparationsHandler は環境準備queryのWails bindingを生成。
func NewPreparationsHandler(listPreparations *usecase.ListPreparations, getPreparation *usecase.GetPreparation, appLogger logger.Logger) *PreparationsHandler {
	return &PreparationsHandler{listPreparations: listPreparations, getPreparation: getPreparation, logger: appLogger}
}

// NewPreparationsHandlerWithStart は環境準備commandを含むWails bindingを生成する。
func NewPreparationsHandlerWithStart(listPreparations *usecase.ListPreparations, getPreparation *usecase.GetPreparation, startPreparation *usecase.StartPreparation, appLogger logger.Logger) *PreparationsHandler {
	return &PreparationsHandler{listPreparations: listPreparations, getPreparation: getPreparation, startPreparation: startPreparation, logger: appLogger}
}

// NewPreparationsHandlerWithStartAndAdopt は開始・候補採用commandを含むWails bindingを生成。
func NewPreparationsHandlerWithStartAndAdopt(listPreparations *usecase.ListPreparations, getPreparation *usecase.GetPreparation, startPreparation *usecase.StartPreparation, adoptCandidate *usecase.AdoptCandidate, appLogger logger.Logger) *PreparationsHandler {
	return &PreparationsHandler{listPreparations: listPreparations, getPreparation: getPreparation, startPreparation: startPreparation, adoptCandidate: adoptCandidate, logger: appLogger}
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

// StartPreparation は環境準備開始を画面向けDTOで返す。
func (h *PreparationsHandler) StartPreparation(requestID string, scope string) StartPreparationResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "start preparation called")
	if h.startPreparation == nil {
		return h.failStartPreparation(ctx, apperr.New(apperr.CodePreparationStartUnavailable))
	}

	start, err := h.startPreparation.Execute(ctx, requestID, scope)
	if err != nil {
		return h.failStartPreparation(ctx, err)
	}

	return StartPreparationResponse{Data: &StartPreparationData{PreparationID: start.PreparationID, State: start.State}}
}

// AdoptCandidate は完了済み環境準備の候補を画面向けDTOで返す。
func (h *PreparationsHandler) AdoptCandidate(requestID string, preparationID string, candidateID string) AdoptCandidateResponse {
	ctx := context.Background()
	h.logger.Info(ctx, "adopt candidate called")
	if h.adoptCandidate == nil {
		return h.failAdoptCandidate(ctx, apperr.New(apperr.CodeCandidateAdoptionUnavailable))
	}

	candidate, err := h.adoptCandidate.Execute(ctx, requestID, preparationID, candidateID)
	if err != nil {
		return h.failAdoptCandidate(ctx, err)
	}

	return AdoptCandidateResponse{Data: &AdoptCandidateData{PreparationID: candidate.PreparationID, CandidateID: candidate.CandidateID, EnvironmentConditions: candidate.EnvironmentConditions}}
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

// failStartPreparation は開始エラーを安全な画面DTOへ変換して記録する。
func (h *PreparationsHandler) failStartPreparation(ctx context.Context, err error) StartPreparationResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	response := StartPreparationResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
	h.logger.ErrorCode(ctx, "start preparation failed", response.Error.Code, slog.String("operation", "start_preparation"))

	return response
}

// failAdoptCandidate は候補採用エラーを安全な画面DTOへ変換して記録する。
func (h *PreparationsHandler) failAdoptCandidate(ctx context.Context, err error) AdoptCandidateResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	response := AdoptCandidateResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
	h.logger.ErrorCode(ctx, "adopt candidate failed", response.Error.Code, slog.String("operation", "adopt_candidate"))

	return response
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
