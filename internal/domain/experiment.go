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

// DerivationBriefingStart は派生実験ブリーフ開始の永続的な結果。
type DerivationBriefingStart struct {
	RequestID          string
	SourceExperimentID string
	BriefingSessionID  string
	OperationID        string
	State              string
	FailureCode        string
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

// ExperimentWorkspace は実験開始後の画面表示に必要な不変条件と進行状況。
type ExperimentWorkspace struct {
	ExperimentID            string
	State                   string
	FixedConditions         ExperimentFixedConditions
	ConditionFixOperationID string
	ConditionFixOperationAt time.Time
	LastConfirmedAt         time.Time
	Runs                    []ExperimentWorkspaceRun
	Evaluations             []ExperimentWorkspaceEvaluation
}

// ExperimentWorkspaceRun は実験ワークスペースに表示するrunの安全な進行状況。
type ExperimentWorkspaceRun struct {
	ID           string
	RetryOfRunID *string
	State        string
	Summary      *string
	UpdatedAt    time.Time
}

// ExperimentRunRetry は終了runから作成した再実行用run。
type ExperimentRunRetry struct {
	RequestID    string
	SourceRunID  string
	ExperimentID string
	RetryRunID   string
	OperationID  string
	State        string
	CreatedAt    time.Time
}

// ExperimentConclusion は確定済み実験結論の安全な不変結果。
type ExperimentConclusion struct {
	RequestID                string
	ExperimentID             string
	ConclusionID             string
	Conclusion               string
	State                    string
	FinalizedAt              time.Time
	EvaluationSnapshotDigest string
}

// ExperimentDerivationSource は派生実験作成前に確認する安全な正本。
type ExperimentDerivationSource struct {
	ExperimentID     string
	Purpose          string
	FixedConditions  *ExperimentFixedConditions
	Conclusion       *ExperimentConclusion
	CanCreateDerived bool
	ReasonCode       string
}

// DerivedExperiment は派生作成の不変結果。
type DerivedExperiment struct {
	RequestID          string
	ExperimentID       string
	SourceExperimentID string
	State              string
	CreatedAt          time.Time
}

// DerivedExperimentChanges は派生元固定条件へ適用する差分。
type DerivedExperimentChanges struct {
	Purpose               *string
	Hypothesis            *string
	EnvironmentConditions *string
	InitialInput          *string
	Prompts               *[]ExperimentPreparationPrompt
	EvaluationAxes        *string
}

// ExperimentComparison は同一実験内の評価結果を比較表示する安全な正本。
type ExperimentComparison struct {
	Experiment      ExperimentComparisonExperiment
	Evaluations     []ExperimentComparisonEvaluation
	Conclusion      *ExperimentConclusion
	LastConfirmedAt time.Time
}

// ExperimentComparisonExperiment は比較対象の固定済み実験条件。
type ExperimentComparisonExperiment struct {
	ID             string
	Purpose        string
	EvaluationAxes string
}

// ExperimentComparisonEvaluation は一件の評価とrun要約の比較用情報。
type ExperimentComparisonEvaluation struct {
	EvaluationID   string
	RunID          string
	State          string
	RunSummary     *string
	Result         ExperimentEvaluationResult
	Reconciliation ExperimentEvaluationReconciliation
	UpdatedAt      time.Time
}

// ExperimentRunDetail はrun詳細画面が再表示する安全な観測結果。
type ExperimentRunDetail struct {
	Run             ExperimentRunFact
	FixedPrompt     ExperimentPreparationPrompt
	Operation       ExperimentRunOperation
	Observations    []ExperimentRunObservation
	Artifacts       ExperimentRunArtifacts
	Failure         *ExperimentRunFailure
	Reconciliation  ExperimentRunReconciliation
	LastConfirmedAt time.Time
}

// ExperimentRunFact はrunの安全な実行事実。
type ExperimentRunFact struct {
	ID           string
	ExperimentID string
	State        string
	Summary      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ExperimentRunOperation はrunを開始した操作の安全な状態。
type ExperimentRunOperation struct {
	ID        string
	State     string
	UpdatedAt time.Time
}

// ExperimentRunObservation は時系列観測の安全な要約。
type ExperimentRunObservation struct {
	SequenceNo int
	Kind       string
	OccurredAt time.Time
	Summary    string
}

// ExperimentRunArtifact は比較可能なartifact差分の安全な識別子。
type ExperimentRunArtifact struct {
	Digest string
	Label  *string
	Status string
}

// ExperimentRunArtifacts はartifact取得の完全性。
type ExperimentRunArtifacts struct {
	Status     string
	Items      []ExperimentRunArtifact
	ReasonCode string
}

// ExperimentRunFailure はrun固有の安全な失敗情報。
type ExperimentRunFailure struct {
	Code           string
	OccurredAt     time.Time
	PartialSummary *string
}

// ExperimentRunReconciliation は保存済み観測との照合状態。
type ExperimentRunReconciliation struct {
	State          string
	LastObservedAt time.Time
}

// ExperimentEvaluationDetail は評価詳細画面が再表示する安全な評価結果。
type ExperimentEvaluationDetail struct {
	Evaluation      ExperimentEvaluationFact
	Operation       ExperimentEvaluationOperation
	Evidence        ExperimentEvaluationEvidence
	Result          ExperimentEvaluationResult
	Failure         *ExperimentEvaluationFailure
	Reconciliation  ExperimentEvaluationReconciliation
	LastConfirmedAt time.Time
}

// ExperimentEvaluationFact は評価の安全な実行事実。
type ExperimentEvaluationFact struct {
	ID           string
	ExperimentID string
	RunID        string
	State        string
	Summary      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ExperimentEvaluationOperation は評価開始操作の安全な状態。
type ExperimentEvaluationOperation struct {
	ID, State string
	UpdatedAt time.Time
}

// ExperimentEvaluationEvidence は評価に用いた安全な根拠。
type ExperimentEvaluationEvidence struct{ RunSummary, EvaluationAxes string }

// ExperimentEvaluationResult は評価結果または評価不能理由。
type ExperimentEvaluationResult struct {
	Status     string
	Summary    *string
	ReasonCode string
}

// ExperimentEvaluationFailure は評価不能の安全な理由。
type ExperimentEvaluationFailure struct {
	Code       string
	OccurredAt time.Time
}

// ExperimentEvaluationReconciliation は評価結果の照合状態。
type ExperimentEvaluationReconciliation struct {
	State          string
	LastObservedAt time.Time
}

// ExperimentStart は実験開始commandの永続的な結果。
type ExperimentStart struct {
	RequestID       string
	ExperimentID    string
	OperationID     string
	State           string
	FailureCode     string
	FixedConditions ExperimentFixedConditions
	Runs            []ExperimentWorkspaceRun
}

// ExperimentRunRequest はDocker runnerへ渡す一件の隔離実行条件。
type ExperimentRunRequest struct {
	ExperimentID          string
	RunID                 string
	Purpose               string
	EnvironmentConditions string
	InitialInput          string
	Prompt                string
	EvaluationAxes        string
}

const (
	ExperimentStartStateStarting           = "starting"
	ExperimentStartStateRunning            = "running"
	ExperimentStartStateFailed             = "failed"
	ExperimentRunStateQueued               = "queued"
	ExperimentRunStateRunning              = "running"
	ExperimentRunStateCompleted            = "completed"
	ExperimentRunStateFailed               = "failed"
	ExperimentEvaluationStateStarting      = "starting"
	ExperimentEvaluationStateCompleted     = "completed"
	ExperimentEvaluationStateFailed        = "failed"
	ExperimentRunArtifactStatusComplete    = "complete"
	ExperimentRunArtifactStatusPartial     = "partial"
	ExperimentRunArtifactStatusNotRecorded = "notRecorded"
)

// ExperimentWorkspaceEvaluation は実験ワークスペースに表示するevaluationの安全な進行状況。
type ExperimentWorkspaceEvaluation struct {
	ID        string
	State     string
	Summary   *string
	UpdatedAt time.Time
}

// ExperimentRunEvaluation は一件のrun評価commandの永続的な結果。
type ExperimentRunEvaluation struct {
	RequestID      string
	ExperimentID   string
	RunID          string
	EvaluationID   string
	OperationID    string
	State          string
	Summary        *string
	FailureCode    string
	UpdatedAt      time.Time
	RunSummary     string
	Purpose        string
	EvaluationAxes string
}

// ExperimentEvaluationRequest は隔離評価runnerへ渡す安全な入力。
type ExperimentEvaluationRequest struct {
	RunID          string
	RunSummary     string
	Purpose        string
	EvaluationAxes string
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
