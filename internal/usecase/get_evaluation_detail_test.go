package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// GetEvaluationDetail は入力を正規化し、安全な評価詳細またはエラーコードを返す。
func TestGetEvaluationDetailExecute(t *testing.T) {
	tests := []struct {
		name       string
		evaluation string
		reader     *getEvaluationDetailReader
		wantCode   apperr.Code
		wantID     string
		wantInput  string
	}{
		{
			name:       "詳細を返す",
			evaluation: " evaluation-1 ",
			reader: &getEvaluationDetailReader{
				found: true,
				detail: domain.ExperimentEvaluationDetail{
					Evaluation: domain.ExperimentEvaluationFact{ID: "evaluation-1"},
				},
			},
			wantID:    "evaluation-1",
			wantInput: "evaluation-1",
		},
		{
			name:       "空IDを拒否する",
			evaluation: " \t",
			wantCode:   apperr.CodeEvaluationDetailRequestInvalid,
		},
		{
			name:       "未検出を返す",
			evaluation: "evaluation-1",
			reader:     &getEvaluationDetailReader{},
			wantCode:   apperr.CodeEvaluationDetailNotFound,
			wantInput:  "evaluation-1",
		},
		{
			name:       "内部エラーを安全に変換する",
			evaluation: "evaluation-1",
			reader: &getEvaluationDetailReader{
				err: errors.New("private database detail"),
			},
			wantCode:  apperr.CodeEvaluationDetailUnavailable,
			wantInput: "evaluation-1",
		},
		{
			name:       "期限超過を返す",
			evaluation: "evaluation-1",
			reader: &getEvaluationDetailReader{
				err: context.DeadlineExceeded,
			},
			wantCode:  apperr.CodeOperationTimeout,
			wantInput: "evaluation-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := tt.reader
			if reader == nil {
				reader = &getEvaluationDetailReader{}
			}
			detail, err := NewGetEvaluationDetail(reader).Execute(context.Background(), tt.evaluation)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}
			} else {
				if err != nil {
					t.Fatalf("Execute() error = %v", err)
				}
				if detail.Evaluation.ID != tt.wantID {
					t.Errorf("Evaluation.ID = %q, want %q", detail.Evaluation.ID, tt.wantID)
				}
			}
			if reader.input != tt.wantInput {
				t.Errorf("reader input = %q, want %q", reader.input, tt.wantInput)
			}
		})
	}
}

// getEvaluationDetailReader は評価詳細readerのtest double。
type getEvaluationDetailReader struct {
	detail domain.ExperimentEvaluationDetail
	found  bool
	err    error
	input  string
}

// GetEvaluationDetail は設定済み詳細を返却する。
func (r *getEvaluationDetailReader) GetEvaluationDetail(_ context.Context, evaluationID string) (domain.ExperimentEvaluationDetail, bool, error) {
	r.input = evaluationID
	return r.detail, r.found, r.err
}
