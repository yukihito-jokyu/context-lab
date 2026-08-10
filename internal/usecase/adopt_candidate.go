package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// AdoptedCandidate は採用確認済み環境候補の安全な値。
type AdoptedCandidate struct {
	PreparationID         string
	CandidateID           string
	EnvironmentConditions string
}

// CandidateAdoptionReader は候補採用確認用の環境準備読み出しport。
type CandidateAdoptionReader interface {
	GetPreparation(context.Context, string) (domain.PreparationDetail, bool, error)
}

// AdoptCandidate は完了済み環境準備の候補を採用可能な値として返すcommand。
type AdoptCandidate struct {
	reader CandidateAdoptionReader
}

// NewAdoptCandidate は環境候補採用commandを生成。
func NewAdoptCandidate(reader CandidateAdoptionReader) *AdoptCandidate {
	return &AdoptCandidate{reader: reader}
}

// Execute は完了済み環境準備から指定候補を安全に取得。
func (u *AdoptCandidate) Execute(ctx context.Context, requestID, preparationID, candidateID string) (AdoptedCandidate, error) {
	if strings.TrimSpace(requestID) == "" || strings.TrimSpace(preparationID) == "" || strings.TrimSpace(candidateID) == "" {
		return AdoptedCandidate{}, apperr.New(apperr.CodeCandidateAdoptionRequestInvalid)
	}

	preparation, found, err := u.reader.GetPreparation(ctx, strings.TrimSpace(preparationID))
	if err != nil {
		return AdoptedCandidate{}, apperr.Wrap(apperr.CodeCandidateAdoptionUnavailable, err)
	}
	if !found {
		return AdoptedCandidate{}, apperr.New(apperr.CodePreparationNotFound)
	}
	if preparation.State != domain.EnvironmentPreparationStateCompleted {
		return AdoptedCandidate{}, apperr.New(apperr.CodePreparationCandidateNotReady)
	}

	for _, candidate := range preparation.Candidates {
		if candidate.ID == strings.TrimSpace(candidateID) {
			return AdoptedCandidate{PreparationID: preparation.ID, CandidateID: candidate.ID, EnvironmentConditions: candidate.EnvironmentConditions}, nil
		}
	}

	return AdoptedCandidate{}, apperr.New(apperr.CodePreparationCandidateNotFound)
}
