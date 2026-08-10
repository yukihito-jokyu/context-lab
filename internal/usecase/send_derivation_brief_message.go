package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// DerivationBriefingMessageStore は派生実験ブリーフ会話の送信結果を記録するport。
type DerivationBriefingMessageStore interface {
	BeginDerivationBriefMessage(context.Context, string, string) (domain.DerivationBriefingMessageOperation, bool, error)
	CompleteDerivationBriefMessage(context.Context, string, string, domain.DerivationBriefingMessageResult) error
	FailDerivationBriefMessage(context.Context, string, string) error
}

// DerivationBriefingMessageSender は派生実験ブリーフ会話をACPへ委譲するport。
type DerivationBriefingMessageSender interface {
	SendDerivationBriefMessage(context.Context, string, string, string) (domain.DerivationBriefingMessageResult, error)
}

// SendDerivationBriefMessage は派生実験ブリーフ会話送信command。
type SendDerivationBriefMessage struct {
	store  DerivationBriefingMessageStore
	sender DerivationBriefingMessageSender
}

// NewSendDerivationBriefMessage は派生実験ブリーフ会話送信commandを生成する。
func NewSendDerivationBriefMessage(store DerivationBriefingMessageStore, sender DerivationBriefingMessageSender) *SendDerivationBriefMessage {
	return &SendDerivationBriefMessage{store: store, sender: sender}
}

// Execute は派生実験ブリーフメッセージを記録してACPへ送信する。
func (u *SendDerivationBriefMessage) Execute(ctx context.Context, requestID, briefingSessionID, message string) (domain.DerivationBriefingMessageOperation, error) {
	requestID = strings.TrimSpace(requestID)
	briefingSessionID = strings.TrimSpace(briefingSessionID)
	message = strings.TrimSpace(message)
	if requestID == "" || briefingSessionID == "" || message == "" {
		return domain.DerivationBriefingMessageOperation{}, apperr.New(apperr.CodeDerivationBriefingMessageInvalid)
	}

	operation, created, err := u.store.BeginDerivationBriefMessage(ctx, requestID, briefingSessionID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.DerivationBriefingMessageOperation{}, appErr
		}

		return domain.DerivationBriefingMessageOperation{}, apperr.Wrap(apperr.CodeDerivationBriefingMessageFailed, err)
	}
	if !created {
		if operation.State == domain.BriefingStartStateFailed {
			return domain.DerivationBriefingMessageOperation{}, derivationBriefingMessageFailure(operation.FailureCode)
		}
		if operation.State == domain.BriefingStartStateStarting {
			return domain.DerivationBriefingMessageOperation{}, apperr.New(apperr.CodeDerivationBriefingMessagePending)
		}

		return operation, nil
	}

	result, err := u.sender.SendDerivationBriefMessage(ctx, operation.BriefingSessionID, operation.OperationID, message)
	if err != nil {
		failure := apperr.As(err)
		if failure == nil {
			failure = apperr.New(apperr.CodeDerivationBriefingMessageFailed)
		}
		if markErr := u.store.FailDerivationBriefMessage(ctx, requestID, string(failure.Code)); markErr != nil {
			return domain.DerivationBriefingMessageOperation{}, apperr.Wrap(apperr.CodeDerivationBriefingMessageFailed, markErr)
		}

		return domain.DerivationBriefingMessageOperation{}, failure
	}
	if err := u.store.CompleteDerivationBriefMessage(ctx, requestID, message, result); err != nil {
		return domain.DerivationBriefingMessageOperation{}, apperr.Wrap(apperr.CodeDerivationBriefingMessageFailed, err)
	}
	operation.State = domain.BriefingStartStateStarted

	return operation, nil
}

// derivationBriefingMessageFailure は永続済みの安全な送信失敗を復元する。
func derivationBriefingMessageFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeACPNotReady:
		return apperr.New(apperr.CodeACPNotReady)
	case apperr.CodeDerivationBriefingMessageInvalid:
		return apperr.New(apperr.CodeDerivationBriefingMessageInvalid)
	case apperr.CodeDerivationBriefingMessageNotFound:
		return apperr.New(apperr.CodeDerivationBriefingMessageNotFound)
	case apperr.CodeDerivationBriefingMessageNotActive:
		return apperr.New(apperr.CodeDerivationBriefingMessageNotActive)
	default:
		return apperr.New(apperr.CodeDerivationBriefingMessageFailed)
	}
}
