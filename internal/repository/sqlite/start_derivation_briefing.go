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

// BeginDerivationBriefing は適格な派生元に紐付く開始操作とsessionを原子的に保存する。
func (s *Store) BeginDerivationBriefing(ctx context.Context, requestID, sourceExperimentID string) (domain.DerivationBriefingStart, bool, error) {
	existing, found, err := s.findDerivationBriefing(ctx, requestID)
	if err != nil {
		return domain.DerivationBriefingStart{}, false, err
	}
	if found {
		return replayDerivationBriefing(existing, sourceExperimentID)
	}

	briefingSessionID, err := newBriefingIdentifier()
	if err != nil {
		return domain.DerivationBriefingStart{}, false, fmt.Errorf("generate derivation briefing session ID: %w", err)
	}
	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.DerivationBriefingStart{}, false, fmt.Errorf("generate derivation briefing operation ID: %w", err)
	}
	start := domain.DerivationBriefingStart{
		RequestID:          requestID,
		SourceExperimentID: sourceExperimentID,
		BriefingSessionID:  briefingSessionID,
		OperationID:        operationID,
		State:              domain.BriefingStartStateStarting,
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.DerivationBriefingStart{}, false, fmt.Errorf("begin derivation briefing: %w", err)
	}
	if err := verifyDerivationBriefingSource(ctx, tx, sourceExperimentID); err != nil {
		_ = tx.Rollback()

		return domain.DerivationBriefingStart{}, false, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", briefingSessionID, "derivation_brief", start.State, now, now); err != nil {
		_ = tx.Rollback()

		return domain.DerivationBriefingStart{}, false, fmt.Errorf("insert derivation briefing session: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO derivation_briefing_operations (request_id, source_experiment_id, preparation_session_id, operation_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", requestID, sourceExperimentID, briefingSessionID, operationID, start.State, now, now); err != nil {
		_ = tx.Rollback()
		if isDerivationBriefingRequestConflict(err) {
			existing, found, findErr := s.findDerivationBriefing(ctx, requestID)
			if findErr != nil {
				return domain.DerivationBriefingStart{}, false, findErr
			}
			if found {
				return replayDerivationBriefing(existing, sourceExperimentID)
			}
		}

		return domain.DerivationBriefingStart{}, false, fmt.Errorf("insert derivation briefing operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.DerivationBriefingStart{}, false, fmt.Errorf("commit derivation briefing: %w", err)
	}

	return start, true, nil
}

// MarkDerivationBriefingStarted は開始済み状態を保存する。
func (s *Store) MarkDerivationBriefingStarted(ctx context.Context, requestID string) error {
	return s.updateDerivationBriefing(ctx, requestID, domain.BriefingStartStateStarted, "")
}

// MarkDerivationBriefingFailed は安全な開始失敗を保存する。
func (s *Store) MarkDerivationBriefingFailed(ctx context.Context, requestID, failureCode string) error {
	return s.updateDerivationBriefing(ctx, requestID, domain.BriefingStartStateFailed, failureCode)
}

// findDerivationBriefing はrequest IDに対応する開始結果を取得する。
func (s *Store) findDerivationBriefing(ctx context.Context, requestID string) (domain.DerivationBriefingStart, bool, error) {
	var start domain.DerivationBriefingStart
	err := s.db.QueryRowContext(ctx, "SELECT request_id, source_experiment_id, preparation_session_id, operation_id, state, failure_code FROM derivation_briefing_operations WHERE request_id=?", requestID).Scan(&start.RequestID, &start.SourceExperimentID, &start.BriefingSessionID, &start.OperationID, &start.State, &start.FailureCode)
	if err == sql.ErrNoRows {
		return domain.DerivationBriefingStart{}, false, nil
	}
	if err != nil {
		return domain.DerivationBriefingStart{}, false, fmt.Errorf("find derivation briefing: %w", err)
	}

	return start, true, nil
}

// verifyDerivationBriefingSource は派生元の不変条件と確定済み結論を検証する。
func verifyDerivationBriefingSource(ctx context.Context, tx briefingTransaction, sourceExperimentID string) error {
	var fixedConditionID, conclusionID string
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(c.id, ''), COALESCE(x.id, '')
		FROM experiments e
		LEFT JOIN experiment_fixed_conditions c ON c.id=e.fixed_condition_id
		LEFT JOIN experiment_conclusions x ON x.experiment_id=e.id AND x.state='finalized'
		WHERE e.id=?`, sourceExperimentID).Scan(&fixedConditionID, &conclusionID)
	if err == sql.ErrNoRows {
		return apperr.New(apperr.CodeDerivedExperimentSourceNotFound)
	}
	if err != nil {
		return fmt.Errorf("find derivation briefing source: %w", err)
	}
	if fixedConditionID == "" || conclusionID == "" {
		return apperr.New(apperr.CodeDerivedExperimentSourceNotEligible)
	}

	return nil
}

// updateDerivationBriefing は開始操作とsession状態を同期する。
func (s *Store) updateDerivationBriefing(ctx context.Context, requestID, state, failureCode string) error {
	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return fmt.Errorf("begin derivation briefing update: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, "UPDATE derivation_briefing_operations SET state=?, failure_code=?, updated_at=? WHERE request_id=?", state, failureCode, now, requestID)
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing operation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("count derivation briefing operation updates: %w", err)
	}
	if affected != 1 {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing operation: request not found")
	}
	if _, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET state=?, updated_at=? WHERE id=(SELECT preparation_session_id FROM derivation_briefing_operations WHERE request_id=?)", state, now, requestID); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit derivation briefing update: %w", err)
	}

	return nil
}

// replayDerivationBriefing は同一request IDの開始結果を再生する。
func replayDerivationBriefing(start domain.DerivationBriefingStart, sourceExperimentID string) (domain.DerivationBriefingStart, bool, error) {
	if start.SourceExperimentID != sourceExperimentID {
		return domain.DerivationBriefingStart{}, false, apperr.New(apperr.CodeDerivedExperimentRequestConflict)
	}

	return start, false, nil
}

// isDerivationBriefingRequestConflict はrequest IDの一意制約違反を判定する。
func isDerivationBriefingRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: derivation_briefing_operations.request_id")
}
