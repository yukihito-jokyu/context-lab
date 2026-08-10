package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// GetDerivationSource は派生元の実験、固定条件、結論を読む。
func (s *Store) GetDerivationSource(ctx context.Context, experimentID string) (domain.ExperimentDerivationSource, bool, error) {
	var source domain.ExperimentDerivationSource
	var fixedID, purpose, hypothesis, environment, input, axes, fixedAt sql.NullString
	var conclusionID, conclusion, state, finalizedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT e.id,e.purpose,c.id,c.purpose,c.hypothesis,c.environment_conditions,c.initial_input,c.evaluation_axes,c.fixed_at,x.id,x.conclusion,x.state,x.finalized_at FROM experiments e LEFT JOIN experiment_fixed_conditions c ON c.id=e.fixed_condition_id LEFT JOIN experiment_conclusions x ON x.experiment_id=e.id WHERE e.id=?`, experimentID).Scan(&source.ExperimentID, &source.Purpose, &fixedID, &purpose, &hypothesis, &environment, &input, &axes, &fixedAt, &conclusionID, &conclusion, &state, &finalizedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentDerivationSource{}, false, nil
	}
	if err != nil {
		return domain.ExperimentDerivationSource{}, false, fmt.Errorf("find derivation source: %w", err)
	}
	if fixedID.Valid {
		at, parseErr := time.Parse(time.RFC3339Nano, fixedAt.String)
		if parseErr != nil {
			return domain.ExperimentDerivationSource{}, false, fmt.Errorf("parse derivation fixed time: %w", parseErr)
		}
		source.FixedConditions = &domain.ExperimentFixedConditions{ExperimentID: source.ExperimentID, FixedConditionID: fixedID.String, Purpose: purpose.String, EnvironmentConditions: environment.String, InitialInput: input.String, EvaluationAxes: axes.String, FixedAt: at}
		if hypothesis.Valid {
			source.FixedConditions.Hypothesis = &hypothesis.String
		}
		rows, queryErr := s.db.QueryContext(ctx, "SELECT sequence_no, content FROM experiment_fixed_condition_prompts WHERE fixed_condition_id = ? ORDER BY sequence_no", fixedID.String)
		if queryErr != nil {
			return domain.ExperimentDerivationSource{}, false, fmt.Errorf("find derivation fixed prompts: %w", queryErr)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var prompt domain.ExperimentPreparationPrompt
			if scanErr := rows.Scan(&prompt.SequenceNo, &prompt.Content); scanErr != nil {
				return domain.ExperimentDerivationSource{}, false, fmt.Errorf("scan derivation fixed prompt: %w", scanErr)
			}
			source.FixedConditions.Prompts = append(source.FixedConditions.Prompts, prompt)
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			return domain.ExperimentDerivationSource{}, false, fmt.Errorf("iterate derivation fixed prompts: %w", rowsErr)
		}
	}
	if conclusionID.Valid && state.String == "finalized" {
		at, parseErr := time.Parse(time.RFC3339Nano, finalizedAt.String)
		if parseErr != nil {
			return domain.ExperimentDerivationSource{}, false, fmt.Errorf("parse derivation conclusion time: %w", parseErr)
		}
		source.Conclusion = &domain.ExperimentConclusion{ExperimentID: source.ExperimentID, ConclusionID: conclusionID.String, Conclusion: conclusion.String, State: state.String, FinalizedAt: at}
	}
	if source.FixedConditions == nil {
		source.ReasonCode = "CONDITIONS_NOT_FIXED"
		return source, true, nil
	}
	if source.Conclusion == nil {
		source.ReasonCode = "CONCLUSION_NOT_FINALIZED"
		return source, true, nil
	}
	source.CanCreateDerived = true
	return source, true, nil
}
