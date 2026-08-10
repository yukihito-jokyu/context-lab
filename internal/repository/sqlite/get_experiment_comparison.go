package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// GetExperimentComparison は実験と所属評価の比較用正本を読み出す。
func (s *Store) GetExperimentComparison(ctx context.Context, experimentID string) (domain.ExperimentComparison, bool, error) {
	var comparison domain.ExperimentComparison
	var experimentUpdatedAt string
	var conclusionID, conclusionContent, conclusionState, conclusionFinalizedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT e.id, c.purpose, c.evaluation_axes, e.updated_at, x.id, x.conclusion, x.state, x.finalized_at
		FROM experiments e JOIN experiment_fixed_conditions c ON c.id = e.fixed_condition_id LEFT JOIN experiment_conclusions x ON x.experiment_id = e.id WHERE e.id = ?`, experimentID).Scan(&comparison.Experiment.ID, &comparison.Experiment.Purpose, &comparison.Experiment.EvaluationAxes, &experimentUpdatedAt, &conclusionID, &conclusionContent, &conclusionState, &conclusionFinalizedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentComparison{}, false, nil
	}
	if err != nil {
		return domain.ExperimentComparison{}, false, fmt.Errorf("find experiment comparison: %w", err)
	}
	comparison.LastConfirmedAt, err = time.Parse(time.RFC3339Nano, experimentUpdatedAt)
	if err != nil {
		return domain.ExperimentComparison{}, false, fmt.Errorf("parse comparison experiment update time: %w", err)
	}
	if conclusionID.Valid {
		finalizedAt, parseErr := time.Parse(time.RFC3339Nano, conclusionFinalizedAt.String)
		if parseErr != nil {
			return domain.ExperimentComparison{}, false, fmt.Errorf("parse comparison conclusion time: %w", parseErr)
		}
		comparison.Conclusion = &domain.ExperimentConclusion{ExperimentID: experimentID, ConclusionID: conclusionID.String, Conclusion: conclusionContent.String, State: conclusionState.String, FinalizedAt: finalizedAt}
		if finalizedAt.After(comparison.LastConfirmedAt) {
			comparison.LastConfirmedAt = finalizedAt
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT e.id, e.run_id, e.state, r.summary, COALESCE(e.result_status, 'notRecorded'), e.summary, COALESCE(e.result_reason_code, ''), COALESCE(e.reconciliation_state, 'confirmed'), e.last_observed_at, e.updated_at FROM experiment_evaluations e JOIN experiment_runs r ON r.id = e.run_id WHERE e.experiment_id = ? ORDER BY e.created_at, e.id`, experimentID)
	if err != nil {
		return domain.ExperimentComparison{}, false, fmt.Errorf("find experiment comparison evaluations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	comparison.Evaluations = make([]domain.ExperimentComparisonEvaluation, 0)
	for rows.Next() {
		var evaluation domain.ExperimentComparisonEvaluation
		var runSummary, resultSummary, reasonCode sql.NullString
		var lastObservedAt sql.NullString
		var updatedAt string
		if err := rows.Scan(&evaluation.EvaluationID, &evaluation.RunID, &evaluation.State, &runSummary, &evaluation.Result.Status, &resultSummary, &reasonCode, &evaluation.Reconciliation.State, &lastObservedAt, &updatedAt); err != nil {
			return domain.ExperimentComparison{}, false, fmt.Errorf("scan experiment comparison evaluation: %w", err)
		}
		if evaluation.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt); err != nil {
			return domain.ExperimentComparison{}, false, fmt.Errorf("parse comparison update time: %w", err)
		}
		if runSummary.Valid && strings.TrimSpace(runSummary.String) != "" {
			evaluation.RunSummary = &runSummary.String
		}
		if resultSummary.Valid && strings.TrimSpace(resultSummary.String) != "" {
			evaluation.Result.Summary = &resultSummary.String
		}
		if reasonCode.Valid {
			evaluation.Result.ReasonCode = reasonCode.String
		}
		if !lastObservedAt.Valid || lastObservedAt.String == "" {
			evaluation.Reconciliation.LastObservedAt = evaluation.UpdatedAt
		} else if evaluation.Reconciliation.LastObservedAt, err = time.Parse(time.RFC3339Nano, lastObservedAt.String); err != nil {
			return domain.ExperimentComparison{}, false, fmt.Errorf("parse comparison last observed time: %w", err)
		}
		if evaluation.UpdatedAt.After(comparison.LastConfirmedAt) {
			comparison.LastConfirmedAt = evaluation.UpdatedAt
		}
		comparison.Evaluations = append(comparison.Evaluations, evaluation)
	}
	if err := rows.Err(); err != nil {
		return domain.ExperimentComparison{}, false, fmt.Errorf("iterate experiment comparison evaluations: %w", err)
	}
	return comparison, true, nil
}
