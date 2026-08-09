package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// JSON-RPC sidecar fixture。
func TestCodexACPFixtureProcess(t *testing.T) {
	if os.Getenv("GO_WANT_ACP_FIXTURE") != "1" {
		return
	}

	mode := os.Getenv("ACP_FIXTURE_MODE")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request rpcEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			os.Exit(2)
		}
		if mode == "initialize-error" && request.Method == "initialize" {
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   &rpcError{Message: "unavailable"},
			})
			continue
		}
		if mode == "new-session-error" && request.Method == "session/new" {
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   &rpcError{Message: "new session failed"},
			})
			continue
		}
		if mode == "prompt-error" && request.Method == "session/prompt" {
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   &rpcError{Message: "prompt failed"},
			})
			continue
		}
		if mode == "close-error" && request.Method == "session/close" {
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   &rpcError{Message: "close failed"},
			})
			continue
		}
		switch request.Method {
		case "initialize":
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result:  json.RawMessage(`{"protocolVersion":1}`),
			})
		case "session/new":
			if mode == "invalid-session" {
				writeFixtureRPC(rpcEnvelope{
					JSONRPC: "2.0",
					ID:      request.ID,
					Result:  json.RawMessage(`{}`),
				})
				continue
			}
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result:  json.RawMessage(`{"sessionId":"fixture-session"}`),
			})
		case "session/prompt":
			writeFixtureRPC(fixturePromptUpdate())
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result:  json.RawMessage(`{}`),
			})
		case "session/close":
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result:  json.RawMessage(`{}`),
			})
			return
		default:
			writeFixtureRPC(rpcEnvelope{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   &rpcError{Message: "unknown method"},
			})
		}
	}
}

// fixture会話通知。
func fixturePromptUpdate() rpcEnvelope {
	params, err := json.Marshal(map[string]any{
		"sessionId": "fixture-session",
		"update": map[string]any{
			"sessionUpdate": "agent_message_chunk",
			"content": map[string]string{
				"type": "text",
				"text": `{"assistantMessage":"実験案を整理しました","brief":{"purpose":"要約比較","decision":"採用判断","hypothesis":"短い要約が良い","candidatePrompts":["短く要約","長く要約"],"evaluationCriteria":"正確性","environmentConditions":"同一モデル","initialInput":"原文","successCriteria":"勝率","requiredConditions":"再現可能","openQuestion":"費用"}}`,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return rpcEnvelope{
		JSONRPC: "2.0",
		Method:  "session/update",
		Params:  params,
	}
}

// JSON-RPC fixture出力。
func writeFixtureRPC(envelope rpcEnvelope) {
	encoded, err := json.Marshal(envelope)
	if err != nil {
		os.Exit(2)
	}
	_, _ = os.Stdout.Write(append(encoded, '\n'))
}

func TestCodexBriefingAdapter(t *testing.T) {
	tests := []struct {
		name          string
		fixtureMode   string
		wantStartCode apperr.Code
		wantSendCode  apperr.Code
		wantStopCode  apperr.Code
	}{
		{
			name:        "実ACPプロトコルで開始送信停止する",
			fixtureMode: "success",
		},
		{
			name:          "開始失敗をACP未準備へ正規化する",
			fixtureMode:   "initialize-error",
			wantStartCode: apperr.CodeACPNotReady,
		},
		{
			name:          "session作成失敗をACP未準備へ正規化する",
			fixtureMode:   "new-session-error",
			wantStartCode: apperr.CodeACPNotReady,
		},
		{
			name:          "session ID不足をACP未準備へ正規化する",
			fixtureMode:   "invalid-session",
			wantStartCode: apperr.CodeACPNotReady,
		},
		{
			name:         "送信失敗を安全な送信エラーへ正規化する",
			fixtureMode:  "prompt-error",
			wantSendCode: apperr.CodeBriefingMessageFailed,
		},
		{
			name:         "停止失敗を安全な停止エラーへ正規化する",
			fixtureMode:  "close-error",
			wantStopCode: apperr.CodeBriefingStopFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GO_WANT_ACP_FIXTURE", "1")
			t.Setenv("ACP_FIXTURE_MODE", tt.fixtureMode)
			adapter := newFixtureAdapter(t)
			ctx := context.Background()
			err := adapter.StartExperimentBriefing(ctx, "briefing-session", "start-operation")
			if tt.wantStartCode != "" {
				assertAppErrorCode(t, err, tt.wantStartCode)
				return
			}
			if err != nil {
				t.Fatalf("StartExperimentBriefing() error = %v", err)
			}
			if err := adapter.StartExperimentBriefing(ctx, "briefing-session", "start-operation"); err != nil {
				t.Errorf("StartExperimentBriefing() duplicate error = %v", err)
			}

			result, err := adapter.SendExperimentBriefMessage(ctx, "briefing-session", "message-operation", "要約を比較したい")
			if tt.wantSendCode != "" {
				assertAppErrorCode(t, err, tt.wantSendCode)
			} else {
				if err != nil {
					appErr := apperr.As(err)
					if appErr == nil {
						t.Fatalf("SendExperimentBriefMessage() error = %v", err)
					}
					t.Fatalf("SendExperimentBriefMessage() error = %v, cause = %v", err, appErr.Unwrap())
				}
				if result.AssistantMessage != "実験案を整理しました" {
					t.Errorf("AssistantMessage = %q, want %q", result.AssistantMessage, "実験案を整理しました")
				}
				if result.Brief == nil {
					t.Fatal("Brief = nil, want brief")
				}
				if result.Brief.Purpose != "要約比較" {
					t.Errorf("Brief.Purpose = %q, want %q", result.Brief.Purpose, "要約比較")
				}
			}

			stopErr := adapter.StopExperimentBriefing(ctx, "briefing-session", "stop-operation")
			if tt.wantStopCode != "" {
				assertAppErrorCode(t, stopErr, tt.wantStopCode)
			} else if stopErr != nil {
				t.Errorf("StopExperimentBriefing() error = %v", stopErr)
			}
			assertAppErrorCode(t, adapter.StopExperimentBriefing(ctx, "briefing-session", "stop-operation"), apperr.CodeBriefingNotActive)
			_, sendErr := adapter.SendExperimentBriefMessage(ctx, "briefing-session", "message-operation", "再送")
			assertAppErrorCode(t, sendErr, apperr.CodeBriefingNotActive)
		})
	}
}

func TestCodexBriefingAdapterStartCommandFailure(t *testing.T) {
	adapter := newCodexBriefingAdapter(t.TempDir(), "context-lab-command-not-found", nil)
	err := adapter.StartExperimentBriefing(context.Background(), "briefing-session", "start-operation")
	assertAppErrorCode(t, err, apperr.CodeACPNotReady)
}

func TestNewCodexBriefingAdapter(t *testing.T) {
	adapter := NewCodexBriefingAdapter("/tmp/context-lab")
	if adapter.workingDirectory != "/tmp/context-lab" {
		t.Errorf("workingDirectory = %q, want %q", adapter.workingDirectory, "/tmp/context-lab")
	}
	if adapter.command != codexACPCommand {
		t.Errorf("command = %q, want %q", adapter.command, codexACPCommand)
	}
	if len(adapter.arguments) != 2 || adapter.arguments[1] != codexACPPackage {
		t.Errorf("arguments = %v, want codex ACP package", adapter.arguments)
	}
}

func TestCodexACPSessionRequestFailures(t *testing.T) {
	t.Run("JSON化できないparameterを返す", func(t *testing.T) {
		session := &codexACPSession{}
		_, err := session.request(context.Background(), "test", math.Inf(1))
		if err == nil {
			t.Fatal("request() error = nil, want JSON marshal error")
		}
	})

	t.Run("stdin書き込み失敗を返す", func(t *testing.T) {
		reader, writer, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}
		_ = reader.Close()
		_ = writer.Close()
		session := &codexACPSession{
			stdin:     writer,
			responses: make(map[string]chan rpcEnvelope),
		}

		_, err = session.request(context.Background(), "test", map[string]string{})
		if err == nil {
			t.Fatal("request() error = nil, want write error")
		}
	})

	t.Run("呼出元のcancelを返す", func(t *testing.T) {
		reader, writer, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}
		defer reader.Close()
		defer writer.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		session := &codexACPSession{
			stdin:     writer,
			responses: make(map[string]chan rpcEnvelope),
		}

		_, err = session.request(ctx, "test", map[string]string{})
		if err != context.Canceled {
			t.Errorf("request() error = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("sidecar終了を返す", func(t *testing.T) {
		writer := &notifyingWriteCloser{writes: make(chan struct{}, 1)}
		session := &codexACPSession{
			stdin:     writer,
			responses: make(map[string]chan rpcEnvelope),
		}
		errors := make(chan error, 1)
		go func() {
			_, err := session.request(context.Background(), "test", map[string]string{})
			errors <- err
		}()
		<-writer.writes
		session.writeMu.Lock()
		for _, response := range session.responses {
			close(response)
		}
		session.writeMu.Unlock()

		if err := <-errors; err == nil || err.Error() != "ACP sidecar stopped" {
			t.Errorf("request() error = %v, want ACP sidecar stopped", err)
		}
	})
}

func TestCodexACPSessionReadAndNotifications(t *testing.T) {
	t.Run("応答と不正JSONを処理し終了時に保留応答を閉じる", func(t *testing.T) {
		response := make(chan rpcEnvelope, 1)
		pending := make(chan rpcEnvelope, 1)
		session := &codexACPSession{
			responses: map[string]chan rpcEnvelope{
				"1": response,
				"3": pending,
			},
		}

		session.read(strings.NewReader("not-json\n{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n{\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{}}\n"))
		got, open := <-response
		if !open || string(got.Result) != `{"ok":true}` {
			t.Errorf("response = %#v, open = %t", got, open)
		}
		if _, open := <-pending; open {
			t.Error("pending response remains open")
		}
	})

	t.Run("会話chunkだけをcollectorへ追加する", func(t *testing.T) {
		var collector strings.Builder
		session := &codexACPSession{collector: &collector}
		session.handleNotification(rpcEnvelope{Method: "other"})
		session.handleNotification(rpcEnvelope{
			Method: "session/update",
			Params: json.RawMessage("not-json"),
		})
		session.handleNotification(rpcEnvelope{
			Method: "session/update",
			Params: json.RawMessage(`{"update":{"sessionUpdate":"tool_call","content":{"type":"text","text":"除外"}}}`),
		})
		session.handleNotification(rpcEnvelope{
			Method: "session/update",
			Params: json.RawMessage(`{"update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"追加"}}}`),
		})

		if collector.String() != "追加" {
			t.Errorf("collector = %q, want %q", collector.String(), "追加")
		}
	})
}

type notifyingWriteCloser struct {
	writes chan struct{}
}

func (w *notifyingWriteCloser) Write(value []byte) (int, error) {
	w.writes <- struct{}{}

	return len(value), nil
}

func (w *notifyingWriteCloser) Close() error {
	return nil
}

// fixture adapter生成。
func newFixtureAdapter(t *testing.T) *CodexBriefingAdapter {
	t.Helper()

	return newCodexBriefingAdapter(
		t.TempDir(),
		os.Args[0],
		[]string{
			"-test.run=^TestCodexACPFixtureProcess$",
			"--",
		},
	)
}

// アプリケーションエラーcode検証。
func assertAppErrorCode(t *testing.T, err error, want apperr.Code) {
	t.Helper()

	got := apperr.As(err)
	if got == nil {
		t.Fatalf("apperr.As() = nil, want code %q (error = %v)", want, err)
	}
	if got.Code != want {
		t.Errorf("Code = %q, want %q", got.Code, want)
	}
}

func TestParseBriefingResponse(t *testing.T) {
	t.Run("JSONの会話とブリーフを安全なdomain値へ変換する", func(t *testing.T) {
		result := parseBriefingResponse("```json\\n{\"assistantMessage\":\"整理しました\",\"brief\":{\"purpose\":\"比較\",\"decision\":\"採用判断\",\"hypothesis\":\"短い方が良い\",\"candidatePrompts\":[\"A\",\" B \"],\"evaluationCriteria\":\"正確性\",\"environmentConditions\":\"同一モデル\",\"initialInput\":\"入力\",\"successCriteria\":\"勝率\",\"requiredConditions\":\"再現可能\",\"openQuestion\":\"費用\"}}\\n```")

		if result.AssistantMessage != "整理しました" {
			t.Errorf("AssistantMessage = %q, want %q", result.AssistantMessage, "整理しました")
		}
		if result.Brief == nil {
			t.Fatal("Brief = nil, want brief")
		}
		want := &domain.ExperimentBrief{
			Purpose:    "比較",
			Decision:   "採用判断",
			Hypothesis: pointer("短い方が良い"),
			CandidatePrompts: []string{
				"A",
				"B",
			},
			EvaluationCriteria:    "正確性",
			EnvironmentConditions: "同一モデル",
			InitialInput:          "入力",
			SuccessCriteria:       "勝率",
			RequiredConditions:    "再現可能",
			OpenQuestion:          pointer("費用"),
		}
		if result.Brief.Purpose != want.Purpose || result.Brief.Decision != want.Decision || *result.Brief.Hypothesis != *want.Hypothesis || len(result.Brief.CandidatePrompts) != 2 || result.Brief.CandidatePrompts[1] != "B" || *result.Brief.OpenQuestion != *want.OpenQuestion {
			t.Errorf("Brief = %#v, want %#v", result.Brief, want)
		}
	})

	t.Run("JSONでない応答は会話だけを保存する", func(t *testing.T) {
		result := parseBriefingResponse("通常の応答")
		if result.AssistantMessage != "通常の応答" {
			t.Errorf("AssistantMessage = %q, want %q", result.AssistantMessage, "通常の応答")
		}
		if result.Brief != nil {
			t.Errorf("Brief = %#v, want nil", result.Brief)
		}
	})

	t.Run("不正なJSONは会話だけを保存する", func(t *testing.T) {
		result := parseBriefingResponse("応答: {not-json}")
		if result.AssistantMessage != "応答: {not-json}" {
			t.Errorf("AssistantMessage = %q, want %q", result.AssistantMessage, "応答: {not-json}")
		}
		if result.Brief != nil {
			t.Errorf("Brief = %#v, want nil", result.Brief)
		}
	})

	t.Run("空の任意項目はnilに正規化する", func(t *testing.T) {
		result := parseBriefingResponse(`{"brief":{"hypothesis":" ","candidatePrompts":[" ","候補"],"openQuestion":""}}`)
		if result.Brief == nil {
			t.Fatal("Brief = nil, want brief")
		}
		if result.Brief.Hypothesis != nil {
			t.Errorf("Hypothesis = %q, want nil", *result.Brief.Hypothesis)
		}
		if result.Brief.OpenQuestion != nil {
			t.Errorf("OpenQuestion = %q, want nil", *result.Brief.OpenQuestion)
		}
		if got := result.Brief.CandidatePrompts; len(got) != 1 || got[0] != "候補" {
			t.Errorf("CandidatePrompts = %v, want %v", got, []string{"候補"})
		}
	})

	t.Run("ブリーフがないJSONは会話だけを保存する", func(t *testing.T) {
		result := parseBriefingResponse(`{"assistantMessage":"確認します"}`)
		if result.AssistantMessage != "確認します" {
			t.Errorf("AssistantMessage = %q, want %q", result.AssistantMessage, "確認します")
		}
		if result.Brief != nil {
			t.Errorf("Brief = %#v, want nil", result.Brief)
		}
	})
}

func TestBriefingPrompt(t *testing.T) {
	prompt := briefingPrompt("目的を比較したい")
	if !containsAll(prompt, "ツール、ファイル、ネットワークを使わず", "candidatePrompts", "目的を比較したい") {
		t.Errorf("briefingPrompt() = %q, want safety and schema instructions", prompt)
	}
}

func pointer(value string) *string {
	return &value
}

func containsAll(text string, values ...string) bool {
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}

	return true
}

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

// ACP未準備adapterの停止安全な失敗返却。
func TestNotReadyBriefingStopperStopExperimentBriefing(t *testing.T) {
	err := (NotReadyBriefingStopper{}).StopExperimentBriefing(context.Background(), "session", "operation")
	if got := apperr.As(err); got == nil {
		t.Fatal("apperr.As() = nil, want app error")
	} else if got.Code != apperr.CodeACPNotReady {
		t.Errorf("Code = %q, want %q", got.Code, apperr.CodeACPNotReady)
	}
}
