package usecase

import (
	"context"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ExperimentBriefingStartStore は実験ブリーフ開始を記録するport。
type ExperimentBriefingStartStore interface {
	BeginExperimentBriefing(context.Context, string) (domain.ExperimentBriefingStart, bool, error)
	MarkExperimentBriefingStarted(context.Context, string) error
	MarkExperimentBriefingFailed(context.Context, string, string) error
}

// BriefingStarter は実験ブリーフ開始を外部へ委譲するport。
type BriefingStarter interface {
	StartExperimentBriefing(context.Context, string, string) error
}

// ExperimentBriefingMessageStore はメッセージ送信の記録と結果保存のport。
type ExperimentBriefingMessageStore interface {
	BeginExperimentBriefMessage(context.Context, string, string) (domain.ExperimentBriefingMessageOperation, bool, error)
	CompleteExperimentBriefMessage(context.Context, string, string, domain.ExperimentBriefingMessageResult) error
	FailExperimentBriefMessage(context.Context, string, string) error
}

// BriefingMessageSender は実験ブリーフ会話をACPへ委譲するport。
type BriefingMessageSender interface {
	SendExperimentBriefMessage(context.Context, string, string, string) (domain.ExperimentBriefingMessageResult, error)
}

// SendExperimentBriefMessage は実験ブリーフ会話送信command。
type SendExperimentBriefMessage struct {
	store  ExperimentBriefingMessageStore
	sender BriefingMessageSender
}

// NewSendExperimentBriefMessage は実験ブリーフ会話送信commandを生成。
func NewSendExperimentBriefMessage(store ExperimentBriefingMessageStore, sender BriefingMessageSender) *SendExperimentBriefMessage {
	return &SendExperimentBriefMessage{store: store, sender: sender}
}

// Execute は実験ブリーフメッセージを記録してACPへ送信。
func (u *SendExperimentBriefMessage) Execute(ctx context.Context, requestID, briefingSessionID, message string) (domain.ExperimentBriefingMessageOperation, error) {
	if strings.TrimSpace(requestID) == "" || strings.TrimSpace(briefingSessionID) == "" || strings.TrimSpace(message) == "" {
		return domain.ExperimentBriefingMessageOperation{}, apperr.New(apperr.CodeBriefingRequestInvalid)
	}

	operation, created, err := u.store.BeginExperimentBriefMessage(ctx, requestID, briefingSessionID)
	if err != nil {
		if appErr := apperr.As(err); appErr != nil {
			return domain.ExperimentBriefingMessageOperation{}, appErr
		}

		return domain.ExperimentBriefingMessageOperation{}, apperr.Wrap(apperr.CodeBriefingMessageFailed, err)
	}
	if !created {
		if operation.State == domain.BriefingStartStateFailed {
			return domain.ExperimentBriefingMessageOperation{}, briefingMessageFailure(operation.FailureCode)
		}
		if operation.State == domain.BriefingStartStateStarting {
			return domain.ExperimentBriefingMessageOperation{}, apperr.New(apperr.CodeBriefingMessagePending)
		}

		return operation, nil
	}

	result, err := u.sender.SendExperimentBriefMessage(ctx, operation.BriefingSessionID, operation.OperationID, strings.TrimSpace(message))
	if err != nil {
		failure := apperr.As(err)
		if failure == nil {
			failure = apperr.New(apperr.CodeBriefingMessageFailed)
		}
		if markErr := u.store.FailExperimentBriefMessage(ctx, requestID, string(failure.Code)); markErr != nil {
			return domain.ExperimentBriefingMessageOperation{}, apperr.Wrap(apperr.CodeBriefingMessageFailed, markErr)
		}

		return domain.ExperimentBriefingMessageOperation{}, failure
	}
	if err := u.store.CompleteExperimentBriefMessage(ctx, requestID, strings.TrimSpace(message), result); err != nil {
		return domain.ExperimentBriefingMessageOperation{}, apperr.Wrap(apperr.CodeBriefingMessageFailed, err)
	}
	operation.State = domain.BriefingStartStateStarted

	return operation, nil
}

// briefingMessageFailure は永続済みの安全な送信失敗を復元。
func briefingMessageFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeACPNotReady:
		return apperr.New(apperr.CodeACPNotReady)
	case apperr.CodeBriefingNotActive:
		return apperr.New(apperr.CodeBriefingNotActive)
	case apperr.CodeBriefingRequestInvalid:
		return apperr.New(apperr.CodeBriefingRequestInvalid)
	default:
		return apperr.New(apperr.CodeBriefingMessageFailed)
	}
}

// ExperimentBriefingReader は実験ブリーフ画面を読み出すport。
type ExperimentBriefingReader interface {
	GetExperimentBriefing(context.Context, string) (domain.ExperimentBriefing, bool, error)
}

// GetExperimentBriefing は実験ブリーフ画面の再読込query。
type GetExperimentBriefing struct {
	reader ExperimentBriefingReader
}

// NewGetExperimentBriefing は実験ブリーフ再読込queryを生成。
func NewGetExperimentBriefing(reader ExperimentBriefingReader) *GetExperimentBriefing {
	return &GetExperimentBriefing{reader: reader}
}

// Execute は保存済み実験ブリーフを取得。
func (u *GetExperimentBriefing) Execute(ctx context.Context, briefingSessionID string) (domain.ExperimentBriefing, error) {
	if strings.TrimSpace(briefingSessionID) == "" {
		return domain.ExperimentBriefing{}, apperr.New(apperr.CodeBriefingRequestInvalid)
	}

	briefing, found, err := u.reader.GetExperimentBriefing(ctx, briefingSessionID)
	if err != nil {
		if appErr := apperr.FromContextError(err); appErr != nil {
			return domain.ExperimentBriefing{}, appErr
		}

		return domain.ExperimentBriefing{}, apperr.Wrap(apperr.CodeBriefingLoadFailed, err)
	}
	if !found {
		return domain.ExperimentBriefing{}, apperr.New(apperr.CodeBriefingNotFound)
	}

	return briefing, nil
}

// StartExperimentBriefing は実験ブリーフ開始command。
type StartExperimentBriefing struct {
	store   ExperimentBriefingStartStore
	starter BriefingStarter
}

// NewStartExperimentBriefing は実験ブリーフ開始commandを生成。
func NewStartExperimentBriefing(store ExperimentBriefingStartStore, starter BriefingStarter) *StartExperimentBriefing {
	return &StartExperimentBriefing{store: store, starter: starter}
}

// Execute は実験ブリーフ開始を記録して外部開始を委譲。
func (u *StartExperimentBriefing) Execute(ctx context.Context, requestID string) (domain.ExperimentBriefingStart, error) {
	if strings.TrimSpace(requestID) == "" {
		return domain.ExperimentBriefingStart{}, apperr.New(apperr.CodeBriefingRequestInvalid)
	}

	start, created, err := u.store.BeginExperimentBriefing(ctx, requestID)
	if err != nil {
		return domain.ExperimentBriefingStart{}, apperr.Wrap(apperr.CodeBriefingStartFailed, err)
	}
	if !created {
		if start.State == domain.BriefingStartStateFailed {
			return domain.ExperimentBriefingStart{}, briefingFailure(start.FailureCode)
		}
		if start.State == domain.BriefingStartStateStarting {
			return domain.ExperimentBriefingStart{}, apperr.New(apperr.CodeBriefingStartPending)
		}

		return start, nil
	}

	if err := u.starter.StartExperimentBriefing(ctx, start.BriefingSessionID, start.OperationID); err != nil {
		failure := apperr.As(err)
		if failure == nil {
			failure = apperr.New(apperr.CodeBriefingStartFailed)
		}
		if markErr := u.store.MarkExperimentBriefingFailed(ctx, requestID, string(failure.Code)); markErr != nil {
			return domain.ExperimentBriefingStart{}, apperr.Wrap(apperr.CodeBriefingStartFailed, markErr)
		}

		return domain.ExperimentBriefingStart{}, failure
	}

	if err := u.store.MarkExperimentBriefingStarted(ctx, requestID); err != nil {
		return domain.ExperimentBriefingStart{}, apperr.Wrap(apperr.CodeBriefingStartFailed, err)
	}
	start.State = domain.BriefingStartStateStarted

	return start, nil
}

// briefingFailure は永続済みの安全な開始失敗を復元。
func briefingFailure(code string) error {
	switch apperr.Code(code) {
	case apperr.CodeACPNotReady:
		return apperr.New(apperr.CodeACPNotReady)
	case apperr.CodeBriefingRequestInvalid:
		return apperr.New(apperr.CodeBriefingRequestInvalid)
	case apperr.CodeBriefingStartPending:
		return apperr.New(apperr.CodeBriefingStartPending)
	default:
		return apperr.New(apperr.CodeBriefingStartFailed)
	}
}

// ExperimentReader は実験一覧を読み出すport。
type ExperimentReader interface {
	ListExperiments(context.Context) (domain.ExperimentCollection, error)
}

// ListExperiments は実験一覧query。
type ListExperiments struct {
	reader ExperimentReader
}

// NewListExperiments は実験一覧queryを生成。
func NewListExperiments(reader ExperimentReader) *ListExperiments {
	return &ListExperiments{reader: reader}
}

// Execute は実験一覧を取得。
func (u *ListExperiments) Execute(ctx context.Context) (domain.ExperimentCollection, error) {
	collection, err := u.reader.ListExperiments(ctx)
	if err == nil {
		return collection, nil
	}
	if appErr := apperr.FromContextError(err); appErr != nil {
		return domain.ExperimentCollection{}, appErr
	}

	return domain.ExperimentCollection{}, apperr.Wrap(apperr.CodeExperimentsLoadFailed, err)
}
