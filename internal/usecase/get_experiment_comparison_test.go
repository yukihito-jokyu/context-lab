package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// GetExperimentComparisonの入力、未検出、エラー契約を検証する。
func TestGetExperimentComparisonExecute(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		reader    comparisonReader
		wantCode  apperr.Code
		wantID    string
		wantInput string
	}{
		{
			name:      "成功",
			id:        " experiment-1 ",
			wantID:    "experiment-1",
			wantInput: "experiment-1",
			reader: comparisonReader{
				found: true,
				comparison: domain.ExperimentComparison{
					Experiment: domain.ExperimentComparisonExperiment{ID: "experiment-1"},
				},
			},
		},
		{
			name:     "空ID",
			wantCode: apperr.CodeExperimentComparisonInvalid,
		},
		{
			name:      "未検出",
			id:        "experiment-1",
			wantCode:  apperr.CodeExperimentComparisonNotFound,
			wantInput: "experiment-1",
		},
		{
			name: "内部エラー",
			id:   "experiment-1",
			reader: comparisonReader{
				err: errors.New("private sqlite"),
			},
			wantCode:  apperr.CodeExperimentComparisonUnavailable,
			wantInput: "experiment-1",
		},
		{
			name: "期限超過",
			id:   "experiment-1",
			reader: comparisonReader{
				err: context.DeadlineExceeded,
			},
			wantCode:  apperr.CodeOperationTimeout,
			wantInput: "experiment-1",
		},
		{
			name: "中止",
			id:   "experiment-1",
			reader: comparisonReader{
				err: context.Canceled,
			},
			wantCode:  apperr.CodeOperationCanceled,
			wantInput: "experiment-1",
		},
		{
			name: "readerアプリケーションエラーを取得不能へ変換する",
			id:   "experiment-1",
			reader: comparisonReader{
				err: apperr.New(apperr.CodeExperimentComparisonNotFound),
			},
			wantCode:  apperr.CodeExperimentComparisonUnavailable,
			wantInput: "experiment-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparison, err := NewGetExperimentComparison(&tt.reader).Execute(context.Background(), tt.id)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want %q", err, tt.wantCode)
				}
			} else if err != nil {
				t.Fatalf("Execute() error = %v", err)
			} else if comparison.Experiment.ID != tt.wantID {
				t.Errorf("Experiment.ID = %q, want %q", comparison.Experiment.ID, tt.wantID)
			}
			if tt.reader.input != tt.wantInput {
				t.Errorf("reader input = %q, want %q", tt.reader.input, tt.wantInput)
			}
		})
	}
}

type comparisonReader struct {
	comparison domain.ExperimentComparison
	found      bool
	err        error
	input      string
}

func (r *comparisonReader) GetExperimentComparison(_ context.Context, experimentID string) (domain.ExperimentComparison, bool, error) {
	r.input = experimentID

	return r.comparison, r.found, r.err
}
