package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// StartDerivationBriefingの入力、再生、開始失敗を確認。
func TestStartDerivationBriefingExecute(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		source   string
		store    *derivationBriefingStore
		starter  derivationBriefingStarter
		wantCode apperr.Code
	}{
		{
			name:     "入力不足",
			wantCode: apperr.CodeDerivedExperimentInvalid,
		},
		{
			name:    "成功",
			request: "request-1",
			source:  "source-1",
			store:   &derivationBriefingStore{},
		},
		{
			name:     "ACP未準備",
			request:  "request-1",
			source:   "source-1",
			store:    &derivationBriefingStore{},
			starter:  derivationBriefingStarter{err: apperr.New(apperr.CodeACPNotReady)},
			wantCode: apperr.CodeACPNotReady,
		},
		{
			name:     "永続化失敗",
			request:  "request-1",
			source:   "source-1",
			store:    &derivationBriefingStore{beginErr: errors.New("private sqlite")},
			wantCode: apperr.CodeDerivationBriefingStartFailed,
		},
		{
			name:     "既知の派生元エラー",
			request:  "request-1",
			source:   "source-1",
			store:    &derivationBriefingStore{beginErr: apperr.New(apperr.CodeDerivedExperimentSourceNotFound)},
			wantCode: apperr.CodeDerivedExperimentSourceNotFound,
		},
		{
			name:    "開始中の再生",
			request: "request-1",
			source:  "source-1",
			store: &derivationBriefingStore{start: domain.DerivationBriefingStart{
				RequestID:          "request-1",
				SourceExperimentID: "source-1",
				State:              domain.BriefingStartStateStarting,
			}},
			wantCode: apperr.CodeDerivationBriefingPending,
		},
		{
			name:    "開始済みの再生",
			request: "request-1",
			source:  "source-1",
			store: &derivationBriefingStore{start: domain.DerivationBriefingStart{
				RequestID:          "request-1",
				SourceExperimentID: "source-1",
				BriefingSessionID:  "session-1",
				OperationID:        "operation-1",
				State:              domain.BriefingStartStateStarted,
			}},
		},
		{
			name:    "失敗済みの再生",
			request: "request-1",
			source:  "source-1",
			store: &derivationBriefingStore{start: domain.DerivationBriefingStart{
				RequestID:          "request-1",
				SourceExperimentID: "source-1",
				State:              domain.BriefingStartStateFailed,
				FailureCode:        string(apperr.CodeACPNotReady),
			}},
			wantCode: apperr.CodeACPNotReady,
		},
		{
			name:     "ACP未知失敗",
			request:  "request-1",
			source:   "source-1",
			store:    &derivationBriefingStore{},
			starter:  derivationBriefingStarter{err: errors.New("private ACP")},
			wantCode: apperr.CodeDerivationBriefingStartFailed,
		},
		{
			name:     "失敗記録エラー",
			request:  "request-1",
			source:   "source-1",
			store:    &derivationBriefingStore{markFailedErr: errors.New("private sqlite")},
			starter:  derivationBriefingStarter{err: apperr.New(apperr.CodeACPNotReady)},
			wantCode: apperr.CodeDerivationBriefingStartFailed,
		},
		{
			name:     "開始記録エラー",
			request:  "request-1",
			source:   "source-1",
			store:    &derivationBriefingStore{markStartedErr: errors.New("private sqlite")},
			wantCode: apperr.CodeDerivationBriefingStartFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			if store == nil {
				store = &derivationBriefingStore{}
			}
			got, err := NewStartDerivationBriefing(store, tt.starter).Execute(context.Background(), tt.request, tt.source)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("error = %v, want code %q", err, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got.SourceExperimentID != tt.source || got.State != domain.BriefingStartStateStarted {
				t.Errorf("result = %+v, want source %q and started", got, tt.source)
			}
		})
	}
}

// derivationBriefingFailureの安全なエラー復元を確認。
func TestDerivationBriefingFailure(t *testing.T) {
	tests := []struct {
		name string
		code string
		want apperr.Code
	}{
		{
			name: "ACP未準備",
			code: string(apperr.CodeACPNotReady),
			want: apperr.CodeACPNotReady,
		},
		{
			name: "入力不正",
			code: string(apperr.CodeDerivedExperimentInvalid),
			want: apperr.CodeDerivedExperimentInvalid,
		},
		{
			name: "派生元なし",
			code: string(apperr.CodeDerivedExperimentSourceNotFound),
			want: apperr.CodeDerivedExperimentSourceNotFound,
		},
		{
			name: "派生元不適格",
			code: string(apperr.CodeDerivedExperimentSourceNotEligible),
			want: apperr.CodeDerivedExperimentSourceNotEligible,
		},
		{
			name: "未知エラー",
			code: "UNKNOWN",
			want: apperr.CodeDerivationBriefingStartFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := derivationBriefingFailure(tt.code); !apperr.IsCode(err, tt.want) {
				t.Errorf("derivationBriefingFailure(%q) = %v, want code %q", tt.code, err, tt.want)
			}
		})
	}
}

// derivationBriefingStore は派生実験ブリーフ開始portのtest double。
type derivationBriefingStore struct {
	start          domain.DerivationBriefingStart
	beginErr       error
	markStartedErr error
	markFailedErr  error
}

// BeginDerivationBriefing は開始結果を生成または再生する。
func (s *derivationBriefingStore) BeginDerivationBriefing(_ context.Context, requestID, sourceExperimentID string) (domain.DerivationBriefingStart, bool, error) {
	if s.beginErr != nil {
		return domain.DerivationBriefingStart{}, false, s.beginErr
	}
	if s.start.RequestID != "" {
		return s.start, false, nil
	}
	s.start = domain.DerivationBriefingStart{
		RequestID:          requestID,
		SourceExperimentID: sourceExperimentID,
		BriefingSessionID:  "session-1",
		OperationID:        "operation-1",
		State:              domain.BriefingStartStateStarting,
	}

	return s.start, true, nil
}

// MarkDerivationBriefingStarted は開始済み状態を記録する。
func (s *derivationBriefingStore) MarkDerivationBriefingStarted(_ context.Context, _ string) error {
	if s.markStartedErr != nil {
		return s.markStartedErr
	}
	s.start.State = domain.BriefingStartStateStarted

	return nil
}

// MarkDerivationBriefingFailed は失敗状態を記録する。
func (s *derivationBriefingStore) MarkDerivationBriefingFailed(_ context.Context, _ string, code string) error {
	if s.markFailedErr != nil {
		return s.markFailedErr
	}
	s.start.State = domain.BriefingStartStateFailed
	s.start.FailureCode = code

	return nil
}

// derivationBriefingStarter はACP開始portのtest double。
type derivationBriefingStarter struct{ err error }

// StartExperimentBriefing は開始結果を返す。
func (s derivationBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	return s.err
}
