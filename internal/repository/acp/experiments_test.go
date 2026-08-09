package acp

import (
	"context"
	"testing"

	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// ACP未準備adapterの安全な失敗返却。
func TestNotReadyBriefingStarterStartExperimentBriefing(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ACP未準備を返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (NotReadyBriefingStarter{}).StartExperimentBriefing(context.Background(), "session", "operation")
			if got := apperr.As(err); got == nil {
				t.Fatal("apperr.As() = nil, want app error")
			} else if got.Code != apperr.CodeACPNotReady {
				t.Errorf("Code = %q, want %q", got.Code, apperr.CodeACPNotReady)
			}
		})
	}
}

// ACP未準備adapterの会話送信安全な失敗返却。
func TestNotReadyBriefingMessageSenderSendExperimentBriefMessage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ACP未準備を返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (NotReadyBriefingMessageSender{}).SendExperimentBriefMessage(context.Background(), "session", "operation", "message")
			if got := apperr.As(err); got == nil {
				t.Fatal("apperr.As() = nil, want app error")
			} else if got.Code != apperr.CodeACPNotReady {
				t.Errorf("Code = %q, want %q", got.Code, apperr.CodeACPNotReady)
			}
		})
	}
}
