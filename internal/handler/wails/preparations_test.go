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
			handler := NewPreparationsHandler(usecase.NewListPreparations(tt.reader), usecase.NewGetPreparation(tt.reader), newTestLogger())

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
	reader := handlerPreparationReader{preparations: []domain.Preparation{}}
	handler := NewPreparationsHandler(usecase.NewListPreparations(reader), usecase.NewGetPreparation(reader), newTestLogger())

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

// Wails環境準備詳細の成功と安全な失敗返却。
func TestPreparationsHandlerGetPreparation(t *testing.T) {
	now := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name     string
		input    string
		reader   handlerPreparationReader
		wantCode apperr.Code
	}{
		{
			name:  "詳細を返す",
			input: "preparation-1",
			reader: handlerPreparationReader{preparation: domain.PreparationDetail{
				ID:             "preparation-1",
				State:          "running",
				StartedAt:      now,
				LastObservedAt: now,
				Candidates: []domain.PreparationCandidate{{
					ID:                    "candidate-1",
					EnvironmentConditions: "macOS",
					Summary:               "利用可能",
					CreatedAt:             now,
				}},
				Diagnostics: []domain.PreparationDiagnostic{{
					ID:          "diagnostic-1",
					Code:        "CHECKED",
					SafeSummary: "確認済み",
					OccurredAt:  now,
				}},
				Failure: &domain.PreparationFailure{
					Code:       "FAILED",
					OccurredAt: now,
				},
				Reconciliation: domain.PreparationReconciliation{
					State:          "reconciling",
					LastObservedAt: now,
				},
			},
				found: true,
			},
		},
		{
			name:     "ID不正を返す",
			input:    " ",
			reader:   handlerPreparationReader{},
			wantCode: apperr.CodePreparationRequestInvalid,
		},
		{
			name:     "内部エラーを安全なコードへ変換する",
			input:    "preparation-1",
			reader:   handlerPreparationReader{err: errors.New("private credential")},
			wantCode: apperr.CodePreparationUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPreparationsHandler(usecase.NewListPreparations(&tt.reader), usecase.NewGetPreparation(&tt.reader), newTestLogger())

			got := handler.GetPreparation(tt.input)
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
				t.Fatal("Data = nil, want preparation detail")
			}
			if got.Data.PreparationID != "preparation-1" {
				t.Errorf("PreparationID = %q, want %q", got.Data.PreparationID, "preparation-1")
			}
			if got.Data.Candidates[0].EnvironmentConditions != "macOS" {
				t.Errorf("Candidate.EnvironmentConditions = %q, want %q", got.Data.Candidates[0].EnvironmentConditions, "macOS")
			}
			if got.Data.Diagnostics[0].Summary != "確認済み" {
				t.Errorf("Diagnostic.Summary = %q, want %q", got.Data.Diagnostics[0].Summary, "確認済み")
			}
			if got.Data.Failure == nil || got.Data.Failure.Code != "FAILED" {
				t.Errorf("Failure = %+v, want FAILED", got.Data.Failure)
			}
			if got.Data.Reconciliation.State != "reconciling" {
				t.Errorf("Reconciliation.State = %q, want %q", got.Data.Reconciliation.State, "reconciling")
			}
		})
	}
}

// Wails環境候補採用の成功と安全な失敗返却。
func TestPreparationsHandlerAdoptCandidate(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		reader    handlerPreparationReader
		wantCode  apperr.Code
	}{
		{
			name:      "完了済み候補を返す",
			requestID: "request-1",
			reader: handlerPreparationReader{preparation: domain.PreparationDetail{
				ID:    "preparation-1",
				State: domain.EnvironmentPreparationStateCompleted,
				Candidates: []domain.PreparationCandidate{{
					ID:                    "candidate-1",
					EnvironmentConditions: "macOS 15",
				}},
			}, found: true},
		},
		{
			name:      "内部エラーを安全なコードへ変換する",
			requestID: "request-1",
			reader:    handlerPreparationReader{err: errors.New("private database credential")},
			wantCode:  apperr.CodeCandidateAdoptionUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := tt.reader
			handler := NewPreparationsHandlerWithStartAndAdopt(usecase.NewListPreparations(&reader), usecase.NewGetPreparation(&reader), nil, usecase.NewAdoptCandidate(&reader), newTestLogger())

			got := handler.AdoptCandidate(tt.requestID, "preparation-1", "candidate-1")
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
				t.Fatal("Data = nil, want candidate")
			}
			if got.Data.PreparationID != "preparation-1" {
				t.Errorf("PreparationID = %q, want %q", got.Data.PreparationID, "preparation-1")
			}
			if got.Data.CandidateID != "candidate-1" {
				t.Errorf("CandidateID = %q, want %q", got.Data.CandidateID, "candidate-1")
			}
			if got.Data.EnvironmentConditions != "macOS 15" {
				t.Errorf("EnvironmentConditions = %q, want %q", got.Data.EnvironmentConditions, "macOS 15")
			}
		})
	}
}

// Wails環境候補採用の未設定command失敗返却。
func TestPreparationsHandlerAdoptCandidateWithoutCommand(t *testing.T) {
	handler := NewPreparationsHandler(nil, nil, newTestLogger())

	got := handler.AdoptCandidate("request-1", "preparation-1", "candidate-1")
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeCandidateAdoptionUnavailable) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeCandidateAdoptionUnavailable)
	}
}

// 環境候補採用の予期しない失敗を安全なDTOへ変換。
func TestFailAdoptCandidate(t *testing.T) {
	handler := NewPreparationsHandler(nil, nil, newTestLogger())

	got := handler.failAdoptCandidate(context.Background(), errors.New("private database credential"))
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeUnexpected)
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

// 環境準備詳細の予期しない失敗を安全なDTOへ変換。
func TestFailGetPreparation(t *testing.T) {
	got := failGetPreparation(errors.New("private sidecar credential"))
	if got.Error == nil {
		t.Fatal("Error = nil, want safe error")
	}
	if got.Error.Code != string(apperr.CodeUnexpected) {
		t.Errorf("Error.Code = %q, want %q", got.Error.Code, apperr.CodeUnexpected)
	}
}

// 環境準備開始のアプリケーションエラーを安全なDTOへ変換。
func TestFailStartPreparation(t *testing.T) {
	handler := NewPreparationsHandler(nil, nil, newTestLogger())
	tests := []struct {
		name string
		err  error
		want apperr.Code
	}{
		{
			name: "アプリケーションエラー",
			err:  apperr.New(apperr.CodePreparationScopeInvalid),
			want: apperr.CodePreparationScopeInvalid,
		},
		{
			name: "予期しないエラー",
			err:  errors.New("private credential"),
			want: apperr.CodeUnexpected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.failStartPreparation(context.Background(), tt.err)
			if got.Error == nil {
				t.Fatal("Error = nil, want safe error")
			}
			if got.Error.Code != string(tt.want) {
				t.Errorf("Error.Code = %q, want %q", got.Error.Code, tt.want)
			}
		})
	}
}

// Wails環境準備開始の成功と安全な失敗返却。
func TestPreparationsHandlerStartPreparation(t *testing.T) {
	tests := []struct {
		name     string
		command  *usecase.StartPreparation
		wantCode apperr.Code
		wantID   string
	}{
		{
			name: "開始結果を返す",
			command: usecase.NewStartPreparation(
				&handlerPreparationStartStore{},
				handlerPreparationScopeValidator{},
				handlerEnvironmentPreparer{},
			),
			wantID: "preparation-1",
		},
		{
			name:     "開始command未設定を安全なエラーへ変換する",
			wantCode: apperr.CodePreparationStartUnavailable,
		},
		{
			name: "開始commandの内部エラーを安全なエラーへ変換する",
			command: usecase.NewStartPreparation(
				&handlerPreparationStartStore{beginErr: errors.New("private sqlite credential")},
				handlerPreparationScopeValidator{},
				handlerEnvironmentPreparer{},
			),
			wantCode: apperr.CodePreparationStartUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPreparationsHandlerWithStart(
				usecase.NewListPreparations(handlerPreparationReader{}),
				usecase.NewGetPreparation(handlerPreparationReader{}),
				tt.command,
				newTestLogger(),
			)

			got := handler.StartPreparation("request-1", ".")
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
				t.Fatal("Data = nil, want start result")
			}
			if got.Data.PreparationID != tt.wantID {
				t.Errorf("PreparationID = %q, want %q", got.Data.PreparationID, tt.wantID)
			}
			if got.Data.State != domain.EnvironmentPreparationStateCompleted {
				t.Errorf("State = %q, want %q", got.Data.State, domain.EnvironmentPreparationStateCompleted)
			}
		})
	}
}

// handlerPreparationReader は環境準備一覧query用のtest double。
type handlerPreparationReader struct {
	preparations []domain.Preparation
	preparation  domain.PreparationDetail
	found        bool
	err          error
}

// ListPreparations は指定済みの一覧またはエラーを返す。
func (h handlerPreparationReader) ListPreparations(context.Context) ([]domain.Preparation, error) {
	return h.preparations, h.err
}

// GetPreparation は指定済み詳細または失敗を返す。
func (h handlerPreparationReader) GetPreparation(context.Context, string) (domain.PreparationDetail, bool, error) {
	return h.preparation, h.found, h.err
}

// handlerPreparationScopeValidator は開始範囲検証用のtest double。
type handlerPreparationScopeValidator struct{}

// ValidatePreparationScope は固定済みの安全な範囲を返す。
func (handlerPreparationScopeValidator) ValidatePreparationScope(string) (string, string, error) {
	return "/workspace", ".", nil
}

// handlerEnvironmentPreparer はACP準備用のtest double。
type handlerEnvironmentPreparer struct{}

// PrepareEnvironment は空の安全な準備結果を返す。
func (handlerEnvironmentPreparer) PrepareEnvironment(context.Context, string) (domain.EnvironmentPreparationResult, error) {
	return domain.EnvironmentPreparationResult{}, nil
}

// handlerPreparationStartStore は開始保存用のtest double。
type handlerPreparationStartStore struct {
	beginErr error
}

// BeginPreparation は新規の開始結果または指定済みエラーを返す。
func (s *handlerPreparationStartStore) BeginPreparation(context.Context, string, string) (domain.EnvironmentPreparationStart, bool, error) {
	if s.beginErr != nil {
		return domain.EnvironmentPreparationStart{}, false, s.beginErr
	}

	return domain.EnvironmentPreparationStart{
		PreparationID: "preparation-1",
		State:         domain.EnvironmentPreparationStateStarting,
	}, true, nil
}

// MarkPreparationRunning は開始中状態を受理する。
func (*handlerPreparationStartStore) MarkPreparationRunning(context.Context, string) error {
	return nil
}

// CompletePreparation は完了状態を受理する。
func (*handlerPreparationStartStore) CompletePreparation(context.Context, string, domain.EnvironmentPreparationResult) error {
	return nil
}

// FailPreparation は失敗状態を受理する。
func (*handlerPreparationStartStore) FailPreparation(context.Context, string, string) error {
	return nil
}
