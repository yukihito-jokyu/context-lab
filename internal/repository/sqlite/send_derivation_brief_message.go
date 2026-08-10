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

// BeginDerivationBriefMessage は開始済み派生壁打ちに紐付く送信意図を原子的に保存する。
func (s *Store) BeginDerivationBriefMessage(ctx context.Context, requestID, briefingSessionID string) (domain.DerivationBriefingMessageOperation, bool, error) {
	existing, found, err := s.findDerivationBriefMessageOperation(ctx, requestID)
	if err != nil {
		return domain.DerivationBriefingMessageOperation{}, false, err
	}
	if found {
		if existing.BriefingSessionID != briefingSessionID {
			return domain.DerivationBriefingMessageOperation{}, false, apperr.New(apperr.CodeDerivationBriefingMessageRequestConflict)
		}

		return existing, false, nil
	}

	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.DerivationBriefingMessageOperation{}, false, fmt.Errorf("generate derivation briefing message operation ID: %w", err)
	}
	operation := domain.DerivationBriefingMessageOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       operationID,
		State:             domain.BriefingStartStateStarting,
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.DerivationBriefingMessageOperation{}, false, fmt.Errorf("begin derivation briefing message: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var state string
	if err := tx.QueryRowContext(ctx, "SELECT state FROM preparation_sessions WHERE id=? AND kind=?", briefingSessionID, "derivation_brief").Scan(&state); err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return domain.DerivationBriefingMessageOperation{}, false, apperr.New(apperr.CodeDerivationBriefingMessageNotFound)
		}

		return domain.DerivationBriefingMessageOperation{}, false, fmt.Errorf("find derivation briefing message session: %w", err)
	}
	if state != domain.BriefingStartStateStarted {
		_ = tx.Rollback()

		return domain.DerivationBriefingMessageOperation{}, false, apperr.New(apperr.CodeDerivationBriefingMessageNotActive)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO derivation_briefing_message_operations (id, request_id, preparation_session_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", operation.OperationID, operation.RequestID, operation.BriefingSessionID, operation.State, now, now); err != nil {
		_ = tx.Rollback()
		if isDerivationBriefingMessageRequestConflict(err) {
			existing, found, findErr := s.findDerivationBriefMessageOperation(ctx, requestID)
			if findErr != nil {
				return domain.DerivationBriefingMessageOperation{}, false, findErr
			}
			if found && existing.BriefingSessionID == briefingSessionID {
				return existing, false, nil
			}
			if found {
				return domain.DerivationBriefingMessageOperation{}, false, apperr.New(apperr.CodeDerivationBriefingMessageRequestConflict)
			}
		}

		return domain.DerivationBriefingMessageOperation{}, false, fmt.Errorf("insert derivation briefing message operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.DerivationBriefingMessageOperation{}, false, fmt.Errorf("commit derivation briefing message: %w", err)
	}

	return operation, true, nil
}

// CompleteDerivationBriefMessage は安全な会話と派生提案を送信完了として保存する。
func (s *Store) CompleteDerivationBriefMessage(ctx context.Context, requestID, message string, result domain.DerivationBriefingMessageResult) error {
	s.derivationBriefingMessageMu.Lock()
	defer s.derivationBriefingMessageMu.Unlock()

	operation, found, err := s.findDerivationBriefMessageOperation(ctx, requestID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("complete derivation briefing message: request not found")
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return fmt.Errorf("begin derivation briefing message completion: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var nextSequence int
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(sequence_no), 0) + 1 FROM derivation_briefing_messages WHERE preparation_session_id=?", operation.BriefingSessionID).Scan(&nextSequence); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("find next derivation briefing message sequence: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO derivation_briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)", operation.BriefingSessionID, nextSequence, "user", message, now); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("insert derivation user briefing message: %w", err)
	}
	if result.AssistantMessage != "" {
		if _, err := tx.ExecContext(ctx, "INSERT INTO derivation_briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)", operation.BriefingSessionID, nextSequence+1, "assistant", result.AssistantMessage, now); err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("insert derivation assistant briefing message: %w", err)
		}
	}
	if result.Suggestion != nil {
		if err := insertDerivationBriefingSuggestion(ctx, tx, operation, *result.Suggestion, now); err != nil {
			_ = tx.Rollback()

			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE derivation_briefing_message_operations SET state=?, failure_code='', updated_at=? WHERE request_id=?", domain.BriefingStartStateStarted, now, requestID); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("complete derivation briefing message operation: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET updated_at=? WHERE id=?", now, operation.BriefingSessionID); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update derivation briefing session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit derivation briefing message completion: %w", err)
	}

	return nil
}

// FailDerivationBriefMessage は安全な送信失敗を保存する。
func (s *Store) FailDerivationBriefMessage(ctx context.Context, requestID, failureCode string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx, "UPDATE derivation_briefing_message_operations SET state=?, failure_code=?, updated_at=? WHERE request_id=?", domain.BriefingStartStateFailed, failureCode, now, requestID)
	if err != nil {
		return fmt.Errorf("fail derivation briefing message operation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count failed derivation briefing message operations: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("fail derivation briefing message operation: request not found")
	}

	return nil
}

// findDerivationBriefMessageOperation はrequest IDに対応する送信結果を取得する。
func (s *Store) findDerivationBriefMessageOperation(ctx context.Context, requestID string) (domain.DerivationBriefingMessageOperation, bool, error) {
	operation := domain.DerivationBriefingMessageOperation{RequestID: requestID}
	err := s.db.QueryRowContext(ctx, "SELECT preparation_session_id, id, state, failure_code FROM derivation_briefing_message_operations WHERE request_id=?", requestID).Scan(&operation.BriefingSessionID, &operation.OperationID, &operation.State, &operation.FailureCode)
	if err == sql.ErrNoRows {
		return domain.DerivationBriefingMessageOperation{}, false, nil
	}
	if err != nil {
		return domain.DerivationBriefingMessageOperation{}, false, fmt.Errorf("find derivation briefing message operation: %w", err)
	}

	return operation, true, nil
}

// insertDerivationBriefingSuggestion は構造化済みの派生提案を一版として保存する。
func insertDerivationBriefingSuggestion(ctx context.Context, tx briefingTransaction, operation domain.DerivationBriefingMessageOperation, suggestion domain.ExperimentBrief, createdAt string) error {
	suggestionID, err := newBriefingIdentifier()
	if err != nil {
		return fmt.Errorf("generate derivation briefing suggestion identifier: %w", err)
	}
	var nextVersion int
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version_no), 0) + 1 FROM derivation_briefing_suggestions WHERE preparation_session_id=?", operation.BriefingSessionID).Scan(&nextVersion); err != nil {
		return fmt.Errorf("find next derivation briefing suggestion version: %w", err)
	}
	prompts, err := json.Marshal(suggestion.CandidatePrompts)
	if err != nil {
		return fmt.Errorf("marshal derivation briefing suggestion prompts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO derivation_briefing_suggestions (id, preparation_session_id, operation_id, version_no, purpose, decision, hypothesis, candidate_prompts, evaluation_criteria, environment_conditions, initial_input, success_criteria, required_conditions, open_question, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", suggestionID, operation.BriefingSessionID, operation.OperationID, nextVersion, suggestion.Purpose, suggestion.Decision, suggestion.Hypothesis, string(prompts), suggestion.EvaluationCriteria, suggestion.EnvironmentConditions, suggestion.InitialInput, suggestion.SuccessCriteria, suggestion.RequiredConditions, suggestion.OpenQuestion, createdAt); err != nil {
		return fmt.Errorf("insert derivation briefing suggestion: %w", err)
	}

	return nil
}

// isDerivationBriefingMessageRequestConflict はrequest IDの一意制約違反を判定する。
func isDerivationBriefingMessageRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: derivation_briefing_message_operations.request_id")
}
