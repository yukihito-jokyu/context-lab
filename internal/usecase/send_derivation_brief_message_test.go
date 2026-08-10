package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SendDerivationBriefMessageの入力検証、再生、失敗状態を確認。
func TestSendDerivationBriefMessageExecute(t *testing.T) {
	tests := []struct {
		name       string
		requestID  string
		operation  domain.DerivationBriefingMessageOperation
		senderErr  error
		wantCode   apperr.Code
		wantCalls  int
		wantFailed bool
	}{
		{
			name:      "空白入力を拒否する",
			requestID: " ",
			wantCode:  apperr.CodeDerivationBriefingMessageInvalid,
		},
		{
			name:      "開始中の送信を保留として返す",
			requestID: "request-pending",
			operation: domain.DerivationBriefingMessageOperation{
				State: domain.BriefingStartStateStarting,
			},
			wantCode: apperr.CodeDerivationBriefingMessagePending,
		},
		{
			name:      "既存失敗を安全に再生する",
			requestID: "request-failed",
			operation: domain.DerivationBriefingMessageOperation{
				State:       domain.BriefingStartStateFailed,
				FailureCode: string(apperr.CodeACPNotReady),
			},
			wantCode: apperr.CodeACPNotReady,
		},
		{
			name:       "ACP未準備を保存して安全に返す",
			requestID:  "request-acp",
			senderErr:  apperr.New(apperr.CodeACPNotReady),
			wantCode:   apperr.CodeACPNotReady,
			wantCalls:  1,
			wantFailed: true,
		},
		{
			name:      "成功を再生する",
			requestID: "request-success",
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &derivationBriefingMessageStore{operation: tt.operation}
			sender := &derivationBriefingMessageSender{err: tt.senderErr}
			command := NewSendDerivationBriefMessage(store, sender)

			got, err := command.Execute(context.Background(), tt.requestID, "session-1", "派生条件を比較したい")
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)
			} else {
				if err != nil {
					t.Fatalf("Execute() error = %v", err)
				}
				second, secondErr := command.Execute(context.Background(), tt.requestID, "session-1", "派生条件を比較したい")
				if secondErr != nil {
					t.Fatalf("second Execute() error = %v", secondErr)
				}
				if second.OperationID != got.OperationID {
					t.Errorf("second OperationID = %q, want %q", second.OperationID, got.OperationID)
				}
			}
			if gotCalls := sender.calls; gotCalls != tt.wantCalls {
				t.Errorf("SendDerivationBriefMessage() calls = %d, want %d", gotCalls, tt.wantCalls)
			}
			if gotFailed := store.failed; gotFailed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", gotFailed, tt.wantFailed)
			}
		})
	}
}

// DerivationBriefingMessageStoreの失敗変換を確認。
func TestSendDerivationBriefMessageFailures(t *testing.T) {
	tests := []struct {
		name     string
		beginErr error
		complete error
		failErr  error
		sender   error
		wantCode apperr.Code
	}{
		{
			name:     "開始記録の未知エラーを変換する",
			beginErr: errors.New("private sqlite"),
			wantCode: apperr.CodeDerivationBriefingMessageFailed,
		},
		{
			name:     "開始記録の安全なエラーを保持する",
			beginErr: apperr.New(apperr.CodeDerivationBriefingMessageNotActive),
			wantCode: apperr.CodeDerivationBriefingMessageNotActive,
		},
		{
			name:     "失敗記録エラーを変換する",
			failErr:  errors.New("private sqlite"),
			sender:   errors.New("private ACP"),
			wantCode: apperr.CodeDerivationBriefingMessageFailed,
		},
		{
			name:     "完了記録エラーを変換する",
			complete: errors.New("private sqlite"),
			wantCode: apperr.CodeDerivationBriefingMessageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &derivationBriefingMessageStore{
				beginErr:    tt.beginErr,
				completeErr: tt.complete,
				failErr:     tt.failErr,
			}
			command := NewSendDerivationBriefMessage(store, &derivationBriefingMessageSender{err: tt.sender})

			_, err := command.Execute(context.Background(), "request-1", "session-1", "message")
			assertBriefingErrorCode(t, err, tt.wantCode)
		})
	}
}

// derivationBriefingMessageFailure は保存済みの安全な失敗コードだけを再生する。
func TestDerivationBriefingMessageFailure(t *testing.T) {
	tests := []struct {
		name string
		code string
		want apperr.Code
	}{
		{
			name: "ACP未準備を再生する",
			code: string(apperr.CodeACPNotReady),
			want: apperr.CodeACPNotReady,
		},
		{
			name: "入力不正を再生する",
			code: string(apperr.CodeDerivationBriefingMessageInvalid),
			want: apperr.CodeDerivationBriefingMessageInvalid,
		},
		{
			name: "対象不在を再生する",
			code: string(apperr.CodeDerivationBriefingMessageNotFound),
			want: apperr.CodeDerivationBriefingMessageNotFound,
		},
		{
			name: "未開始を再生する",
			code: string(apperr.CodeDerivationBriefingMessageNotActive),
			want: apperr.CodeDerivationBriefingMessageNotActive,
		},
		{
			name: "未知の失敗を安全な失敗へ変換する",
			code: "private failure",
			want: apperr.CodeDerivationBriefingMessageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBriefingErrorCode(t, derivationBriefingMessageFailure(tt.code), tt.want)
		})
	}
}

// derivationBriefingMessageStore は派生会話送信portのtest double。
type derivationBriefingMessageStore struct {
	operation   domain.DerivationBriefingMessageOperation
	beginErr    error
	completeErr error
	failErr     error
	failed      bool
}

// BeginDerivationBriefMessage は送信操作を生成または再利用する。
func (s *derivationBriefingMessageStore) BeginDerivationBriefMessage(_ context.Context, requestID, briefingSessionID string) (domain.DerivationBriefingMessageOperation, bool, error) {
	if s.beginErr != nil {
		return domain.DerivationBriefingMessageOperation{}, false, s.beginErr
	}
	if s.operation.State != "" {
		return s.operation, false, nil
	}
	s.operation = domain.DerivationBriefingMessageOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "operation-" + requestID,
		State:             domain.BriefingStartStateStarting,
	}

	return s.operation, true, nil
}

// CompleteDerivationBriefMessage は送信完了を記録する。
func (s *derivationBriefingMessageStore) CompleteDerivationBriefMessage(_ context.Context, _ string, _ string, _ domain.DerivationBriefingMessageResult) error {
	if s.completeErr != nil {
		return s.completeErr
	}
	s.operation.State = domain.BriefingStartStateStarted

	return nil
}

// FailDerivationBriefMessage は送信失敗を記録する。
func (s *derivationBriefingMessageStore) FailDerivationBriefMessage(_ context.Context, _ string, _ string) error {
	s.failed = true

	return s.failErr
}

// derivationBriefingMessageSender は派生会話送信portのtest double。
type derivationBriefingMessageSender struct {
	err   error
	calls int
}

// SendDerivationBriefMessage は指定済み応答を返す。
func (s *derivationBriefingMessageSender) SendDerivationBriefMessage(context.Context, string, string, string) (domain.DerivationBriefingMessageResult, error) {
	s.calls++

	return domain.DerivationBriefingMessageResult{}, s.err
}
