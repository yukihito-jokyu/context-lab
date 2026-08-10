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
