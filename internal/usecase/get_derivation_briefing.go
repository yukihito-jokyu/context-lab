package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// DerivationBriefingReader は派生実験ブリーフ画面を読み出すport。
type DerivationBriefingReader interface {
	GetDerivationBriefing(context.Context, string) (domain.DerivationBriefing, bool, error)
}

// GetDerivationBriefing は派生実験ブリーフの再読込query。
type GetDerivationBriefing struct {
	reader DerivationBriefingReader
}

// NewGetDerivationBriefing は派生実験ブリーフ再読込queryを生成する。
func NewGetDerivationBriefing(reader DerivationBriefingReader) *GetDerivationBriefing {
	return &GetDerivationBriefing{reader: reader}
}

// Execute は保存済み派生実験ブリーフを取得する。
func (u *GetDerivationBriefing) Execute(ctx context.Context, briefingSessionID string) (domain.DerivationBriefing, error) {
	briefingSessionID = strings.TrimSpace(briefingSessionID)
	if briefingSessionID == "" {
		return domain.DerivationBriefing{}, apperr.New(apperr.CodeDerivationBriefingInvalid)
	}

	briefing, found, err := u.reader.GetDerivationBriefing(ctx, briefingSessionID)
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.DerivationBriefing{}, appErr
		}

		return domain.DerivationBriefing{}, apperr.Wrap(apperr.CodeDerivationBriefingUnavailable, err)
	}
	if !found {
		return domain.DerivationBriefing{}, apperr.New(apperr.CodeDerivationBriefingNotFound)
	}

	return briefing, nil
}
