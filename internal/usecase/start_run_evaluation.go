package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// RunEvaluationStore は評価操作と評価状態を永続化するport。
type RunEvaluationStore interface {
	BeginRunEvaluation(context.Context, string, string) (domain.ExperimentRunEvaluation, bool, error)
	CompleteRunEvaluation(context.Context, string, string) error
	FailRunEvaluation(context.Context, string, string) error
}

// RunEvaluator は隔離評価をDocker adapterへ委譲するport。
type RunEvaluator interface {
	EvaluateRun(context.Context, domain.ExperimentEvaluationRequest) (string, error)
}

// StartRunEvaluation はrun評価開始command。
type StartRunEvaluation struct {
	store     RunEvaluationStore
	evaluator RunEvaluator
}

// NewStartRunEvaluation はrun評価開始commandを生成。
func NewStartRunEvaluation(store RunEvaluationStore, evaluator RunEvaluator) *StartRunEvaluation {
	return &StartRunEvaluation{store: store, evaluator: evaluator}
}

// Execute は完了済みrunの隔離評価を開始する。
func (u *StartRunEvaluation) Execute(ctx context.Context, requestID, runID string) (domain.ExperimentRunEvaluation, error) {
	requestID = strings.TrimSpace(requestID)
	runID = strings.TrimSpace(runID)
	if requestID == "" || runID == "" {
		return domain.ExperimentRunEvaluation{}, apperr.New(apperr.CodeRunEvaluationRequestInvalid)
	}

	evaluation, created, err := u.store.BeginRunEvaluation(ctx, requestID, runID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentRunEvaluation{}, appErr
		}

		return domain.ExperimentRunEvaluation{}, apperr.Wrap(apperr.CodeRunEvaluationFailed, err)
	}
	if !created {
		return replayRunEvaluation(evaluation)
	}

	summary, err := u.evaluator.EvaluateRun(ctx, domain.ExperimentEvaluationRequest{
		RunID:          evaluation.RunID,
		RunSummary:     evaluation.RunSummary,
		Purpose:        evaluation.Purpose,
		EvaluationAxes: evaluation.EvaluationAxes,
	})
	if err != nil {
		if markErr := u.store.FailRunEvaluation(ctx, evaluation.EvaluationID, string(runEvaluationFailureCode(err))); markErr != nil {
			return domain.ExperimentRunEvaluation{}, apperr.Wrap(apperr.CodeRunEvaluationFailed, markErr)
		}

		return domain.ExperimentRunEvaluation{}, runEvaluationFailure(err)
	}
	if err := u.store.CompleteRunEvaluation(ctx, evaluation.EvaluationID, summary); err != nil {
		return domain.ExperimentRunEvaluation{}, apperr.Wrap(apperr.CodeRunEvaluationFailed, err)
	}

	return evaluation, nil
}

// replayRunEvaluation は再送時に永続済み評価結果を復元。
func replayRunEvaluation(evaluation domain.ExperimentRunEvaluation) (domain.ExperimentRunEvaluation, error) {
	switch evaluation.State {
	case domain.ExperimentEvaluationStateFailed:
		return domain.ExperimentRunEvaluation{}, runEvaluationPersistedFailure(evaluation.FailureCode)
	case domain.ExperimentEvaluationStateStarting:
		return domain.ExperimentRunEvaluation{}, apperr.New(apperr.CodeRunEvaluationPending)
	default:
		return evaluation, nil
	}
}

// runEvaluationFailureCode はrunner失敗を安全なコードへ正規化。
func runEvaluationFailureCode(err error) apperr.Code {
	if appErr := apperr.As(err); appErr != nil {
		return appErr.Code
	}

	return apperr.CodeRunEvaluationFailed
}

// runEvaluationFailure はrunner失敗を画面安全なエラーへ正規化。
func runEvaluationFailure(err error) error {
	if appErr := apperr.As(err); appErr != nil {
		return appErr
	}

	return apperr.New(apperr.CodeRunEvaluationFailed)
}

// runEvaluationPersistedFailure は永続済み評価失敗を安全に復元。
func runEvaluationPersistedFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeOperationCanceled:
		return apperr.New(apperr.CodeOperationCanceled)
	case apperr.CodeOperationTimeout:
		return apperr.New(apperr.CodeOperationTimeout)
	default:
		return apperr.New(apperr.CodeRunEvaluationFailed)
	}
}
