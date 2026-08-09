package acp

import (
	"context"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// NotReadyBriefingStarter はACP検証完了前の安全な開始adapter。
type NotReadyBriefingStarter struct{}

// StartExperimentBriefing はACP未準備エラーを返却。
func (NotReadyBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	return apperr.New(apperr.CodeACPNotReady)
}

// NotReadyBriefingMessageSender はACP検証完了前の安全な送信adapter。
type NotReadyBriefingMessageSender struct{}

// SendExperimentBriefMessage はACP未準備エラーを返却。
func (NotReadyBriefingMessageSender) SendExperimentBriefMessage(context.Context, string, string, string) (domain.ExperimentBriefingMessageResult, error) {
	return domain.ExperimentBriefingMessageResult{}, apperr.New(apperr.CodeACPNotReady)
}
