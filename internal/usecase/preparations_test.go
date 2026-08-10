package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ListPreparationsのport委譲。
func TestListPreparationsExecute(t *testing.T) {
	repositoryFailure := errors.New("repository failure")
	tests := []struct {
		name      string
		reader    fakePreparationReader
		wantCode  apperr.Code
		wantCause error
	}{
		{
			name:   "環境準備一覧を返す",
			reader: fakePreparationReader{preparations: []domain.Preparation{}},
		},
		{
			name:      "portの失敗をアプリケーションエラーへ正規化する",
			reader:    fakePreparationReader{err: repositoryFailure},
			wantCode:  apperr.CodePreparationsLoadFailed,
			wantCause: repositoryFailure,
		},
		{
			name:      "取消も一覧取得失敗へ正規化する",
			reader:    fakePreparationReader{err: context.Canceled},
			wantCode:  apperr.CodePreparationsLoadFailed,
			wantCause: context.Canceled,
		},
		{
			name:      "タイムアウトも一覧取得失敗へ正規化する",
			reader:    fakePreparationReader{err: context.DeadlineExceeded},
			wantCode:  apperr.CodePreparationsLoadFailed,
			wantCause: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listPreparations := NewListPreparations(tt.reader)

			_, err := listPreparations.Execute(context.Background())
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
			if got := appErr.Code; got != tt.wantCode {
				t.Errorf("Code = %q, want %q", got, tt.wantCode)
			}
			if got := errors.Is(err, tt.wantCause); !got {
				t.Errorf("Execute() errors.Is(error, wantCause) = %v, want true", got)
			}
		})
	}
}

// fakePreparationReader は環境準備一覧query用のtest double。
type fakePreparationReader struct {
	preparations []domain.Preparation
	err          error
}

// ListPreparations は指定済みの一覧またはエラーを返す。
func (f fakePreparationReader) ListPreparations(context.Context) ([]domain.Preparation, error) {
	return f.preparations, f.err
}
