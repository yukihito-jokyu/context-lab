package domain

import "time"

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
