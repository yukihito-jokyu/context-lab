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

// PreparationsHandler は環境準備session一覧queryのWails binding。
type PreparationsHandler struct {
	listPreparations *usecase.ListPreparations
	logger           logger.Logger
}

// NewPreparationsHandler は環境準備session一覧bindingを生成。
func NewPreparationsHandler(listPreparations *usecase.ListPreparations, appLogger logger.Logger) *PreparationsHandler {
	return &PreparationsHandler{listPreparations: listPreparations, logger: appLogger}
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

// failListPreparations は内部エラーを安全な画面エラーへ変換。
func failListPreparations(err error) ListPreparationsResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}

	return ListPreparationsResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
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
