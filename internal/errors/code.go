package apperr

type Code string
type message string

const (
	CodeExperimentsLoadFailed            Code = "EXPERIMENTS_UNAVAILABLE"
	CodePreparationsLoadFailed           Code = "PREPARATIONS_UNAVAILABLE"
	CodeBriefingRequestInvalid           Code = "INVALID_REQUEST"
	CodeACPNotReady                      Code = "ACP_NOT_READY"
	CodeBriefingStartFailed              Code = "EXPERIMENT_BRIEFING_START_UNAVAILABLE"
	CodeBriefingStartPending             Code = "EXPERIMENT_BRIEFING_PENDING"
	CodeBriefingNotFound                 Code = "EXPERIMENT_BRIEFING_NOT_FOUND"
	CodeBriefingLoadFailed               Code = "EXPERIMENT_BRIEFING_UNAVAILABLE"
	CodeBriefingMessageFailed            Code = "EXPERIMENT_BRIEFING_MESSAGE_UNAVAILABLE"
	CodeBriefingNotActive                Code = "EXPERIMENT_BRIEFING_NOT_ACTIVE"
	CodeBriefingMessagePending           Code = "EXPERIMENT_BRIEFING_MESSAGE_PENDING"
	CodeBriefingStopFailed               Code = "EXPERIMENT_BRIEFING_STOP_UNAVAILABLE"
	CodeBriefingStopPending              Code = "EXPERIMENT_BRIEFING_STOP_PENDING"
	CodeExperimentCreateFailed           Code = "EXPERIMENT_CREATE_UNAVAILABLE"
	CodeExperimentBriefIncomplete        Code = "EXPERIMENT_BRIEF_INCOMPLETE"
	CodeExperimentBriefVersionNotFound   Code = "EXPERIMENT_BRIEF_VERSION_NOT_FOUND"
	CodeExperimentCreateRequestConflict  Code = "EXPERIMENT_CREATE_REQUEST_CONFLICT"
	CodeExperimentPreparationNotFound    Code = "EXPERIMENT_PREPARATION_NOT_FOUND"
	CodeExperimentPreparationNotEditable Code = "EXPERIMENT_PREPARATION_NOT_EDITABLE"
	CodeExperimentPreparationUnavailable Code = "EXPERIMENT_PREPARATION_UNAVAILABLE"
	CodeDraftRequestInvalid              Code = "DRAFT_REQUEST_INVALID"
	CodeDraftSaveFailed                  Code = "DRAFT_SAVE_FAILED"
	CodeOperationTimeout                 Code = "OPERATION_TIMEOUT"
	CodeOperationCanceled                Code = "OPERATION_CANCELED"
	CodeUnexpected                       Code = "UNEXPECTED"
)

var defaultMessages = map[Code]message{
	CodeExperimentsLoadFailed:            "実験一覧を取得できませんでした",
	CodePreparationsLoadFailed:           "準備session一覧を取得できませんでした",
	CodeBriefingRequestInvalid:           "開始要求が不正です",
	CodeACPNotReady:                      "実験ブリーフを開始する準備ができていません",
	CodeBriefingStartFailed:              "実験ブリーフを開始できませんでした",
	CodeBriefingStartPending:             "実験ブリーフの開始を確認中です",
	CodeBriefingNotFound:                 "実験ブリーフが見つかりません",
	CodeBriefingLoadFailed:               "実験ブリーフを取得できませんでした",
	CodeBriefingMessageFailed:            "実験ブリーフへメッセージを送信できませんでした",
	CodeBriefingNotActive:                "実験ブリーフは送信できる状態ではありません",
	CodeBriefingMessagePending:           "実験ブリーフのメッセージ送信を確認中です",
	CodeBriefingStopFailed:               "実験ブリーフを終了できませんでした",
	CodeBriefingStopPending:              "実験ブリーフの終了を確認中です",
	CodeExperimentCreateFailed:           "実験を作成できませんでした",
	CodeExperimentBriefIncomplete:        "採用する実験ブリーフの条件が不足しています",
	CodeExperimentBriefVersionNotFound:   "採用する実験ブリーフ版が見つかりません",
	CodeExperimentCreateRequestConflict:  "この作成要求は別のブリーフに使用されています",
	CodeExperimentPreparationNotFound:    "実験準備が見つかりません",
	CodeExperimentPreparationNotEditable: "この実験準備は編集できる状態ではありません",
	CodeExperimentPreparationUnavailable: "実験準備を取得できませんでした",
	CodeDraftRequestInvalid:              "下書き保存要求が不正です",
	CodeDraftSaveFailed:                  "下書きを保存できませんでした",
	CodeOperationTimeout:                 "実験一覧の取得がタイムアウトしました",
	CodeOperationCanceled:                "実験一覧の取得を中止しました",
	CodeUnexpected:                       "予期しないエラーが発生しました",
}
