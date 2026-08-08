package wails

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/logger"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
	"io"
	"log/slog"
	"strings"
	"time"
)

// Wails実験ブリーフ開始の成功と安全な失敗返却。
func TestExperimentBriefingsHandlerStartExperimentBriefing(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		storeErr  error
		wantCode  apperr.Code
	}{
		{
			name:      "開始識別子だけを返す",
			requestID: "request-1",
		},
		{
			name:      "内部エラーを安全なコードへ変換する",
			requestID: "request-2",
			storeErr:  errors.New("database credential leaked"),
			wantCode:  apperr.CodeBriefingStartFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &handlerBriefingStore{beginErr: tt.storeErr}
			handler := NewExperimentBriefingsHandler(usecase.NewStartExperimentBriefing(store, handlerBriefingStarter{}), newTestLogger())

			got := handler.StartExperimentBriefing(tt.requestID)
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
				t.Fatal("Data = nil, want start identifiers")
			}
			if got.Data.BriefingSessionID != "session-1" {
				t.Errorf("BriefingSessionID = %q, want %q", got.Data.BriefingSessionID, "session-1")
			}
			if got.Data.OperationID != "operation-1" {
				t.Errorf("OperationID = %q, want %q", got.Data.OperationID, "operation-1")
			}
		})
	}
}

// Wails実験ブリーフ開始の想定外エラー変換。
func TestFailStartExperimentBriefing(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode apperr.Code
	}{
		{
			name:     "想定外エラーを安全なコードへ変換する",
			err:      errors.New("internal detail"),
			wantCode: apperr.CodeUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := failStartExperimentBriefing(tt.err)
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
			}
		})
	}
}

// handlerBriefingStore はhandler用開始記録portのtest double。
type handlerBriefingStore struct {
	beginErr error
}

// BeginExperimentBriefing は指定済み開始結果を返却。
func (s *handlerBriefingStore) BeginExperimentBriefing(context.Context, string) (domain.ExperimentBriefingStart, bool, error) {
	if s.beginErr != nil {
		return domain.ExperimentBriefingStart{}, false, s.beginErr
	}

	return domain.ExperimentBriefingStart{
		BriefingSessionID: "session-1",
		OperationID:       "operation-1",
		State:             domain.BriefingStartStateStarting,
	}, true, nil
}

// MarkExperimentBriefingStarted は開始済み状態を受理。
func (*handlerBriefingStore) MarkExperimentBriefingStarted(context.Context, string) error {
	return nil
}

// MarkExperimentBriefingFailed は失敗状態を受理。
func (*handlerBriefingStore) MarkExperimentBriefingFailed(context.Context, string, string) error {
	return nil
}

// handlerBriefingStarter はhandler用外部開始portのtest double。
type handlerBriefingStarter struct{}

// StartExperimentBriefing は開始を受理。
func (handlerBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	return nil
}

// ListExperimentsの成功と安全な失敗返却。
func TestExperimentsHandlerListExperiments(t *testing.T) {
	updatedAt := time.Date(2026, time.August, 8, 3, 0, 0, 0, time.UTC)
	derivedFromExperimentID := "experiment-1"
	tests := []struct {
		name          string
		reader        fakeExperimentReader
		wantErrorCode string
		wantData      bool
	}{
		{
			name: "実験一覧を画面DTOへ変換する",
			reader: fakeExperimentReader{collection: domain.ExperimentCollection{
				Experiments: []domain.Experiment{
					{
						ID:                      "experiment-2",
						Purpose:                 "評価",
						State:                   "running",
						ProgressSummary:         "実行中",
						DerivedFromExperimentID: &derivedFromExperimentID,
						UpdatedAt:               updatedAt,
					},
				},
				CancelledExperiments: []domain.Experiment{},
			}},
			wantData: true,
		},
		{
			name:          "repository失敗を安全なエラーへ変換する",
			reader:        fakeExperimentReader{err: errors.New("SELECT secret_column")},
			wantErrorCode: "EXPERIMENTS_UNAVAILABLE",
		},
		{
			name:          "取消を安全なエラーへ変換する",
			reader:        fakeExperimentReader{err: context.Canceled},
			wantErrorCode: "OPERATION_CANCELED",
		},
		{
			name:          "タイムアウトを安全なエラーへ変換する",
			reader:        fakeExperimentReader{err: context.DeadlineExceeded},
			wantErrorCode: "OPERATION_TIMEOUT",
		},
		{
			name:          "未分類の失敗を安全なエラーへ変換する",
			reader:        fakeExperimentReader{err: errors.New("internal detail")},
			wantErrorCode: "EXPERIMENTS_UNAVAILABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewExperimentsHandler(usecase.NewListExperiments(tt.reader), newTestLogger())

			got := handler.ListExperiments()
			if gotData := got.Data != nil; gotData != tt.wantData {
				t.Fatalf("Data available = %v, want %v", gotData, tt.wantData)
			}
			if tt.wantData {
				if got := got.Data.ResumeSummary.RecommendedExperimentID; got == nil || *got != "experiment-2" {
					t.Errorf("RecommendedExperimentID = %v, want %q", got, "experiment-2")
				}
				if got := got.Data.ResumeSummary.StatusCounts["running"]; got != 1 {
					t.Errorf("StatusCounts[running] = %d, want %d", got, 1)
				}
				if got := len(got.Data.CancelledExperiments); got != 0 {
					t.Errorf("CancelledExperiments length = %d, want %d", got, 0)
				}

				return
			}
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != tt.wantErrorCode {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantErrorCode)
			}
			if strings.Contains(got.Error.Message, "secret_column") {
				t.Errorf("Error.Message = %q, want no SQL detail", got.Error.Message)
			}
		})
	}
}

// failListExperimentsは公開可能なエラーだけをDTOへ変換する。
func TestFailListExperiments(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    string
		wantMessage string
	}{
		{
			name:        "アプリケーションエラーを変換する",
			err:         apperr.New(apperr.CodeExperimentsLoadFailed),
			wantCode:    "EXPERIMENTS_UNAVAILABLE",
			wantMessage: "実験一覧を取得できませんでした",
		},
		{
			name:        "通常エラーを想定外エラーへ変換する",
			err:         errors.New("private database detail"),
			wantCode:    "UNEXPECTED",
			wantMessage: "予期しないエラーが発生しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := failListExperiments(tt.err)
			if got.Error == nil {
				t.Fatal("Error = nil, want error response")
			}
			if got.Error.Code != tt.wantCode {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if got.Error.Message != tt.wantMessage {
				t.Errorf("Error.Message = %q, want %q", got.Error.Message, tt.wantMessage)
			}
		})
	}
}

// 成功時刻の画面側保持契約。
func TestExperimentsHandlerListExperimentsRetainsLastSuccessForClient(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 8, 3, 4, 5, 0, time.UTC)
	reader := &sequenceExperimentReader{results: []readerResult{
		{
			collection: domain.ExperimentCollection{
				Experiments:          []domain.Experiment{},
				CancelledExperiments: []domain.Experiment{},
				LastConfirmedAt:      &confirmedAt,
			},
		},
		{
			err: errors.New("SELECT raw_sql"),
		},
	}}
	handler := NewExperimentsHandler(usecase.NewListExperiments(reader), newTestLogger())

	first := handler.ListExperiments()
	if first.Data == nil || first.Data.LastConfirmedAt == nil {
		t.Fatal("first ListExperiments() data = nil, want confirmed timestamp")
	}
	lastSuccessfulConfirmation := *first.Data.LastConfirmedAt

	second := handler.ListExperiments()
	if second.Error == nil {
		t.Fatal("second ListExperiments() error = nil, want safe error")
	}
	if got := second.Error.Code; got != "EXPERIMENTS_UNAVAILABLE" {
		t.Errorf("second Error.Code = %q, want %q", got, "EXPERIMENTS_UNAVAILABLE")
	}
	if got := lastSuccessfulConfirmation; !got.Equal(confirmedAt) {
		t.Errorf("retained LastConfirmedAt = %s, want %s", got, confirmedAt)
	}
}

// newTestLogger はテスト出力を抑止するロガー。
func newTestLogger() logger.Logger {
	return logger.NewWithWriter(io.Discard, slog.LevelDebug)
}

// fakeExperimentReader は一覧読み出しportのtest double。
type fakeExperimentReader struct {
	collection domain.ExperimentCollection
	err        error
}

// ListExperiments は指定済みの一覧またはエラーを返す。
func (f fakeExperimentReader) ListExperiments(context.Context) (domain.ExperimentCollection, error) {
	return f.collection, f.err
}

// readerResult は順番に返す一覧結果。
type readerResult struct {
	collection domain.ExperimentCollection
	err        error
}

// sequenceExperimentReader は連続した一覧結果を返すtest double。
type sequenceExperimentReader struct {
	results []readerResult
	index   int
}

// ListExperiments は次の指定済み一覧またはエラーを返す。
func (s *sequenceExperimentReader) ListExperiments(context.Context) (domain.ExperimentCollection, error) {
	result := s.results[s.index]
	s.index++

	return result.collection, result.err
}
