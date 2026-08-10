package usecase

import (
	"context"
	"sort"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// InsightCreator は知見作成の永続化port。
type InsightCreator interface {
	CreateInsight(context.Context, string, []domain.InsightEvidence, string, string, string) (domain.Insight, bool, error)
}

// CreateInsight は知見作成のapplication service。
type CreateInsight struct{ creator InsightCreator }

// NewCreateInsight は知見作成usecaseを生成する。
func NewCreateInsight(creator InsightCreator) *CreateInsight { return &CreateInsight{creator: creator} }

// Execute は正規化した知見を保存する。
func (u *CreateInsight) Execute(ctx context.Context, requestID string, evidences []domain.InsightEvidence, statement, conditions, gaps string) (domain.Insight, error) {
	requestID, statement, conditions, gaps = strings.TrimSpace(requestID), strings.TrimSpace(statement), strings.TrimSpace(conditions), strings.TrimSpace(gaps)
	if requestID == "" || statement == "" || conditions == "" || gaps == "" {
		return domain.Insight{}, apperr.New(apperr.CodeInsightCreateInvalid)
	}

	canonical, err := canonicalInsightEvidences(evidences)
	if err != nil {
		return domain.Insight{}, err
	}
	if len(canonical) < 2 {
		return domain.Insight{}, apperr.New(apperr.CodeInsightCreateEvidenceInsufficient)
	}

	got, _, err := u.creator.CreateInsight(ctx, requestID, canonical, statement, conditions, gaps)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.Insight{}, appErr
		}
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.Insight{}, appErr
		}
		return domain.Insight{}, apperr.Wrap(apperr.CodeInsightCreateUnavailable, err)
	}

	return got, nil
}

// canonicalInsightEvidences は根拠を検証し安定順へ整列する。
func canonicalInsightEvidences(evidences []domain.InsightEvidence) ([]domain.InsightEvidence, error) {
	canonical := make([]domain.InsightEvidence, len(evidences))
	seen := make(map[string]struct{}, len(evidences))
	for index, evidence := range evidences {
		evidence.ExperimentID = strings.TrimSpace(evidence.ExperimentID)
		evidence.ConclusionID = strings.TrimSpace(evidence.ConclusionID)
		if evidence.ExperimentID == "" || evidence.ConclusionID == "" {
			return nil, apperr.New(apperr.CodeInsightCreateInvalid)
		}
		if _, exists := seen[evidence.ExperimentID]; exists {
			return nil, apperr.New(apperr.CodeInsightCreateInvalid)
		}
		seen[evidence.ExperimentID] = struct{}{}
		canonical[index] = evidence
	}
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].ExperimentID == canonical[j].ExperimentID {
			return canonical[i].ConclusionID < canonical[j].ConclusionID
		}
		return canonical[i].ExperimentID < canonical[j].ExperimentID
	})

	return canonical, nil
}
