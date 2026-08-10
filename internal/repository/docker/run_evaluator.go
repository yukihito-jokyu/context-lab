package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// CodexRunEvaluator は隔離Dockerでrun評価を起動するadapter。
type CodexRunEvaluator struct {
	image string
}

// NewCodexRunEvaluator は既定イメージの評価runnerを生成。
func NewCodexRunEvaluator() *CodexRunEvaluator {
	return &CodexRunEvaluator{image: defaultExperimentImage}
}

// EvaluateRun は安全なrun結果と評価軸をDockerへ渡す。
func (r *CodexRunEvaluator) EvaluateRun(ctx context.Context, request domain.ExperimentEvaluationRequest) (string, error) {
	if strings.TrimSpace(request.RunID) == "" || strings.TrimSpace(request.RunSummary) == "" || strings.TrimSpace(request.EvaluationAxes) == "" {
		return "", fmt.Errorf("invalid isolated run evaluation")
	}
	prompt := strings.Join([]string{
		"実験目的: " + strings.TrimSpace(request.Purpose),
		"評価軸: " + strings.TrimSpace(request.EvaluationAxes),
		"run結果: " + strings.TrimSpace(request.RunSummary),
		"上記のrun結果を評価し、利用者向けの簡潔な要約を返してください。",
	}, "\n")
	command := commandContext(ctx, "docker", "run", "--rm", "--network", "none", "--read-only", "--tmpfs", "/tmp", r.image, "codex", "exec", "--skip-git-repo-check", prompt)
	if _, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("run isolated evaluation: %w", err)
	}

	return "評価を開始しました", nil
}
