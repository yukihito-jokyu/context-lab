package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// StartExperimentの入力検証、全run開始、失敗永続化。
func TestStartExperimentExecute(t *testing.T) {
	tests := []struct {
		name          string
		requestID     string
		experimentID  string
		start         domain.ExperimentStart
		created       bool
		beginErr      error
		runnerErr     error
		wantCode      apperr.Code
		wantRunner    int
		wantCompleted int
		wantFailed    int
		markErr       error
		completeErr   error
		failErr       error
		finishErr     error
	}{
		{
			name:         "永続化済みのアプリケーションエラーを返す",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			beginErr:     apperr.New(apperr.CodeExperimentStartNotReady),
			wantCode:     apperr.CodeExperimentStartNotReady,
		},
		{
			name:         "永続化失敗を画面安全なエラーへ変換する",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			beginErr:     errors.New("database private detail"),
			wantCode:     apperr.CodeExperimentStartFailed,
		},
		{
			name:         "run開始状態の更新失敗を返す",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			created:      true,
			markErr:      errors.New("update failed"),
			wantCode:     apperr.CodeExperimentStartFailed,
		},
		{
			name:         "run完了更新失敗を返す",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			created:      true,
			completeErr:  errors.New("complete failed"),
			wantCode:     apperr.CodeExperimentStartFailed,
			wantRunner:   1,
		},
		{
			name:         "run失敗更新失敗を返す",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			created:      true,
			runnerErr:    errors.New("runner failed"),
			failErr:      errors.New("failure update failed"),
			wantCode:     apperr.CodeExperimentStartFailed,
			wantRunner:   1,
		},
		{
			name:          "開始完了更新失敗を返す",
			requestID:     "request-1",
			experimentID:  "experiment-1",
			start:         testExperimentStart(),
			created:       true,
			finishErr:     errors.New("finish failed"),
			wantCode:      apperr.CodeExperimentStartFailed,
			wantRunner:    2,
			wantCompleted: 2,
		},
		{
			name:          "全promptのrunを開始する",
			requestID:     "request-1",
			experimentID:  "experiment-1",
			start:         testExperimentStart(),
			created:       true,
			wantRunner:    2,
			wantCompleted: 2,
		},
		{
			name:     "request ID不足を拒否する",
			start:    testExperimentStart(),
			wantCode: apperr.CodeExperimentStartRequestInvalid,
		},
		{
			name:         "runner失敗を永続化する",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			created:      true,
			runnerErr:    errors.New("private docker detail"),
			wantCode:     apperr.CodeExperimentStartFailed,
			wantRunner:   2,
			wantFailed:   2,
		},
		{
			name:         "開始済み結果を再利用する",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start:        testExperimentStart(),
			created:      false,
		},
		{
			name:         "開始失敗を再現する",
			requestID:    "request-1",
			experimentID: "experiment-1",
			start: domain.ExperimentStart{
				State:       domain.ExperimentStartStateFailed,
				FailureCode: string(apperr.CodeExperimentStartFailed),
			},
			created:  false,
			wantCode: apperr.CodeExperimentStartFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeExperimentStartStore{
				start:       tt.start,
				created:     tt.created,
				err:         tt.beginErr,
				markErr:     tt.markErr,
				completeErr: tt.completeErr,
				failErr:     tt.failErr,
				finishErr:   tt.finishErr,
			}
			runner := &fakeExperimentRunner{err: tt.runnerErr}
			got, err := NewStartExperiment(store, runner).Execute(context.Background(), tt.requestID, tt.experimentID)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}
			} else if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if gotCalls := runner.calls; gotCalls != tt.wantRunner {
				t.Errorf("RunExperiment() calls = %d, want %d", gotCalls, tt.wantRunner)
			}
			if gotCompleted := store.completed; gotCompleted != tt.wantCompleted {
				t.Errorf("CompleteExperimentRun() calls = %d, want %d", gotCompleted, tt.wantCompleted)
			}
			if gotFailed := store.failed; gotFailed != tt.wantFailed {
				t.Errorf("FailExperimentRun() calls = %d, want %d", gotFailed, tt.wantFailed)
			}
			if tt.wantCode == "" && tt.created && got.State != domain.ExperimentStartStateRunning {
				t.Errorf("State = %q, want %q", got.State, domain.ExperimentStartStateRunning)
			}
		})
	}
}

// 開始結果の再送、runner失敗、失敗コードの安全な正規化。
func TestExperimentStartHelpers(t *testing.T) {
	tests := []struct {
		name     string
		start    domain.ExperimentStart
		runner   error
		wantCode apperr.Code
	}{
		{
			name: "開始中の再送を保留する",
			start: domain.ExperimentStart{
				State: domain.ExperimentStartStateStarting,
			},
			wantCode: apperr.CodeExperimentStartPending,
		},
		{
			name: "中止済み開始を再現する",
			start: domain.ExperimentStart{
				State:       domain.ExperimentStartStateFailed,
				FailureCode: string(apperr.CodeOperationCanceled),
			},
			wantCode: apperr.CodeOperationCanceled,
		},
		{
			name: "timeout済み開始を再現する",
			start: domain.ExperimentStart{
				State:       domain.ExperimentStartStateFailed,
				FailureCode: string(apperr.CodeOperationTimeout),
			},
			wantCode: apperr.CodeOperationTimeout,
		},
		{
			name:     "アプリケーションrunner失敗を維持する",
			runner:   apperr.New(apperr.CodeOperationCanceled),
			wantCode: apperr.CodeOperationCanceled,
		},
		{
			name:     "未知runner失敗を開始失敗へ正規化する",
			runner:   errors.New("private detail"),
			wantCode: apperr.CodeExperimentStartFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.start.State != "" {
				_, err := replayExperimentStart(tt.start)
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("replayExperimentStart() error = %v, want code %q", err, tt.wantCode)
				}

				return
			}
			if got := experimentRunFailureCode(tt.runner); got != tt.wantCode {
				t.Errorf("experimentRunFailureCode() = %q, want %q", got, tt.wantCode)
			}
			if !apperr.IsCode(experimentRunFailure(tt.runner), tt.wantCode) {
				t.Errorf("experimentRunFailure() = %v, want code %q", experimentRunFailure(tt.runner), tt.wantCode)
			}
		})
	}
}

// runner失敗後も、残りの固定promptをすべて試行して状態を残す。
func TestStartExperimentExecuteAttemptsEveryPromptAfterRunnerFailure(t *testing.T) {
	store := &fakeExperimentStartStore{
		start:   testExperimentStart(),
		created: true,
	}
	runner := &fakeExperimentRunner{
		errors: []error{
			errors.New("private docker detail"),
			nil,
		},
	}

	_, err := NewStartExperiment(store, runner).Execute(context.Background(), "request-1", "experiment-1")
	if !apperr.IsCode(err, apperr.CodeExperimentStartFailed) {
		t.Fatalf("Execute() error = %v, want code %q", err, apperr.CodeExperimentStartFailed)
	}
	if runner.calls != 2 {
		t.Errorf("RunExperiment() calls = %d, want 2", runner.calls)
	}
	if store.failed != 1 || store.completed != 1 {
		t.Errorf("run persistence = failed:%d completed:%d, want failed:1 completed:1", store.failed, store.completed)
	}
	if got, want := store.failedRunIDs, []string{"run-1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("failed run IDs = %v, want %v", got, want)
	}
	if got, want := store.completedRunIDs, []string{"run-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("completed run IDs = %v, want %v", got, want)
	}
}

// testExperimentStart は全prompt開始用の固定条件を返す。
func testExperimentStart() domain.ExperimentStart {
	return domain.ExperimentStart{
		RequestID:    "request-1",
		ExperimentID: "experiment-1",
		OperationID:  "operation-1",
		FixedConditions: domain.ExperimentFixedConditions{
			Purpose:               "目的",
			EnvironmentConditions: "環境",
			InitialInput:          "入力",
			EvaluationAxes:        "評価",
			Prompts: []domain.ExperimentPreparationPrompt{
				{
					SequenceNo: 1,
					Content:    "prompt 1",
				},
				{
					SequenceNo: 2,
					Content:    "prompt 2",
				},
			},
		},
		Runs: []domain.ExperimentWorkspaceRun{
			{
				ID:    "run-1",
				State: domain.ExperimentRunStateQueued,
			},
			{
				ID:    "run-2",
				State: domain.ExperimentRunStateQueued,
			},
		},
	}
}

// fakeExperimentStartStore は開始操作portのtest double。
type fakeExperimentStartStore struct {
	start           domain.ExperimentStart
	created         bool
	err             error
	markErr         error
	completeErr     error
	failErr         error
	finishErr       error
	completed       int
	failed          int
	failedRunIDs    []string
	completedRunIDs []string
}

// BeginExperiment は指定済み開始結果を返却。
func (s *fakeExperimentStartStore) BeginExperiment(context.Context, string, string) (domain.ExperimentStart, bool, error) {
	return s.start, s.created, s.err
}

// MarkExperimentRunRunning はrun開始状態を受理。
func (s *fakeExperimentStartStore) MarkExperimentRunRunning(context.Context, string) error {
	return s.markErr
}

// CompleteExperimentRun は完了回数を記録。
func (s *fakeExperimentStartStore) CompleteExperimentRun(_ context.Context, runID, _ string) error {
	if s.completeErr != nil {
		return s.completeErr
	}
	s.completed++
	s.completedRunIDs = append(s.completedRunIDs, runID)

	return nil
}

// FailExperimentRun は失敗回数を記録。
func (s *fakeExperimentStartStore) FailExperimentRun(_ context.Context, runID, _ string) error {
	if s.failErr != nil {
		return s.failErr
	}
	s.failed++
	s.failedRunIDs = append(s.failedRunIDs, runID)

	return nil
}

// CompleteExperimentStart は開始完了を受理。
func (s *fakeExperimentStartStore) CompleteExperimentStart(context.Context, string) error {
	return s.finishErr
}

// fakeExperimentRunner はrunner portのtest double。
type fakeExperimentRunner struct {
	calls  int
	err    error
	errors []error
}

// RunExperiment は指定済みの成功または失敗を返却。
func (r *fakeExperimentRunner) RunExperiment(context.Context, domain.ExperimentRunRequest) (string, error) {
	r.calls++
	if len(r.errors) >= r.calls && r.errors[r.calls-1] != nil {
		return "", r.errors[r.calls-1]
	}
	if r.err != nil {
		return "", r.err
	}

	return "開始しました", nil
}
