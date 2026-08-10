package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

const insightWorkspaceCandidatesQuery = `SELECT e.id, e.purpose, c.evaluation_axes, x.id, x.conclusion, x.finalized_at
	FROM experiment_conclusions x
	JOIN experiments e ON e.id = x.experiment_id
	JOIN experiment_fixed_conditions c ON c.id = e.fixed_condition_id
	WHERE x.state = 'finalized'
	ORDER BY x.finalized_at, x.id`

// GetInsightWorkspace は知見作成画面の確定済み比較結論を読み出す。
func (s *Store) GetInsightWorkspace(ctx context.Context) (domain.InsightWorkspace, error) {
	rows, err := s.db.QueryContext(ctx, insightWorkspaceCandidatesQuery)
	if err != nil {
		return domain.InsightWorkspace{}, fmt.Errorf("find insight workspace candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	workspace := domain.InsightWorkspace{
		EvidenceCandidates:  make([]domain.InsightEvidenceCandidate, 0),
		SavedConsiderations: make([]domain.InsightSavedConsideration, 0),
		Insights:            make([]domain.InsightSummary, 0),
	}
	for rows.Next() {
		var candidate domain.InsightEvidenceCandidate
		var finalizedAt string
		if err := rows.Scan(&candidate.ExperimentID, &candidate.Purpose, &candidate.EvaluationAxes, &candidate.ConclusionID, &candidate.Conclusion, &finalizedAt); err != nil {
			return domain.InsightWorkspace{}, fmt.Errorf("scan insight workspace candidate: %w", err)
		}
		parsedFinalizedAt, parseErr := time.Parse(time.RFC3339Nano, finalizedAt)
		if parseErr != nil {
			return domain.InsightWorkspace{}, fmt.Errorf("parse insight workspace candidate time: %w", parseErr)
		}
		candidate.FinalizedAt = parsedFinalizedAt
		workspace.EvidenceCandidates = append(workspace.EvidenceCandidates, candidate)
		workspace.SavedConsiderations = append(workspace.SavedConsiderations, domain.InsightSavedConsideration{
			ExperimentID: candidate.ExperimentID,
			ConclusionID: candidate.ConclusionID,
			Content:      candidate.Conclusion,
			FinalizedAt:  candidate.FinalizedAt,
		})
		if workspace.LastConfirmedAt == nil || candidate.FinalizedAt.After(*workspace.LastConfirmedAt) {
			lastConfirmedAt := candidate.FinalizedAt
			workspace.LastConfirmedAt = &lastConfirmedAt
		}
	}
	if err := rows.Err(); err != nil {
		return domain.InsightWorkspace{}, fmt.Errorf("iterate insight workspace candidates: %w", err)
	}

	return workspace, nil
}
