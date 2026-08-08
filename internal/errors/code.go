package apperr

type Code string
type message string

const (
	CodeExperimentsLoadFailed Code = "EXPERIMENTS_UNAVAILABLE"
	CodeOperationTimeout      Code = "OPERATION_TIMEOUT"
	CodeOperationCanceled     Code = "OPERATION_CANCELED"
	CodeUnexpected            Code = "UNEXPECTED"
)

var defaultMessages = map[Code]message{
	CodeExperimentsLoadFailed: "実験一覧を取得できませんでした",
	CodeOperationTimeout:      "実験一覧の取得がタイムアウトしました",
	CodeOperationCanceled:     "実験一覧の取得を中止しました",
	CodeUnexpected:            "予期しないエラーが発生しました",
}
