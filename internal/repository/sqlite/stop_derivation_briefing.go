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

// BeginStopDerivationBriefing は派生壁打ち停止意図を原子的に保存する。
func (s *Store) BeginStopDerivationBriefing(ctx context.Context, requestID, briefingSessionID string) (domain.DerivationBriefingStopOperation, bool, error) {
	existing, found, err := s.findStopDerivationBriefing(ctx, requestID)
	if err != nil {
		return domain.DerivationBriefingStopOperation{}, false, err
	}
	if found {
		if existing.BriefingSessionID != briefingSessionID {
			return domain.DerivationBriefingStopOperation{}, false, apperr.New(apperr.CodeDerivationBriefingStopRequestConflict)
		}

		return existing, false, nil
	}

	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("generate derivation briefing stop operation ID: %w", err)
	}
	operation := domain.DerivationBriefingStopOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       operationID,
		State:             domain.BriefingStartStateStarting,
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("begin stop derivation briefing: %w", err)
	}
	result, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET state=?, updated_at=? WHERE id=? AND kind=? AND state=?", domain.BriefingStartStateStopping, time.Now().UTC().Format(time.RFC3339Nano), briefingSessionID, "derivation_brief", domain.BriefingStartStateStarted)
	if err != nil {
		_ = tx.Rollback()

		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("mark derivation briefing stopping: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("count derivation briefing stopping updates: %w", err)
	}
	if affected != 1 {
		_ = tx.Rollback()
		existing, found, findErr := s.findStopDerivationBriefing(ctx, requestID)
		if findErr != nil {
			return domain.DerivationBriefingStopOperation{}, false, findErr
		}
		if found && existing.BriefingSessionID == briefingSessionID {
			return existing, false, nil
		}
		if found {
			return domain.DerivationBriefingStopOperation{}, false, apperr.New(apperr.CodeDerivationBriefingStopRequestConflict)
		}
		var exists bool
		if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM preparation_sessions WHERE id=? AND kind=?)", briefingSessionID, "derivation_brief").Scan(&exists); err != nil {
			return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("find stop derivation briefing session: %w", err)
		}
		if !exists {
			return domain.DerivationBriefingStopOperation{}, false, apperr.New(apperr.CodeDerivationBriefingStopNotFound)
		}

		return domain.DerivationBriefingStopOperation{}, false, apperr.New(apperr.CodeDerivationBriefingStopNotActive)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO derivation_briefing_stop_operations (id, request_id, preparation_session_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", operation.OperationID, operation.RequestID, operation.BriefingSessionID, operation.State, now, now); err != nil {
		_ = tx.Rollback()
		if isDerivationBriefingStopRequestConflict(err) {
			existing, found, findErr := s.findStopDerivationBriefing(ctx, requestID)
			if findErr != nil {
				return domain.DerivationBriefingStopOperation{}, false, findErr
			}
			if found && existing.BriefingSessionID == briefingSessionID {
				return existing, false, nil
			}
			if found {
				return domain.DerivationBriefingStopOperation{}, false, apperr.New(apperr.CodeDerivationBriefingStopRequestConflict)
			}
		}

		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("insert derivation briefing stop operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("commit stop derivation briefing: %w", err)
	}

	return operation, true, nil
}

// CompleteStopDerivationBriefing は停止確認後に操作とsessionを停止済みに保存する。
func (s *Store) CompleteStopDerivationBriefing(ctx context.Context, requestID string) error {
	return s.updateStopDerivationBriefing(ctx, requestID, domain.BriefingStartStateStopped, "")
}

// FailStopDerivationBriefing は安全な停止失敗を保存する。
func (s *Store) FailStopDerivationBriefing(ctx context.Context, requestID, failureCode string) error {
	return s.updateStopDerivationBriefing(ctx, requestID, domain.BriefingStartStateFailed, failureCode, domain.BriefingStartStateStarted)
}

// findStopDerivationBriefing はrequest IDに対応する停止結果を取得する。
func (s *Store) findStopDerivationBriefing(ctx context.Context, requestID string) (domain.DerivationBriefingStopOperation, bool, error) {
	operation := domain.DerivationBriefingStopOperation{RequestID: requestID}
	err := s.db.QueryRowContext(ctx, "SELECT preparation_session_id, id, state, failure_code FROM derivation_briefing_stop_operations WHERE request_id=?", requestID).Scan(&operation.BriefingSessionID, &operation.OperationID, &operation.State, &operation.FailureCode)
	if err == sql.ErrNoRows {
		return domain.DerivationBriefingStopOperation{}, false, nil
	}
	if err != nil {
		return domain.DerivationBriefingStopOperation{}, false, fmt.Errorf("find derivation briefing stop operation: %w", err)
	}

	return operation, true, nil
}

// updateStopDerivationBriefing は停止操作とsession状態を同期する。
func (s *Store) updateStopDerivationBriefing(ctx context.Context, requestID, state, failureCode string, sessionState ...string) error {
	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return fmt.Errorf("begin derivation briefing stop update: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, "UPDATE derivation_briefing_stop_operations SET state=?, failure_code=?, updated_at=? WHERE request_id=?", state, failureCode, now, requestID)
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing stop operation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("count derivation briefing stop operation updates: %w", err)
	}
	if affected != 1 {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing stop operation: request not found")
	}
	stateForSession := state
	if len(sessionState) != 0 {
		stateForSession = sessionState[0]
	}
	result, err = tx.ExecContext(ctx, "UPDATE preparation_sessions SET state=?, updated_at=? WHERE id=(SELECT preparation_session_id FROM derivation_briefing_stop_operations WHERE request_id=?)", stateForSession, now, requestID)
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing stop session: %w", err)
	}
	affected, err = result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("count derivation briefing stop session updates: %w", err)
	}
	if affected != 1 {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing stop session: session not found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit derivation briefing stop update: %w", err)
	}

	return nil
}

// isDerivationBriefingStopRequestConflict はrequest IDの一意制約違反を判定する。
func isDerivationBriefingStopRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: derivation_briefing_stop_operations.request_id")
}
