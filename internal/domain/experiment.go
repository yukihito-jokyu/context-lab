package domain

import (
	"strings"
	"time"
)

const (
	BriefingStartStateStarting = "starting"
	BriefingStartStateStarted  = "started"
	BriefingStartStateFailed   = "failed"
	BriefingStartStateStopped  = "stopped"
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

// ExperimentBriefingStopOperation は実験ブリーフ終了の永続的な結果。
type ExperimentBriefingStopOperation struct {
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
	VersionID             string
	Purpose               string
	Decision              string
	Hypothesis            *string
	CandidatePrompts      []string
	EvaluationCriteria    string
	EnvironmentConditions string
	InitialInput          string
	SuccessCriteria       string
	RequiredConditions    string
	OpenQuestion          *string
}

// ExperimentCreation はブリーフ採用で作成した準備中実験。
type ExperimentCreation struct {
	ExperimentID string
	State        string
}

// ExperimentPreparation は準備中実験の編集前条件。
type ExperimentPreparation struct {
	ExperimentID          string
	State                 string
	Purpose               string
	Hypothesis            *string
	EnvironmentConditions string
	InitialInput          string
	Prompts               []ExperimentPreparationPrompt
	EvaluationAxes        string
	Source                ExperimentPreparationSource
	LastConfirmedAt       time.Time
}

// ExperimentPreparationPrompt は準備中実験の表示用prompt。
type ExperimentPreparationPrompt struct {
	SequenceNo int
	Content    string
}

// ExperimentPreparationDraft は準備中実験へ保存する編集下書き。
type ExperimentPreparationDraft struct {
	RequestID             string
	ExperimentID          string
	Purpose               string
	Hypothesis            *string
	EnvironmentConditions string
	InitialInput          string
	Prompts               []ExperimentPreparationPrompt
	EvaluationAxes        string
	SavedAt               time.Time
}

// ExperimentFixedConditions は実験開始前に不変として記録する条件。
type ExperimentFixedConditions struct {
	RequestID             string
	ExperimentID          string
	Purpose               string
	Hypothesis            *string
	EnvironmentConditions string
	InitialInput          string
	Prompts               []ExperimentPreparationPrompt
	EvaluationAxes        string
	FixedConditionID      string
	OperationID           string
	FixedAt               time.Time
}

// Valid は固定できる必須条件が揃っているかを返す。
func (c ExperimentFixedConditions) Valid() bool {
	if strings.TrimSpace(c.Purpose) == "" || strings.TrimSpace(c.EnvironmentConditions) == "" || strings.TrimSpace(c.InitialInput) == "" || strings.TrimSpace(c.EvaluationAxes) == "" || len(c.Prompts) == 0 {
		return false
	}
	for _, prompt := range c.Prompts {
		if strings.TrimSpace(prompt.Content) == "" {
			return false
		}
	}

	return true
}

// ExperimentPreparationSource は採用済みブリーフの安全な表示用情報。
type ExperimentPreparationSource struct {
	State     string
	VersionID string
}

// ExperimentPreparationRequiredFields は条件固定に必要な入力の充足状態。
type ExperimentPreparationRequiredFields struct {
	Purpose               bool
	EnvironmentConditions bool
	InitialInput          bool
	Prompts               bool
	EvaluationAxes        bool
}

// RequiredFields は条件固定に必要な入力の充足状態を返す。
func (p ExperimentPreparation) RequiredFields() ExperimentPreparationRequiredFields {
	promptsComplete := len(p.Prompts) >= 2
	for _, prompt := range p.Prompts {
		if strings.TrimSpace(prompt.Content) == "" {
			promptsComplete = false
		}
	}

	return ExperimentPreparationRequiredFields{
		Purpose:               strings.TrimSpace(p.Purpose) != "",
		EnvironmentConditions: strings.TrimSpace(p.EnvironmentConditions) != "",
		InitialInput:          strings.TrimSpace(p.InitialInput) != "",
		Prompts:               promptsComplete,
		EvaluationAxes:        strings.TrimSpace(p.EvaluationAxes) != "",
	}
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
