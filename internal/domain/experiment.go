package domain

import "time"

const (
	BriefingStartStateStarting = "starting"
	BriefingStartStateStarted  = "started"
	BriefingStartStateFailed   = "failed"
)

// ExperimentBriefingStart は実験ブリーフ開始の永続的な結果。
type ExperimentBriefingStart struct {
	RequestID         string
	BriefingSessionID string
	OperationID       string
	State             string
	FailureCode       string
}

// Experiment は一覧表示に必要な実験の安全な属性。
type Experiment struct {
	ID                      string
	Purpose                 string
	State                   string
	ProgressSummary         string
	DerivedFromExperimentID *string
	UpdatedAt               time.Time
}

// ExperimentCollection は実験一覧と再開判断用の集計。
type ExperimentCollection struct {
	Experiments          []Experiment
	CancelledExperiments []Experiment
	LastConfirmedAt      *time.Time
}
