package wails

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// RetryEndedRun handlerの成功DTOと安全な失敗返却を検証する。
func TestExperimentRunRetriesHandlerRetryEndedRun(t *testing.T) {
	createdAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	tests := []struct {
		name     string
		request  RetryEndedRunRequest
		store    handlerRetryEndedRunStore
		wantCode apperr.Code
	}{
		{
			name: "安全なqueued runを返す",
			request: RetryEndedRunRequest{
				RequestID: "request-1",
				RunID:     "run-1",
			},
			store: handlerRetryEndedRunStore{
				retry: domain.ExperimentRunRetry{
					SourceRunID:  "run-1",
					ExperimentID: "experiment-1",
					RetryRunID:   "retry-1",
					OperationID:  "operation-1",
					State:        "queued",
					CreatedAt:    createdAt,
				},
			},
		},
		{
			name: "内部詳細を漏らさない",
			request: RetryEndedRunRequest{
				RequestID: "request-1",
				RunID:     "run-1",
			},
			store: handlerRetryEndedRunStore{
				err: errors.New("private docker ID"),
			},
			wantCode: apperr.CodeRunRetryUnavailable,
		},
		{
			name:     "入力不正を返す",
			request:  RetryEndedRunRequest{},
			wantCode: apperr.CodeRunRetryRequestInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExperimentRunRetriesHandler(usecase.NewRetryEndedRun(&tt.store), newTestLogger()).RetryEndedRun(tt.request)
			if tt.wantCode != "" {
				if got.Error == nil || got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error = %+v, want code %q", got.Error, tt.wantCode)
				}
				if got.Error != nil && strings.Contains(got.Error.Message, "docker") {
					t.Errorf("Error.Message = %q, want no private detail", got.Error.Message)
				}

				return
			}
			if got.Data == nil {
				t.Fatal("Data = nil, want retry result")
			}
			if got.Data.RetryRunID != "retry-1" || got.Data.State != "queued" || got.Data.ExperimentID != "experiment-1" {
				t.Errorf("Data = %+v, want safe retry result", got.Data)
			}
			if got.Data.CreatedAt.Location() != time.UTC {
				t.Errorf("CreatedAt location = %s, want UTC", got.Data.CreatedAt.Location())
			}
		})
	}
}

// RetryEndedRun handlerの依存欠落を検証する。
func TestExperimentRunRetriesHandlerFailureFallback(t *testing.T) {
	got := NewExperimentRunRetriesHandler(nil, newTestLogger()).RetryEndedRun(RetryEndedRunRequest{
		RequestID: "request-1",
		RunID:     "run-1",
	})
	if got.Error == nil || got.Error.Code != string(apperr.CodeRunRetryUnavailable) {
		t.Errorf("Error = %+v, want unavailable error", got.Error)
	}
}

// failRetryEndedRunの未知エラー変換を検証する。
func TestFailRetryEndedRun(t *testing.T) {
	got := failRetryEndedRun(errors.New("private credential"))
	if got.Error == nil || got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error = %+v, want unexpected error", got.Error)
	}
}

// handlerRetryEndedRunStore はRetryEndedRun handler用test double。
type handlerRetryEndedRunStore struct {
	retry domain.ExperimentRunRetry
	err   error
}

// RetryEndedRun は設定済みの再実行結果を返す。
func (s *handlerRetryEndedRunStore) RetryEndedRun(context.Context, string, string) (domain.ExperimentRunRetry, bool, error) {
	return s.retry, true, s.err
}
