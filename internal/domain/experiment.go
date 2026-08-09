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

// ExperimentBriefingMessageOperation は実験ブリーフ会話送信の永続的な結果。
type ExperimentBriefingMessageOperation struct {
	RequestID         string
	BriefingSessionID string
	OperationID       string
	State             string
	FailureCode       string
}

// ExperimentBriefingMessageResult はACPから受け取る安全な会話と下書き候補。
type ExperimentBriefingMessageResult struct {
	AssistantMessage string
	Brief            *ExperimentBrief
}

// ExperimentBriefing は実験ブリーフ画面の再表示に必要な永続状態。
type ExperimentBriefing struct {
	State           string
	Messages        []ExperimentBriefingMessage
	LatestBrief     *ExperimentBrief
	LastConfirmedAt time.Time
}

// ExperimentBriefingMessage は利用者表示用の実験ブリーフ会話。
type ExperimentBriefingMessage struct {
	Role       string
	Content    string
	SequenceNo int
	CreatedAt  time.Time
}

// ExperimentBrief は実験ブリーフの一版。
type ExperimentBrief struct {
	VersionID          string
	Decision           string
	Hypothesis         *string
	SuccessCriteria    string
	RequiredConditions string
	OpenQuestion       *string
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
