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
			handler := NewExperimentBriefingsHandler(usecase.NewStartExperimentBriefing(store, handlerBriefingStarter{}), usecase.NewSendExperimentBriefMessage(store, handlerBriefingMessageSender{}), usecase.NewGetExperimentBriefing(store), usecase.NewCreateExperimentFromBrief(store), newTestLogger())

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

// Wails実験ブリーフ会話送信の成功返却。
func TestExperimentBriefingsHandlerSendExperimentBriefMessage(t *testing.T) {
	store := &handlerBriefingStore{}
	handler := NewExperimentBriefingsHandler(
		usecase.NewStartExperimentBriefing(store, handlerBriefingStarter{}),
		usecase.NewSendExperimentBriefMessage(store, handlerBriefingMessageSender{}),
		usecase.NewGetExperimentBriefing(store),
		usecase.NewCreateExperimentFromBrief(store),
		newTestLogger(),
	)

	got := handler.SendExperimentBriefMessage("request-1", "session-1", "目的を確認したい")
	if got.Error != nil {
		t.Fatalf("Error = %+v, want nil", got.Error)
	}
	if got.Data == nil {
		t.Fatal("Data = nil, want operation identifier")
	}
	if got := got.Data.OperationID; got != "message-operation-1" {
		t.Errorf("OperationID = %q, want %q", got, "message-operation-1")
	}
}

// Wails実験ブリーフ終了の成功と安全な失敗返却。
func TestExperimentBriefingsHandlerStopExperimentBriefing(t *testing.T) {
	tests := []struct {
		name       string
		stopperErr error
		wantCode   apperr.Code
	}{
		{
			name: "終了操作識別子だけを返す",
		},
		{
			name:       "内部エラーを安全なコードへ変換する",
			stopperErr: errors.New("private ACP credential"),
			wantCode:   apperr.CodeBriefingStopFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &handlerBriefingStore{}
			stopper := handlerBriefingStopper{err: tt.stopperErr}
			handler := NewExperimentBriefingsHandler(
				usecase.NewStartExperimentBriefing(store, handlerBriefingStarter{}),
				usecase.NewSendExperimentBriefMessage(store, handlerBriefingMessageSender{}),
				usecase.NewGetExperimentBriefing(store),
				usecase.NewCreateExperimentFromBrief(store),
				newTestLogger(),
				usecase.NewStopExperimentBriefing(store, stopper),
			)

			got := handler.StopExperimentBriefing("request-1", "session-1")
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
				t.Fatal("Data = nil, want operation identifier")
			}
			if got.Data.OperationID != "stop-operation-1" {
				t.Errorf("OperationID = %q, want %q", got.Data.OperationID, "stop-operation-1")
			}
		})
	}
}

// 実験ブリーフ終了失敗の安全な変換。
func TestFailStopExperimentBriefing(t *testing.T) {
	got := failStopExperimentBriefing(errors.New("private ACP credential"))
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeUnexpected)
	}
}

// Wails実験ブリーフ採用の成功と安全な失敗返却。
func TestExperimentBriefingsHandlerCreateExperimentFromBrief(t *testing.T) {
	tests := []struct {
		name      string
		createErr error
		wantCode  apperr.Code
	}{
		{
			name: "準備中実験を返す",
		},
		{
			name:      "内部エラーを安全に返す",
			createErr: errors.New("private database detail"),
			wantCode:  apperr.CodeExperimentCreateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &handlerBriefingStore{
				createErr: tt.createErr,
			}
			handler := NewExperimentBriefingsHandler(usecase.NewStartExperimentBriefing(store, handlerBriefingStarter{}), usecase.NewSendExperimentBriefMessage(store, handlerBriefingMessageSender{}), usecase.NewGetExperimentBriefing(store), usecase.NewCreateExperimentFromBrief(store), newTestLogger())

			got := handler.CreateExperimentFromBrief("request-1", "session-1", "version-1")
			if tt.wantCode != "" {
				if got.Error == nil {
					t.Fatal("Error = nil, want safe error")
				}
				if got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
				}

				return
			}
			if got.Error != nil {
				t.Fatalf("Error = %+v, want nil", got.Error)
			}
			if got.Data == nil {
				t.Fatal("Data = nil, want experiment")
			}
			if got.Data.ExperimentID != "experiment-1" || got.Data.State != "preparing" {
				t.Errorf("Data = %+v, want preparing experiment", got.Data)
			}
		})
	}
}

// 実験ブリーフ採用失敗の安全な変換。
func TestFailCreateExperimentFromBrief(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode apperr.Code
	}{
		{
			name:     "想定外エラーを安全に変換する",
			err:      errors.New("private database detail"),
			wantCode: apperr.CodeUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := failCreateExperimentFromBrief(tt.err)
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
			}
		})
	}
}

// Wails実験ブリーフ会話送信の安全な失敗返却。
func TestExperimentBriefingsHandlerSendExperimentBriefMessageFailure(t *testing.T) {
	store := &handlerBriefingStore{}
	handler := NewExperimentBriefingsHandler(
		usecase.NewStartExperimentBriefing(store, handlerBriefingStarter{}),
		usecase.NewSendExperimentBriefMessage(store, handlerBriefingMessageSender{err: errors.New("private ACP credential")}),
		usecase.NewGetExperimentBriefing(store),
		usecase.NewCreateExperimentFromBrief(store),
		newTestLogger(),
	)

	got := handler.SendExperimentBriefMessage("request-1", "session-1", "目的を確認したい")
	if got.Data != nil {
		t.Errorf("Data = %+v, want nil", got.Data)
	}
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeBriefingMessageFailed) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeBriefingMessageFailed)
	}
}

// 実験ブリーフ会話送信失敗の安全な変換。
func TestFailSendExperimentBriefMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode apperr.Code
	}{
		{
			name:     "通常エラーを想定外エラーへ変換する",
			err:      errors.New("private ACP credential"),
			wantCode: apperr.CodeUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := failSendExperimentBriefMessage(tt.err)
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
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

// 実験ブリーフ再読込失敗の安全な変換。
func TestFailGetExperimentBriefing(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode apperr.Code
	}{
		{
			name:     "アプリケーションエラーを変換する",
			err:      apperr.New(apperr.CodeBriefingLoadFailed),
			wantCode: apperr.CodeBriefingLoadFailed,
		},
		{
			name:     "通常エラーを想定外エラーへ変換する",
			err:      errors.New("private database detail"),
			wantCode: apperr.CodeUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := failGetExperimentBriefing(tt.err)
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
			}
		})
	}
}

// Wails実験ブリーフ再読込の成功と安全な失敗返却。
func TestExperimentBriefingsHandlerGetExperimentBriefing(t *testing.T) {
	confirmedAt := time.Date(2026, time.August, 9, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	tests := []struct {
		name      string
		store     handlerBriefingStore
		sessionID string
		wantCode  apperr.Code
		wantData  bool
	}{
		{
			name: "会話と最新ブリーフを画面DTOへ変換する",
			store: handlerBriefingStore{
				found: true,
				briefing: domain.ExperimentBriefing{
					State: "started",
					Messages: []domain.ExperimentBriefingMessage{{
						Role:       "user",
						Content:    "目的",
						SequenceNo: 1,
						CreatedAt:  confirmedAt,
					}},
					LatestBrief: &domain.ExperimentBrief{
						VersionID:          "version-1",
						Decision:           "比較する",
						SuccessCriteria:    "正確性",
						RequiredConditions: "固定条件",
					},
					LastConfirmedAt: confirmedAt,
				},
			},
			sessionID: "session-1",
			wantData:  true,
		},
		{
			name:      "未知セッションを安全なコードへ変換する",
			store:     handlerBriefingStore{},
			sessionID: "missing",
			wantCode:  apperr.CodeBriefingNotFound,
		},
		{
			name: "内部エラーを安全なコードへ変換する",
			store: handlerBriefingStore{
				getErr: errors.New("SELECT private_content"),
			},
			sessionID: "session-2",
			wantCode:  apperr.CodeBriefingLoadFailed,
		},
		{
			name: "ブリーフ未作成時も空sliceとnilを返す",
			store: handlerBriefingStore{
				found: true,
				briefing: domain.ExperimentBriefing{
					State:           "started",
					LastConfirmedAt: confirmedAt,
				},
			},
			sessionID: "session-3",
			wantData:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewExperimentBriefingsHandler(
				usecase.NewStartExperimentBriefing(&tt.store, handlerBriefingStarter{}),
				usecase.NewSendExperimentBriefMessage(&tt.store, handlerBriefingMessageSender{}),
				usecase.NewGetExperimentBriefing(&tt.store),
				usecase.NewCreateExperimentFromBrief(&tt.store),
				newTestLogger(),
			)

			got := handler.GetExperimentBriefing(tt.sessionID)
			if gotData := got.Data != nil; gotData != tt.wantData {
				t.Fatalf("Data available = %v, want %v", gotData, tt.wantData)
			}
			if tt.wantData {
				if tt.sessionID == "session-3" {
					if got := got.Data.Messages; len(got) != 0 {
						t.Errorf("Messages = %+v, want empty slice", got)
					}
					if got := got.Data.LatestBrief; got != nil {
						t.Errorf("LatestBrief = %+v, want nil", got)
					}

					return
				}
				if got := got.Data.Messages; len(got) != 1 || got[0].Content != "目的" {
					t.Errorf("Messages = %+v, want one user message", got)
				}
				if got := got.Data.LatestBrief; got == nil || got.VersionID != "version-1" {
					t.Errorf("LatestBrief = %+v, want version-1", got)
				}
				if got := got.Data.LastConfirmedAt; !got.Equal(confirmedAt.UTC()) || got.Location() != time.UTC {
					t.Errorf("LastConfirmedAt = %s, want UTC %s", got, confirmedAt.UTC())
				}

				return
			}
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != string(tt.wantCode) {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if strings.Contains(got.Error.Message, "private_content") {
				t.Errorf("Error.Message = %q, want no SQL detail", got.Error.Message)
			}
		})
	}
}

// handlerBriefingStore はhandler用開始記録portのtest double。
type handlerBriefingStore struct {
	beginErr  error
	briefing  domain.ExperimentBriefing
	getErr    error
	createErr error
	found     bool
}

// BeginStopExperimentBriefing は指定済み終了操作を返却。
func (*handlerBriefingStore) BeginStopExperimentBriefing(_ context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingStopOperation, bool, error) {
	return domain.ExperimentBriefingStopOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "stop-operation-1",
		State:             domain.BriefingStartStateStarting,
	}, true, nil
}

// CompleteStopExperimentBriefing は終了完了を受理。
func (*handlerBriefingStore) CompleteStopExperimentBriefing(context.Context, string) error {
	return nil
}

// FailStopExperimentBriefing は終了失敗を受理。
func (*handlerBriefingStore) FailStopExperimentBriefing(context.Context, string, string) error {
	return nil
}

// 指定済み採用結果返却。
func (s *handlerBriefingStore) CreateExperimentFromBrief(context.Context, string, string, string) (domain.ExperimentCreation, bool, error) {
	if s.createErr != nil {
		return domain.ExperimentCreation{}, false, s.createErr
	}

	return domain.ExperimentCreation{
		ExperimentID: "experiment-1",
		State:        "preparing",
	}, true, nil
}

// GetExperimentBriefing は指定済み実験ブリーフを返却。
func (s *handlerBriefingStore) GetExperimentBriefing(context.Context, string) (domain.ExperimentBriefing, bool, error) {
	return s.briefing, s.found, s.getErr
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

// BeginExperimentBriefMessage は指定済み送信結果を返却。
func (*handlerBriefingStore) BeginExperimentBriefMessage(_ context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingMessageOperation, bool, error) {
	return domain.ExperimentBriefingMessageOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       "message-operation-1",
		State:             domain.BriefingStartStateStarting,
	}, true, nil
}

// CompleteExperimentBriefMessage は送信結果を受理。
func (*handlerBriefingStore) CompleteExperimentBriefMessage(context.Context, string, string, domain.ExperimentBriefingMessageResult) error {
	return nil
}

// FailExperimentBriefMessage は送信失敗を受理。
func (*handlerBriefingStore) FailExperimentBriefMessage(context.Context, string, string) error {
	return nil
}

// handlerBriefingStarter はhandler用外部開始portのtest double。
type handlerBriefingStarter struct{}

// StartExperimentBriefing は開始を受理。
func (handlerBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	return nil
}

// handlerBriefingMessageSender はhandler用会話送信portのtest double。
type handlerBriefingMessageSender struct {
	err error
}

// handlerBriefingStopper はhandler用外部停止portのtest double。
type handlerBriefingStopper struct {
	err error
}

// StopExperimentBriefing は指定済み結果を返却。
func (s handlerBriefingStopper) StopExperimentBriefing(context.Context, string, string) error {
	return s.err
}

// SendExperimentBriefMessage は安全な応答を返却。
func (s handlerBriefingMessageSender) SendExperimentBriefMessage(context.Context, string, string, string) (domain.ExperimentBriefingMessageResult, error) {
	return domain.ExperimentBriefingMessageResult{}, s.err
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
