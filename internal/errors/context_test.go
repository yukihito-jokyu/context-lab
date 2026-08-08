package apperr

import (
	"context"
	stderrors "errors"
	"testing"
)

// コンテキストエラーをアプリケーションエラーへ変換する。
func TestFromContextError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode Code
	}{
		{
			name:     "取消を変換する",
			err:      context.Canceled,
			wantCode: CodeOperationCanceled,
		},
		{
			name:     "期限超過を変換する",
			err:      context.DeadlineExceeded,
			wantCode: CodeOperationTimeout,
		},
		{
			name: "通常エラーは変換しない",
			err:  stderrors.New("other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromContextError(tt.err)
			if tt.wantCode == "" {
				if got != nil {
					t.Errorf("FromContextError() = %v, want nil", got)
				}

				return
			}
			if got == nil {
				t.Fatal("FromContextError() = nil, want application error")
			}
			if got.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", got.Code, tt.wantCode)
			}
			if !stderrors.Is(got, tt.err) {
				t.Errorf("errors.Is(error, source) = false, want true")
			}
		})
	}
}
