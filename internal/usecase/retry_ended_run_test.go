package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// RetryEndedRunの入力検証、idempotency、エラー契約を検証する。
func TestRetryEndedRunExecute(t *testing.T) {
	createdAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name     string
		request  string
		runID    string
		store    retryEndedRunStore
		wantCode apperr.Code
		wantID   string
	}{
		{
			name:    "queued runを返す",
			request: " request-1 ",
			runID:   " run-1 ",
			store: retryEndedRunStore{
				retry: domain.ExperimentRunRetry{
					RequestID:  "request-1",
					RetryRunID: "retry-1",
					State:      domain.ExperimentRunStateQueued,
					CreatedAt:  createdAt,
				},
			},
			wantID: "retry-1",
		},
		{
			name:     "空requestを拒否する",
			runID:    "run-1",
			wantCode: apperr.CodeRunRetryRequestInvalid,
		},
		{
			name:     "空runを拒否する",
			request:  "request-1",
			wantCode: apperr.CodeRunRetryRequestInvalid,
		},
		{
			name:    "アプリケーションエラーを維持する",
			request: "request-1",
			runID:   "run-1",
			store: retryEndedRunStore{
				err: apperr.New(apperr.CodeRunRetryNotAllowed),
			},
			wantCode: apperr.CodeRunRetryNotAllowed,
		},
		{
			name:    "内部エラーを安全に変換する",
			request: "request-1",
			runID:   "run-1",
			store: retryEndedRunStore{
				err: errors.New("private sqlite error"),
			},
			wantCode: apperr.CodeRunRetryUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retry, err := NewRetryEndedRun(&tt.store).Execute(context.Background(), tt.request, tt.runID)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if retry.RetryRunID != tt.wantID {
				t.Errorf("RetryRunID = %q, want %q", retry.RetryRunID, tt.wantID)
			}
			if tt.store.requestID != "request-1" || tt.store.runID != "run-1" {
				t.Errorf("store arguments = (%q, %q), want trimmed values", tt.store.requestID, tt.store.runID)
			}
		})
	}
}

// retryEndedRunStore はRetryEndedRun用test double。
type retryEndedRunStore struct {
	retry     domain.ExperimentRunRetry
	err       error
	requestID string
	runID     string
}

// RetryEndedRun は設定済みの再実行結果を返す。
func (s *retryEndedRunStore) RetryEndedRun(_ context.Context, requestID, runID string) (domain.ExperimentRunRetry, bool, error) {
	s.requestID = requestID
	s.runID = runID

	return s.retry, true, s.err
}
