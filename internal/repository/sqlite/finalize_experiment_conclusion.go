package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

var experimentConclusionRetryLimit = 14

// FinalizeExperimentConclusion は評価snapshotを根拠に実験結論を原子的に確定する。
func (s *Store) FinalizeExperimentConclusion(ctx context.Context, requestID, experimentID, conclusion string) (domain.ExperimentConclusion, bool, error) {
	var lastErr error
	for attempt := range experimentConclusionRetryLimit {
		finalized, created, err := s.finalizeExperimentConclusion(ctx, requestID, experimentID, conclusion)
		if !isSQLiteBusy(err) && !isExperimentConclusionRequestConflict(err) {
			return finalized, created, err
		}
		lastErr = err
		if existing, found, findErr := findExperimentConclusionOperation(ctx, s.db, requestID); findErr == nil && found {
			if existing.ExperimentID != experimentID || existing.Conclusion != conclusion {
				return domain.ExperimentConclusion{}, false, apperr.New(apperr.CodeExperimentConclusionRequestConflict)
			}
			return existing, false, nil
		} else if findErr != nil && !isSQLiteBusy(findErr) {
			return domain.ExperimentConclusion{}, false, findErr
		}
		if attempt == experimentConclusionRetryLimit-1 {
			break
		}
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.ExperimentConclusion{}, false, err
		}
	}
	return domain.ExperimentConclusion{}, false, lastErr
}

// finalizeExperimentConclusion は一回のSQLite transactionで実験結論を確定する。
func (s *Store) finalizeExperimentConclusion(ctx context.Context, requestID, experimentID, conclusion string) (domain.ExperimentConclusion, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("begin experiment conclusion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, found, err := findExperimentConclusionOperation(ctx, tx, requestID)
	if err != nil {
		return domain.ExperimentConclusion{}, false, err
	}
	if found {
		if existing.ExperimentID != experimentID || existing.Conclusion != conclusion {
			return domain.ExperimentConclusion{}, false, apperr.New(apperr.CodeExperimentConclusionRequestConflict)
		}
		return existing, false, nil
	}

	var fixedConditionID sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT fixed_condition_id FROM experiments WHERE id = ?", experimentID).Scan(&fixedConditionID)
	if err == sql.ErrNoRows {
		return domain.ExperimentConclusion{}, false, apperr.New(apperr.CodeExperimentConclusionNotFound)
	}
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("find conclusion experiment: %w", err)
	}
	if !fixedConditionID.Valid {
		return domain.ExperimentConclusion{}, false, apperr.New(apperr.CodeExperimentConclusionNotReady)
	}

	persisted, foundConclusion, err := findExperimentConclusion(ctx, tx, experimentID)
	if err != nil {
		return domain.ExperimentConclusion{}, false, err
	}
	if foundConclusion {
		if persisted.Conclusion != conclusion {
			return domain.ExperimentConclusion{}, false, apperr.New(apperr.CodeExperimentConclusionAlreadyFinalized)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_conclusion_operations (request_id, experiment_id, conclusion_id, conclusion, evaluation_snapshot_digest, finalized_at) VALUES (?, ?, ?, ?, ?, ?)", requestID, experimentID, persisted.ConclusionID, conclusion, persisted.EvaluationSnapshotDigest, persisted.FinalizedAt.Format(time.RFC3339Nano)); err != nil {
			return domain.ExperimentConclusion{}, false, fmt.Errorf("insert conclusion replay operation: %w", err)
		}
		persisted.RequestID = requestID
		if err := tx.Commit(); err != nil {
			return domain.ExperimentConclusion{}, false, fmt.Errorf("commit conclusion replay: %w", err)
		}
		return persisted, false, nil
	}

	digest, ready, err := experimentEvaluationSnapshotDigest(ctx, tx, experimentID)
	if err != nil {
		return domain.ExperimentConclusion{}, false, err
	}
	if !ready {
		return domain.ExperimentConclusion{}, false, apperr.New(apperr.CodeExperimentConclusionNotReady)
	}
	conclusionID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("generate experiment conclusion ID: %w", err)
	}
	now := time.Now().UTC()
	nowValue := now.Format(time.RFC3339Nano)
	finalized := domain.ExperimentConclusion{RequestID: requestID, ExperimentID: experimentID, ConclusionID: conclusionID, Conclusion: conclusion, State: "finalized", FinalizedAt: now, EvaluationSnapshotDigest: digest}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_conclusions (id, experiment_id, conclusion, evaluation_snapshot_digest, state, finalized_at) VALUES (?, ?, ?, ?, ?, ?)", conclusionID, experimentID, conclusion, digest, finalized.State, nowValue); err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("insert experiment conclusion: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_conclusion_operations (request_id, experiment_id, conclusion_id, conclusion, evaluation_snapshot_digest, finalized_at) VALUES (?, ?, ?, ?, ?, ?)", requestID, experimentID, conclusionID, conclusion, digest, nowValue); err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("insert experiment conclusion operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("commit experiment conclusion: %w", err)
	}
	return finalized, true, nil
}

// findExperimentConclusionOperation はrequest IDに対応する結論snapshotを取得する。
func findExperimentConclusionOperation(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, requestID string) (domain.ExperimentConclusion, bool, error) {
	var conclusion domain.ExperimentConclusion
	var finalizedAt string
	err := queryer.QueryRowContext(ctx, "SELECT request_id, experiment_id, conclusion_id, conclusion, evaluation_snapshot_digest, finalized_at FROM experiment_conclusion_operations WHERE request_id = ?", requestID).Scan(&conclusion.RequestID, &conclusion.ExperimentID, &conclusion.ConclusionID, &conclusion.Conclusion, &conclusion.EvaluationSnapshotDigest, &finalizedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentConclusion{}, false, nil
	}
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("find experiment conclusion operation: %w", err)
	}
	conclusion.FinalizedAt, err = time.Parse(time.RFC3339Nano, finalizedAt)
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("parse conclusion finalized time: %w", err)
	}
	conclusion.State = "finalized"
	return conclusion, true, nil
}

// findExperimentConclusion は実験に確定済みの結論を取得する。
func findExperimentConclusion(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, experimentID string) (domain.ExperimentConclusion, bool, error) {
	var conclusion domain.ExperimentConclusion
	var finalizedAt string
	err := queryer.QueryRowContext(ctx, "SELECT id, conclusion, evaluation_snapshot_digest, state, finalized_at FROM experiment_conclusions WHERE experiment_id = ?", experimentID).Scan(&conclusion.ConclusionID, &conclusion.Conclusion, &conclusion.EvaluationSnapshotDigest, &conclusion.State, &finalizedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentConclusion{}, false, nil
	}
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("find experiment conclusion: %w", err)
	}
	conclusion.ExperimentID = experimentID
	conclusion.FinalizedAt, err = time.Parse(time.RFC3339Nano, finalizedAt)
	if err != nil {
		return domain.ExperimentConclusion{}, false, fmt.Errorf("parse experiment conclusion time: %w", err)
	}
	return conclusion, true, nil
}

// isExperimentConclusionRequestConflict は同一結論へ収束可能な一意競合を判定する。
func isExperimentConclusionRequestConflict(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "UNIQUE constraint failed: experiment_conclusion_operations.request_id") ||
		strings.Contains(message, "UNIQUE constraint failed: experiment_conclusions.experiment_id")
}

// experimentEvaluationSnapshotDigest は結論確定可能な評価正本のdigestを返す。
func experimentEvaluationSnapshotDigest(ctx context.Context, queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, experimentID string) (string, bool, error) {
	rows, err := queryer.QueryContext(ctx, "SELECT id, run_id, state, COALESCE(result_status, 'notRecorded'), COALESCE(summary, ''), COALESCE(result_reason_code, ''), COALESCE(reconciliation_state, 'confirmed'), COALESCE(last_observed_at, '') FROM experiment_evaluations WHERE experiment_id = ? ORDER BY created_at, id", experimentID)
	if err != nil {
		return "", false, fmt.Errorf("find conclusion evaluations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var values []string
	recorded := false
	for rows.Next() {
		var id, runID, state, status, summary, reason, reconciliation, observed string
		if err := rows.Scan(&id, &runID, &state, &status, &summary, &reason, &reconciliation, &observed); err != nil {
			return "", false, fmt.Errorf("scan conclusion evaluation: %w", err)
		}
		if reconciliation != "confirmed" || (state != "completed" && state != "failed") {
			return "", false, nil
		}
		if status != "notRecorded" {
			recorded = true
		}
		values = append(values, strings.Join([]string{id, runID, state, status, summary, reason, reconciliation, observed}, "\x1f"))
	}
	if err := rows.Err(); err != nil {
		return "", false, fmt.Errorf("iterate conclusion evaluations: %w", err)
	}
	if len(values) == 0 || !recorded {
		return "", false, nil
	}
	sum := sha256.Sum256([]byte(strings.Join(values, "\x1e")))
	return hex.EncodeToString(sum[:]), true, nil
}
