package apperr

type Code string
type message string

const (
	CodeExperimentsLoadFailed                    Code = "EXPERIMENTS_UNAVAILABLE"
	CodePreparationsLoadFailed                   Code = "PREPARATIONS_UNAVAILABLE"
	CodePreparationRequestInvalid                Code = "PREPARATION_REQUEST_INVALID"
	CodePreparationNotFound                      Code = "PREPARATION_NOT_FOUND"
	CodePreparationUnavailable                   Code = "PREPARATION_UNAVAILABLE"
	CodePreparationStartRequestInvalid           Code = "PREPARATION_START_REQUEST_INVALID"
	CodePreparationScopeInvalid                  Code = "PREPARATION_SCOPE_INVALID"
	CodePreparationStartPending                  Code = "PREPARATION_START_PENDING"
	CodePreparationStartRequestConflict          Code = "PREPARATION_START_REQUEST_CONFLICT"
	CodePreparationStartUnavailable              Code = "PREPARATION_START_UNAVAILABLE"
	CodeBriefingRequestInvalid                   Code = "INVALID_REQUEST"
	CodeACPNotReady                              Code = "ACP_NOT_READY"
	CodeBriefingStartFailed                      Code = "EXPERIMENT_BRIEFING_START_UNAVAILABLE"
	CodeBriefingStartPending                     Code = "EXPERIMENT_BRIEFING_PENDING"
	CodeBriefingNotFound                         Code = "EXPERIMENT_BRIEFING_NOT_FOUND"
	CodeBriefingLoadFailed                       Code = "EXPERIMENT_BRIEFING_UNAVAILABLE"
	CodeBriefingMessageFailed                    Code = "EXPERIMENT_BRIEFING_MESSAGE_UNAVAILABLE"
	CodeBriefingNotActive                        Code = "EXPERIMENT_BRIEFING_NOT_ACTIVE"
	CodeBriefingMessagePending                   Code = "EXPERIMENT_BRIEFING_MESSAGE_PENDING"
	CodeBriefingStopFailed                       Code = "EXPERIMENT_BRIEFING_STOP_UNAVAILABLE"
	CodeBriefingStopPending                      Code = "EXPERIMENT_BRIEFING_STOP_PENDING"
	CodeExperimentCreateFailed                   Code = "EXPERIMENT_CREATE_UNAVAILABLE"
	CodeExperimentBriefIncomplete                Code = "EXPERIMENT_BRIEF_INCOMPLETE"
	CodeExperimentBriefVersionNotFound           Code = "EXPERIMENT_BRIEF_VERSION_NOT_FOUND"
	CodeExperimentCreateRequestConflict          Code = "EXPERIMENT_CREATE_REQUEST_CONFLICT"
	CodeExperimentPreparationNotFound            Code = "EXPERIMENT_PREPARATION_NOT_FOUND"
	CodeExperimentPreparationNotEditable         Code = "EXPERIMENT_PREPARATION_NOT_EDITABLE"
	CodeExperimentPreparationUnavailable         Code = "EXPERIMENT_PREPARATION_UNAVAILABLE"
	CodeDraftRequestInvalid                      Code = "DRAFT_REQUEST_INVALID"
	CodeDraftSaveFailed                          Code = "DRAFT_SAVE_FAILED"
	CodeFixConditionsRequestInvalid              Code = "FIX_CONDITIONS_REQUEST_INVALID"
	CodeExperimentConditionsAlreadyFixed         Code = "EXPERIMENT_CONDITIONS_ALREADY_FIXED"
	CodeExperimentConditionsConflict             Code = "EXPERIMENT_CONDITIONS_CONFLICT"
	CodeFixConditionsSaveFailed                  Code = "FIX_CONDITIONS_SAVE_FAILED"
	CodeConditionsInvalid                        Code = "CONDITIONS_INVALID"
	CodeExperimentWorkspaceRequestInvalid        Code = "EXPERIMENT_WORKSPACE_REQUEST_INVALID"
	CodeExperimentWorkspaceNotFound              Code = "EXPERIMENT_WORKSPACE_NOT_FOUND"
	CodeExperimentWorkspaceNotReady              Code = "EXPERIMENT_WORKSPACE_NOT_READY"
	CodeExperimentWorkspaceUnavailable           Code = "EXPERIMENT_WORKSPACE_UNAVAILABLE"
	CodeExperimentStartRequestInvalid            Code = "EXPERIMENT_START_REQUEST_INVALID"
	CodeExperimentStartNotReady                  Code = "EXPERIMENT_START_NOT_READY"
	CodeExperimentStartPending                   Code = "EXPERIMENT_START_PENDING"
	CodeExperimentStartFailed                    Code = "EXPERIMENT_START_UNAVAILABLE"
	CodeRunEvaluationRequestInvalid              Code = "RUN_EVALUATION_REQUEST_INVALID"
	CodeRunEvaluationNotReady                    Code = "RUN_EVALUATION_NOT_READY"
	CodeRunEvaluationAlreadyExists               Code = "RUN_EVALUATION_ALREADY_EXISTS"
	CodeRunEvaluationPending                     Code = "RUN_EVALUATION_PENDING"
	CodeRunEvaluationFailed                      Code = "RUN_EVALUATION_UNAVAILABLE"
	CodeRunDetailRequestInvalid                  Code = "RUN_DETAIL_REQUEST_INVALID"
	CodeRunDetailNotFound                        Code = "RUN_DETAIL_NOT_FOUND"
	CodeRunDetailUnavailable                     Code = "RUN_DETAIL_UNAVAILABLE"
	CodeEvaluationDetailRequestInvalid           Code = "EVALUATION_DETAIL_REQUEST_INVALID"
	CodeEvaluationDetailNotFound                 Code = "EVALUATION_DETAIL_NOT_FOUND"
	CodeEvaluationDetailUnavailable              Code = "EVALUATION_DETAIL_UNAVAILABLE"
	CodeRunRetryRequestInvalid                   Code = "RUN_RETRY_REQUEST_INVALID"
	CodeRunRetryNotFound                         Code = "RUN_RETRY_NOT_FOUND"
	CodeRunRetryNotAllowed                       Code = "RUN_RETRY_NOT_ALLOWED"
	CodeRunRetryRequestConflict                  Code = "RUN_RETRY_REQUEST_CONFLICT"
	CodeRunRetryUnavailable                      Code = "RUN_RETRY_UNAVAILABLE"
	CodeExperimentComparisonInvalid              Code = "EXPERIMENT_COMPARISON_INVALID"
	CodeExperimentComparisonNotFound             Code = "EXPERIMENT_COMPARISON_NOT_FOUND"
	CodeExperimentComparisonUnavailable          Code = "EXPERIMENT_COMPARISON_UNAVAILABLE"
	CodeExperimentDerivationSourceInvalid        Code = "EXPERIMENT_DERIVATION_SOURCE_INVALID"
	CodeExperimentDerivationSourceNotFound       Code = "EXPERIMENT_DERIVATION_SOURCE_NOT_FOUND"
	CodeExperimentDerivationSourceUnavailable    Code = "EXPERIMENT_DERIVATION_SOURCE_UNAVAILABLE"
	CodeDerivedExperimentInvalid                 Code = "DERIVED_EXPERIMENT_INVALID"
	CodeDerivedExperimentSourceNotFound          Code = "DERIVED_EXPERIMENT_SOURCE_NOT_FOUND"
	CodeDerivedExperimentSourceNotEligible       Code = "DERIVED_EXPERIMENT_SOURCE_NOT_ELIGIBLE"
	CodeDerivedExperimentRequestConflict         Code = "DERIVED_EXPERIMENT_REQUEST_CONFLICT"
	CodeDerivedExperimentUnavailable             Code = "DERIVED_EXPERIMENT_UNAVAILABLE"
	CodeDerivationBriefingStartFailed            Code = "DERIVATION_BRIEFING_START_UNAVAILABLE"
	CodeDerivationBriefingPending                Code = "DERIVATION_BRIEFING_PENDING"
	CodeDerivationBriefingInvalid                Code = "DERIVATION_BRIEFING_INVALID"
	CodeDerivationBriefingNotFound               Code = "DERIVATION_BRIEFING_NOT_FOUND"
	CodeDerivationBriefingUnavailable            Code = "DERIVATION_BRIEFING_UNAVAILABLE"
	CodeDerivationBriefingMessageInvalid         Code = "DERIVATION_BRIEFING_MESSAGE_INVALID"
	CodeDerivationBriefingMessageNotFound        Code = "DERIVATION_BRIEFING_MESSAGE_NOT_FOUND"
	CodeDerivationBriefingMessageNotActive       Code = "DERIVATION_BRIEFING_MESSAGE_NOT_ACTIVE"
	CodeDerivationBriefingMessageRequestConflict Code = "DERIVATION_BRIEFING_MESSAGE_REQUEST_CONFLICT"
	CodeDerivationBriefingMessagePending         Code = "DERIVATION_BRIEFING_MESSAGE_PENDING"
	CodeDerivationBriefingMessageFailed          Code = "DERIVATION_BRIEFING_MESSAGE_UNAVAILABLE"
	CodeDerivationBriefingStopInvalid            Code = "DERIVATION_BRIEFING_STOP_INVALID"
	CodeDerivationBriefingStopNotFound           Code = "DERIVATION_BRIEFING_STOP_NOT_FOUND"
	CodeDerivationBriefingStopNotActive          Code = "DERIVATION_BRIEFING_STOP_NOT_ACTIVE"
	CodeDerivationBriefingStopRequestConflict    Code = "DERIVATION_BRIEFING_STOP_REQUEST_CONFLICT"
	CodeDerivationBriefingStopPending            Code = "DERIVATION_BRIEFING_STOP_PENDING"
	CodeDerivationBriefingStopFailed             Code = "DERIVATION_BRIEFING_STOP_UNAVAILABLE"
	CodeExperimentConclusionInvalid              Code = "EXPERIMENT_CONCLUSION_INVALID"
	CodeExperimentConclusionNotFound             Code = "EXPERIMENT_CONCLUSION_NOT_FOUND"
	CodeExperimentConclusionNotReady             Code = "EXPERIMENT_CONCLUSION_NOT_READY"
	CodeExperimentConclusionRequestConflict      Code = "EXPERIMENT_CONCLUSION_REQUEST_CONFLICT"
	CodeExperimentConclusionAlreadyFinalized     Code = "EXPERIMENT_CONCLUSION_ALREADY_FINALIZED"
	CodeExperimentConclusionUnavailable          Code = "EXPERIMENT_CONCLUSION_UNAVAILABLE"
	CodeInsightWorkspaceUnavailable              Code = "INSIGHT_WORKSPACE_UNAVAILABLE"
	CodeInsightCreateInvalid                     Code = "INSIGHT_CREATE_INVALID"
	CodeInsightCreateEvidenceInsufficient        Code = "INSIGHT_CREATE_EVIDENCE_INSUFFICIENT"
	CodeInsightCreateEvidenceNotFound            Code = "INSIGHT_CREATE_EVIDENCE_NOT_FOUND"
	CodeInsightCreateRequestConflict             Code = "INSIGHT_CREATE_REQUEST_CONFLICT"
	CodeInsightCreateUnavailable                 Code = "INSIGHT_CREATE_UNAVAILABLE"
	CodeOperationTimeout                         Code = "OPERATION_TIMEOUT"
	CodeOperationCanceled                        Code = "OPERATION_CANCELED"
	CodeUnexpected                               Code = "UNEXPECTED"
)

var defaultMessages = map[Code]message{
	CodeExperimentsLoadFailed:                    "実験一覧を取得できませんでした",
	CodePreparationsLoadFailed:                   "準備session一覧を取得できませんでした",
	CodePreparationRequestInvalid:                "準備sessionの指定が不正です",
	CodePreparationNotFound:                      "準備sessionが見つかりません",
	CodePreparationUnavailable:                   "準備sessionを取得できませんでした",
	CodePreparationStartRequestInvalid:           "環境準備の開始要求が不正です",
	CodePreparationScopeInvalid:                  "環境準備の対象範囲が不正です",
	CodePreparationStartPending:                  "この範囲の環境準備はすでに実行中です",
	CodePreparationStartRequestConflict:          "この開始要求は別の範囲に使用されています",
	CodePreparationStartUnavailable:              "環境準備を開始できませんでした",
	CodeBriefingRequestInvalid:                   "開始要求が不正です",
	CodeACPNotReady:                              "実験ブリーフを開始する準備ができていません",
	CodeBriefingStartFailed:                      "実験ブリーフを開始できませんでした",
	CodeBriefingStartPending:                     "実験ブリーフの開始を確認中です",
	CodeBriefingNotFound:                         "実験ブリーフが見つかりません",
	CodeBriefingLoadFailed:                       "実験ブリーフを取得できませんでした",
	CodeBriefingMessageFailed:                    "実験ブリーフへメッセージを送信できませんでした",
	CodeBriefingNotActive:                        "実験ブリーフは送信できる状態ではありません",
	CodeBriefingMessagePending:                   "実験ブリーフのメッセージ送信を確認中です",
	CodeBriefingStopFailed:                       "実験ブリーフを終了できませんでした",
	CodeBriefingStopPending:                      "実験ブリーフの終了を確認中です",
	CodeExperimentCreateFailed:                   "実験を作成できませんでした",
	CodeExperimentBriefIncomplete:                "採用する実験ブリーフの条件が不足しています",
	CodeExperimentBriefVersionNotFound:           "採用する実験ブリーフ版が見つかりません",
	CodeExperimentCreateRequestConflict:          "この作成要求は別のブリーフに使用されています",
	CodeExperimentPreparationNotFound:            "実験準備が見つかりません",
	CodeExperimentPreparationNotEditable:         "この実験準備は編集できる状態ではありません",
	CodeExperimentPreparationUnavailable:         "実験準備を取得できませんでした",
	CodeDraftRequestInvalid:                      "下書き保存要求が不正です",
	CodeDraftSaveFailed:                          "下書きを保存できませんでした",
	CodeFixConditionsRequestInvalid:              "条件固定要求が不正です",
	CodeExperimentConditionsAlreadyFixed:         "この実験の条件はすでに固定されています",
	CodeExperimentConditionsConflict:             "固定する条件が現在の下書きと一致しません",
	CodeFixConditionsSaveFailed:                  "条件を固定できませんでした",
	CodeConditionsInvalid:                        "固定する条件が不足しています",
	CodeExperimentWorkspaceRequestInvalid:        "実験ワークスペース取得要求が不正です",
	CodeExperimentWorkspaceNotFound:              "実験ワークスペースが見つかりません",
	CodeExperimentWorkspaceNotReady:              "実験ワークスペースはまだ利用できません",
	CodeExperimentWorkspaceUnavailable:           "実験ワークスペースを取得できませんでした",
	CodeExperimentStartRequestInvalid:            "実験開始要求が不正です",
	CodeExperimentStartNotReady:                  "この実験は開始できる状態ではありません",
	CodeExperimentStartPending:                   "実験の開始を確認中です",
	CodeExperimentStartFailed:                    "実験を開始できませんでした",
	CodeRunEvaluationRequestInvalid:              "run評価の開始要求が不正です",
	CodeRunEvaluationNotReady:                    "このrunは評価できる状態ではありません",
	CodeRunEvaluationAlreadyExists:               "このrunはすでに評価済みです",
	CodeRunEvaluationPending:                     "run評価の開始を確認中です",
	CodeRunEvaluationFailed:                      "run評価を開始できませんでした",
	CodeRunDetailRequestInvalid:                  "run詳細の取得要求が不正です",
	CodeRunDetailNotFound:                        "runが見つかりません",
	CodeRunDetailUnavailable:                     "run詳細を取得できませんでした",
	CodeEvaluationDetailRequestInvalid:           "評価詳細の取得要求が不正です",
	CodeEvaluationDetailNotFound:                 "評価が見つかりません",
	CodeEvaluationDetailUnavailable:              "評価詳細を取得できませんでした",
	CodeRunRetryRequestInvalid:                   "run再実行要求が不正です",
	CodeRunRetryNotFound:                         "再実行するrunが見つかりません",
	CodeRunRetryNotAllowed:                       "失敗したrunだけを再実行できます",
	CodeRunRetryRequestConflict:                  "この再実行要求は別のrunに使用されています",
	CodeRunRetryUnavailable:                      "再実行用runを作成できませんでした",
	CodeExperimentComparisonInvalid:              "実験比較要求が不正です",
	CodeExperimentComparisonNotFound:             "比較する実験が見つかりません",
	CodeExperimentComparisonUnavailable:          "実験比較を取得できませんでした",
	CodeDerivationBriefingStartFailed:            "派生実験の壁打ちを開始できませんでした",
	CodeDerivationBriefingPending:                "派生実験の壁打ち開始を確認中です",
	CodeDerivationBriefingInvalid:                "派生実験の壁打ち取得要求が不正です",
	CodeDerivationBriefingNotFound:               "派生実験の壁打ちが見つかりません",
	CodeDerivationBriefingUnavailable:            "派生実験の壁打ちを取得できませんでした",
	CodeDerivationBriefingMessageInvalid:         "派生実験の壁打ちメッセージが不正です",
	CodeDerivationBriefingMessageNotFound:        "派生実験の壁打ちが見つかりません",
	CodeDerivationBriefingMessageNotActive:       "派生実験の壁打ちは送信できる状態ではありません",
	CodeDerivationBriefingMessageRequestConflict: "この送信要求は別の壁打ちに使用されています",
	CodeDerivationBriefingMessagePending:         "派生実験の壁打ちメッセージ送信を確認中です",
	CodeDerivationBriefingMessageFailed:          "派生実験の壁打ちへメッセージを送信できませんでした",
	CodeDerivationBriefingStopInvalid:            "派生実験の壁打ち終了要求が不正です",
	CodeDerivationBriefingStopNotFound:           "派生実験の壁打ちが見つかりません",
	CodeDerivationBriefingStopNotActive:          "派生実験の壁打ちは終了できる状態ではありません",
	CodeDerivationBriefingStopRequestConflict:    "この終了要求は別の壁打ちに使用されています",
	CodeDerivationBriefingStopPending:            "派生実験の壁打ち終了を確認中です",
	CodeDerivationBriefingStopFailed:             "派生実験の壁打ちを終了できませんでした",
	CodeExperimentConclusionInvalid:              "結論確定要求が不正です",
	CodeExperimentConclusionNotFound:             "結論を確定する実験が見つかりません",
	CodeExperimentConclusionNotReady:             "この実験はまだ結論を確定できません",
	CodeExperimentConclusionRequestConflict:      "この結論確定要求は別の内容に使用されています",
	CodeExperimentConclusionAlreadyFinalized:     "この実験の結論はすでに確定しています",
	CodeExperimentConclusionUnavailable:          "実験の結論を確定できませんでした",
	CodeInsightWorkspaceUnavailable:              "知見ワークスペースを取得できませんでした",
	CodeInsightCreateInvalid:                     "知見作成要求が不正です",
	CodeInsightCreateEvidenceInsufficient:        "知見には異なる実験の根拠が二件以上必要です",
	CodeInsightCreateEvidenceNotFound:            "確定済みの根拠が見つかりません",
	CodeInsightCreateRequestConflict:             "この作成要求は別の知見に使用されています",
	CodeInsightCreateUnavailable:                 "知見を記録できませんでした",
	CodeOperationTimeout:                         "実験一覧の取得がタイムアウトしました",
	CodeOperationCanceled:                        "実験一覧の取得を中止しました",
	CodeUnexpected:                               "予期しないエラーが発生しました",
}
