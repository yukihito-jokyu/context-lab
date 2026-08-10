package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

const listPreparationsQuery = `SELECT id, state, created_at, updated_at
FROM preparation_sessions
WHERE kind = ?
ORDER BY updated_at DESC, id ASC`

const environmentPreparationKind = "environment_preparation"

type preparationQueryer interface {
	QueryContext(context.Context, string, ...any) (preparationRows, error)
	QueryRowContext(context.Context, string, ...any) preparationRow
}

// preparationRow は環境準備詳細の単一行読み出し境界。
type preparationRow interface {
	Scan(...any) error
}

// sqlitePreparationQueryer はdatabase/sqlの環境準備詳細query adapter。
type sqlitePreparationQueryer struct {
	db *sql.DB
}

// QueryContext は複数行queryをSQLiteへ委譲。
func (q sqlitePreparationQueryer) QueryContext(ctx context.Context, query string, arguments ...any) (preparationRows, error) {
	rows, err := q.db.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, err
	}

	return sqlitePreparationRows{rows: rows}, nil
}

// QueryRowContext は単一行queryをSQLiteへ委譲。
func (q sqlitePreparationQueryer) QueryRowContext(ctx context.Context, query string, arguments ...any) preparationRow {
	return q.db.QueryRowContext(ctx, query, arguments...)
}

// preparationRows は環境準備session一覧の行読み出し境界。
type preparationRows interface {
	Close() error
	Err() error
	Next() bool
	Scan(...any) error
}

// sqlitePreparationRows はdatabase/sqlの行読み出しadapter。
type sqlitePreparationRows struct {
	rows *sql.Rows
}

// Close は行読み出しを閉じる。
func (r sqlitePreparationRows) Close() error {
	return r.rows.Close()
}

// Err は行読み出しエラーを返す。
func (r sqlitePreparationRows) Err() error {
	return r.rows.Err()
}

// Next は次の行へ進む。
func (r sqlitePreparationRows) Next() bool {
	return r.rows.Next()
}

// Scan は現在行を読み出す。
func (r sqlitePreparationRows) Scan(destinations ...any) error {
	return r.rows.Scan(destinations...)
}

// ListPreparations は環境準備sessionだけを最終観測時刻順で読み出す。
func (s *Store) ListPreparations(ctx context.Context) ([]domain.Preparation, error) {
	rows, err := s.listPreparations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list preparations: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	preparations := make([]domain.Preparation, 0)
	for rows.Next() {
		var preparation domain.Preparation
		var startedAt string
		var lastObservedAt string
		if err := rows.Scan(&preparation.ID, &preparation.State, &startedAt, &lastObservedAt); err != nil {
			return nil, fmt.Errorf("scan preparation: %w", err)
		}
		parsedStartedAt, err := time.Parse(time.RFC3339Nano, startedAt)
		if err != nil {
			return nil, fmt.Errorf("parse preparation start time: %w", err)
		}
		parsedLastObservedAt, err := time.Parse(time.RFC3339Nano, lastObservedAt)
		if err != nil {
			return nil, fmt.Errorf("parse preparation observation time: %w", err)
		}
		preparation.StartedAt = parsedStartedAt.UTC()
		preparation.LastObservedAt = parsedLastObservedAt.UTC()
		preparations = append(preparations, preparation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate preparations: %w", err)
	}

	return preparations, nil
}

// GetPreparation は環境準備sessionの詳細を安全な情報だけで読み出す。
func (s *Store) GetPreparation(ctx context.Context, preparationID string) (domain.PreparationDetail, bool, error) {
	return s.getPreparation(ctx, preparationID)
}

// getPreparation は環境準備session詳細のSQLite読み出しを実行。
func getPreparation(ctx context.Context, queryer preparationQueryer, preparationID string) (domain.PreparationDetail, bool, error) {
	preparation, found, err := getPreparationSession(ctx, queryer, preparationID)
	if err != nil || !found {
		return preparation, found, err
	}

	candidates, err := getPreparationCandidates(ctx, queryer, preparationID)
	if err != nil {
		return domain.PreparationDetail{}, false, err
	}
	diagnostics, err := getPreparationDiagnostics(ctx, queryer, preparationID)
	if err != nil {
		return domain.PreparationDetail{}, false, err
	}
	operation, err := getPreparationOperation(ctx, queryer, preparationID)
	if err != nil {
		return domain.PreparationDetail{}, false, err
	}

	preparation.Candidates = candidates
	preparation.Diagnostics = diagnostics
	preparation.Reconciliation = domain.PreparationReconciliation{State: preparation.State, LastObservedAt: preparation.LastObservedAt}
	if operation != nil && (operation.State == "starting" || operation.State == "running") {
		preparation.Reconciliation.State = "reconciling"
	}
	if operation != nil && operation.State == "failed" {
		preparation.Failure = &domain.PreparationFailure{Code: operation.FailureCode, OccurredAt: operation.UpdatedAt}
	}

	return preparation, true, nil
}

// getPreparationSession は環境準備session本体を読み出す。
func getPreparationSession(ctx context.Context, queryer preparationQueryer, preparationID string) (domain.PreparationDetail, bool, error) {
	var preparation domain.PreparationDetail
	var startedAt string
	var lastObservedAt string
	err := queryer.QueryRowContext(ctx, `SELECT id, state, created_at, updated_at FROM preparation_sessions WHERE id = ? AND kind = ?`, preparationID, environmentPreparationKind).Scan(&preparation.ID, &preparation.State, &startedAt, &lastObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PreparationDetail{}, false, nil
	}
	if err != nil {
		return domain.PreparationDetail{}, false, fmt.Errorf("get preparation session: %w", err)
	}
	parsedStartedAt, err := time.Parse(time.RFC3339Nano, startedAt)
	if err != nil {
		return domain.PreparationDetail{}, false, fmt.Errorf("parse preparation start time: %w", err)
	}
	parsedLastObservedAt, err := time.Parse(time.RFC3339Nano, lastObservedAt)
	if err != nil {
		return domain.PreparationDetail{}, false, fmt.Errorf("parse preparation observation time: %w", err)
	}
	preparation.StartedAt = parsedStartedAt.UTC()
	preparation.LastObservedAt = parsedLastObservedAt.UTC()

	return preparation, true, nil
}

// getPreparationCandidates は候補を作成順で読み出す。
func getPreparationCandidates(ctx context.Context, queryer preparationQueryer, preparationID string) ([]domain.PreparationCandidate, error) {
	rows, err := queryer.QueryContext(ctx, `SELECT id, environment_conditions, safe_summary, created_at FROM environment_preparation_candidates WHERE preparation_session_id = ? ORDER BY created_at ASC, id ASC`, preparationID)
	if err != nil {
		return nil, fmt.Errorf("get preparation candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	candidates := make([]domain.PreparationCandidate, 0)
	for rows.Next() {
		var candidate domain.PreparationCandidate
		var createdAt string
		if err := rows.Scan(&candidate.ID, &candidate.EnvironmentConditions, &candidate.Summary, &createdAt); err != nil {
			return nil, fmt.Errorf("scan preparation candidate: %w", err)
		}
		parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse preparation candidate time: %w", err)
		}
		candidate.CreatedAt = parsedCreatedAt.UTC()
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate preparation candidates: %w", err)
	}

	return candidates, nil
}

// getPreparationDiagnostics は診断を発生順で読み出す。
func getPreparationDiagnostics(ctx context.Context, queryer preparationQueryer, preparationID string) ([]domain.PreparationDiagnostic, error) {
	rows, err := queryer.QueryContext(ctx, `SELECT id, code, safe_summary, occurred_at FROM environment_preparation_diagnostics WHERE preparation_session_id = ? ORDER BY occurred_at ASC, id ASC`, preparationID)
	if err != nil {
		return nil, fmt.Errorf("get preparation diagnostics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	diagnostics := make([]domain.PreparationDiagnostic, 0)
	for rows.Next() {
		var diagnostic domain.PreparationDiagnostic
		var occurredAt string
		if err := rows.Scan(&diagnostic.ID, &diagnostic.Code, &diagnostic.SafeSummary, &occurredAt); err != nil {
			return nil, fmt.Errorf("scan preparation diagnostic: %w", err)
		}
		parsedOccurredAt, err := time.Parse(time.RFC3339Nano, occurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse preparation diagnostic time: %w", err)
		}
		diagnostic.OccurredAt = parsedOccurredAt.UTC()
		diagnostics = append(diagnostics, diagnostic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate preparation diagnostics: %w", err)
	}

	return diagnostics, nil
}

// preparationOperation は最後に観測した環境準備操作。
type preparationOperation struct {
	State       string
	FailureCode string
	UpdatedAt   time.Time
}

// getPreparationOperation は最後に観測した環境準備操作を読み出す。
func getPreparationOperation(ctx context.Context, queryer preparationQueryer, preparationID string) (*preparationOperation, error) {
	var operation preparationOperation
	var updatedAt string
	err := queryer.QueryRowContext(ctx, `SELECT state, COALESCE(failure_code, ''), updated_at FROM environment_preparation_operations WHERE preparation_session_id = ? ORDER BY updated_at DESC, request_id ASC LIMIT 1`, preparationID).Scan(&operation.State, &operation.FailureCode, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get preparation operation: %w", err)
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse preparation operation time: %w", err)
	}
	operation.UpdatedAt = parsedUpdatedAt.UTC()

	return &operation, nil
}
