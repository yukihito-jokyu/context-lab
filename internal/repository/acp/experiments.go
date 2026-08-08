package acp

import (
	"context"

	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// NotReadyBriefingStarter はACP検証完了前の安全な開始adapter。
type NotReadyBriefingStarter struct{}

// StartExperimentBriefing はACP未準備エラーを返却。
func (NotReadyBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	return apperr.New(apperr.CodeACPNotReady)
}
