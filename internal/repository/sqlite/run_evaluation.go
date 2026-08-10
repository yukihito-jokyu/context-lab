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

const runEvaluationRetryLimit = 14

// BeginRunEvaluation は完了済みrunの評価操作を原子的に記録する。
func (s *Store) BeginRunEvaluation(ctx context.Context, requestID, runID string) (domain.ExperimentRunEvaluation, bool, error) {
	var lastErr error
	for attempt := range runEvaluationRetryLimit {
		existing, found, err := s.findRunEvaluation(ctx, requestID)
		if isSQLiteBusy(err) {
			lastErr = err
		} else if err != nil {
			return domain.ExperimentRunEvaluation{}, false, err
		} else if found {
			if existing.RunID != runID {
				return domain.ExperimentRunEvaluation{}, false, apperr.New(apperr.CodeRunEvaluationRequestInvalid)
			}

			return existing, false, nil
		} else {
			evaluation, created, beginErr := s.beginRunEvaluation(ctx, requestID, runID)
			if !isSQLiteBusy(beginErr) {
				return evaluation, created, beginErr
			}
			lastErr = beginErr
		}
		if attempt == runEvaluationRetryLimit-1 {
			break
		}
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.ExperimentRunEvaluation{}, false, err
		}
	}

	return domain.ExperimentRunEvaluation{}, false, lastErr
}

// beginRunEvaluation は一回のSQLite transactionで評価操作を生成する。
func (s *Store) beginRunEvaluation(ctx context.Context, requestID, runID string) (domain.ExperimentRunEvaluation, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("begin run evaluation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	evaluation, err := findEvaluableRun(ctx, tx, runID)
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, err
	}
	evaluation.RequestID = requestID
	evaluation.EvaluationID, err = newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("generate run evaluation ID: %w", err)
	}
	evaluation.OperationID, err = newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("generate run evaluation operation ID: %w", err)
	}
	evaluation.State = domain.ExperimentEvaluationStateStarting
	now := time.Now().UTC()
	nowValue := now.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_evaluations (id, experiment_id, run_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", evaluation.EvaluationID, evaluation.ExperimentID, evaluation.RunID, evaluation.State, nowValue, nowValue); err != nil {
		if isRunEvaluationAlreadyExists(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("rollback duplicate run evaluation: %w", rollbackErr)
			}
			existing, found, findErr := s.findRunEvaluation(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentRunEvaluation{}, false, findErr
			}
			if found && existing.RunID == runID {
				return existing, false, nil
			}

			return domain.ExperimentRunEvaluation{}, false, apperr.New(apperr.CodeRunEvaluationAlreadyExists)
		}

		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("insert run evaluation: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_evaluation_operations (request_id, run_id, evaluation_id, operation_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", requestID, evaluation.RunID, evaluation.EvaluationID, evaluation.OperationID, evaluation.State, nowValue, nowValue); err != nil {
		if isRunEvaluationRequestConflict(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("rollback run evaluation conflict: %w", rollbackErr)
			}
			existing, found, findErr := s.findRunEvaluation(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentRunEvaluation{}, false, findErr
			}
			if found && existing.RunID == runID {
				return existing, false, nil
			}
			if found {
				return domain.ExperimentRunEvaluation{}, false, apperr.New(apperr.CodeRunEvaluationRequestInvalid)
			}
		}
		if isRunEvaluationAlreadyExists(err) {
			return domain.ExperimentRunEvaluation{}, false, apperr.New(apperr.CodeRunEvaluationAlreadyExists)
		}

		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("insert run evaluation operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("commit run evaluation: %w", err)
	}
	persisted, found, err := s.findRunEvaluation(ctx, requestID)
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, err
	}
	if !found {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("find committed run evaluation")
	}

	return persisted, true, nil
}

// CompleteRunEvaluation は評価の安全な要約と完了状態を記録する。
func (s *Store) CompleteRunEvaluation(ctx context.Context, evaluationID, summary string) error {
	return s.updateRunEvaluation(ctx, evaluationID, domain.ExperimentEvaluationStateCompleted, summary, "")
}

// FailRunEvaluation は評価runner失敗と操作失敗を記録する。
func (s *Store) FailRunEvaluation(ctx context.Context, evaluationID, failureCode string) error {
	return s.updateRunEvaluation(ctx, evaluationID, domain.ExperimentEvaluationStateFailed, "", failureCode)
}

// updateRunEvaluation はevaluationとoperationの状態を一貫して更新する。
func (s *Store) updateRunEvaluation(ctx context.Context, evaluationID, state, summary, failureCode string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update run evaluation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var summaryValue any
	if strings.TrimSpace(summary) == "" {
		summaryValue = nil
	} else {
		summaryValue = summary
	}
	if _, err := tx.ExecContext(ctx, "UPDATE experiment_evaluations SET state = ?, summary = ?, updated_at = ? WHERE id = ?", state, summaryValue, now, evaluationID); err != nil {
		return fmt.Errorf("update run evaluation: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE experiment_evaluation_operations SET state = ?, failure_code = ?, updated_at = ? WHERE evaluation_id = ?", state, nullableFailureCode(failureCode), now, evaluationID); err != nil {
		return fmt.Errorf("update run evaluation operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update run evaluation: %w", err)
	}

	return nil
}

// findEvaluableRun は評価可能な完了済みrunと固定評価軸を取得する。
func findEvaluableRun(ctx context.Context, tx *sql.Tx, runID string) (domain.ExperimentRunEvaluation, error) {
	var evaluation domain.ExperimentRunEvaluation
	var summary sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT r.id, r.experiment_id, r.state, r.summary, c.purpose, c.evaluation_axes FROM experiment_runs r JOIN experiments e ON e.id = r.experiment_id JOIN experiment_fixed_conditions c ON c.id = e.fixed_condition_id WHERE r.id = ?", runID).Scan(&evaluation.RunID, &evaluation.ExperimentID, &evaluation.State, &summary, &evaluation.Purpose, &evaluation.EvaluationAxes)
	if err == sql.ErrNoRows {
		return domain.ExperimentRunEvaluation{}, apperr.New(apperr.CodeRunEvaluationNotReady)
	}
	if err != nil {
		return domain.ExperimentRunEvaluation{}, fmt.Errorf("find evaluable run: %w", err)
	}
	if evaluation.State != domain.ExperimentRunStateCompleted || !summary.Valid || strings.TrimSpace(summary.String) == "" {
		return domain.ExperimentRunEvaluation{}, apperr.New(apperr.CodeRunEvaluationNotReady)
	}
	var existing string
	err = tx.QueryRowContext(ctx, "SELECT id FROM experiment_evaluations WHERE run_id = ?", runID).Scan(&existing)
	if err == nil {
		return domain.ExperimentRunEvaluation{}, apperr.New(apperr.CodeRunEvaluationAlreadyExists)
	}
	if err != sql.ErrNoRows {
		return domain.ExperimentRunEvaluation{}, fmt.Errorf("find existing run evaluation: %w", err)
	}
	evaluation.RunSummary = summary.String

	return evaluation, nil
}

// findRunEvaluation はrequest IDに対応する評価結果を取得する。
func (s *Store) findRunEvaluation(ctx context.Context, requestID string) (domain.ExperimentRunEvaluation, bool, error) {
	var evaluation domain.ExperimentRunEvaluation
	var summary sql.NullString
	var updatedAt string
	err := s.db.QueryRowContext(ctx, "SELECT o.request_id, o.run_id, o.evaluation_id, o.operation_id, o.state, COALESCE(o.failure_code, ''), e.summary, e.updated_at FROM experiment_evaluation_operations o JOIN experiment_evaluations e ON e.id = o.evaluation_id WHERE o.request_id = ?", requestID).Scan(&evaluation.RequestID, &evaluation.RunID, &evaluation.EvaluationID, &evaluation.OperationID, &evaluation.State, &evaluation.FailureCode, &summary, &updatedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentRunEvaluation{}, false, nil
	}
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("find run evaluation: %w", err)
	}
	if summary.Valid {
		evaluation.Summary = &summary.String
	}
	evaluation.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.ExperimentRunEvaluation{}, false, fmt.Errorf("parse run evaluation update time: %w", err)
	}

	return evaluation, true, nil
}

// nullableFailureCode は空の失敗コードをNULLとして保存する。
func nullableFailureCode(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return value
}

// isRunEvaluationRequestConflict は評価request IDの一意制約競合を判定する。
func isRunEvaluationRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_evaluation_operations.request_id")
}

// isRunEvaluationAlreadyExists はrunごとの評価一意制約競合を判定する。
func isRunEvaluationAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_evaluations.run_id")
}
