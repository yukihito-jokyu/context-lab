package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// EndedRunRetryStore は終了runから再実行用runを作成するport。
type EndedRunRetryStore interface {
	RetryEndedRun(context.Context, string, string) (domain.ExperimentRunRetry, bool, error)
}

// RetryEndedRun は終了runから再実行用runを作成するcommand。
type RetryEndedRun struct {
	store EndedRunRetryStore
}

// NewRetryEndedRun は再実行用run作成commandを生成する。
func NewRetryEndedRun(store EndedRunRetryStore) *RetryEndedRun {
	return &RetryEndedRun{store: store}
}

// Execute は失敗済みrunと同じ固定条件を持つqueued runを作成する。
func (u *RetryEndedRun) Execute(ctx context.Context, requestID, runID string) (domain.ExperimentRunRetry, error) {
	requestID = strings.TrimSpace(requestID)
	runID = strings.TrimSpace(runID)
	if requestID == "" || runID == "" {
		return domain.ExperimentRunRetry{}, apperr.New(apperr.CodeRunRetryRequestInvalid)
	}

	retry, _, err := u.store.RetryEndedRun(ctx, requestID, runID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentRunRetry{}, appErr
		}

		return domain.ExperimentRunRetry{}, apperr.Wrap(apperr.CodeRunRetryUnavailable, err)
	}

	return retry, nil
}
