package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// GetRunDetailの入力検証と安全なエラー変換。
func TestGetRunDetailExecute(t *testing.T) {
	tests := []struct {
		name     string
		runID    string
		reader   getRunDetailReader
		wantCode apperr.Code
		wantID   string
	}{
		{
			name:  "詳細を返す",
			runID: " run-1 ",
			reader: getRunDetailReader{
				found: true,
				detail: domain.ExperimentRunDetail{
					Run: domain.ExperimentRunFact{ID: "run-1"},
				},
			},
			wantID: "run-1",
		},
		{
			name:     "空IDを拒否する",
			wantCode: apperr.CodeRunDetailRequestInvalid,
		},
		{
			name:     "未検出を返す",
			runID:    "run-1",
			wantCode: apperr.CodeRunDetailNotFound,
		},
		{
			name:  "内部エラーを安全に変換する",
			runID: "run-1",
			reader: getRunDetailReader{
				err: errors.New("private database detail"),
			},
			wantCode: apperr.CodeRunDetailUnavailable,
		},
		{
			name:  "期限超過を返す",
			runID: "run-1",
			reader: getRunDetailReader{
				err: context.DeadlineExceeded,
			},
			wantCode: apperr.CodeOperationTimeout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail, err := NewGetRunDetail(tt.reader).Execute(context.Background(), tt.runID)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if detail.Run.ID != tt.wantID {
				t.Errorf("Run.ID = %q, want %q", detail.Run.ID, tt.wantID)
			}
		})
	}
}

// getRunDetailReader はrun詳細readerのtest double。
type getRunDetailReader struct {
	detail domain.ExperimentRunDetail
	found  bool
	err    error
}

// GetRunDetail は設定済み詳細を返却。
func (r getRunDetailReader) GetRunDetail(context.Context, string) (domain.ExperimentRunDetail, bool, error) {
	return r.detail, r.found, r.err
}

// GetRunDetailの時刻値維持。
func TestGetRunDetailDetailTime(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	reader := getRunDetailReader{
		found: true,
		detail: domain.ExperimentRunDetail{
			LastConfirmedAt: confirmedAt,
		},
	}
	detail, err := NewGetRunDetail(reader).Execute(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !detail.LastConfirmedAt.Equal(confirmedAt) {
		t.Errorf("LastConfirmedAt = %s, want %s", detail.LastConfirmedAt, confirmedAt)
	}
}
