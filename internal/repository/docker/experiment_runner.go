package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

const defaultExperimentImage = "context-lab/codex-experiment:latest"

var commandContext = exec.CommandContext

// CodexExperimentRunner は隔離Dockerでcodex execを起動するadapter。
type CodexExperimentRunner struct {
	image string
}

// NewCodexExperimentRunner は既定イメージのDocker runnerを生成する。
func NewCodexExperimentRunner() *CodexExperimentRunner {
	return &CodexExperimentRunner{image: defaultExperimentImage}
}

// RunExperiment は安全な固定条件をDockerへ渡してrunを開始する。
func (r *CodexExperimentRunner) RunExperiment(ctx context.Context, request domain.ExperimentRunRequest) (string, error) {
	prompt := strings.TrimSpace(request.Prompt)
	if prompt == "" || strings.TrimSpace(request.RunID) == "" || strings.TrimSpace(request.ExperimentID) == "" {
		return "", fmt.Errorf("invalid isolated experiment run")
	}
	commandPrompt := strings.Join([]string{
		"実験目的: " + strings.TrimSpace(request.Purpose),
		"環境条件: " + strings.TrimSpace(request.EnvironmentConditions),
		"初期入力: " + strings.TrimSpace(request.InitialInput),
		"評価軸: " + strings.TrimSpace(request.EvaluationAxes),
		"実行prompt: " + prompt,
	}, "\n")
	command := commandContext(ctx, "docker", "run", "--rm", "--network", "none", "--read-only", "--tmpfs", "/tmp", r.image, "codex", "exec", "--skip-git-repo-check", commandPrompt)
	if output, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("run isolated codex experiment: %w", err)
	} else if len(output) == 0 {
		return "実行を開始しました", nil
	}

	return "実行を開始しました", nil
}
