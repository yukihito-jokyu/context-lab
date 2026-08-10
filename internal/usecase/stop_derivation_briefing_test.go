package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// StopDerivationBriefingの入力、再生、停止失敗を確認。
func TestStopDerivationBriefingExecute(t *testing.T) {
	tests := []struct {
		name       string
		requestID  string
		sessionID  string
		store      *derivationBriefingStopStore
		stopper    derivationBriefingStopper
		wantCode   apperr.Code
		wantState  string
		wantFinish int
		wantFail   int
	}{
		{
			name:      "入力不足を拒否する",
			requestID: " ",
			sessionID: "session-1",
			wantCode:  apperr.CodeDerivationBriefingStopInvalid,
		},
		{
			name:       "停止確認を保存する",
			requestID:  "request-1",
			sessionID:  "session-1",
			store:      &derivationBriefingStopStore{created: true},
			wantState:  domain.BriefingStartStateStopped,
			wantFinish: 1,
		},
		{
			name:      "停止済みの同一要求を再生する",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{operation: domain.DerivationBriefingStopOperation{
				OperationID: "operation-1",
				State:       domain.BriefingStartStateStopped,
			}},
			wantState: domain.BriefingStartStateStopped,
		},
		{
			name:      "永続済みエラーを返す",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{
				beginErr: apperr.New(apperr.CodeDerivationBriefingStopNotFound),
			},
			wantCode: apperr.CodeDerivationBriefingStopNotFound,
		},
		{
			name:      "未知の開始保存失敗を安全に正規化する",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{
				beginErr: errors.New("private sqlite failure"),
			},
			wantCode: apperr.CodeDerivationBriefingStopFailed,
		},
		{
			name:      "開始中の同一要求を保留として返す",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{operation: domain.DerivationBriefingStopOperation{
				State: domain.BriefingStartStateStarting,
			}},
			wantCode: apperr.CodeDerivationBriefingStopPending,
		},
		{
			name:      "失敗済み要求を安全な失敗として再生する",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{operation: domain.DerivationBriefingStopOperation{
				State:       domain.BriefingStartStateFailed,
				FailureCode: string(apperr.CodeDerivationBriefingStopNotActive),
			}},
			wantCode: apperr.CodeDerivationBriefingStopNotActive,
		},
		{
			name:      "ACP停止失敗を保存する",
			requestID: "request-1",
			sessionID: "session-1",
			store:     &derivationBriefingStopStore{created: true},
			stopper:   derivationBriefingStopper{err: apperr.New(apperr.CodeACPNotReady)},
			wantCode:  apperr.CodeACPNotReady,
			wantFail:  1,
		},
		{
			name:      "未知の停止失敗を安全に正規化する",
			requestID: "request-1",
			sessionID: "session-1",
			store:     &derivationBriefingStopStore{created: true},
			stopper:   derivationBriefingStopper{err: errors.New("private ACP failure")},
			wantCode:  apperr.CodeDerivationBriefingStopFailed,
			wantFail:  1,
		},
		{
			name:      "停止失敗の保存失敗を安全に正規化する",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{
				created: true,
				failErr: errors.New("private sqlite failure"),
			},
			stopper:  derivationBriefingStopper{err: errors.New("private ACP failure")},
			wantCode: apperr.CodeDerivationBriefingStopFailed,
			wantFail: 1,
		},
		{
			name:      "停止確認の保存失敗を安全に正規化する",
			requestID: "request-1",
			sessionID: "session-1",
			store: &derivationBriefingStopStore{
				created:   true,
				finishErr: errors.New("private sqlite failure"),
			},
			wantCode:   apperr.CodeDerivationBriefingStopFailed,
			wantFinish: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			if store == nil {
				store = &derivationBriefingStopStore{}
			}
			got, err := NewStopDerivationBriefing(store, tt.stopper).Execute(context.Background(), tt.requestID, tt.sessionID)
			if tt.wantCode != "" {
				appErr := apperr.As(err)
				if appErr == nil {
					t.Fatalf("error = %v, want code %q", err, tt.wantCode)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("Code = %q, want %q", appErr.Code, tt.wantCode)
				}
			} else if err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}
			if got.State != tt.wantState {
				t.Errorf("State = %q, want %q", got.State, tt.wantState)
			}
			if store.finished != tt.wantFinish {
				t.Errorf("finished = %d, want %d", store.finished, tt.wantFinish)
			}
			if store.failed != tt.wantFail {
				t.Errorf("failed = %d, want %d", store.failed, tt.wantFail)
			}
		})
	}
}

// 永続済み派生壁打ち終了失敗の安全な復元を確認。
func TestDerivationBriefingStopFailure(t *testing.T) {
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
			name: "入力不正を復元する",
			code: string(apperr.CodeDerivationBriefingStopInvalid),
			want: apperr.CodeDerivationBriefingStopInvalid,
		},
		{
			name: "壁打ち不在を復元する",
			code: string(apperr.CodeDerivationBriefingStopNotFound),
			want: apperr.CodeDerivationBriefingStopNotFound,
		},
		{
			name: "終了不可能状態を復元する",
			code: string(apperr.CodeDerivationBriefingStopNotActive),
			want: apperr.CodeDerivationBriefingStopNotActive,
		},
		{
			name: "未知の失敗を安全に正規化する",
			code: "private failure",
			want: apperr.CodeDerivationBriefingStopFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := apperr.As(derivationBriefingStopFailure(tt.code))
			if appErr == nil {
				t.Fatal("apperr.As() = nil, want app error")
			}
			if appErr.Code != tt.want {
				t.Errorf("Code = %q, want %q", appErr.Code, tt.want)
			}
		})
	}
}

// derivationBriefingStopStore は派生壁打ち停止の永続化結果を固定する。
type derivationBriefingStopStore struct {
	operation domain.DerivationBriefingStopOperation
	created   bool
	beginErr  error
	finishErr error
	failErr   error
	finished  int
	failed    int
}

// BeginStopDerivationBriefing は停止操作を返す。
func (s *derivationBriefingStopStore) BeginStopDerivationBriefing(_ context.Context, requestID, briefingSessionID string) (domain.DerivationBriefingStopOperation, bool, error) {
	if s.beginErr != nil {
		return domain.DerivationBriefingStopOperation{}, false, s.beginErr
	}
	if s.operation.RequestID == "" && s.operation.State == "" {
		s.operation = domain.DerivationBriefingStopOperation{
			RequestID:         requestID,
			BriefingSessionID: briefingSessionID,
			OperationID:       "operation-1",
		}
	}

	return s.operation, s.created, nil
}

// CompleteStopDerivationBriefing は停止確認回数を記録する。
func (s *derivationBriefingStopStore) CompleteStopDerivationBriefing(context.Context, string) error {
	s.finished++

	return s.finishErr
}

// FailStopDerivationBriefing は停止失敗回数を記録する。
func (s *derivationBriefingStopStore) FailStopDerivationBriefing(context.Context, string, string) error {
	s.failed++

	return s.failErr
}

// derivationBriefingStopper はACP停止結果を固定する。
type derivationBriefingStopper struct {
	err error
}

// StopDerivationBriefing は停止呼出しを記録する。
func (s derivationBriefingStopper) StopDerivationBriefing(context.Context, string, string) error {
	return s.err
}
