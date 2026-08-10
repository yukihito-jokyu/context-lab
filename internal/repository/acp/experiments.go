package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

const (
	codexACPCommand = "npx"
	codexACPPackage = "@agentclientprotocol/codex-acp@1.1.14"
	acpProtocolV1   = 1
)

// CodexBriefingAdapter はCodex ACP sidecarを実験ブリーフ専用に接続するadapter。
// ACPのsession IDやsidecarのプロセス情報はadapter内だけに保持し、外部へ公開しない。
type CodexBriefingAdapter struct {
	workingDirectory string
	command          string
	arguments        []string

	mu       sync.Mutex
	sessions map[string]*codexACPSession
}

// NewCodexBriefingAdapter は実ACPを使う実験ブリーフadapterを生成する。
func NewCodexBriefingAdapter(workingDirectory string) *CodexBriefingAdapter {
	return newCodexBriefingAdapter(workingDirectory, codexACPCommand, []string{"-y", codexACPPackage})
}

func newCodexBriefingAdapter(workingDirectory, command string, arguments []string) *CodexBriefingAdapter {
	return &CodexBriefingAdapter{
		workingDirectory: workingDirectory,
		command:          command,
		arguments:        append([]string(nil), arguments...),
		sessions:         make(map[string]*codexACPSession),
	}
}

// StartExperimentBriefing はACP sidecarを起動してブリーフ会話sessionを開始する。
func (a *CodexBriefingAdapter) StartExperimentBriefing(ctx context.Context, briefingSessionID, _ string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, found := a.sessions[briefingSessionID]; found {
		return nil
	}

	session, err := startCodexACPSession(ctx, a.workingDirectory, a.command, a.arguments)
	if err != nil {
		return apperr.Wrap(apperr.CodeACPNotReady, err)
	}
	a.sessions[briefingSessionID] = session

	return nil
}

// SendExperimentBriefMessage は利用者入力を実ACPへ送り、安全な会話とブリーフ候補を返す。
func (a *CodexBriefingAdapter) SendExperimentBriefMessage(ctx context.Context, briefingSessionID, _ string, message string) (domain.ExperimentBriefingMessageResult, error) {
	session, err := a.session(briefingSessionID)
	if err != nil {
		return domain.ExperimentBriefingMessageResult{}, err
	}

	response, err := session.prompt(ctx, briefingPrompt(message))
	if err != nil {
		return domain.ExperimentBriefingMessageResult{}, apperr.Wrap(apperr.CodeBriefingMessageFailed, err)
	}

	return parseBriefingResponse(response), nil
}

// SendDerivationBriefMessage は利用者入力を派生実験のACP壁打ちへ送り、安全な会話と提案を返す。
func (a *CodexBriefingAdapter) SendDerivationBriefMessage(ctx context.Context, briefingSessionID, _ string, message string) (domain.DerivationBriefingMessageResult, error) {
	session, err := a.session(briefingSessionID)
	if err != nil {
		return domain.DerivationBriefingMessageResult{}, apperr.New(apperr.CodeDerivationBriefingMessageNotActive)
	}

	response, err := session.prompt(ctx, derivationBriefingPrompt(message))
	if err != nil {
		return domain.DerivationBriefingMessageResult{}, apperr.Wrap(apperr.CodeDerivationBriefingMessageFailed, err)
	}
	parsed := parseDerivationBriefingResponse(response)

	return parsed, nil
}

// StopExperimentBriefing は実ACP sessionを閉じ、sidecarを終了する。
func (a *CodexBriefingAdapter) StopExperimentBriefing(ctx context.Context, briefingSessionID, _ string) error {
	a.mu.Lock()
	session, found := a.sessions[briefingSessionID]
	if found {
		delete(a.sessions, briefingSessionID)
	}
	a.mu.Unlock()
	if !found {
		return apperr.New(apperr.CodeBriefingNotActive)
	}

	if err := session.close(ctx); err != nil {
		return apperr.Wrap(apperr.CodeBriefingStopFailed, err)
	}

	return nil
}

// StopDerivationBriefing は派生実験のACP sessionを閉じ、sidecarを終了する。
func (a *CodexBriefingAdapter) StopDerivationBriefing(ctx context.Context, briefingSessionID, _ string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, found := a.sessions[briefingSessionID]
	if !found {
		return apperr.New(apperr.CodeDerivationBriefingStopNotActive)
	}

	if err := session.close(ctx); err != nil {
		return apperr.Wrap(apperr.CodeDerivationBriefingStopFailed, err)
	}
	delete(a.sessions, briefingSessionID)

	return nil
}

func (a *CodexBriefingAdapter) session(briefingSessionID string) (*codexACPSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	session, found := a.sessions[briefingSessionID]
	if !found {
		return nil, apperr.New(apperr.CodeBriefingNotActive)
	}

	return session, nil
}

type codexACPSession struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	sessionID string
	responses map[string]chan rpcEnvelope

	operationMu sync.Mutex
	writeMu     sync.Mutex
	responseMu  sync.Mutex
	nextID      int
	collector   *strings.Builder
}

func startCodexACPSession(ctx context.Context, workingDirectory, command string, arguments []string) (*codexACPSession, error) {
	cmd := exec.CommandContext(ctx, command, arguments...)
	cmd.Dir = workingDirectory
	cmd.Env = append(cmd.Environ(), "INITIAL_AGENT_MODE=read-only", "NO_BROWSER=1")
	stdin, err := cmd.StdinPipe()
	// 単体テスト到達不可: exec.CommandContext直後でStdin未設定のため、標準ライブラリはStdinPipeエラーを返さない。
	if err != nil {
		return nil, fmt.Errorf("open ACP stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	// 単体テスト到達不可: exec.CommandContext直後でStdout未設定のため、標準ライブラリはStdoutPipeエラーを返さない。
	if err != nil {
		return nil, fmt.Errorf("open ACP stdout: %w", err)
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ACP sidecar: %w", err)
	}

	session := &codexACPSession{
		cmd:       cmd,
		stdin:     stdin,
		responses: make(map[string]chan rpcEnvelope),
	}
	go session.read(stdout)

	if _, err := session.request(ctx, "initialize", map[string]any{
		"protocolVersion":    acpProtocolV1,
		"clientCapabilities": map[string]any{},
		"clientInfo": map[string]string{
			"name":    "context-lab",
			"version": "0.1.0",
		},
	}); err != nil {
		session.terminate()
		return nil, fmt.Errorf("initialize ACP: %w", err)
	}
	newSession, err := session.request(ctx, "session/new", map[string]any{
		"cwd":        workingDirectory,
		"mcpServers": []any{},
	})
	if err != nil {
		session.terminate()
		return nil, fmt.Errorf("create ACP session: %w", err)
	}
	var created struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(newSession, &created); err != nil || strings.TrimSpace(created.SessionID) == "" {
		session.terminate()
		return nil, errors.New("ACP did not return a session ID")
	}
	session.sessionID = created.SessionID

	return session, nil
}

func (s *codexACPSession) prompt(ctx context.Context, text string) (string, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()

	var response strings.Builder
	s.responseMu.Lock()
	s.collector = &response
	s.responseMu.Unlock()
	defer func() {
		s.responseMu.Lock()
		s.collector = nil
		s.responseMu.Unlock()
	}()

	if _, err := s.request(ctx, "session/prompt", map[string]any{
		"sessionId": s.sessionID,
		"prompt":    []map[string]string{{"type": "text", "text": text}},
	}); err != nil {
		return "", err
	}

	return response.String(), nil
}

func (s *codexACPSession) close(ctx context.Context) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()

	_, requestErr := s.request(ctx, "session/close", map[string]any{"sessionId": s.sessionID})
	if requestErr != nil {
		return requestErr
	}
	s.terminate()

	return nil
}

func (s *codexACPSession) request(ctx context.Context, method string, params any) (json.RawMessage, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	s.writeMu.Lock()
	s.nextID++
	id := fmt.Sprintf("%d", s.nextID)
	response := make(chan rpcEnvelope, 1)
	s.responses[id] = response
	request := rpcEnvelope{JSONRPC: "2.0", ID: json.RawMessage(id), Method: method, Params: paramsJSON}
	encoded, err := json.Marshal(request)
	if err == nil {
		_, err = s.stdin.Write(append(encoded, '\n'))
	}
	if err != nil {
		delete(s.responses, id)
		s.writeMu.Unlock()
		return nil, err
	}
	s.writeMu.Unlock()

	select {
	case received, open := <-response:
		if !open {
			return nil, errors.New("ACP sidecar stopped")
		}
		if received.Error != nil {
			return nil, fmt.Errorf("ACP %s: %s", method, received.Error.Message)
		}
		return received.Result, nil
	case <-ctx.Done():
		s.writeMu.Lock()
		delete(s.responses, id)
		s.writeMu.Unlock()
		return nil, ctx.Err()
	}
}

func (s *codexACPSession) read(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var envelope rpcEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &envelope); err != nil {
			continue
		}
		if envelope.Method != "" {
			s.handleNotification(envelope)
			continue
		}
		id := string(envelope.ID)
		s.writeMu.Lock()
		response, found := s.responses[id]
		if found {
			delete(s.responses, id)
		}
		s.writeMu.Unlock()
		if found {
			response <- envelope
			close(response)
		}
	}

	s.writeMu.Lock()
	for id, response := range s.responses {
		delete(s.responses, id)
		close(response)
	}
	s.writeMu.Unlock()
}

func (s *codexACPSession) handleNotification(envelope rpcEnvelope) {
	if envelope.Method != "session/update" {
		return
	}
	var update struct {
		Update struct {
			SessionUpdate string `json:"sessionUpdate"`
			Content       struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"update"`
	}
	if err := json.Unmarshal(envelope.Params, &update); err != nil || update.Update.SessionUpdate != "agent_message_chunk" || update.Update.Content.Type != "text" {
		return
	}
	s.responseMu.Lock()
	defer s.responseMu.Unlock()
	if s.collector != nil {
		s.collector.WriteString(update.Update.Content.Text)
	}
}

func (s *codexACPSession) terminate() {
	_ = s.stdin.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	_ = s.cmd.Wait()
}

type rpcEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Message string `json:"message"`
}

func briefingPrompt(message string) string {
	return "あなたは実験設計を整理する壁打ち助手です。ツール、ファイル、ネットワークを使わず、日本語で回答してください。回答はMarkdownや説明を付けず、次のJSON objectだけにしてください。未確定の文字列は空文字、候補がなければ配列は空にします。\\n" +
		`{"assistantMessage":"利用者へ表示する短い応答","brief":{"purpose":"実験の目的","decision":"比較・判断したいこと","hypothesis":"仮説","candidatePrompts":["候補1","候補2"],"evaluationCriteria":"評価基準","environmentConditions":"環境条件","initialInput":"初期入力","successCriteria":"成功条件","requiredConditions":"必須条件","openQuestion":"未解決事項"}}` +
		"\\n利用者入力:\\n" + message
}

// derivationBriefingPrompt は派生実験の安全な構造化提案を要求する。
func derivationBriefingPrompt(message string) string {
	return "あなたは確定済み実験から派生実験を検討する壁打ち助手です。ツール、ファイル、ネットワークを使わず、日本語で回答してください。資格情報、内部推論、プロセス情報を出力せず、回答はMarkdownや説明を付けない次のJSON objectだけにしてください。未確定の文字列は空文字、候補がなければ配列は空にします。\n" +
		`{"assistantMessage":"利用者へ表示する短い応答","brief":{"purpose":"派生実験の目的","decision":"比較・判断したいこと","hypothesis":"仮説","candidatePrompts":["候補1","候補2"],"evaluationCriteria":"評価基準","environmentConditions":"環境条件","initialInput":"初期入力","successCriteria":"成功条件","requiredConditions":"必須条件","openQuestion":"未解決事項"}}` +
		"\n利用者入力:\n" + message
}

// parseDerivationBriefingResponse は構造化JSONだけを安全な派生壁打ち結果へ変換する。
func parseDerivationBriefingResponse(response string) domain.DerivationBriefingMessageResult {
	const fallbackMessage = "提案を安全に解析できませんでした。もう一度お試しください。"

	jsonResponse := extractJSONObject(response)
	if jsonResponse == "" {
		return domain.DerivationBriefingMessageResult{AssistantMessage: fallbackMessage}
	}
	var payload struct {
		AssistantMessage string `json:"assistantMessage"`
	}
	if err := json.Unmarshal([]byte(jsonResponse), &payload); err != nil {
		return domain.DerivationBriefingMessageResult{AssistantMessage: fallbackMessage}
	}
	parsed := parseBriefingResponse(jsonResponse)
	assistantMessage := strings.TrimSpace(payload.AssistantMessage)
	if assistantMessage == "" {
		assistantMessage = fallbackMessage
	}

	return domain.DerivationBriefingMessageResult{
		AssistantMessage: assistantMessage,
		Suggestion:       parsed.Brief,
	}
}

func parseBriefingResponse(response string) domain.ExperimentBriefingMessageResult {
	result := domain.ExperimentBriefingMessageResult{AssistantMessage: strings.TrimSpace(response)}
	jsonResponse := extractJSONObject(response)
	if jsonResponse == "" {
		return result
	}
	var payload struct {
		AssistantMessage string `json:"assistantMessage"`
		Brief            *struct {
			Purpose               string   `json:"purpose"`
			Decision              string   `json:"decision"`
			Hypothesis            string   `json:"hypothesis"`
			CandidatePrompts      []string `json:"candidatePrompts"`
			EvaluationCriteria    string   `json:"evaluationCriteria"`
			EnvironmentConditions string   `json:"environmentConditions"`
			InitialInput          string   `json:"initialInput"`
			SuccessCriteria       string   `json:"successCriteria"`
			RequiredConditions    string   `json:"requiredConditions"`
			OpenQuestion          string   `json:"openQuestion"`
		} `json:"brief"`
	}
	if err := json.Unmarshal([]byte(jsonResponse), &payload); err != nil {
		return result
	}
	if strings.TrimSpace(payload.AssistantMessage) != "" {
		result.AssistantMessage = strings.TrimSpace(payload.AssistantMessage)
	}
	if payload.Brief == nil {
		return result
	}
	result.Brief = &domain.ExperimentBrief{
		Purpose:               strings.TrimSpace(payload.Brief.Purpose),
		Decision:              strings.TrimSpace(payload.Brief.Decision),
		Hypothesis:            optionalText(payload.Brief.Hypothesis),
		CandidatePrompts:      trimTexts(payload.Brief.CandidatePrompts),
		EvaluationCriteria:    strings.TrimSpace(payload.Brief.EvaluationCriteria),
		EnvironmentConditions: strings.TrimSpace(payload.Brief.EnvironmentConditions),
		InitialInput:          strings.TrimSpace(payload.Brief.InitialInput),
		SuccessCriteria:       strings.TrimSpace(payload.Brief.SuccessCriteria),
		RequiredConditions:    strings.TrimSpace(payload.Brief.RequiredConditions),
		OpenQuestion:          optionalText(payload.Brief.OpenQuestion),
	}

	return result
}

func extractJSONObject(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return ""
	}

	return text[start : end+1]
}

func optionalText(text string) *string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func trimTexts(texts []string) []string {
	trimmed := make([]string, 0, len(texts))
	for _, text := range texts {
		if value := strings.TrimSpace(text); value != "" {
			trimmed = append(trimmed, value)
		}
	}

	return trimmed
}

// NotReadyBriefingStarter はACP未準備エラーを返すテスト・代替adapter。
type NotReadyBriefingStarter struct{}

// StartExperimentBriefing はACP未準備エラーを返却。
func (NotReadyBriefingStarter) StartExperimentBriefing(context.Context, string, string) error {
	return apperr.New(apperr.CodeACPNotReady)
}

// NotReadyBriefingMessageSender はACP未準備エラーを返すテスト・代替adapter。
type NotReadyBriefingMessageSender struct{}

// SendExperimentBriefMessage はACP未準備エラーを返却。
func (NotReadyBriefingMessageSender) SendExperimentBriefMessage(context.Context, string, string, string) (domain.ExperimentBriefingMessageResult, error) {
	return domain.ExperimentBriefingMessageResult{}, apperr.New(apperr.CodeACPNotReady)
}

// SendDerivationBriefMessage はACP未準備エラーを返却。
func (NotReadyBriefingMessageSender) SendDerivationBriefMessage(context.Context, string, string, string) (domain.DerivationBriefingMessageResult, error) {
	return domain.DerivationBriefingMessageResult{}, apperr.New(apperr.CodeACPNotReady)
}

// NotReadyBriefingStopper はACP未準備エラーを返すテスト・代替adapter。
type NotReadyBriefingStopper struct{}

// StopExperimentBriefing はACP未準備エラーを返却。
func (NotReadyBriefingStopper) StopExperimentBriefing(context.Context, string, string) error {
	return apperr.New(apperr.CodeACPNotReady)
}
