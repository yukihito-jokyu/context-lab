package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

const listPreparationsQuery = `SELECT id, state, created_at, updated_at
FROM preparation_sessions
WHERE kind = ?
ORDER BY updated_at DESC, id ASC`

const environmentPreparationKind = "environment_preparation"

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
