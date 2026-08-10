package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// GetRunDetail は指定runの安全な実行事実と観測結果を取得する。
func (s *Store) GetRunDetail(ctx context.Context, runID string) (domain.ExperimentRunDetail, bool, error) {
	detail, found, err := s.findRunDetail(ctx, runID)
	if err != nil || !found {
		return detail, found, err
	}
	if err := s.populateRunDetail(ctx, &detail); err != nil {
		return domain.ExperimentRunDetail{}, false, err
	}

	return detail, true, nil
}

// findRunDetail はrun本体と固定prompt、開始操作を取得する。
func (s *Store) findRunDetail(ctx context.Context, runID string) (domain.ExperimentRunDetail, bool, error) {
	var detail domain.ExperimentRunDetail
	var summary sql.NullString
	var operationID sql.NullString
	var operationState sql.NullString
	var operationUpdatedAt sql.NullString
	var prompt sql.NullString
	var promptSequenceNo sql.NullInt64
	var lastObservedAt sql.NullString
	var reconciliationState sql.NullString
	var createdAt string
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `SELECT r.id, r.experiment_id, r.state, r.summary, r.created_at, r.updated_at,
		COALESCE(r.operation_id, ro.operation_id, o.operation_id), COALESCE(ro.state, o.state), COALESCE(ro.updated_at, o.updated_at), p.sequence_no, p.content, r.last_observed_at, r.reconciliation_state
		FROM experiment_runs r
		LEFT JOIN experiment_start_operations o ON o.experiment_id = r.experiment_id
			AND (r.operation_id IS NULL OR o.operation_id = r.operation_id)
		LEFT JOIN experiment_run_retry_operations ro ON ro.run_id = r.id
		LEFT JOIN experiments e ON e.id = r.experiment_id
		LEFT JOIN experiment_fixed_conditions c ON c.id = e.fixed_condition_id
		LEFT JOIN experiment_fixed_condition_prompts p ON p.fixed_condition_id = c.id AND p.sequence_no = r.prompt_sequence_no
		WHERE r.id = ? ORDER BY o.updated_at DESC LIMIT 1`, runID).Scan(
		&detail.Run.ID,
		&detail.Run.ExperimentID,
		&detail.Run.State,
		&summary,
		&createdAt,
		&updatedAt,
		&operationID,
		&operationState,
		&operationUpdatedAt,
		&promptSequenceNo,
		&prompt,
		&lastObservedAt,
		&reconciliationState,
	)
	if err == sql.ErrNoRows {
		return domain.ExperimentRunDetail{}, false, nil
	}
	if err != nil {
		return domain.ExperimentRunDetail{}, false, fmt.Errorf("find run detail: %w", err)
	}
	if summary.Valid && strings.TrimSpace(summary.String) != "" {
		detail.Run.Summary = &summary.String
	}
	if prompt.Valid {
		detail.FixedPrompt.Content = prompt.String
	}
	if promptSequenceNo.Valid {
		detail.FixedPrompt.SequenceNo = int(promptSequenceNo.Int64)
	}
	if operationID.Valid {
		detail.Operation.ID = operationID.String
	}
	if operationState.Valid {
		detail.Operation.State = operationState.String
	}
	if operationUpdatedAt.Valid {
		if detail.Operation.UpdatedAt, err = time.Parse(time.RFC3339Nano, operationUpdatedAt.String); err != nil {
			return domain.ExperimentRunDetail{}, false, fmt.Errorf("parse run operation update time: %w", err)
		}
	}
	if detail.Run.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt); err != nil {
		return domain.ExperimentRunDetail{}, false, fmt.Errorf("parse run creation time: %w", err)
	}
	if detail.Run.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt); err != nil {
		return domain.ExperimentRunDetail{}, false, fmt.Errorf("parse run update time: %w", err)
	}
	if reconciliationState.Valid && reconciliationState.String != "" {
		detail.Reconciliation.State = reconciliationState.String
	} else {
		detail.Reconciliation.State = "confirmed"
	}
	if lastObservedAt.Valid && lastObservedAt.String != "" {
		if detail.Reconciliation.LastObservedAt, err = time.Parse(time.RFC3339Nano, lastObservedAt.String); err != nil {
			return domain.ExperimentRunDetail{}, false, fmt.Errorf("parse run last observed time: %w", err)
		}
	} else {
		detail.Reconciliation.LastObservedAt = detail.Run.UpdatedAt
	}

	return detail, true, nil
}

// populateRunDetail は観測、artifact、失敗、照合状態を取得する。
func (s *Store) populateRunDetail(ctx context.Context, detail *domain.ExperimentRunDetail) error {
	observations, err := s.findRunObservations(ctx, detail.Run.ID)
	if err != nil {
		return err
	}
	detail.Observations = observations
	artifacts, err := s.findRunArtifacts(ctx, detail.Run.ID)
	if err != nil {
		return err
	}
	detail.Artifacts = artifacts
	failure, err := s.findRunFailure(ctx, detail.Run.ID)
	if err != nil {
		return err
	}
	detail.Failure = failure
	detail.LastConfirmedAt = detail.Run.UpdatedAt

	return nil
}

// findRunObservations はrunの観測要約を順番に取得する。
func (s *Store) findRunObservations(ctx context.Context, runID string) ([]domain.ExperimentRunObservation, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT sequence_no, kind, occurred_at, summary FROM experiment_run_observations WHERE run_id = ? ORDER BY sequence_no", runID)
	if err != nil {
		return nil, fmt.Errorf("find run observations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	observations := make([]domain.ExperimentRunObservation, 0)
	for rows.Next() {
		var observation domain.ExperimentRunObservation
		var occurredAt string
		if err := rows.Scan(&observation.SequenceNo, &observation.Kind, &occurredAt, &observation.Summary); err != nil {
			return nil, fmt.Errorf("scan run observation: %w", err)
		}
		if observation.OccurredAt, err = time.Parse(time.RFC3339Nano, occurredAt); err != nil {
			return nil, fmt.Errorf("parse run observation time: %w", err)
		}
		observations = append(observations, observation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run observations: %w", err)
	}

	return observations, nil
}

// findRunArtifacts はartifact差分と未記録状態を取得する。
func (s *Store) findRunArtifacts(ctx context.Context, runID string) (domain.ExperimentRunArtifacts, error) {
	artifacts := domain.ExperimentRunArtifacts{Items: make([]domain.ExperimentRunArtifact, 0)}
	var reasonCode sql.NullString
	if err := s.db.QueryRowContext(ctx, "SELECT artifact_status, artifact_reason_code FROM experiment_runs WHERE id = ?", runID).Scan(&artifacts.Status, &reasonCode); err != nil {
		return domain.ExperimentRunArtifacts{}, fmt.Errorf("find run artifact status: %w", err)
	}
	if reasonCode.Valid {
		artifacts.ReasonCode = reasonCode.String
	}
	if artifacts.Status == domain.ExperimentRunArtifactStatusPartial && strings.TrimSpace(artifacts.ReasonCode) == "" {
		artifacts.ReasonCode = "ARTIFACT_PARTIAL"
	}

	rows, err := s.db.QueryContext(ctx, "SELECT digest, label, status FROM experiment_run_artifacts WHERE run_id = ? ORDER BY digest", runID)
	if err != nil {
		return domain.ExperimentRunArtifacts{}, fmt.Errorf("find run artifacts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var artifact domain.ExperimentRunArtifact
		var label sql.NullString
		if err := rows.Scan(&artifact.Digest, &label, &artifact.Status); err != nil {
			return domain.ExperimentRunArtifacts{}, fmt.Errorf("scan run artifact: %w", err)
		}
		if label.Valid {
			artifact.Label = &label.String
		}
		artifacts.Items = append(artifacts.Items, artifact)
	}
	if err := rows.Err(); err != nil {
		return domain.ExperimentRunArtifacts{}, fmt.Errorf("iterate run artifacts: %w", err)
	}
	return artifacts, nil
}

// findRunFailure はrun固有の安全な失敗事実を取得する。
func (s *Store) findRunFailure(ctx context.Context, runID string) (*domain.ExperimentRunFailure, error) {
	var failure domain.ExperimentRunFailure
	var occurredAt string
	var partialSummary sql.NullString
	err := s.db.QueryRowContext(ctx, "SELECT code, occurred_at, partial_summary FROM experiment_run_failures WHERE run_id = ?", runID).Scan(&failure.Code, &occurredAt, &partialSummary)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find run failure: %w", err)
	}
	if failure.OccurredAt, err = time.Parse(time.RFC3339Nano, occurredAt); err != nil {
		return nil, fmt.Errorf("parse run failure time: %w", err)
	}
	if partialSummary.Valid && strings.TrimSpace(partialSummary.String) != "" {
		failure.PartialSummary = &partialSummary.String
	}

	return &failure, nil
}
