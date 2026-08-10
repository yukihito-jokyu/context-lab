# WI-027 StartPreparation

## 対象

- GitHub Issue: #30
- Wails関数: `PreparationsHandler.StartPreparation(requestID string, scope string)`
- 画面: SCR-008 環境準備一覧

## 契約

- `requestID`が空白のみの場合は`PREPARATION_START_REQUEST_INVALID`を返す。
- `scope`は作業root配下の相対directoryだけを受理し、実体解決後のcanonicalなroot相対値で排他・再送を判定する。
- 同一requestは同一結果を再生し、別scopeなら`PREPARATION_START_REQUEST_CONFLICT`、同scope実行中なら`PREPARATION_START_PENDING`を返す。
- 成功時は安全な`preparationId`と`completed`状態だけを返す。ACPの生応答、資格情報、内部推論、sidecar識別子は返さない。

## 実装

- SQLite transactionでsessionとoperationを開始し、候補・診断と完了状態を同一transactionで記録する。
- ACP adapterは範囲を実行直前にも再検証し、one-shot sessionを閉じる。
- 失敗時は安全なエラーコードだけをoperationとsessionへ記録する。

## 検証

- Go: `go test ./...`、usecase・SQLite・ACP・Wails handlerの対象関数 raw/effective coverage 100%。
- Frontend: `task frontend:check`、`task frontend:build`、`task frontend:storybook:build`、`task frontend:e2e`（55件）。
- 静的検査: `task test:style`、`git diff --check`。

## レビュー

- 独立レビューでscopeのcanonical化、ACP出力の安全フィルタ、通信不確実性と終端失敗を分けたrequest ID再試行を確認・修正した。
- 未解決のP1/P2はない。
