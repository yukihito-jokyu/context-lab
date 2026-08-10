package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// GetDerivationBriefingの入力検証とport失敗正規化。
func TestGetDerivationBriefingExecute(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name      string
		sessionID string
		reader    fakeDerivationBriefingReader
		wantCode  apperr.Code
		wantFound bool
	}{
		{
			name:      "空のsession IDを拒否する",
			sessionID: " ",
			wantCode:  apperr.CodeDerivationBriefingInvalid,
		},
		{
			name:      "未知sessionを区別する",
			sessionID: "missing",
			wantCode:  apperr.CodeDerivationBriefingNotFound,
		},
		{
			name:      "永続化失敗を取得失敗へ正規化する",
			sessionID: "session-1",
			reader:    fakeDerivationBriefingReader{err: errors.New("repository failure")},
			wantCode:  apperr.CodeDerivationBriefingUnavailable,
		},
		{
			name:      "取消を呼び出し元へ返す",
			sessionID: "session-1",
			reader:    fakeDerivationBriefingReader{err: context.Canceled},
			wantCode:  apperr.CodeOperationCanceled,
		},
		{
			name:      "開始直後の空状態を返す",
			sessionID: "session-1",
			reader: fakeDerivationBriefingReader{
				found: true,
				briefing: domain.DerivationBriefing{
					State:           domain.BriefingStartStateStarted,
					Messages:        []domain.DerivationBriefingMessage{},
					LastConfirmedAt: confirmedAt,
				},
			},
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGetDerivationBriefing(tt.reader).Execute(context.Background(), tt.sessionID)
			if tt.wantCode != "" {
				assertBriefingErrorCode(t, err, tt.wantCode)

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if gotFound := !got.LastConfirmedAt.IsZero(); gotFound != tt.wantFound {
				t.Errorf("briefing found = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

// fakeDerivationBriefingReader は派生実験ブリーフqueryの読み出し結果を固定する。
type fakeDerivationBriefingReader struct {
	briefing domain.DerivationBriefing
	found    bool
	err      error
}

// GetDerivationBriefing は固定の派生実験ブリーフ読み出し結果を返す。
func (f fakeDerivationBriefingReader) GetDerivationBriefing(context.Context, string) (domain.DerivationBriefing, bool, error) {
	return f.briefing, f.found, f.err
}
