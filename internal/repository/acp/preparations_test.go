package acp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 環境準備adapterの生成、範囲検証、ACP起動失敗を確認。
func TestCodexPreparationAdapter(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "target"), 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	adapter := newCodexPreparationAdapter(root, "context-lab-command-not-found", nil)
	resolved, canonical, err := adapter.ValidatePreparationScope("target")
	if err != nil {
		t.Fatalf("ValidatePreparationScope() error = %v", err)
	}
	if canonical != "target" {
		t.Errorf("canonical = %q, want target", canonical)
	}
	if resolved == "" {
		t.Error("resolved = empty, want directory")
	}
	if _, err := adapter.PrepareEnvironment(context.Background(), resolved); !apperr.IsCode(err, apperr.CodeACPNotReady) {
		t.Errorf("PrepareEnvironment() error = %v, want ACP not ready", err)
	}
}

// ACP環境準備の成功と安全な失敗変換を確認。
func TestCodexPreparationAdapterPrepareEnvironment(t *testing.T) {
	root := t.TempDir()
	adapter := newCodexPreparationAdapter(root, "unused", nil)
	tests := []struct {
		name      string
		starter   preparationSessionStarter
		wantCode  apperr.Code
		wantCount int
	}{
		{
			name: "安全な候補を返す",
			starter: func(context.Context, string, string, []string) (preparationSession, error) {
				return preparationSessionStub{response: `{"candidates":[{"environmentConditions":"macOS","summary":"利用可能"}]}`}, nil
			},
			wantCount: 1,
		},
		{
			name: "prompt失敗を安全に変換する",
			starter: func(context.Context, string, string, []string) (preparationSession, error) {
				return preparationSessionStub{err: errors.New("private sidecar")}, nil
			},
			wantCode: apperr.CodePreparationStartUnavailable,
		},
		{
			name: "不正応答を安全に変換する",
			starter: func(context.Context, string, string, []string) (preparationSession, error) {
				return preparationSessionStub{response: "{"}, nil
			},
			wantCode: apperr.CodePreparationStartUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter.start = tt.starter
			got, err := adapter.PrepareEnvironment(context.Background(), root)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("PrepareEnvironment() error = %v, want code %q", err, tt.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("PrepareEnvironment() error = %v", err)
			}
			if gotCount := len(got.Candidates); gotCount != tt.wantCount {
				t.Errorf("Candidates length = %d, want %d", gotCount, tt.wantCount)
			}
		})
	}
	if _, err := adapter.PrepareEnvironment(context.Background(), filepath.Join(root, "missing")); !apperr.IsCode(err, apperr.CodePreparationScopeInvalid) {
		t.Errorf("invalid scope error = %v, want scope invalid", err)
	}
}

// 実ACP環境準備adapterの既定値を確認。
func TestNewCodexPreparationAdapter(t *testing.T) {
	adapter := NewCodexPreparationAdapter("/workspace")
	if adapter.workingRoot != "/workspace" {
		t.Errorf("workingRoot = %q, want /workspace", adapter.workingRoot)
	}
	if adapter.command != codexACPCommand {
		t.Errorf("command = %q, want %q", adapter.command, codexACPCommand)
	}
}

// 環境準備範囲の不正入力を確認。
func TestResolvePreparationScopeInvalid(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("file"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	tests := []struct {
		name  string
		scope string
	}{
		{
			name:  "空",
			scope: " ",
		},
		{
			name:  "絶対パス",
			scope: root,
		},
		{
			name:  "親directory",
			scope: "../outside",
		},
		{
			name:  "存在しない",
			scope: "missing",
		},
		{
			name:  "通常file",
			scope: "file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := resolvePreparationScope(root, tt.scope); err == nil {
				t.Error("resolvePreparationScope() error = nil, want invalid scope")
			}
		})
	}
	if _, canonical, err := resolvePreparationScope(root, "."); err != nil || canonical != "." {
		t.Errorf("resolvePreparationScope(root) = (%q, %v), want canonical .", canonical, err)
	}
	if _, _, err := resolvePreparationScope(filepath.Join(root, "missing-root"), "target"); err == nil {
		t.Error("resolvePreparationScope(missing root) error = nil, want root resolution error")
	}
	previous := preparationRelativePath
	preparationRelativePath = func(string, string) (string, error) { return "", errors.New("relative failed") }
	t.Cleanup(func() { preparationRelativePath = previous })
	if _, _, err := resolvePreparationScope(root, "."); err == nil {
		t.Error("resolvePreparationScope(relative failure) error = nil, want relative error")
	}
}

// 解決済み環境準備範囲の再検証を確認。
func TestValidateResolvedPreparationScope(t *testing.T) {
	root := t.TempDir()
	if err := validateResolvedPreparationScope(root, root); err != nil {
		t.Errorf("validateResolvedPreparationScope(root) error = %v", err)
	}
	if err := validateResolvedPreparationScope(root, filepath.Join(root, "missing")); err == nil {
		t.Error("validateResolvedPreparationScope() error = nil, want invalid scope")
	}
	if err := validateResolvedPreparationScope(filepath.Join(root, "missing-root"), root); err == nil {
		t.Error("validateResolvedPreparationScope(missing root) error = nil, want root resolution error")
	}
}

// 環境準備ACP応答の安全な抽出を確認。
func TestParseEnvironmentPreparationResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "候補と診断",
			response: `{"candidates":[{"environmentConditions":" macOS ","summary":" 利用可能 "}],"diagnostics":[{"code":" CHECKED ","summary":" 確認済み "}]}`,
		},
		{
			name:     "JSON不正",
			response: "{",
			wantErr:  true,
		},
		{
			name:     "候補不正",
			response: `{"candidates":[{"environmentConditions":"","summary":"x"}]}`,
			wantErr:  true,
		},
		{
			name:     "診断不正",
			response: `{"diagnostics":[{"code":"","summary":"x"}]}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnvironmentPreparationResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseEnvironmentPreparationResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Candidates[0].EnvironmentConditions != "macOS" {
				t.Errorf("candidate conditions = %q, want macOS", got.Candidates[0].EnvironmentConditions)
			}
			if got.Diagnostics[0].SafeSummary != "確認済み" {
				t.Errorf("diagnostic summary = %q, want 確認済み", got.Diagnostics[0].SafeSummary)
			}
		})
	}
}

// 環境準備ACP応答の機微情報拒否を確認。
func TestParseEnvironmentPreparationResponseRejectsUnsafeText(t *testing.T) {
	unsafe := []string{
		"/Users/name/private",
		"credential",
		"password",
		"secret",
		"token",
		"api key",
		"PID 123",
		"docker ID abc",
		"container ID abc",
		"sidecar PID 123",
		"内部推論",
		"reasoning",
		"chain of thought",
	}
	for _, value := range unsafe {
		t.Run(value, func(t *testing.T) {
			response := `{"candidates":[{"environmentConditions":"macOS","summary":"` + value + `"}]}`
			if _, err := parseEnvironmentPreparationResponse(response); err == nil {
				t.Error("parseEnvironmentPreparationResponse() error = nil, want unsafe text rejection")
			}
		})
	}
	if isSafeEnvironmentPreparationText("") {
		t.Error("isSafeEnvironmentPreparationText(empty) = true, want false")
	}
	if isSafeEnvironmentPreparationText(string(make([]byte, 2_001))) {
		t.Error("isSafeEnvironmentPreparationText(long) = true, want false")
	}
	if !isSafeEnvironmentPreparationText("安全な環境要約") {
		t.Error("isSafeEnvironmentPreparationText(safe) = false, want true")
	}
}

// 環境準備promptの安全要求を確認。
func TestEnvironmentPreparationPrompt(t *testing.T) {
	if prompt := environmentPreparationPrompt(); prompt == "" {
		t.Error("environmentPreparationPrompt() = empty, want safety prompt")
	}
}

// 環境準備範囲外判定を確認。
func TestValidatePreparationDirectory(t *testing.T) {
	root := t.TempDir()
	if err := validatePreparationDirectory(root, filepath.Dir(root)); err == nil {
		t.Error("validatePreparationDirectory() error = nil, want outside error")
	}
	if err := validatePreparationDirectory(root, filepath.Join(root, "missing")); err == nil {
		t.Error("validatePreparationDirectory(missing) error = nil, want stat error")
	}
}

// preparationSessionStub はACP session用のtest double。
type preparationSessionStub struct {
	response string
	err      error
}

// prompt は指定済み応答または失敗を返す。
func (s preparationSessionStub) prompt(context.Context, string) (string, error) {
	return s.response, s.err
}

// close はsession終了を受理する。
func (preparationSessionStub) close(context.Context) error { return nil }
