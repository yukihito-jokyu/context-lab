package acp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// CodexPreparationAdapter はACPを使う環境準備adapter。
type CodexPreparationAdapter struct {
	workingRoot string
	command     string
	arguments   []string
	start       preparationSessionStarter
}

type preparationSession interface {
	prompt(context.Context, string) (string, error)
	close(context.Context) error
}

type preparationSessionStarter func(context.Context, string, string, []string) (preparationSession, error)

var preparationRelativePath = filepath.Rel

// NewCodexPreparationAdapter は実ACPを使う環境準備adapterを生成する。
func NewCodexPreparationAdapter(workingRoot string) *CodexPreparationAdapter {
	return newCodexPreparationAdapter(workingRoot, codexACPCommand, []string{"-y", codexACPPackage})
}

// newCodexPreparationAdapter はテスト可能な環境準備adapterを生成する。
func newCodexPreparationAdapter(workingRoot, command string, arguments []string) *CodexPreparationAdapter {
	return &CodexPreparationAdapter{
		workingRoot: workingRoot,
		command:     command,
		arguments:   append([]string(nil), arguments...),
		start: func(ctx context.Context, scope, binary string, args []string) (preparationSession, error) {
			return startCodexACPSession(ctx, scope, binary, args)
		},
	}
}

// ValidatePreparationScope は作業root配下の相対directoryを実体確認して解決する。
func (a *CodexPreparationAdapter) ValidatePreparationScope(scope string) (string, string, error) {
	return resolvePreparationScope(a.workingRoot, scope)
}

// PrepareEnvironment は解決済み範囲をACPで一度だけ照合する。
func (a *CodexPreparationAdapter) PrepareEnvironment(ctx context.Context, resolvedScope string) (domain.EnvironmentPreparationResult, error) {
	if err := validateResolvedPreparationScope(a.workingRoot, resolvedScope); err != nil {
		return domain.EnvironmentPreparationResult{}, apperr.New(apperr.CodePreparationScopeInvalid)
	}

	session, err := a.start(ctx, resolvedScope, a.command, a.arguments)
	if err != nil {
		return domain.EnvironmentPreparationResult{}, apperr.Wrap(apperr.CodeACPNotReady, err)
	}
	defer func() { _ = session.close(context.Background()) }()

	response, err := session.prompt(ctx, environmentPreparationPrompt())
	if err != nil {
		return domain.EnvironmentPreparationResult{}, apperr.Wrap(apperr.CodePreparationStartUnavailable, err)
	}
	result, err := parseEnvironmentPreparationResponse(response)
	if err != nil {
		return domain.EnvironmentPreparationResult{}, apperr.Wrap(apperr.CodePreparationStartUnavailable, err)
	}

	return result, nil
}

// resolvePreparationScope は未解決scopeをroot内の実directoryへ解決する。
func resolvePreparationScope(workingRoot, scope string) (string, string, error) {
	scope = strings.TrimSpace(scope)
	if scope == "" || len(scope) > 512 || strings.ContainsRune(scope, 0) || filepath.IsAbs(scope) {
		return "", "", errors.New("invalid preparation scope")
	}
	for _, part := range strings.FieldsFunc(filepath.ToSlash(scope), func(value rune) bool { return value == '/' }) {
		if part == ".." {
			return "", "", errors.New("preparation scope escapes root")
		}
	}

	root, err := filepath.EvalSymlinks(workingRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve working root: %w", err)
	}
	target, err := filepath.EvalSymlinks(filepath.Join(root, scope))
	if err != nil {
		return "", "", fmt.Errorf("resolve preparation scope: %w", err)
	}
	if err := validatePreparationDirectory(root, target); err != nil {
		return "", "", err
	}
	canonical, err := preparationRelativePath(root, target)
	if err != nil {
		return "", "", fmt.Errorf("resolve canonical preparation scope: %w", err)
	}
	if canonical == "." {
		return target, canonical, nil
	}
	return target, filepath.ToSlash(canonical), nil
}

// validateResolvedPreparationScope はACP実行直前にroot内directoryを再確認する。
func validateResolvedPreparationScope(workingRoot, resolvedScope string) error {
	root, err := filepath.EvalSymlinks(workingRoot)
	if err != nil {
		return fmt.Errorf("resolve working root: %w", err)
	}
	target, err := filepath.EvalSymlinks(resolvedScope)
	if err != nil {
		return fmt.Errorf("resolve preparation scope: %w", err)
	}

	return validatePreparationDirectory(root, target)
}

// validatePreparationDirectory はdirectoryが作業root内に残ることを確認する。
func validatePreparationDirectory(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return errors.New("preparation scope escapes root")
	}
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("stat preparation scope: %w", err)
	}
	if !info.IsDir() {
		return errors.New("preparation scope is not directory")
	}

	return nil
}

// environmentPreparationPrompt は安全な構造化候補だけを依頼する。
func environmentPreparationPrompt() string {
	return "作業範囲を読み取り専用で確認し、環境準備候補をJSONだけで返してください。形式: {\"candidates\":[{\"environmentConditions\":\"安全な環境条件\",\"summary\":\"安全な要約\"}],\"diagnostics\":[{\"code\":\"安全な診断コード\",\"summary\":\"安全な診断要約\"}]}。資格情報、内部推論、プロセスID、絶対パスは含めないでください。"
}

// parseEnvironmentPreparationResponse はACP生応答から安全DTOだけを抽出する。
func parseEnvironmentPreparationResponse(response string) (domain.EnvironmentPreparationResult, error) {
	var payload struct {
		Candidates []struct {
			EnvironmentConditions string `json:"environmentConditions"`
			Summary               string `json:"summary"`
		} `json:"candidates"`
		Diagnostics []struct {
			Code    string `json:"code"`
			Summary string `json:"summary"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(response), &payload); err != nil {
		return domain.EnvironmentPreparationResult{}, fmt.Errorf("parse environment preparation response: %w", err)
	}

	result := domain.EnvironmentPreparationResult{Candidates: make([]domain.EnvironmentPreparationCandidate, 0, len(payload.Candidates)), Diagnostics: make([]domain.EnvironmentPreparationDiagnostic, 0, len(payload.Diagnostics))}
	for _, candidate := range payload.Candidates {
		conditions := strings.TrimSpace(candidate.EnvironmentConditions)
		summary := strings.TrimSpace(candidate.Summary)
		if !isSafeEnvironmentPreparationText(conditions) || !isSafeEnvironmentPreparationText(summary) {
			return domain.EnvironmentPreparationResult{}, errors.New("invalid environment preparation candidate")
		}
		result.Candidates = append(result.Candidates, domain.EnvironmentPreparationCandidate{EnvironmentConditions: conditions, Summary: summary})
	}
	for _, diagnostic := range payload.Diagnostics {
		code := strings.TrimSpace(diagnostic.Code)
		summary := strings.TrimSpace(diagnostic.Summary)
		if !isSafeEnvironmentPreparationText(code) || !isSafeEnvironmentPreparationText(summary) {
			return domain.EnvironmentPreparationResult{}, errors.New("invalid environment preparation diagnostic")
		}
		result.Diagnostics = append(result.Diagnostics, domain.EnvironmentPreparationDiagnostic{Code: code, SafeSummary: summary})
	}

	return result, nil
}

// isSafeEnvironmentPreparationText は画面・DBへ保存してよい安全な要約だけを許可する。
func isSafeEnvironmentPreparationText(value string) bool {
	if value == "" || len(value) > 2_000 {
		return false
	}
	lower := strings.ToLower(value)
	for _, forbidden := range []string{
		"/users/",
		"/home/",
		"/tmp/",
		"c:\\",
		"credential",
		"password",
		"secret",
		"token",
		"api key",
		"pid",
		"プロセスid",
		"docker id",
		"container id",
		"sidecar",
		"内部推論",
		"reasoning",
		"chain of thought",
	} {
		if strings.Contains(lower, forbidden) {
			return false
		}
	}

	return true
}
