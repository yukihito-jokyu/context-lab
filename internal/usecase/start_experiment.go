package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentStartStore は開始操作とrun状態を永続化するport。
type ExperimentStartStore interface {
	BeginExperiment(context.Context, string, string) (domain.ExperimentStart, bool, error)
	MarkExperimentRunRunning(context.Context, string) error
	CompleteExperimentRun(context.Context, string, string) error
	FailExperimentRun(context.Context, string, string) error
	CompleteExperimentStart(context.Context, string) error
}

// ExperimentRunner は隔離runをDocker adapterへ委譲するport。
type ExperimentRunner interface {
	RunExperiment(context.Context, domain.ExperimentRunRequest) (string, error)
}

// StartExperiment は実験開始command。
type StartExperiment struct {
	store  ExperimentStartStore
	runner ExperimentRunner
}

// NewStartExperiment は実験開始commandを生成。
func NewStartExperiment(store ExperimentStartStore, runner ExperimentRunner) *StartExperiment {
	return &StartExperiment{store: store, runner: runner}
}

// Execute は固定済み全promptのrunを開始する。
func (u *StartExperiment) Execute(ctx context.Context, requestID, experimentID string) (domain.ExperimentStart, error) {
	requestID = strings.TrimSpace(requestID)
	experimentID = strings.TrimSpace(experimentID)
	if requestID == "" || experimentID == "" {
		return domain.ExperimentStart{}, apperr.New(apperr.CodeExperimentStartRequestInvalid)
	}

	start, created, err := u.store.BeginExperiment(ctx, requestID, experimentID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentStart{}, appErr
		}

		return domain.ExperimentStart{}, apperr.Wrap(apperr.CodeExperimentStartFailed, err)
	}
	if !created {
		return replayExperimentStart(start)
	}

	var runFailure error
	for index := range start.Runs {
		run := &start.Runs[index]
		if err := u.store.MarkExperimentRunRunning(ctx, run.ID); err != nil {
			return domain.ExperimentStart{}, apperr.Wrap(apperr.CodeExperimentStartFailed, err)
		}
		run.State = domain.ExperimentRunStateRunning

		prompt := start.FixedConditions.Prompts[index]
		summary, err := u.runner.RunExperiment(ctx, domain.ExperimentRunRequest{
			ExperimentID:          start.ExperimentID,
			RunID:                 run.ID,
			Purpose:               start.FixedConditions.Purpose,
			EnvironmentConditions: start.FixedConditions.EnvironmentConditions,
			InitialInput:          start.FixedConditions.InitialInput,
			Prompt:                prompt.Content,
			EvaluationAxes:        start.FixedConditions.EvaluationAxes,
		})
		if err != nil {
			if markErr := u.store.FailExperimentRun(ctx, run.ID, string(experimentRunFailureCode(err))); markErr != nil {
				return domain.ExperimentStart{}, apperr.Wrap(apperr.CodeExperimentStartFailed, markErr)
			}
			run.State = domain.ExperimentRunStateFailed
			if runFailure == nil {
				runFailure = experimentRunFailure(err)
			}
			continue
		}
		if err := u.store.CompleteExperimentRun(ctx, run.ID, summary); err != nil {
			return domain.ExperimentStart{}, apperr.Wrap(apperr.CodeExperimentStartFailed, err)
		}
		run.State = domain.ExperimentRunStateCompleted
		if strings.TrimSpace(summary) != "" {
			run.Summary = &summary
		}
	}
	if runFailure != nil {
		return domain.ExperimentStart{}, runFailure
	}
	if err := u.store.CompleteExperimentStart(ctx, start.RequestID); err != nil {
		return domain.ExperimentStart{}, apperr.Wrap(apperr.CodeExperimentStartFailed, err)
	}
	start.State = domain.ExperimentStartStateRunning

	return start, nil
}

// replayExperimentStart は再送時に永続済み開始結果を復元。
func replayExperimentStart(start domain.ExperimentStart) (domain.ExperimentStart, error) {
	switch start.State {
	case domain.ExperimentStartStateFailed:
		return domain.ExperimentStart{}, experimentStartFailure(start.FailureCode)
	case domain.ExperimentStartStateStarting:
		return domain.ExperimentStart{}, apperr.New(apperr.CodeExperimentStartPending)
	default:
		return start, nil
	}
}

// experimentRunFailureCode はrunner失敗を永続化する安全なコードへ正規化。
func experimentRunFailureCode(err error) apperr.Code {
	if appErr := apperr.As(err); appErr != nil {
		return appErr.Code
	}

	return apperr.CodeExperimentStartFailed
}

// experimentRunFailure はrunner失敗を画面安全なエラーへ正規化。
func experimentRunFailure(err error) error {
	if appErr := apperr.As(err); appErr != nil {
		return appErr
	}

	return apperr.New(apperr.CodeExperimentStartFailed)
}

// experimentStartFailure は永続済み開始失敗を安全に復元。
func experimentStartFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeOperationCanceled:
		return apperr.New(apperr.CodeOperationCanceled)
	case apperr.CodeOperationTimeout:
		return apperr.New(apperr.CodeOperationTimeout)
	default:
		return apperr.New(apperr.CodeExperimentStartFailed)
	}
}
