package apperr

type Code string
type message string

const (
	CodeExperimentsLoadFailed  Code = "EXPERIMENTS_UNAVAILABLE"
	CodeBriefingRequestInvalid Code = "INVALID_REQUEST"
	CodeACPNotReady            Code = "ACP_NOT_READY"
	CodeBriefingStartFailed    Code = "EXPERIMENT_BRIEFING_UNAVAILABLE"
	CodeBriefingStartPending   Code = "EXPERIMENT_BRIEFING_PENDING"
	CodeOperationTimeout       Code = "OPERATION_TIMEOUT"
	CodeOperationCanceled      Code = "OPERATION_CANCELED"
	CodeUnexpected             Code = "UNEXPECTED"
)

var defaultMessages = map[Code]message{
	CodeExperimentsLoadFailed:  "実験一覧を取得できませんでした",
	CodeBriefingRequestInvalid: "開始要求が不正です",
	CodeACPNotReady:            "実験ブリーフを開始する準備ができていません",
	CodeBriefingStartFailed:    "実験ブリーフを開始できませんでした",
	CodeBriefingStartPending:   "実験ブリーフの開始を確認中です",
	CodeOperationTimeout:       "実験一覧の取得がタイムアウトしました",
	CodeOperationCanceled:      "実験一覧の取得を中止しました",
	CodeUnexpected:             "予期しないエラーが発生しました",
}
