package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// ListExperiments は取消済みを分離して実験一覧を読み出す。
func (s *Store) ListExperiments(ctx context.Context) (domain.ExperimentCollection, error) {
	experiments, err := s.listByCancellation(ctx, false)
	if err != nil {
		return domain.ExperimentCollection{}, err
	}

	cancelledExperiments, err := s.listByCancellation(ctx, true)
	if err != nil {
		return domain.ExperimentCollection{}, err
	}

	lastConfirmedAt, err := s.confirmList(ctx)
	if err != nil {
		return domain.ExperimentCollection{}, err
	}

	return domain.ExperimentCollection{
		Experiments:          experiments,
		CancelledExperiments: cancelledExperiments,
		LastConfirmedAt:      lastConfirmedAt,
	}, nil
}

// listByCancellation は取消状態で実験を絞り込む。
func (s *Store) listByCancellation(ctx context.Context, cancelled bool) (experiments []domain.Experiment, err error) {
	operator := "<>"
	if cancelled {
		operator = "="
	}

	query := "SELECT id, purpose, state, progress_summary, derived_from_experiment_id, updated_at FROM experiments WHERE state " + operator + " ? ORDER BY updated_at DESC, id ASC"
	rows, err := s.db.QueryContext(ctx, query, "cancelled")
	if err != nil {
		return nil, fmt.Errorf("query experiments: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close experiments rows: %w", closeErr)
		}
	}()

	experiments = make([]domain.Experiment, 0)
	for rows.Next() {
		experiment, err := scanExperiment(rows)
		if err != nil {
			return nil, err
		}

		experiments = append(experiments, experiment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate experiments: %w", err)
	}

	return experiments, nil
}

// scanExperiment はSQLite行をdomain実験へ変換。
func scanExperiment(rows *sql.Rows) (domain.Experiment, error) {
	var experiment domain.Experiment
	var derivedFromExperimentID sql.NullString
	var updatedAt string
	if err := rows.Scan(
		&experiment.ID,
		&experiment.Purpose,
		&experiment.State,
		&experiment.ProgressSummary,
		&derivedFromExperimentID,
		&updatedAt,
	); err != nil {
		return domain.Experiment{}, fmt.Errorf("scan experiment: %w", err)
	}

	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("parse experiment update time: %w", err)
	}
	experiment.UpdatedAt = parsedUpdatedAt
	if derivedFromExperimentID.Valid {
		experiment.DerivedFromExperimentID = &derivedFromExperimentID.String
	}

	return experiment, nil
}

// confirmList は成功した一覧取得時刻を記録して返す。
func (s *Store) confirmList(ctx context.Context) (*time.Time, error) {
	confirmedAt := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, "INSERT INTO application_metadata (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", "last_confirmed_at", confirmedAt.Format(time.RFC3339Nano)); err != nil {
		return nil, fmt.Errorf("record list confirmation: %w", err)
	}

	return &confirmedAt, nil
}
