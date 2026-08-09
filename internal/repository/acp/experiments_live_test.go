//go:build integration

package acp

import (
	"context"
	"os"
	"testing"
	"time"
)

// 実Codex ACP sidecarへ接続する確認。明示指定時だけ実行する。
func TestCodexBriefingAdapterLive(t *testing.T) {
	if os.Getenv("CONTEXT_LAB_ACP_LIVE_TEST") != "1" {
		t.Skip("set CONTEXT_LAB_ACP_LIVE_TEST=1 to use the real Codex ACP sidecar")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	adapter := NewCodexBriefingAdapter(t.TempDir())
	if err := adapter.StartExperimentBriefing(ctx, "live-session", "start-operation"); err != nil {
		t.Fatalf("StartExperimentBriefing() error = %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.StopExperimentBriefing(context.Background(), "live-session", "stop-operation"); err != nil {
			t.Errorf("StopExperimentBriefing() error = %v", err)
		}
	})

	result, err := adapter.SendExperimentBriefMessage(ctx, "live-session", "message-operation", "要約の長さを比較する実験を設計したい")
	if err != nil {
		t.Fatalf("SendExperimentBriefMessage() error = %v", err)
	}
	if result.AssistantMessage == "" {
		t.Fatal("AssistantMessage is empty")
	}
	if result.Brief == nil {
		t.Fatal("Brief is nil")
	}
}
