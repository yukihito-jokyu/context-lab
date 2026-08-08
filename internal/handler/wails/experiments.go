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
