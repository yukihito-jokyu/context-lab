package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// SaveExperimentPreparationDraft は編集下書きと冪等結果snapshotを同一transactionで保存。
func (s *Store) SaveExperimentPreparationDraft(ctx context.Context, draft domain.ExperimentPreparationDraft) (domain.ExperimentPreparationDraft, error) {
	var lastErr error
	for attempt := range 5 {
		saved, err := s.saveExperimentPreparationDraft(ctx, draft)
		if !isSQLiteBusy(err) {
			return saved, err
		}
		lastErr = err
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.ExperimentPreparationDraft{}, err
		}
	}

	return domain.ExperimentPreparationDraft{}, lastErr
}

// saveExperimentPreparationDraft は一回のSQLite transactionで下書きを保存。
func (s *Store) saveExperimentPreparationDraft(ctx context.Context, draft domain.ExperimentPreparationDraft) (domain.ExperimentPreparationDraft, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("begin draft transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, found, err := findDraftOperation(ctx, tx, draft.RequestID)
	if err != nil {
		return domain.ExperimentPreparationDraft{}, err
	}
	if found {
		if existing.ExperimentID != draft.ExperimentID {
			return domain.ExperimentPreparationDraft{}, apperr.New(apperr.CodeDraftRequestInvalid)
		}

		return existing, nil
	}

	var state string
	err = tx.QueryRowContext(ctx, "SELECT state FROM experiments WHERE id = ?", draft.ExperimentID).Scan(&state)
	if err == sql.ErrNoRows {
		return domain.ExperimentPreparationDraft{}, apperr.New(apperr.CodeExperimentPreparationNotFound)
	}
	if err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("find draft experiment: %w", err)
	}
	if state != "preparing" {
		return domain.ExperimentPreparationDraft{}, apperr.New(apperr.CodeExperimentPreparationNotEditable)
	}

	var preparationExists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM experiment_preparations WHERE experiment_id = ?)", draft.ExperimentID).Scan(&preparationExists); err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("find draft preparation: %w", err)
	}
	if !preparationExists {
		return domain.ExperimentPreparationDraft{}, apperr.New(apperr.CodeExperimentPreparationNotFound)
	}

	draft.SavedAt = time.Now().UTC()
	promptsJSON, err := json.Marshal(draft.Prompts)
	if err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("marshal draft prompts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE experiments SET purpose = ?, updated_at = ? WHERE id = ?", draft.Purpose, draft.SavedAt.Format(time.RFC3339Nano), draft.ExperimentID); err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("update draft experiment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE experiment_preparations SET hypothesis = ?, environment_conditions = ?, initial_input = ?, evaluation_criteria = ?, updated_at = ? WHERE experiment_id = ?", draft.Hypothesis, draft.EnvironmentConditions, draft.InitialInput, draft.EvaluationAxes, draft.SavedAt.Format(time.RFC3339Nano), draft.ExperimentID); err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("update draft preparation: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM experiment_preparation_prompts WHERE experiment_id = ?", draft.ExperimentID); err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("delete draft prompts: %w", err)
	}
	for index, prompt := range draft.Prompts {
		if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_preparation_prompts (experiment_id, sequence_no, content) VALUES (?, ?, ?)", draft.ExperimentID, index+1, prompt.Content); err != nil {
			return domain.ExperimentPreparationDraft{}, fmt.Errorf("insert draft prompt: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_preparation_draft_operations (request_id, experiment_id, purpose, hypothesis, environment_conditions, initial_input, evaluation_axes, prompts_json, saved_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", draft.RequestID, draft.ExperimentID, draft.Purpose, draft.Hypothesis, draft.EnvironmentConditions, draft.InitialInput, draft.EvaluationAxes, string(promptsJSON), draft.SavedAt.Format(time.RFC3339Nano)); err != nil {
		if isDraftRequestConflict(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.ExperimentPreparationDraft{}, fmt.Errorf("rollback draft transaction after request conflict: %w", rollbackErr)
			}
			existing, found, findErr := findDraftOperation(ctx, s.db, draft.RequestID)
			if findErr != nil {
				return domain.ExperimentPreparationDraft{}, findErr
			}
			if found && existing.ExperimentID == draft.ExperimentID {
				return existing, nil
			}
			if found {
				return domain.ExperimentPreparationDraft{}, apperr.New(apperr.CodeDraftRequestInvalid)
			}
		}

		return domain.ExperimentPreparationDraft{}, fmt.Errorf("insert draft operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentPreparationDraft{}, fmt.Errorf("commit draft transaction: %w", err)
	}

	return draft, nil
}

// findDraftOperation はrequest IDに対応する下書きsnapshotを取得。
func findDraftOperation(ctx context.Context, queryer draftOperationQueryer, requestID string) (domain.ExperimentPreparationDraft, bool, error) {
	var draft domain.ExperimentPreparationDraft
	var hypothesis sql.NullString
	var promptsJSON string
	var savedAt string
	err := queryer.QueryRowContext(ctx, "SELECT request_id, experiment_id, purpose, hypothesis, environment_conditions, initial_input, evaluation_axes, prompts_json, saved_at FROM experiment_preparation_draft_operations WHERE request_id = ?", requestID).Scan(&draft.RequestID, &draft.ExperimentID, &draft.Purpose, &hypothesis, &draft.EnvironmentConditions, &draft.InitialInput, &draft.EvaluationAxes, &promptsJSON, &savedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentPreparationDraft{}, false, nil
	}
	if err != nil {
		return domain.ExperimentPreparationDraft{}, false, fmt.Errorf("find draft operation: %w", err)
	}
	if hypothesis.Valid {
		draft.Hypothesis = &hypothesis.String
	}
	if err := json.Unmarshal([]byte(promptsJSON), &draft.Prompts); err != nil {
		return domain.ExperimentPreparationDraft{}, false, fmt.Errorf("unmarshal draft prompts: %w", err)
	}
	if err := parseDraftSavedAt(savedAt, &draft); err != nil {
		return domain.ExperimentPreparationDraft{}, false, err
	}

	return draft, true, nil
}

// draftOperationQueryer は下書き冪等snapshotの読込境界。
type draftOperationQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// isDraftRequestConflict は下書きrequest IDの一意制約競合を判定。
func isDraftRequestConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_preparation_draft_operations.request_id")
}

// isSQLiteBusy は再試行で解消できるSQLiteの一時的な競合を判定。
func isSQLiteBusy(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "database is locked") || strings.Contains(message, "database is busy") || strings.Contains(message, "SQLITE_BUSY")
}

// waitDraftSaveRetry はSQLite競合後の再試行をcontextに従って待機。
func waitDraftSaveRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// parseDraftSavedAt は保存時刻をUTCへ正規化。
func parseDraftSavedAt(value string, draft *domain.ExperimentPreparationDraft) error {
	savedAt, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return fmt.Errorf("parse draft saved time: %w", err)
	}
	draft.SavedAt = savedAt.UTC()

	return nil
}
