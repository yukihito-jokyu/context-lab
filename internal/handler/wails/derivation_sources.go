package wails

import (
	"context"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/logger"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
	"log/slog"
)

type GetDerivationSourceResponse struct {
	Data  *GetDerivationSourceData `json:"data,omitempty"`
	Error *ErrorResponse           `json:"error,omitempty"`
}
type GetDerivationSourceData struct {
	Source      DerivationSourceData      `json:"source"`
	Eligibility DerivationEligibilityData `json:"eligibility"`
}
type DerivationSourceData struct {
	ExperimentID    string                                  `json:"experimentId"`
	Purpose         string                                  `json:"purpose"`
	FixedConditions *ExperimentWorkspaceFixedConditionsData `json:"fixedConditions,omitempty"`
	Conclusion      *ExperimentConclusionData               `json:"conclusion,omitempty"`
}
type DerivationEligibilityData struct {
	CanCreateDerivedExperiment bool   `json:"canCreateDerivedExperiment"`
	ReasonCode                 string `json:"reasonCode,omitempty"`
}
type ExperimentDerivationSourcesHandler struct {
	query  *usecase.GetDerivationSource
	logger logger.Logger
}

func NewExperimentDerivationSourcesHandler(query *usecase.GetDerivationSource, appLogger logger.Logger) *ExperimentDerivationSourcesHandler {
	return &ExperimentDerivationSourcesHandler{query: query, logger: appLogger}
}
func (h *ExperimentDerivationSourcesHandler) GetDerivationSource(experimentID string) GetDerivationSourceResponse {
	ctx := context.Background()
	h.logger.Debug(ctx, "get derivation source called")
	if h.query == nil {
		return h.fail(ctx, apperr.New(apperr.CodeExperimentDerivationSourceUnavailable))
	}
	source, err := h.query.Execute(ctx, experimentID)
	if err != nil {
		return h.fail(ctx, err)
	}
	data := GetDerivationSourceData{Source: DerivationSourceData{ExperimentID: source.ExperimentID, Purpose: source.Purpose}, Eligibility: DerivationEligibilityData{CanCreateDerivedExperiment: source.CanCreateDerived, ReasonCode: source.ReasonCode}}
	if source.FixedConditions != nil {
		fixed := source.FixedConditions
		prompts := make([]ExperimentPreparationPromptResponse, 0, len(fixed.Prompts))
		for _, prompt := range fixed.Prompts {
			prompts = append(prompts, ExperimentPreparationPromptResponse{SequenceNo: prompt.SequenceNo, Content: prompt.Content})
		}
		data.Source.FixedConditions = &ExperimentWorkspaceFixedConditionsData{FixedConditionID: fixed.FixedConditionID, Purpose: fixed.Purpose, Hypothesis: fixed.Hypothesis, EnvironmentConditions: fixed.EnvironmentConditions, InitialInput: fixed.InitialInput, Prompts: prompts, EvaluationAxes: fixed.EvaluationAxes, FixedAt: fixed.FixedAt.UTC()}
	}
	if source.Conclusion != nil && source.Conclusion.State == "finalized" {
		conclusion := source.Conclusion
		data.Source.Conclusion = &ExperimentConclusionData{ID: conclusion.ConclusionID, Content: conclusion.Conclusion, State: conclusion.State, FinalizedAt: conclusion.FinalizedAt.UTC()}
	}
	return GetDerivationSourceResponse{Data: &data}
}
func (h *ExperimentDerivationSourcesHandler) fail(ctx context.Context, err error) GetDerivationSourceResponse {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.NewUnexpected(err)
	}
	h.logger.ErrorCode(ctx, "get derivation source failed", string(appErr.Code), slog.String("operation", "get_derivation_source"))
	return GetDerivationSourceResponse{Error: &ErrorResponse{Code: string(appErr.Code), Message: appErr.Error()}}
}
