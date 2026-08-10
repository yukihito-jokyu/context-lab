package wails

import (
	"context"
	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/logger"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
	"log/slog"
)

type CreateDerivedExperimentRequest struct {
	RequestID          string                          `json:"requestId"`
	SourceExperimentID string                          `json:"sourceExperimentId"`
	Changes            domain.DerivedExperimentChanges `json:"changes"`
	Reason             string                          `json:"reason"`
}
type CreateDerivedExperimentResponse struct {
	Data  *CreateDerivedExperimentData `json:"data,omitempty"`
	Error *ErrorResponse               `json:"error,omitempty"`
}
type CreateDerivedExperimentData struct {
	RequestID          string `json:"requestId"`
	ExperimentID       string `json:"experimentId"`
	SourceExperimentID string `json:"sourceExperimentId"`
	State              string `json:"state"`
	CreatedAt          string `json:"createdAt"`
}
type CreateDerivedExperimentsHandler struct {
	command *usecase.CreateDerivedExperiment
	logger  logger.Logger
}

func NewCreateDerivedExperimentsHandler(command *usecase.CreateDerivedExperiment, appLogger logger.Logger) *CreateDerivedExperimentsHandler {
	return &CreateDerivedExperimentsHandler{command: command, logger: appLogger}
}
func (h *CreateDerivedExperimentsHandler) CreateDerivedExperiment(request CreateDerivedExperimentRequest) CreateDerivedExperimentResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "create derived experiment called")
	if h.command == nil {
		return h.fail(ctx, apperr.New(apperr.CodeDerivedExperimentUnavailable))
	}
	result, err := h.command.Execute(ctx, request.RequestID, request.SourceExperimentID, request.Changes, request.Reason)
	if err != nil {
		return h.fail(ctx, err)
	}
	return CreateDerivedExperimentResponse{Data: &CreateDerivedExperimentData{RequestID: result.RequestID, ExperimentID: result.ExperimentID, SourceExperimentID: result.SourceExperimentID, State: result.State, CreatedAt: result.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")}}
}
func (h *CreateDerivedExperimentsHandler) fail(ctx context.Context, err error) CreateDerivedExperimentResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "create derived experiment failed", string(appErr.Code), slog.String("operation", "create_derived_experiment"))
	return CreateDerivedExperimentResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}
