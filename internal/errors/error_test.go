package apperr

import (
	stderrors "errors"
	"fmt"
	"testing"
)

// アプリケーションエラー生成と原因保持を検証する。
func TestErrorCreation(t *testing.T) {
	cause := stderrors.New("database unavailable")
	tests := []struct {
		name      string
		err       *Error
		wantCode  Code
		wantText  string
		wantCause error
	}{
		{
			name:     "原因なしのエラーを生成する",
			err:      New(CodeExperimentsLoadFailed),
			wantCode: CodeExperimentsLoadFailed,
			wantText: "実験一覧を取得できませんでした",
		},
		{
			name:      "原因付きエラーを生成する",
			err:       Wrap(CodeOperationTimeout, cause),
			wantCode:  CodeOperationTimeout,
			wantText:  "実験一覧の取得がタイムアウトしました",
			wantCause: cause,
		},
		{
			name:      "想定外エラーを生成する",
			err:       NewUnexpected(cause),
			wantCode:  CodeUnexpected,
			wantText:  "予期しないエラーが発生しました",
			wantCause: cause,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Code; got != tt.wantCode {
				t.Errorf("Code = %q, want %q", got, tt.wantCode)
			}
			if got := tt.err.Error(); got != tt.wantText {
				t.Errorf("Error() = %q, want %q", got, tt.wantText)
			}
			if got := stderrors.Is(tt.err, tt.wantCause); got != (tt.wantCause != nil) {
				t.Errorf("errors.Is(error, cause) = %v, want %v", got, tt.wantCause != nil)
			}
		})
	}
}

// nilエラーとラップされたアプリケーションエラーを検証する。
func TestErrorHelpers(t *testing.T) {
	cause := stderrors.New("cause")
	appErr := Wrap(CodeExperimentsLoadFailed, cause)
	tests := []struct {
		name       string
		err        error
		wantAppErr *Error
		wantCode   bool
	}{
		{
			name:       "直接のアプリケーションエラーを抽出する",
			err:        appErr,
			wantAppErr: appErr,
			wantCode:   true,
		},
		{
			name:       "ラップされたアプリケーションエラーを抽出する",
			err:        fmt.Errorf("request failed: %w", appErr),
			wantAppErr: appErr,
			wantCode:   true,
		},
		{
			name: "通常エラーは抽出しない",
			err:  cause,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := As(tt.err); got != tt.wantAppErr {
				t.Errorf("As() = %v, want %v", got, tt.wantAppErr)
			}
			if got := IsCode(tt.err, CodeExperimentsLoadFailed); got != tt.wantCode {
				t.Errorf("IsCode() = %v, want %v", got, tt.wantCode)
			}
		})
	}

	var nilError *Error
	if got := nilError.Error(); got != "" {
		t.Errorf("nil Error() = %q, want empty string", got)
	}
	if got := nilError.Unwrap(); got != nil {
		t.Errorf("nil Unwrap() = %v, want nil", got)
	}
}
