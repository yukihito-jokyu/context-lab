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
