package domain

import "time"

// Preparation は環境準備sessionの安全な一覧表示情報。
type Preparation struct {
	ID             string
	State          string
	StartedAt      time.Time
	LastObservedAt time.Time
}

// PreparationDetail は環境準備sessionの安全な詳細表示情報。
type PreparationDetail struct {
	ID             string
	State          string
	StartedAt      time.Time
	LastObservedAt time.Time
	Candidates     []PreparationCandidate
	Diagnostics    []PreparationDiagnostic
	Failure        *PreparationFailure
	Reconciliation PreparationReconciliation
}

// PreparationCandidate は環境準備で検出した安全な候補情報。
type PreparationCandidate struct {
	ID                    string
	EnvironmentConditions string
	Summary               string
	CreatedAt             time.Time
}

// PreparationDiagnostic は環境準備の安全な診断情報。
type PreparationDiagnostic struct {
	ID          string
	Code        string
	SafeSummary string
	OccurredAt  time.Time
}

// PreparationFailure は環境準備の安全な失敗情報。
type PreparationFailure struct {
	Code       string
	OccurredAt time.Time
}

// PreparationReconciliation は環境準備の再照合状態。
type PreparationReconciliation struct {
	State          string
	LastObservedAt time.Time
}

// EnvironmentPreparationStart は環境準備開始操作の安全な状態。
type EnvironmentPreparationStart struct {
	RequestID     string
	PreparationID string
	Scope         string
	State         string
	FailureCode   string
}

const (
	// EnvironmentPreparationStateStarting は開始記録直後の状態。
	EnvironmentPreparationStateStarting = "starting"
	// EnvironmentPreparationStateRunning はACP照合中の状態。
	EnvironmentPreparationStateRunning = "running"
	// EnvironmentPreparationStateCompleted は候補の保存済み状態。
	EnvironmentPreparationStateCompleted = "completed"
	// EnvironmentPreparationStateFailed は開始または照合の失敗状態。
	EnvironmentPreparationStateFailed = "failed"
)

// EnvironmentPreparationResult はACPが返す安全な環境準備結果。
type EnvironmentPreparationResult struct {
	Candidates  []EnvironmentPreparationCandidate
	Diagnostics []EnvironmentPreparationDiagnostic
}

// EnvironmentPreparationCandidate は保存前の安全な環境候補。
type EnvironmentPreparationCandidate struct {
	EnvironmentConditions string
	Summary               string
}

// EnvironmentPreparationDiagnostic は保存前の安全な診断情報。
type EnvironmentPreparationDiagnostic struct {
	Code        string
	SafeSummary string
}
