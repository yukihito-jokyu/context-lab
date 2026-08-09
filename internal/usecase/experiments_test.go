package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// StartExperimentBriefingの入力検証と冪等開始。
func TestStartExperimentBriefingExecute(t *testing.T) {
	tests := []struct {
		name          string
		requestID     string
		starterError  error
		callTwice     bool
		wantCode      apperr.Code
		wantStart     bool
		wantStartRuns int
	}{
		{
			name:      "空のrequest IDを拒否する",
			requestID: " ",
			wantCode:  apperr.CodeBriefingRequestInvalid,
		},
		{
			name:          "開始結果を同じrequest IDで再利用する",
			requestID:     "request-1",
			callTwice:     true,
			wantStart:     true,
			wantStartRuns: 1,
		},
		{
			name:          "開始失敗を同じrequest IDで再利用する",
			requestID:     "request-2",
			starterError:  apperr.New(apperr.CodeACPNotReady),
			callTwice:     true,
			wantCode:      apperr.CodeACPNotReady,
			wantStartRuns: 1,
		},
		{
			name:          "開始失敗の記録失敗を開始失敗へ変換する",
			requestID:     "request-3",
			starterError:  errors.New("starter failed"),
			wantCode:      apperr.CodeBriefingStartFailed,
			wantStartRuns: 1,
		},
		{
			name:          "開始済み状態の記録失敗を開始失敗へ変換する",
			requestID:     "request-4",
			wantCode:      apperr.CodeBriefingStartFailed,
			wantStartRuns: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeExperimentBriefingStore()
			if tt.requestID == "request-3" {
				store.markFailedError = errors.New("mark failed")
			}
			if tt.requestID == "request-4" {
				store.markStartedError = errors.New("mark started")
			}
			starter := &fakeBriefingStarter{err: tt.starterError}
			usecase := NewStartExperimentBriefing(store, starter)

			got, err := usecase.Execute(context.Background(), tt.requestID)
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)
			} else if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if tt.wantStart {
				if got.BriefingSessionID == "" {
					t.Error("BriefingSessionID = empty, want identifier")
				}
				if got.OperationID == "" {
					t.Error("OperationID = empty, want identifier")
				}
			}
			if tt.callTwice {
				second, secondErr := usecase.Execute(context.Background(), tt.requestID)
				if tt.wantCode != "" {
					assertBriefingErrorCode(t, secondErr, tt.wantCode)
				} else if secondErr != nil {
					t.Fatalf("second Execute() error = %v", secondErr)
				}
				if tt.wantStart && (second.BriefingSessionID != got.BriefingSessionID || second.OperationID != got.OperationID) {
					t.Errorf("second start = %+v, want %+v", second, got)
				}
			}
			if gotRuns := starter.calls; gotRuns != tt.wantStartRuns {
				t.Errorf("StartExperimentBriefing() calls = %d, want %d", gotRuns, tt.wantStartRuns)
			}
		})
	}
}

// StopExperimentBriefingの入力検証、冪等停止、失敗記録。
func TestStopExperimentBriefingExecute(t *testing.T) {
	tests := []struct {
		name        string
		requestID   string
		sessionID   string
		stopErr     error
		completeErr error
		wantCode    apperr.Code
		wantCalls   int
		callTwice   bool
	}{
		{
			name:      "空の入力を拒否する",
			requestID: " ",
			sessionID: "session-1",
			wantCode:  apperr.CodeBriefingRequestInvalid,
		},
		{
			name:      "同じrequest IDで停止結果を再利用する",
			requestID: "request-1",
			sessionID: "session-1",
			wantCalls: 1,
			callTwice: true,
		},
		{
			name:      "ACP停止失敗を記録する",
			requestID: "request-2",
			sessionID: "session-1",
			stopErr:   apperr.New(apperr.CodeACPNotReady),
			wantCode:  apperr.CodeACPNotReady,
			wantCalls: 1,
			callTwice: true,
		},
		{
			name:        "停止確認の保存失敗を正規化する",
			requestID:   "request-3",
			sessionID:   "session-1",
			completeErr: errors.New("store unavailable"),
			wantCode:    apperr.CodeBriefingStopFailed,
			wantCalls:   1,
		},
		{
			name:      "通常の停止失敗を安全に正規化する",
			requestID: "request-4",
			sessionID: "session-1",
			stopErr:   errors.New("stop unavailable"),
			wantCode:  apperr.CodeBriefingStopFailed,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeBriefingStopStore(tt.completeErr)
			stopper := &fakeBriefingStopper{err: tt.stopErr}
			command := NewStopExperimentBriefing(store, stopper)

			operation, err := command.Execute(context.Background(), tt.requestID, tt.sessionID)
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)
			} else if err != nil {
				t.Fatalf("Execute() error = %v", err)
			} else if operation.OperationID != "stop-operation-"+tt.requestID {
				t.Errorf("OperationID = %q, want %q", operation.OperationID, "stop-operation-"+tt.requestID)
			}
			if tt.callTwice {
				_, secondErr := command.Execute(context.Background(), tt.requestID, tt.sessionID)
				if tt.wantCode != "" {
					assertBriefingErrorCode(t, secondErr, tt.wantCode)
				} else if secondErr != nil {
					t.Fatalf("second Execute() error = %v", secondErr)
				}
			}
			if got := stopper.calls; got != tt.wantCalls {
				t.Errorf("StopExperimentBriefing() calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

// 永続済み停止状態の復元。
func TestBriefingStopFailure(t *testing.T) {
	tests := []struct {
		name string
		code string
		want apperr.Code
	}{
		{
			name: "ACP未準備",
			code: string(apperr.CodeACPNotReady),
			want: apperr.CodeACPNotReady,
		},
		{
			name: "非active",
			code: string(apperr.CodeBriefingNotActive),
			want: apperr.CodeBriefingNotActive,
		},
		{
			name: "入力不正",
			code: string(apperr.CodeBriefingRequestInvalid),
			want: apperr.CodeBriefingRequestInvalid,
		},
		{
			name: "未知",
			code: "unknown",
			want: apperr.CodeBriefingStopFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertBriefingErrorCode(t, briefingStopFailure(tt.code), tt.want) })
	}
}

// 停止commandの永続化失敗と保留状態。
func TestStopExperimentBriefingExecuteStoreBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		store     *fakeBriefingStopStore
		wantCode  apperr.Code
		wantCalls int
	}{
		{
			name: "通常の開始保存失敗",
			store: &fakeBriefingStopStore{
				operations: map[string]domain.ExperimentBriefingStopOperation{},
				beginErr:   errors.New("store"),
			},
			wantCode: apperr.CodeBriefingStopFailed,
		},
		{
			name: "アプリケーション開始保存失敗",
			store: &fakeBriefingStopStore{
				operations: map[string]domain.ExperimentBriefingStopOperation{},
				beginErr:   apperr.New(apperr.CodeBriefingNotActive),
			},
			wantCode: apperr.CodeBriefingNotActive,
		},
		{
			name: "保留中を返す",
			store: &fakeBriefingStopStore{
				operations: map[string]domain.ExperimentBriefingStopOperation{
					"request-1": {
						RequestID:         "request-1",
						BriefingSessionID: "session-1",
						State:             domain.BriefingStartStateStarting,
					},
				},
			},
			wantCode: apperr.CodeBriefingStopPending,
		},
		{
			name: "失敗記録失敗",
			store: &fakeBriefingStopStore{
				operations: map[string]domain.ExperimentBriefingStopOperation{},
				failErr:    errors.New("store"),
			},
			wantCode:  apperr.CodeBriefingStopFailed,
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopper := &fakeBriefingStopper{err: errors.New("stop")}
			_, err := NewStopExperimentBriefing(tt.store, stopper).Execute(context.Background(), "request-1", "session-1")
			assertBriefingErrorCode(t, err, tt.wantCode)
			if stopper.calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", stopper.calls, tt.wantCalls)
			}
		})
	}
}

// CreateExperimentFromBriefの入力検証とエラー正規化。
func TestCreateExperimentFromBriefExecute(t *testing.T) {
	tests := []struct {
		name       string
		requestID  string
		sessionID  string
		versionID  string
		creatorErr error
		wantCode   apperr.Code
		wantCalls  int
	}{
		{
			name:      "準備中実験を作成する",
			requestID: "request-1",
			sessionID: "session-1",
			versionID: "version-1",
			wantCalls: 1,
		},
		{
			name:      "入力不足を拒否する",
			sessionID: "session-1",
			versionID: "version-1",
			wantCode:  apperr.CodeBriefingRequestInvalid,
		},
		{
			name:       "アプリケーションエラーを保持する",
			requestID:  "request-1",
			sessionID:  "session-1",
			versionID:  "version-1",
			creatorErr: apperr.New(apperr.CodeExperimentBriefIncomplete),
			wantCode:   apperr.CodeExperimentBriefIncomplete,
			wantCalls:  1,
		},
		{
			name:       "内部エラーを安全な作成失敗へ変換する",
			requestID:  "request-1",
			sessionID:  "session-1",
			versionID:  "version-1",
			creatorErr: errors.New("private database detail"),
			wantCode:   apperr.CodeExperimentCreateFailed,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &fakeExperimentBriefCreator{
				err: tt.creatorErr,
			}

			got, err := NewCreateExperimentFromBrief(creator).Execute(context.Background(), tt.requestID, tt.sessionID, tt.versionID)
			if gotCalls := creator.calls; gotCalls != tt.wantCalls {
				t.Errorf("CreateExperimentFromBrief() calls = %d, want %d", gotCalls, tt.wantCalls)
			}
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got.ExperimentID != "experiment-1" || got.State != "preparing" {
				t.Errorf("creation = %+v, want preparing experiment", got)
			}
		})
	}
}

// fakeExperimentBriefCreator はブリーフ採用portのtest double。
type fakeExperimentBriefCreator struct {
	err   error
	calls int
}

// 指定済み採用結果返却。
func (f *fakeExperimentBriefCreator) CreateExperimentFromBrief(context.Context, string, string, string) (domain.ExperimentCreation, bool, error) {
	f.calls++
	if f.err != nil {
		return domain.ExperimentCreation{}, false, f.err
	}

	return domain.ExperimentCreation{
		ExperimentID: "experiment-1",
		State:        "preparing",
	}, true, nil
}

// 永続済み開始失敗の安全な復元。
func TestBriefingFailure(t *testing.T) {
	tests := []struct {
		name string
		code string
		want apperr.Code
	}{
		{
			name: "ACP未準備を復元する",
			code: string(apperr.CodeACPNotReady),
			want: apperr.CodeACPNotReady,
		},
		{
			name: "入力エラーを復元する",
			code: string(apperr.CodeBriefingRequestInvalid),
			want: apperr.CodeBriefingRequestInvalid,
		},
		{
			name: "未知のエラーを開始失敗へ変換する",
			code: "UNKNOWN",
			want: apperr.CodeBriefingStartFailed,
		},
		{
			name: "開始確認中を復元する",
			code: string(apperr.CodeBriefingStartPending),
			want: apperr.CodeBriefingStartPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBriefingErrorCode(t, briefingFailure(tt.code), tt.want)
		})
	}
}

// StartExperimentBriefingの既存開始確認中返却。
func TestStartExperimentBriefingExecuteStarting(t *testing.T) {
	store := newFakeExperimentBriefingStore()
	store.starts["request-1"] = domain.ExperimentBriefingStart{
		RequestID:         "request-1",
		BriefingSessionID: "session-1",
		OperationID:       "operation-1",
		State:             domain.BriefingStartStateStarting,
	}
	starter := &fakeBriefingStarter{}
	usecase := NewStartExperimentBriefing(store, starter)

	_, err := usecase.Execute(context.Background(), "request-1")
	assertBriefingErrorCode(t, err, apperr.CodeBriefingStartPending)
	if got := starter.calls; got != 0 {
		t.Errorf("StartExperimentBriefing() calls = %d, want 0", got)
	}
}

// StartExperimentBriefingの永続化失敗変換。
func TestStartExperimentBriefingExecuteStoreFailure(t *testing.T) {
	store := newFakeExperimentBriefingStore()
	store.beginError = errors.New("database unavailable")
	usecase := NewStartExperimentBriefing(store, &fakeBriefingStarter{})

	_, err := usecase.Execute(context.Background(), "request-1")
	assertBriefingErrorCode(t, err, apperr.CodeBriefingStartFailed)
}

// GetExperimentBriefingの入力検証とport失敗正規化。
func TestGetExperimentBriefingExecute(t *testing.T) {
	repositoryFailure := errors.New("repository failure")
	confirmedAt := time.Date(2026, time.August, 9, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name      string
		sessionID string
		reader    fakeExperimentBriefingReader
		wantCode  apperr.Code
		wantFound bool
	}{
		{
			name:      "空のsession IDを拒否する",
			sessionID: " ",
			wantCode:  apperr.CodeBriefingRequestInvalid,
		},
		{
			name:      "未知sessionを区別する",
			sessionID: "missing",
			reader:    fakeExperimentBriefingReader{},
			wantCode:  apperr.CodeBriefingNotFound,
		},
		{
			name:      "永続化失敗を取得失敗へ正規化する",
			sessionID: "session-1",
			reader: fakeExperimentBriefingReader{
				err: repositoryFailure,
			},
			wantCode: apperr.CodeBriefingLoadFailed,
		},
		{
			name:      "取消を呼び出し元へ返す",
			sessionID: "session-1",
			reader: fakeExperimentBriefingReader{
				err: context.Canceled,
			},
			wantCode: apperr.CodeOperationCanceled,
		},
		{
			name:      "保存済みsessionを返す",
			sessionID: "session-1",
			reader: fakeExperimentBriefingReader{
				found: true,
				briefing: domain.ExperimentBriefing{
					State:           domain.BriefingStartStateStarted,
					Messages:        []domain.ExperimentBriefingMessage{},
					LastConfirmedAt: confirmedAt,
				},
			},
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getExperimentBriefing := NewGetExperimentBriefing(tt.reader)

			got, err := getExperimentBriefing.Execute(context.Background(), tt.sessionID)
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if gotFound := !got.LastConfirmedAt.IsZero(); gotFound != tt.wantFound {
				t.Errorf("briefing found = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

// fakeExperimentBriefingReader は再読込portのtest double。
type fakeExperimentBriefingReader struct {
	briefing domain.ExperimentBriefing
	found    bool
	err      error
}

// GetExperimentBriefing は指定済み実験ブリーフを返却。
func (f fakeExperimentBriefingReader) GetExperimentBriefing(context.Context, string) (domain.ExperimentBriefing, bool, error) {
	return f.briefing, f.found, f.err
}

// 安全な開始エラーコード検証。
func assertBriefingErrorCode(t *testing.T, err error, want apperr.Code) {
	t.Helper()

	got := apperr.As(err)
	if got == nil {
		t.Fatal("apperr.As() = nil, want app error")
	}
	if got.Code != want {
		t.Errorf("Code = %q, want %q", got.Code, want)
	}
}

// fakeExperimentBriefingStore は開始記録portのtest double。
type fakeExperimentBriefingStore struct {
	starts           map[string]domain.ExperimentBriefingStart
	beginError       error
	markStartedError error
	markFailedError  error
}

// newFakeExperimentBriefingStore は開始記録test doubleを生成。
func newFakeExperimentBriefingStore() *fakeExperimentBriefingStore {
	return &fakeExperimentBriefingStore{starts: make(map[string]domain.ExperimentBriefingStart)}
}

// BeginExperimentBriefing は開始記録を生成または再利用。
func (f *fakeExperimentBriefingStore) BeginExperimentBriefing(_ context.Context, requestID string) (domain.ExperimentBriefingStart, bool, error) {
	if f.beginError != nil {
		return domain.ExperimentBriefingStart{}, false, f.beginError
	}
	if start, found := f.starts[requestID]; found {
		return start, false, nil
	}
	start := domain.ExperimentBriefingStart{
		RequestID:         requestID,
		BriefingSessionID: "session-" + requestID,
		OperationID:       "operation-" + requestID,
		State:             domain.BriefingStartStateStarting,
	}
	f.starts[requestID] = start

	return start, true, nil
}

// MarkExperimentBriefingStarted は開始済み状態を記録。
func (f *fakeExperimentBriefingStore) MarkExperimentBriefingStarted(_ context.Context, requestID string) error {
	if f.markStartedError != nil {
		return f.markStartedError
	}
	start := f.starts[requestID]
	start.State = domain.BriefingStartStateStarted
	f.starts[requestID] = start

	return nil
}

// MarkExperimentBriefingFailed は失敗状態を記録。
func (f *fakeExperimentBriefingStore) MarkExperimentBriefingFailed(_ context.Context, requestID, failureCode string) error {
	if f.markFailedError != nil {
		return f.markFailedError
	}
	start := f.starts[requestID]
	start.State = domain.BriefingStartStateFailed
	start.FailureCode = failureCode
	f.starts[requestID] = start

	return nil
}

// fakeBriefingStarter は外部開始portのtest double。
type fakeBriefingStarter struct {
	err   error
	calls int
}

// fakeBriefingStopStore は終了記録portのtest double。
type fakeBriefingStopStore struct {
	operations  map[string]domain.ExperimentBriefingStopOperation
	completeErr error
	beginErr    error
	failErr     error
}

// newFakeBriefingStopStore は終了記録test doubleを生成。
func newFakeBriefingStopStore(completeErr error) *fakeBriefingStopStore {
	return &fakeBriefingStopStore{
		operations:  make(map[string]domain.ExperimentBriefingStopOperation),
		completeErr: completeErr,
	}
}

// BeginStopExperimentBriefing は終了操作を生成または再利用。
func (f *fakeBriefingStopStore) BeginStopExperimentBriefing(_ context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingStopOperation, bool, error) {
	if f.beginErr != nil {
		return domain.ExperimentBriefingStopOperation{}, false, f.beginErr
	}
	if operation, found := f.operations[requestID]; found {
		if operation.BriefingSessionID != briefingSessionID {
			return domain.ExperimentBriefingStopOperation{}, false, apperr.New(apperr.CodeBriefingRequestInvalid)
		}

		return operation, false, nil
	}
	operation := domain.ExperimentBriefingStopOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "stop-operation-" + requestID,
		State:             domain.BriefingStartStateStarting,
	}
	f.operations[requestID] = operation

	return operation, true, nil
}

// CompleteStopExperimentBriefing は終了済み状態を記録。
func (f *fakeBriefingStopStore) CompleteStopExperimentBriefing(_ context.Context, requestID string) error {
	if f.completeErr != nil {
		return f.completeErr
	}
	operation := f.operations[requestID]
	operation.State = domain.BriefingStartStateStopped
	f.operations[requestID] = operation

	return nil
}

// FailStopExperimentBriefing は終了失敗を記録。
func (f *fakeBriefingStopStore) FailStopExperimentBriefing(_ context.Context, requestID, failureCode string) error {
	if f.failErr != nil {
		return f.failErr
	}
	operation := f.operations[requestID]
	operation.State = domain.BriefingStartStateFailed
	operation.FailureCode = failureCode
	f.operations[requestID] = operation

	return nil
}

// fakeBriefingStopper は外部停止portのtest double。
type fakeBriefingStopper struct {
	err   error
	calls int
}

// StopExperimentBriefing は指定済み結果を返却。
func (f *fakeBriefingStopper) StopExperimentBriefing(context.Context, string, string) error {
	f.calls++

	return f.err
}

// StartExperimentBriefing は指定済み結果を返却。
func (f *fakeBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	f.calls++

	return f.err
}

// ListExperimentsのport委譲。
func TestListExperimentsExecute(t *testing.T) {
	repositoryFailure := errors.New("repository failure")
	tests := []struct {
		name      string
		reader    fakeExperimentReader
		wantCode  apperr.Code
		wantCause error
	}{
		{
			name: "一覧を返す",
			reader: fakeExperimentReader{collection: domain.ExperimentCollection{
				Experiments: []domain.Experiment{},
			}},
		},
		{
			name:      "portの失敗をアプリケーションエラーへ正規化する",
			reader:    fakeExperimentReader{err: repositoryFailure},
			wantCode:  apperr.CodeExperimentsLoadFailed,
			wantCause: repositoryFailure,
		},
		{
			name:      "取消は呼び出し元へ返す",
			reader:    fakeExperimentReader{err: context.Canceled},
			wantCode:  apperr.CodeOperationCanceled,
			wantCause: context.Canceled,
		},
		{
			name:      "タイムアウトは呼び出し元へ返す",
			reader:    fakeExperimentReader{err: context.DeadlineExceeded},
			wantCode:  apperr.CodeOperationTimeout,
			wantCause: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listExperiments := NewListExperiments(tt.reader)

			_, err := listExperiments.Execute(context.Background())
			if tt.wantCode == "" {
				if err != nil {
					t.Errorf("Execute() error = %v, want nil", err)
				}

				return
			}
			appErr := apperr.As(err)
			if appErr == nil {
				t.Fatal("As(error) = nil, want application error")
			}
			if gotCode := appErr.Code; gotCode != tt.wantCode {
				t.Errorf("Code = %q, want %q", gotCode, tt.wantCode)
			}
			if gotCause := errors.Is(err, tt.wantCause); !gotCause {
				t.Errorf("Execute() errors.Is(error, wantCause) = %v, want true", gotCause)
			}
		})
	}
}

// fakeExperimentReader は一覧読み出しportのtest double。
type fakeExperimentReader struct {
	collection domain.ExperimentCollection
	err        error
}

// ListExperiments は指定済みの一覧またはエラーを返す。
func (f fakeExperimentReader) ListExperiments(context.Context) (domain.ExperimentCollection, error) {
	return f.collection, f.err
}

// SendExperimentBriefMessageの入力検証と冪等送信。
func TestSendExperimentBriefMessageExecute(t *testing.T) {
	tests := []struct {
		name        string
		requestID   string
		sessionID   string
		message     string
		senderError error
		wantCode    apperr.Code
		wantCalls   int
	}{
		{
			name:      "空白入力を拒否する",
			requestID: "request-1",
			sessionID: "session-1",
			message:   " ",
			wantCode:  apperr.CodeBriefingRequestInvalid,
		},
		{
			name:        "ACP未準備を安全に返し新requestで再送可能にする",
			requestID:   "request-2",
			sessionID:   "session-1",
			message:     "目的を確認したい",
			senderError: apperr.New(apperr.CodeACPNotReady),
			wantCode:    apperr.CodeACPNotReady,
			wantCalls:   1,
		},
		{
			name:      "同じrequestで既存operationを返す",
			requestID: "request-3",
			sessionID: "session-1",
			message:   "目的を確認したい",
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeBriefingMessageStore()
			sender := &fakeBriefingMessageSender{err: tt.senderError}
			usecase := NewSendExperimentBriefMessage(store, sender)

			got, err := usecase.Execute(context.Background(), tt.requestID, tt.sessionID, tt.message)
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)
				if got.OperationID != "" {
					t.Errorf("OperationID = %q, want empty", got.OperationID)
				}
			} else {
				if err != nil {
					t.Fatalf("Execute() error = %v", err)
				}
				second, secondErr := usecase.Execute(context.Background(), tt.requestID, tt.sessionID, tt.message)
				if secondErr != nil {
					t.Fatalf("second Execute() error = %v", secondErr)
				}
				if second.OperationID != got.OperationID {
					t.Errorf("second OperationID = %q, want %q", second.OperationID, got.OperationID)
				}
			}
			if got := sender.calls; got != tt.wantCalls {
				t.Errorf("SendExperimentBriefMessage() calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

// fakeBriefingMessageStore は会話送信portのtest double。
type fakeBriefingMessageStore struct {
	operations map[string]domain.ExperimentBriefingMessageOperation
}

// newFakeBriefingMessageStore は会話送信test doubleを生成。
func newFakeBriefingMessageStore() *fakeBriefingMessageStore {
	return &fakeBriefingMessageStore{operations: make(map[string]domain.ExperimentBriefingMessageOperation)}
}

// BeginExperimentBriefMessage は送信操作を生成または再利用。
func (f *fakeBriefingMessageStore) BeginExperimentBriefMessage(_ context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingMessageOperation, bool, error) {
	if operation, found := f.operations[requestID]; found {
		return operation, false, nil
	}
	operation := domain.ExperimentBriefingMessageOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "operation-" + requestID,
		State:             domain.BriefingStartStateStarting,
	}
	f.operations[requestID] = operation

	return operation, true, nil
}

// CompleteExperimentBriefMessage は送信完了を記録。
func (f *fakeBriefingMessageStore) CompleteExperimentBriefMessage(_ context.Context, requestID, _ string, _ domain.ExperimentBriefingMessageResult) error {
	operation := f.operations[requestID]
	operation.State = domain.BriefingStartStateStarted
	f.operations[requestID] = operation

	return nil
}

// FailExperimentBriefMessage は送信失敗を記録。
func (f *fakeBriefingMessageStore) FailExperimentBriefMessage(_ context.Context, requestID, failureCode string) error {
	operation := f.operations[requestID]
	operation.State = domain.BriefingStartStateFailed
	operation.FailureCode = failureCode
	f.operations[requestID] = operation

	return nil
}

// fakeBriefingMessageSender は会話送信portのtest double。
type fakeBriefingMessageSender struct {
	err   error
	calls int
}

// SendExperimentBriefMessage は指定済み応答を返却。
func (f *fakeBriefingMessageSender) SendExperimentBriefMessage(context.Context, string, string, string) (domain.ExperimentBriefingMessageResult, error) {
	f.calls++

	return domain.ExperimentBriefingMessageResult{}, f.err
}

// SendExperimentBriefMessageの失敗分岐。
func TestSendExperimentBriefMessageExecuteFailures(t *testing.T) {
	tests := []struct {
		name       string
		store      briefingMessageStoreScenario
		senderErr  error
		wantCode   apperr.Code
		wantFailed bool
	}{
		{
			name: "開始記録の通常失敗を変換する",
			store: briefingMessageStoreScenario{
				beginErr: errors.New("store unavailable"),
			},
			wantCode: apperr.CodeBriefingMessageFailed,
		},
		{
			name: "開始記録の安全な失敗を保持する",
			store: briefingMessageStoreScenario{
				beginErr: apperr.New(apperr.CodeBriefingNotFound),
			},
			wantCode: apperr.CodeBriefingNotFound,
		},
		{
			name: "既存失敗を復元する",
			store: briefingMessageStoreScenario{
				operation: domain.ExperimentBriefingMessageOperation{
					State:       domain.BriefingStartStateFailed,
					FailureCode: string(apperr.CodeBriefingNotActive),
				},
			},
			wantCode: apperr.CodeBriefingNotActive,
		},
		{
			name: "既存送信中を返す",
			store: briefingMessageStoreScenario{
				operation: domain.ExperimentBriefingMessageOperation{State: domain.BriefingStartStateStarting},
			},
			wantCode: apperr.CodeBriefingMessagePending,
		},
		{
			name: "送信失敗の記録失敗を変換する",
			store: briefingMessageStoreScenario{
				failErr: errors.New("mark unavailable"),
			},
			senderErr:  errors.New("ACP unavailable"),
			wantCode:   apperr.CodeBriefingMessageFailed,
			wantFailed: true,
		},
		{
			name: "完了記録失敗を変換する",
			store: briefingMessageStoreScenario{
				completeErr: errors.New("save unavailable"),
			},
			wantCode: apperr.CodeBriefingMessageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			sender := fakeBriefingMessageSender{err: tt.senderErr}
			usecase := NewSendExperimentBriefMessage(&store, &sender)

			_, err := usecase.Execute(context.Background(), "request-1", "session-1", "message")
			assertBriefingErrorCode(t, err, tt.wantCode)
			if got := store.failed; got != tt.wantFailed {
				t.Errorf("failed = %v, want %v", got, tt.wantFailed)
			}
		})
	}
}

// briefingMessageStoreScenario は会話送信失敗再現用port。
type briefingMessageStoreScenario struct {
	operation   domain.ExperimentBriefingMessageOperation
	beginErr    error
	completeErr error
	failErr     error
	failed      bool
}

// BeginExperimentBriefMessage は指定済み送信開始結果を返却。
func (s *briefingMessageStoreScenario) BeginExperimentBriefMessage(_ context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingMessageOperation, bool, error) {
	if s.beginErr != nil {
		return domain.ExperimentBriefingMessageOperation{}, false, s.beginErr
	}
	if s.operation.State != "" {
		return s.operation, false, nil
	}

	return domain.ExperimentBriefingMessageOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "operation-1",
		State:             domain.BriefingStartStateStarting,
	}, true, nil
}

// CompleteExperimentBriefMessage は指定済み完了記録結果を返却。
func (s *briefingMessageStoreScenario) CompleteExperimentBriefMessage(context.Context, string, string, domain.ExperimentBriefingMessageResult) error {
	return s.completeErr
}

// FailExperimentBriefMessage は指定済み失敗記録結果を返却。
func (s *briefingMessageStoreScenario) FailExperimentBriefMessage(context.Context, string, string) error {
	s.failed = true

	return s.failErr
}

// briefingMessageFailureの安全な失敗復元。
func TestBriefingMessageFailure(t *testing.T) {
	tests := []struct {
		name string
		code string
		want apperr.Code
	}{
		{
			name: "ACP未準備を復元する",
			code: string(apperr.CodeACPNotReady),
			want: apperr.CodeACPNotReady,
		},
		{
			name: "非開始を復元する",
			code: string(apperr.CodeBriefingNotActive),
			want: apperr.CodeBriefingNotActive,
		},
		{
			name: "不正入力を復元する",
			code: string(apperr.CodeBriefingRequestInvalid),
			want: apperr.CodeBriefingRequestInvalid,
		},
		{
			name: "未知の失敗を正規化する",
			code: "UNKNOWN",
			want: apperr.CodeBriefingMessageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBriefingErrorCode(t, briefingMessageFailure(tt.code), tt.want)
		})
	}
}
