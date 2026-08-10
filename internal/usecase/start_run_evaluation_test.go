package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// run評価開始の成功、入力検証、runner失敗永続化。
func TestStartRunEvaluationExecute(t *testing.T) {
	tests := []struct {
		name        string
		requestID   string
		runID       string
		created     bool
		beginErr    error
		runError    error
		failErr     error
		completeErr error
		state       string
		failureCode string
		wantCode    apperr.Code
		wantState   string
	}{
		{
			name:      "開始結果を返す",
			requestID: "request-1",
			runID:     "run-1",
			created:   true,
		},
		{
			name:     "入力不足を拒否する",
			wantCode: apperr.CodeRunEvaluationRequestInvalid,
		},
		{
			name:      "runner失敗を永続化する",
			requestID: "request-1",
			runID:     "run-1",
			created:   true,
			runError:  errors.New("private runner error"),
			wantCode:  apperr.CodeRunEvaluationFailed,
		},
		{
			name:      "開始中の再送を保留として返す",
			requestID: "request-1",
			runID:     "run-1",
			state:     domain.ExperimentEvaluationStateStarting,
			wantCode:  apperr.CodeRunEvaluationPending,
		},
		{
			name:      "永続化済みアプリケーションエラーを返す",
			requestID: "request-1",
			runID:     "run-1",
			beginErr:  apperr.New(apperr.CodeRunEvaluationNotReady),
			wantCode:  apperr.CodeRunEvaluationNotReady,
		},
		{
			name:      "永続化済み想定外エラーを安全な開始失敗へ変換する",
			requestID: "request-1",
			runID:     "run-1",
			beginErr:  errors.New("private database error"),
			wantCode:  apperr.CodeRunEvaluationFailed,
		},
		{
			name:      "runnerアプリケーションエラーを返す",
			requestID: "request-1",
			runID:     "run-1",
			created:   true,
			runError:  apperr.New(apperr.CodeOperationTimeout),
			wantCode:  apperr.CodeOperationTimeout,
		},
		{
			name:      "失敗状態更新失敗を開始失敗へ変換する",
			requestID: "request-1",
			runID:     "run-1",
			created:   true,
			runError:  errors.New("runner failed"),
			failErr:   errors.New("save failed"),
			wantCode:  apperr.CodeRunEvaluationFailed,
		},
		{
			name:        "完了状態更新失敗を開始失敗へ変換する",
			requestID:   "request-1",
			runID:       "run-1",
			created:     true,
			completeErr: errors.New("save failed"),
			wantCode:    apperr.CodeRunEvaluationFailed,
		},
		{
			name:        "失敗済み再送はtimeoutを返す",
			requestID:   "request-1",
			runID:       "run-1",
			state:       domain.ExperimentEvaluationStateFailed,
			failureCode: string(apperr.CodeOperationTimeout),
			wantCode:    apperr.CodeOperationTimeout,
		},
		{
			name:        "失敗済み再送はcanceledを返す",
			requestID:   "request-1",
			runID:       "run-1",
			state:       domain.ExperimentEvaluationStateFailed,
			failureCode: string(apperr.CodeOperationCanceled),
			wantCode:    apperr.CodeOperationCanceled,
		},
		{
			name:        "失敗済み再送は不明コードを開始失敗へ変換する",
			requestID:   "request-1",
			runID:       "run-1",
			state:       domain.ExperimentEvaluationStateFailed,
			failureCode: "unknown",
			wantCode:    apperr.CodeRunEvaluationFailed,
		},
		{
			name:      "完了済み再送は評価結果を返す",
			requestID: "request-1",
			runID:     "run-1",
			state:     domain.ExperimentEvaluationStateCompleted,
			wantState: domain.ExperimentEvaluationStateCompleted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &runEvaluationStore{
				created:     tt.created,
				beginErr:    tt.beginErr,
				evaluation:  testRunEvaluation(tt.state),
				failErr:     tt.failErr,
				completeErr: tt.completeErr,
			}
			store.evaluation.FailureCode = tt.failureCode
			result, err := NewStartRunEvaluation(store, runEvaluator{err: tt.runError}).Execute(context.Background(), tt.requestID, tt.runID)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			wantState := tt.wantState
			if wantState == "" {
				wantState = domain.ExperimentEvaluationStateStarting
			}
			if result.State != wantState {
				t.Errorf("result.State = %q, want %q", result.State, wantState)
			}
			if wantState == domain.ExperimentEvaluationStateStarting && store.completed != 1 {
				t.Errorf("completed = %d, want 1", store.completed)
			}
		})
	}
}

// testRunEvaluation は評価開始用のdomain結果を返す。
func testRunEvaluation(state string) domain.ExperimentRunEvaluation {
	if state == "" {
		state = domain.ExperimentEvaluationStateStarting
	}

	return domain.ExperimentRunEvaluation{
		RequestID:      "request-1",
		RunID:          "run-1",
		EvaluationID:   "evaluation-1",
		OperationID:    "operation-1",
		State:          state,
		RunSummary:     "run要約",
		Purpose:        "目的",
		EvaluationAxes: "評価軸",
	}
}

// runEvaluationStore は評価永続化portのtest double。
type runEvaluationStore struct {
	evaluation  domain.ExperimentRunEvaluation
	created     bool
	beginErr    error
	completed   int
	failed      int
	completeErr error
	failErr     error
}

// BeginRunEvaluation は指定済み評価結果を返却。
func (s *runEvaluationStore) BeginRunEvaluation(context.Context, string, string) (domain.ExperimentRunEvaluation, bool, error) {
	return s.evaluation, s.created, s.beginErr
}

// CompleteRunEvaluation は評価完了を記録。
func (s *runEvaluationStore) CompleteRunEvaluation(context.Context, string, string) error {
	s.completed++

	return s.completeErr
}

// FailRunEvaluation は評価失敗を記録。
func (s *runEvaluationStore) FailRunEvaluation(context.Context, string, string) error {
	s.failed++

	return s.failErr
}

// runEvaluator は評価runner portのtest double。
type runEvaluator struct {
	err error
}

// EvaluateRun は指定済み結果を返却。
func (r runEvaluator) EvaluateRun(context.Context, domain.ExperimentEvaluationRequest) (string, error) {
	return "評価要約", r.err
}
