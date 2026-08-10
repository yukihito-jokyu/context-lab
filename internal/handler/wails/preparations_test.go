package wails

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// Wails環境準備一覧の成功と安全な失敗返却。
func TestPreparationsHandlerListPreparations(t *testing.T) {
	startedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	lastObservedAt := startedAt.Add(time.Minute)
	tests := []struct {
		name      string
		reader    handlerPreparationReader
		wantCode  apperr.Code
		wantCount int
	}{
		{
			name: "環境準備一覧を返す",
			reader: handlerPreparationReader{preparations: []domain.Preparation{
				{
					ID:             "preparation-1",
					State:          "running",
					StartedAt:      startedAt,
					LastObservedAt: lastObservedAt,
				},
			}},
			wantCount: 1,
		},
		{
			name:     "内部エラーを安全なコードへ変換する",
			reader:   handlerPreparationReader{err: errors.New("private sidecar credential")},
			wantCode: apperr.CodePreparationsLoadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPreparationsHandler(usecase.NewListPreparations(tt.reader), newTestLogger())

			got := handler.ListPreparations()
			if tt.wantCode != "" {
				if got.Error == nil {
					t.Fatal("Error = nil, want safe error")
				}
				if got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
				}
				if got.Data != nil {
					t.Errorf("Data = %+v, want nil", got.Data)
				}

				return
			}
			if got.Error != nil {
				t.Fatalf("Error = %+v, want nil", got.Error)
			}
			if got.Data == nil {
				t.Fatal("Data = nil, want preparation list")
			}
			if gotCount := len(got.Data.Preparations); gotCount != tt.wantCount {
				t.Fatalf("Preparations length = %d, want %d", gotCount, tt.wantCount)
			}
			preparation := got.Data.Preparations[0]
			if preparation.PreparationID != "preparation-1" {
				t.Errorf("PreparationID = %q, want %q", preparation.PreparationID, "preparation-1")
			}
			if preparation.State != "running" {
				t.Errorf("State = %q, want %q", preparation.State, "running")
			}
			if !preparation.StartedAt.Equal(startedAt.UTC()) {
				t.Errorf("StartedAt = %s, want %s", preparation.StartedAt, startedAt.UTC())
			}
			if !preparation.LastObservedAt.Equal(lastObservedAt.UTC()) {
				t.Errorf("LastObservedAt = %s, want %s", preparation.LastObservedAt, lastObservedAt.UTC())
			}
		})
	}
}

// Wails環境準備一覧の空配列返却。
func TestPreparationsHandlerListPreparationsReturnsEmptySlice(t *testing.T) {
	handler := NewPreparationsHandler(usecase.NewListPreparations(handlerPreparationReader{preparations: []domain.Preparation{}}), newTestLogger())

	got := handler.ListPreparations()
	if got.Error != nil {
		t.Fatalf("Error = %+v, want nil", got.Error)
	}
	if got.Data == nil {
		t.Fatal("Data = nil, want empty preparation list")
	}
	if got.Data.Preparations == nil {
		t.Error("Preparations = nil, want empty slice")
	}
	if gotCount := len(got.Data.Preparations); gotCount != 0 {
		t.Errorf("Preparations length = %d, want 0", gotCount)
	}
}

// 環境準備一覧の予期しない失敗を安全なDTOへ変換。
func TestFailListPreparations(t *testing.T) {
	got := failListPreparations(errors.New("private sidecar credential"))
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeUnexpected)
	}
}

// handlerPreparationReader は環境準備一覧query用のtest double。
type handlerPreparationReader struct {
	preparations []domain.Preparation
	err          error
}

// ListPreparations は指定済みの一覧またはエラーを返す。
func (h handlerPreparationReader) ListPreparations(context.Context) ([]domain.Preparation, error) {
	return h.preparations, h.err
}
