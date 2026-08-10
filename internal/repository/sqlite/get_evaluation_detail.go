package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yukihito-jokyu/context-lab/internal/domain"
	"time"
)

// GetEvaluationDetail は評価の安全な根拠と結果を取得する。
func (s *Store) GetEvaluationDetail(ctx context.Context, evaluationID string) (domain.ExperimentEvaluationDetail, bool, error) {
	var d domain.ExperimentEvaluationDetail
	var summary sql.NullString
	var resultSummary sql.NullString
	var reason sql.NullString
	var observed sql.NullString
	var reconciliation sql.NullString
	var created, updated, operationUpdated string
	err := s.db.QueryRowContext(ctx, `SELECT e.id,e.experiment_id,e.run_id,e.state,e.summary,e.created_at,e.updated_at,o.operation_id,o.state,o.updated_at,COALESCE(r.summary,''),COALESCE(c.evaluation_axes,''),COALESCE(e.result_status,'notRecorded'),e.summary,COALESCE(e.result_reason_code,''),e.last_observed_at,e.reconciliation_state FROM experiment_evaluations e JOIN experiment_evaluation_operations o ON o.evaluation_id=e.id LEFT JOIN experiment_runs r ON r.id=e.run_id LEFT JOIN experiments x ON x.id=e.experiment_id LEFT JOIN experiment_fixed_conditions c ON c.id=x.fixed_condition_id WHERE e.id=?`, evaluationID).Scan(&d.Evaluation.ID, &d.Evaluation.ExperimentID, &d.Evaluation.RunID, &d.Evaluation.State, &summary, &created, &updated, &d.Operation.ID, &d.Operation.State, &operationUpdated, &d.Evidence.RunSummary, &d.Evidence.EvaluationAxes, &d.Result.Status, &resultSummary, &reason, &observed, &reconciliation)
	if err == sql.ErrNoRows {
		return d, false, nil
	}
	if err != nil {
		return d, false, fmt.Errorf("get evaluation detail: %w", err)
	}
	var parseErr error
	d.Evaluation.CreatedAt, parseErr = time.Parse(time.RFC3339Nano, created)
	if parseErr != nil {
		return d, false, parseErr
	}
	d.Evaluation.UpdatedAt, parseErr = time.Parse(time.RFC3339Nano, updated)
	if parseErr != nil {
		return d, false, parseErr
	}
	d.Operation.UpdatedAt, parseErr = time.Parse(time.RFC3339Nano, operationUpdated)
	if parseErr != nil {
		return d, false, parseErr
	}
	if summary.Valid {
		d.Evaluation.Summary = &summary.String
	}
	if resultSummary.Valid {
		d.Result.Summary = &resultSummary.String
	}
	if reason.Valid {
		d.Result.ReasonCode = reason.String
	}
	d.Reconciliation.State = "confirmed"
	if reconciliation.Valid && reconciliation.String != "" {
		d.Reconciliation.State = reconciliation.String
	}
	d.Reconciliation.LastObservedAt = d.Evaluation.UpdatedAt
	if observed.Valid && observed.String != "" {
		d.Reconciliation.LastObservedAt, parseErr = time.Parse(time.RFC3339Nano, observed.String)
		if parseErr != nil {
			return d, false, parseErr
		}
	}
	d.LastConfirmedAt = d.Evaluation.UpdatedAt
	if d.Evaluation.State == domain.ExperimentEvaluationStateFailed {
		d.Failure = &domain.ExperimentEvaluationFailure{Code: d.Result.ReasonCode, OccurredAt: d.Evaluation.UpdatedAt}
	}
	return d, true, nil
}
