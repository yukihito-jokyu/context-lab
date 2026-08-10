package docker

import (
	"context"
	"os/exec"
	"reflect"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// Docker起動はホスト共有をせず、ネットワークを遮断したargvだけを構築する。
func TestCodexExperimentRunnerRunExperimentBuildsIsolatedDockerArguments(t *testing.T) {
	originalCommandContext := commandContext
	t.Cleanup(func() { commandContext = originalCommandContext })
	var gotName string
	var gotArguments []string
	commandContext = func(ctx context.Context, name string, arguments ...string) *exec.Cmd {
		gotName = name
		gotArguments = append([]string(nil), arguments...)

		return exec.CommandContext(ctx, "sh", "-c", "printf ok")
	}

	summary, err := NewCodexExperimentRunner().RunExperiment(context.Background(), domain.ExperimentRunRequest{
		ExperimentID:          "experiment-1",
		RunID:                 "run-1",
		Purpose:               "目的",
		EnvironmentConditions: "環境",
		InitialInput:          "入力",
		Prompt:                "prompt",
		EvaluationAxes:        "評価",
	})
	if err != nil {
		t.Fatalf("RunExperiment() error = %v", err)
	}
	if summary != "実行を開始しました" {
		t.Errorf("summary = %q, want safe summary", summary)
	}
	if gotName != "docker" {
		t.Errorf("command = %q, want docker", gotName)
	}
	wantPrefix := []string{
		"run",
		"--rm",
		"--network",
		"none",
		"--read-only",
		"--tmpfs",
		"/tmp",
		defaultExperimentImage,
		"codex",
		"exec",
		"--skip-git-repo-check",
	}
	if len(gotArguments) != len(wantPrefix)+1 || !reflect.DeepEqual(gotArguments[:len(wantPrefix)], wantPrefix) {
		t.Errorf("docker args = %q, want isolated prefix %q plus prompt", gotArguments, wantPrefix)
	}
	if len(gotArguments) > len(wantPrefix) && gotArguments[len(wantPrefix)] != "実験目的: 目的\n環境条件: 環境\n初期入力: 入力\n評価軸: 評価\n実行prompt: prompt" {
		t.Errorf("prompt argument = %q, want normalized fixed conditions", gotArguments[len(wantPrefix)])
	}
}

// 空の識別子やpromptではDockerを起動しない。
func TestCodexExperimentRunnerRunExperimentRejectsInvalidRequest(t *testing.T) {
	originalCommandContext := commandContext
	t.Cleanup(func() { commandContext = originalCommandContext })
	called := false
	commandContext = func(ctx context.Context, name string, arguments ...string) *exec.Cmd {
		called = true

		return exec.CommandContext(ctx, name, arguments...)
	}

	_, err := NewCodexExperimentRunner().RunExperiment(context.Background(), domain.ExperimentRunRequest{})
	if err == nil {
		t.Fatal("RunExperiment() error = nil, want invalid request")
	}
	if called {
		t.Error("commandContext called for invalid request")
	}
}

// Docker実行失敗と空出力を画面安全な結果へ変換する。
func TestCodexExperimentRunnerRunExperimentHandlesCommandResults(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:    "空出力でも安全な開始要約を返す",
			command: "exit 0",
		},
		{
			name:      "Docker失敗を返す",
			command:   "exit 1",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCommandContext := commandContext
			t.Cleanup(func() { commandContext = originalCommandContext })
			commandContext = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
				return exec.CommandContext(ctx, "sh", "-c", tt.command)
			}

			summary, err := NewCodexExperimentRunner().RunExperiment(context.Background(), domain.ExperimentRunRequest{
				ExperimentID: "experiment-1",
				RunID:        "run-1",
				Prompt:       "prompt",
			})
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("RunExperiment() error = %v, want error = %v", err, tt.wantError)
			}
			if tt.wantError {
				return
			}
			if summary != "実行を開始しました" {
				t.Errorf("summary = %q, want safe summary", summary)
			}
		})
	}
}
