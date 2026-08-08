package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ListExperimentsのport委譲。
func TestListExperimentsExecute(t *testing.T) {
	repositoryFailure := errors.New("repository failure")
	tests := []struct {
		name      string
		reader    fakeExperimentReader
		wantCode  apperr.Code
		wantCause error
	}{
		{
			name: "一覧を返す",
			reader: fakeExperimentReader{collection: domain.ExperimentCollection{
				Experiments: []domain.Experiment{},
			}},
		},
		{
			name:      "portの失敗をアプリケーションエラーへ正規化する",
			reader:    fakeExperimentReader{err: repositoryFailure},
			wantCode:  apperr.CodeExperimentsLoadFailed,
			wantCause: repositoryFailure,
		},
		{
			name:      "取消は呼び出し元へ返す",
			reader:    fakeExperimentReader{err: context.Canceled},
			wantCode:  apperr.CodeOperationCanceled,
			wantCause: context.Canceled,
		},
		{
			name:      "タイムアウトは呼び出し元へ返す",
			reader:    fakeExperimentReader{err: context.DeadlineExceeded},
			wantCode:  apperr.CodeOperationTimeout,
			wantCause: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listExperiments := NewListExperiments(tt.reader)

			_, err := listExperiments.Execute(context.Background())
			if tt.wantCode == "" {
				if err != nil {
					t.Errorf("Execute() error = %v, want nil", err)
				}

				return
			}
			appErr := apperr.As(err)
			if appErr == nil {
				t.Fatal("As(error) = nil, want application error")
			}
			if gotCode := appErr.Code; gotCode != tt.wantCode {
				t.Errorf("Code = %q, want %q", gotCode, tt.wantCode)
			}
			if gotCause := errors.Is(err, tt.wantCause); !gotCause {
				t.Errorf("Execute() errors.Is(error, wantCause) = %v, want true", gotCause)
			}
		})
	}
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
