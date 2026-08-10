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
			listPreparations := NewListPreparations(&tt.reader)

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

// GetPreparationの入力検証とport委譲。
func TestGetPreparationExecute(t *testing.T) {
	repositoryFailure := errors.New("repository failure")
	tests := []struct {
		name      string
		input     string
		reader    fakePreparationReader
		wantCode  apperr.Code
		wantFound bool
	}{
		{
			name:     "空白IDを拒否する",
			input:    "  ",
			wantCode: apperr.CodePreparationRequestInvalid,
		},
		{
			name:     "存在しないIDを返す",
			input:    "preparation-1",
			reader:   fakePreparationReader{found: false},
			wantCode: apperr.CodePreparationNotFound,
		},
		{
			name:     "repository失敗を安全なコードへ変換する",
			input:    "preparation-1",
			reader:   fakePreparationReader{err: repositoryFailure},
			wantCode: apperr.CodePreparationUnavailable,
		},
		{
			name:  "空白を除去して詳細を返す",
			input: " preparation-1 ",
			reader: fakePreparationReader{
				preparation: domain.PreparationDetail{ID: "preparation-1"},
				found:       true,
			},
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPreparation := NewGetPreparation(&tt.reader)

			got, err := getPreparation.Execute(context.Background(), tt.input)
			if tt.wantCode != "" {
				if gotCode := apperr.As(err).Code; gotCode != tt.wantCode {
					t.Errorf("Code = %q, want %q", gotCode, tt.wantCode)
				}

				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got.ID != "preparation-1" {
				t.Errorf("ID = %q, want %q", got.ID, "preparation-1")
			}
			if !tt.wantFound {
				t.Error("wantFound = false, want true")
			}
			if tt.reader.gotID != "preparation-1" {
				t.Errorf("GetPreparation ID = %q, want %q", tt.reader.gotID, "preparation-1")
			}
		})
	}
}

// fakePreparationReader は環境準備一覧query用のtest double。
type fakePreparationReader struct {
	preparations []domain.Preparation
	preparation  domain.PreparationDetail
	found        bool
	gotID        string
	err          error
}

// ListPreparations は指定済みの一覧またはエラーを返す。
func (f fakePreparationReader) ListPreparations(context.Context) ([]domain.Preparation, error) {
	return f.preparations, f.err
}

// GetPreparation は詳細未設定を返す。
func (f *fakePreparationReader) GetPreparation(_ context.Context, preparationID string) (domain.PreparationDetail, bool, error) {
	f.gotID = preparationID

	return f.preparation, f.found, f.err
}
