package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 結論確定commandの入力とエラー契約を検証する。
func TestFinalizeExperimentConclusionExecute(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		experiment string
		conclusion string
		err        error
		wantCode   apperr.Code
	}{
		{
			name:       "成功",
			request:    " request ",
			experiment: " experiment ",
			conclusion: " 結論 ",
		},
		{
			name:     "空入力",
			wantCode: apperr.CodeExperimentConclusionInvalid,
		},
		{
			name:       "期限超過",
			request:    "request",
			experiment: "experiment",
			conclusion: "結論",
			err:        context.DeadlineExceeded,
			wantCode:   apperr.CodeOperationTimeout,
		},
		{
			name:       "キャンセル",
			request:    "request",
			experiment: "experiment",
			conclusion: "結論",
			err:        context.Canceled,
			wantCode:   apperr.CodeOperationCanceled,
		},
		{
			name:       "既知アプリケーションエラー",
			request:    "request",
			experiment: "experiment",
			conclusion: "結論",
			err:        apperr.New(apperr.CodeExperimentConclusionNotReady),
			wantCode:   apperr.CodeExperimentConclusionNotReady,
		},
		{
			name:       "内部エラー",
			request:    "request",
			experiment: "experiment",
			conclusion: "結論",
			err:        errors.New("private"),
			wantCode:   apperr.CodeExperimentConclusionUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &conclusionFinalizerStub{err: tt.err}
			got, err := NewFinalizeExperimentConclusion(store).Execute(context.Background(), tt.request, tt.experiment, tt.conclusion)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want %q", err, tt.wantCode)
				}
				return
			}
			if err != nil || got.ExperimentID != "experiment" || store.requestID != "request" {
				t.Errorf("Execute() = (%+v, %v), input = %q, want normalized success", got, err, store.requestID)
			}
		})
	}
}

// conclusionFinalizerStub は結論確定portのtest double。
type conclusionFinalizerStub struct {
	requestID string
	err       error
}

// FinalizeExperimentConclusion は設定済み結果を返す。
func (s *conclusionFinalizerStub) FinalizeExperimentConclusion(_ context.Context, requestID, experimentID, conclusion string) (domain.ExperimentConclusion, bool, error) {
	s.requestID = requestID
	return domain.ExperimentConclusion{
		RequestID:    requestID,
		ExperimentID: experimentID,
		Conclusion:   conclusion,
	}, true, s.err
}
