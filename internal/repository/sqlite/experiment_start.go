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

// BeginExperiment は固定済み全promptのrunと開始操作を原子的に記録する。
func (s *Store) BeginExperiment(ctx context.Context, requestID, experimentID string) (domain.ExperimentStart, bool, error) {
	var lastErr error
	for attempt := range 12 {
		start, created, err := s.beginExperiment(ctx, requestID, experimentID)
		if !isSQLiteBusy(err) {
			return start, created, err
		}
		lastErr = err
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.ExperimentStart{}, false, err
		}
	}

	return domain.ExperimentStart{}, false, lastErr
}

// beginExperiment は一回のSQLite transactionで開始操作とrunを記録する。
func (s *Store) beginExperiment(ctx context.Context, requestID, experimentID string) (domain.ExperimentStart, bool, error) {
	existing, found, err := s.findExperimentStart(ctx, requestID)
	if err != nil {
		return domain.ExperimentStart{}, false, err
	}
	if found {
		if existing.ExperimentID != experimentID {
			return domain.ExperimentStart{}, false, apperr.New(apperr.CodeExperimentStartRequestInvalid)
		}

		return existing, false, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ExperimentStart{}, false, fmt.Errorf("begin experiment start: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	start, err := findStartableExperiment(ctx, tx, experimentID)
	if err != nil {
		return domain.ExperimentStart{}, false, err
	}
	start.RequestID = requestID
	start.OperationID, err = newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentStart{}, false, fmt.Errorf("generate experiment start operation ID: %w", err)
	}
	start.State = domain.ExperimentStartStateStarting
	now := time.Now().UTC()
	nowValue := now.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_start_operations (request_id, experiment_id, operation_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", requestID, experimentID, start.OperationID, start.State, nowValue, nowValue); err != nil {
		if isExperimentStartRequestConflict(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.ExperimentStart{}, false, fmt.Errorf("rollback experiment start conflict: %w", rollbackErr)
			}
			existing, found, findErr := s.findExperimentStart(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentStart{}, false, findErr
			}
			if found && existing.ExperimentID == experimentID {
				return existing, false, nil
			}
			if found {
				return domain.ExperimentStart{}, false, apperr.New(apperr.CodeExperimentStartRequestInvalid)
			}
		}

		return domain.ExperimentStart{}, false, fmt.Errorf("insert experiment start operation: %w", err)
	}
	for index := range start.FixedConditions.Prompts {
		runID, err := newBriefingIdentifier()
		if err != nil {
			return domain.ExperimentStart{}, false, fmt.Errorf("generate experiment run ID: %w", err)
		}
		run := domain.ExperimentWorkspaceRun{ID: runID, State: domain.ExperimentRunStateQueued, UpdatedAt: now}
		if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_runs (id, experiment_id, state, prompt_sequence_no, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", run.ID, experimentID, run.State, index+1, nowValue, nowValue); err != nil {
			return domain.ExperimentStart{}, false, fmt.Errorf("insert experiment run: %w", err)
		}
		start.Runs = append(start.Runs, run)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE experiments SET state = ?, updated_at = ? WHERE id = ?", domain.ExperimentStartStateRunning, nowValue, experimentID); err != nil {
		return domain.ExperimentStart{}, false, fmt.Errorf("transition experiment to running: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentStart{}, false, fmt.Errorf("commit experiment start: %w", err)
	}
	persisted, found, err := s.findExperimentStart(ctx, requestID)
	if err != nil {
		return domain.ExperimentStart{}, false, err
	}
	if !found {
		return domain.ExperimentStart{}, false, fmt.Errorf("find committed experiment start")
	}

	return persisted, true, nil
}

// MarkExperimentRunRunning は一件のrunをrunner呼出し中へ更新する。
func (s *Store) MarkExperimentRunRunning(ctx context.Context, runID string) error {
	return s.updateExperimentRun(ctx, runID, domain.ExperimentRunStateRunning, nil)
}

// CompleteExperimentRun は一件のrunの安全な要約を記録する。
func (s *Store) CompleteExperimentRun(ctx context.Context, runID, summary string) error {
	return s.updateExperimentRun(ctx, runID, domain.ExperimentRunStateCompleted, &summary)
}

// FailExperimentRun は一件のrunner失敗と開始操作失敗を記録する。
func (s *Store) FailExperimentRun(ctx context.Context, runID, failureCode string) error {
	if err := s.updateExperimentRun(ctx, runID, domain.ExperimentRunStateFailed, nil); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "UPDATE experiment_start_operations SET state = ?, failure_code = ?, updated_at = ? WHERE experiment_id = (SELECT experiment_id FROM experiment_runs WHERE id = ?)", domain.ExperimentStartStateFailed, failureCode, time.Now().UTC().Format(time.RFC3339Nano), runID); err != nil {
		return fmt.Errorf("fail experiment start operation: %w", err)
	}

	return nil
}

// CompleteExperimentStart はすべてのrunの開始要求完了を記録する。
func (s *Store) CompleteExperimentStart(ctx context.Context, requestID string) error {
	if _, err := s.db.ExecContext(ctx, "UPDATE experiment_start_operations SET state = ?, updated_at = ? WHERE request_id = ?", domain.ExperimentStartStateRunning, time.Now().UTC().Format(time.RFC3339Nano), requestID); err != nil {
		return fmt.Errorf("complete experiment start operation: %w", err)
	}

	return nil
}

// updateExperimentRun は一件のrun状態と安全な要約を更新する。
func (s *Store) updateExperimentRun(ctx context.Context, runID, state string, summary *string) error {
	if _, err := s.db.ExecContext(ctx, "UPDATE experiment_runs SET state = ?, summary = ?, updated_at = ? WHERE id = ?", state, summary, time.Now().UTC().Format(time.RFC3339Nano), runID); err != nil {
		return fmt.Errorf("update experiment run: %w", err)
	}

	return nil
}

// findStartableExperiment はready状態の固定条件をrun生成用に取得する。
func findStartableExperiment(ctx context.Context, tx *sql.Tx, experimentID string) (domain.ExperimentStart, error) {
	var start domain.ExperimentStart
	var fixedConditionID sql.NullString
	var hypothesis sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT e.state, e.fixed_condition_id, COALESCE(c.purpose, ''), c.hypothesis, COALESCE(c.environment_conditions, ''), COALESCE(c.initial_input, ''), COALESCE(c.evaluation_axes, '') FROM experiments e LEFT JOIN experiment_fixed_conditions c ON c.id = e.fixed_condition_id WHERE e.id = ?", experimentID).Scan(&start.State, &fixedConditionID, &start.FixedConditions.Purpose, &hypothesis, &start.FixedConditions.EnvironmentConditions, &start.FixedConditions.InitialInput, &start.FixedConditions.EvaluationAxes)
	if err == sql.ErrNoRows {
		return domain.ExperimentStart{}, apperr.New(apperr.CodeExperimentWorkspaceNotFound)
	}
	if err != nil {
		return domain.ExperimentStart{}, fmt.Errorf("find startable experiment: %w", err)
	}
	if start.State != "ready" || !fixedConditionID.Valid {
		return domain.ExperimentStart{}, apperr.New(apperr.CodeExperimentStartNotReady)
	}
	start.ExperimentID = experimentID
	start.FixedConditions.ExperimentID = experimentID
	start.FixedConditions.FixedConditionID = fixedConditionID.String
	if hypothesis.Valid {
		start.FixedConditions.Hypothesis = &hypothesis.String
	}
	rows, err := tx.QueryContext(ctx, "SELECT sequence_no, content FROM experiment_fixed_condition_prompts WHERE fixed_condition_id = ? ORDER BY sequence_no", fixedConditionID.String)
	if err != nil {
		return domain.ExperimentStart{}, fmt.Errorf("find startable experiment prompts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var prompt domain.ExperimentPreparationPrompt
		if err := rows.Scan(&prompt.SequenceNo, &prompt.Content); err != nil {
			return domain.ExperimentStart{}, fmt.Errorf("scan startable experiment prompt: %w", err)
		}
		start.FixedConditions.Prompts = append(start.FixedConditions.Prompts, prompt)
	}
	if err := rows.Err(); err != nil {
		return domain.ExperimentStart{}, fmt.Errorf("iterate startable experiment prompts: %w", err)
	}
	if len(start.FixedConditions.Prompts) == 0 {
		return domain.ExperimentStart{}, apperr.New(apperr.CodeExperimentStartNotReady)
	}

	return start, nil
}

// findExperimentStart はrequest IDに対応する開始結果とrun一覧を取得する。
func (s *Store) findExperimentStart(ctx context.Context, requestID string) (domain.ExperimentStart, bool, error) {
	var start domain.ExperimentStart
	err := s.db.QueryRowContext(ctx, "SELECT request_id, experiment_id, operation_id, state, COALESCE(failure_code, '') FROM experiment_start_operations WHERE request_id = ?", requestID).Scan(&start.RequestID, &start.ExperimentID, &start.OperationID, &start.State, &start.FailureCode)
	if err == sql.ErrNoRows {
		return domain.ExperimentStart{}, false, nil
	}
	if err != nil {
		return domain.ExperimentStart{}, false, fmt.Errorf("find experiment start: %w", err)
	}
	runs, err := s.findExperimentWorkspaceRuns(ctx, start.ExperimentID)
	if err != nil {
		return domain.ExperimentStart{}, false, err
	}
	start.Runs = runs
	workspace, found, err := s.GetExperimentWorkspace(ctx, start.ExperimentID)
	if err != nil {
		return domain.ExperimentStart{}, false, err
	}
	if !found {
		return domain.ExperimentStart{}, false, apperr.New(apperr.CodeExperimentWorkspaceNotFound)
	}
	start.FixedConditions = workspace.FixedConditions

	return start, true, nil
}

// isExperimentStartRequestConflict は開始request IDの一意制約競合を判定する。
func isExperimentStartRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_start_operations.request_id")
}
