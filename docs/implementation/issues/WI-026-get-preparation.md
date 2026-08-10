# WI-026 GetPreparation

## 対象

- GitHub Issue: #29
- Wails 関数: `PreparationsHandler.GetPreparation(preparationID string)`
- 画面: SCR-008 環境準備詳細

## 契約

- 空白のみのIDは `PREPARATION_REQUEST_INVALID` を返す。
- `kind = environment_preparation` のsessionだけを取得し、存在しない場合は `PREPARATION_NOT_FOUND` を返す。
- repository由来の失敗は `PREPARATION_UNAVAILABLE`、想定外の失敗は安全な `UNEXPECTED` へ変換する。
- 応答には安全な候補、診断、失敗コード、再照合状態だけを含め、資格情報・内部推論・実行識別子は含めない。

## 永続化

- `environment_preparation_operations` は準備操作の状態と安全な失敗コードを記録する。
- `environment_preparation_candidates` と `environment_preparation_diagnostics` はsessionに紐付く安全な表示情報を記録する。
- 実際の書込みは #30 `StartPreparation` が正本として担当する。

## 検証

- usecase、SQLite、Wails handlerの単体テストで入力検証、種別分離、時刻・候補・診断・失敗情報、DTO安全化を確認する。
