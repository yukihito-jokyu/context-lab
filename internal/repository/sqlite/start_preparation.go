package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// BeginPreparation は開始意図、session、operationを同一transactionで保存する。
func (s *Store) BeginPreparation(ctx context.Context, requestID, scope string) (domain.EnvironmentPreparationStart, bool, error) {
	existing, found, err := s.loadPreparationStart(ctx, requestID)
	if err != nil {
		return domain.EnvironmentPreparationStart{}, false, err
	}
	if found {
		if existing.Scope != scope {
			return domain.EnvironmentPreparationStart{}, false, apperr.New(apperr.CodePreparationStartRequestConflict)
		}

		return existing, false, nil
	}

	preparationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.EnvironmentPreparationStart{}, false, fmt.Errorf("generate preparation identifier: %w", err)
	}
	start := domain.EnvironmentPreparationStart{RequestID: requestID, PreparationID: preparationID, Scope: scope, State: domain.EnvironmentPreparationStateStarting}
	if err := s.savePreparationStart(ctx, start); err != nil {
		if isPreparationRequestConflict(err) {
			existing, found, findErr := s.loadPreparationStart(ctx, requestID)
			if findErr != nil {
				return domain.EnvironmentPreparationStart{}, false, findErr
			}
			if found {
				if existing.Scope != scope {
					return domain.EnvironmentPreparationStart{}, false, apperr.New(apperr.CodePreparationStartRequestConflict)
				}

				return existing, false, nil
			}
		}
		if isPreparationScopeConflict(err) {
			return domain.EnvironmentPreparationStart{}, false, apperr.New(apperr.CodePreparationStartPending)
		}

		return domain.EnvironmentPreparationStart{}, false, err
	}

	return start, true, nil
}

func (s *Store) loadPreparationStart(ctx context.Context, requestID string) (domain.EnvironmentPreparationStart, bool, error) {
	if s.findPreparationStartOverride != nil {
		return s.findPreparationStartOverride(ctx, requestID)
	}

	return s.findPreparationStart(ctx, requestID)
}

func (s *Store) savePreparationStart(ctx context.Context, start domain.EnvironmentPreparationStart) error {
	if s.insertPreparationStartOverride != nil {
		return s.insertPreparationStartOverride(ctx, start)
	}

	return s.insertPreparationStart(ctx, start)
}

// insertPreparationStart は開始記録を原子的に書き込む。
func (s *Store) insertPreparationStart(ctx context.Context, start domain.EnvironmentPreparationStart) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin preparation start: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", start.PreparationID, environmentPreparationKind, start.State, now, now); err != nil {
		return fmt.Errorf("insert preparation session: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO environment_preparation_operations (request_id, preparation_session_id, scope, state, failure_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", start.RequestID, start.PreparationID, start.Scope, start.State, "", now, now); err != nil {
		return fmt.Errorf("insert preparation operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit preparation start: %w", err)
	}

	return nil
}

// MarkPreparationRunning はACP照合開始を記録する。
func (s *Store) MarkPreparationRunning(ctx context.Context, requestID string) error {
	return s.updatePreparationStart(ctx, requestID, domain.EnvironmentPreparationStateRunning, "", nil)
}

// CompletePreparation は安全な候補と診断を保存して開始を完了する。
func (s *Store) CompletePreparation(ctx context.Context, requestID string, result domain.EnvironmentPreparationResult) error {
	return s.updatePreparationStart(ctx, requestID, domain.EnvironmentPreparationStateCompleted, "", &result)
}

// FailPreparation は安全な失敗コードを保存して開始を失敗にする。
func (s *Store) FailPreparation(ctx context.Context, requestID, failureCode string) error {
	return s.updatePreparationStart(ctx, requestID, domain.EnvironmentPreparationStateFailed, failureCode, nil)
}

// updatePreparationStart はoperationとsessionを同じ状態へ更新する。
func (s *Store) updatePreparationStart(ctx context.Context, requestID, state, failureCode string, result *domain.EnvironmentPreparationResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin preparation update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var preparationID string
	if err := tx.QueryRowContext(ctx, "SELECT preparation_session_id FROM environment_preparation_operations WHERE request_id = ?", requestID).Scan(&preparationID); errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find preparation operation: not found")
	} else if err != nil {
		return fmt.Errorf("find preparation operation: %w", err)
	}
	if result != nil {
		if err := insertPreparationResult(ctx, tx, preparationID, *result); err != nil {
			return err
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "UPDATE environment_preparation_operations SET state = ?, failure_code = ?, updated_at = ? WHERE request_id = ?", state, failureCode, now, requestID); err != nil {
		return fmt.Errorf("update preparation operation: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET state = ?, updated_at = ? WHERE id = ?", state, now, preparationID); err != nil {
		return fmt.Errorf("update preparation session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit preparation update: %w", err)
	}

	return nil
}

// insertPreparationResult は安全な候補と診断を保存する。
func insertPreparationResult(ctx context.Context, tx *sql.Tx, preparationID string, result domain.EnvironmentPreparationResult) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, candidate := range result.Candidates {
		candidateID, err := newBriefingIdentifier()
		if err != nil {
			return fmt.Errorf("generate preparation candidate identifier: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO environment_preparation_candidates (id, preparation_session_id, environment_conditions, safe_summary, created_at) VALUES (?, ?, ?, ?, ?)", candidateID, preparationID, candidate.EnvironmentConditions, candidate.Summary, now); err != nil {
			return fmt.Errorf("insert preparation candidate: %w", err)
		}
	}
	for _, diagnostic := range result.Diagnostics {
		diagnosticID, err := newBriefingIdentifier()
		if err != nil {
			return fmt.Errorf("generate preparation diagnostic identifier: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO environment_preparation_diagnostics (id, preparation_session_id, code, safe_summary, occurred_at) VALUES (?, ?, ?, ?, ?)", diagnosticID, preparationID, diagnostic.Code, diagnostic.SafeSummary, now); err != nil {
			return fmt.Errorf("insert preparation diagnostic: %w", err)
		}
	}

	return nil
}

// findPreparationStart はrequest IDの開始結果を読み出す。
func (s *Store) findPreparationStart(ctx context.Context, requestID string) (domain.EnvironmentPreparationStart, bool, error) {
	var start domain.EnvironmentPreparationStart
	err := s.db.QueryRowContext(ctx, "SELECT request_id, preparation_session_id, scope, state, COALESCE(failure_code, '') FROM environment_preparation_operations WHERE request_id = ?", requestID).Scan(&start.RequestID, &start.PreparationID, &start.Scope, &start.State, &start.FailureCode)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.EnvironmentPreparationStart{}, false, nil
	}
	if err != nil {
		return domain.EnvironmentPreparationStart{}, false, fmt.Errorf("find preparation start: %w", err)
	}

	return start, true, nil
}

// isPreparationRequestConflict はrequest IDの一意制約競合を判定する。
func isPreparationRequestConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: environment_preparation_operations.request_id")
}

// isPreparationScopeConflict は実行中scopeの一意制約競合を判定する。
func isPreparationScopeConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: environment_preparation_operations.scope")
}
