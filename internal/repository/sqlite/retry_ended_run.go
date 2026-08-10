package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

const runRetryLimit = 14

// RetryEndedRun は失敗済みrunと同じ固定promptを持つqueued runを原子的に作成する。
func (s *Store) RetryEndedRun(ctx context.Context, requestID, sourceRunID string) (domain.ExperimentRunRetry, bool, error) {
	var lastErr error
	for attempt := range runRetryLimit {
		existing, found, err := s.findRunRetry(ctx, requestID)
		if isSQLiteBusy(err) {
			lastErr = err
		} else if err != nil {
			return domain.ExperimentRunRetry{}, false, err
		} else if found {
			if existing.SourceRunID != sourceRunID {
				return domain.ExperimentRunRetry{}, false, apperr.New(apperr.CodeRunRetryRequestConflict)
			}

			return existing, false, nil
		} else {
			retry, created, createErr := s.createRunRetry(ctx, requestID, sourceRunID)
			if !isSQLiteBusy(createErr) {
				return retry, created, createErr
			}
			lastErr = createErr
		}
		if attempt == runRetryLimit-1 {
			break
		}
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.ExperimentRunRetry{}, false, err
		}
	}

	return domain.ExperimentRunRetry{}, false, lastErr
}

// createRunRetry は一回のSQLite transactionで再実行用runと操作を記録する。
func (s *Store) createRunRetry(ctx context.Context, requestID, sourceRunID string) (domain.ExperimentRunRetry, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("begin run retry: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var experimentID string
	var state string
	var promptSequence sql.NullInt64
	var isolationKind sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT experiment_id, state, prompt_sequence_no, isolation_kind FROM experiment_runs WHERE id = ?", sourceRunID).Scan(&experimentID, &state, &promptSequence, &isolationKind)
	if err == sql.ErrNoRows {
		return domain.ExperimentRunRetry{}, false, apperr.New(apperr.CodeRunRetryNotFound)
	}
	if err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("find retry source run: %w", err)
	}
	if state != domain.ExperimentRunStateFailed {
		return domain.ExperimentRunRetry{}, false, apperr.New(apperr.CodeRunRetryNotAllowed)
	}

	retryRunID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("generate retry run ID: %w", err)
	}
	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("generate run retry operation ID: %w", err)
	}
	now := time.Now().UTC()
	nowValue := now.Format(time.RFC3339Nano)
	retry := domain.ExperimentRunRetry{
		RequestID:    requestID,
		SourceRunID:  sourceRunID,
		ExperimentID: experimentID,
		RetryRunID:   retryRunID,
		OperationID:  operationID,
		State:        domain.ExperimentRunStateQueued,
		CreatedAt:    now,
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO experiment_runs
		(id, experiment_id, operation_id, retry_of_run_id, state, prompt_sequence_no, isolation_kind, last_observed_at, reconciliation_state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		retry.RetryRunID, retry.ExperimentID, retry.OperationID, retry.SourceRunID, retry.State, promptSequence,
		isolationKind, nowValue, "confirmed", nowValue, nowValue); err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("insert retry run: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO experiment_run_retry_operations
		(request_id, source_run_id, experiment_id, run_id, operation_id, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		retry.RequestID, retry.SourceRunID, retry.ExperimentID, retry.RetryRunID, retry.OperationID, retry.State, nowValue, nowValue); err != nil {
		if isRunRetryRequestConflict(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.ExperimentRunRetry{}, false, fmt.Errorf("rollback run retry conflict: %w", rollbackErr)
			}
			existing, found, findErr := s.findRunRetry(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentRunRetry{}, false, findErr
			}
			if found && existing.SourceRunID == sourceRunID {
				return existing, false, nil
			}
			if found {
				return domain.ExperimentRunRetry{}, false, apperr.New(apperr.CodeRunRetryRequestConflict)
			}
		}

		return domain.ExperimentRunRetry{}, false, fmt.Errorf("insert run retry operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("commit run retry: %w", err)
	}

	persisted, found, err := s.findRunRetry(ctx, requestID)
	if err != nil {
		return domain.ExperimentRunRetry{}, false, err
	}
	if !found {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("find committed run retry")
	}

	return persisted, true, nil
}

// isRunRetryRequestConflict は再実行request IDの一意制約競合を判定する。
func isRunRetryRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_run_retry_operations.request_id")
}

// findRunRetry はrequest IDに対応する再実行結果を取得する。
func (s *Store) findRunRetry(ctx context.Context, requestID string) (domain.ExperimentRunRetry, bool, error) {
	var retry domain.ExperimentRunRetry
	var createdAt string
	err := s.db.QueryRowContext(ctx, `SELECT o.request_id, o.source_run_id, r.experiment_id, o.run_id, o.operation_id, o.state, o.created_at
		FROM experiment_run_retry_operations o
		JOIN experiment_runs r ON r.id = o.run_id
		WHERE o.request_id = ?`, requestID).Scan(
		&retry.RequestID,
		&retry.SourceRunID,
		&retry.ExperimentID,
		&retry.RetryRunID,
		&retry.OperationID,
		&retry.State,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return domain.ExperimentRunRetry{}, false, nil
	}
	if err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("find run retry: %w", err)
	}
	if retry.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt); err != nil {
		return domain.ExperimentRunRetry{}, false, fmt.Errorf("parse run retry creation time: %w", err)
	}

	return retry, true, nil
}
