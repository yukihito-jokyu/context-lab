# GitHub #69 Wails bridge障害の識別と回復案内

## 契約

- 対象Wails query: `ExperimentPreparationsHandler.GetExperimentPreparation(experimentID string)`。
- Wails bindingを呼ぶfrontend serviceは、Go handlerの安全なエラー応答とは別に、次のhandler到達前後の状態を安全なエラーDTOへ変換する。
  - `WAILS_BRIDGE_UNAVAILABLE`: `window.go` または `window.go.wails.ExperimentPreparationsHandler` が未注入。
  - `WAILS_HANDLER_UNAVAILABLE`: handlerはあるが `GetExperimentPreparation` が未登録。
  - `WAILS_BRIDGE_CALL_FAILED`: binding呼出しが拒否された。
- 例外本文、資格情報、Docker ID、sidecar PID、内部推論は画面へ出さない。
- bridge障害時はDB初期化を案内せず、アプリまたは開発サーバーを完全終了して起動し直す回復手順を表示する。
- WebViewへのbridge注入が画面初期化より遅れる場合は、最大1秒間待機してから未注入と判定する。待機中は既存の読込状態を維持する。

## 開発時の版確認

- `task app:bridge:versions` はWails CLIとGo moduleの版を表示し、不一致なら失敗する。
- `task dev` はこの確認を完了してから `wails dev` を起動する。Go handlerまたは生成bindingを更新した後は、Vite HMRだけに頼らずこのプロセスを完全停止して再起動する。

## 自動検証の範囲

- Playwright E2Eは通常、`window.go` をmockする。このためnative WebViewにWails bridgeが注入されることそのものは検証しない。
- 本Issueでは、bridge未注入、handler未登録、Promise拒否を画面E2EとStorybookで確認する。
- native bridgeの注入・再起動回復は下記の人間確認で検証する。

## 人間確認ゲート（マージ必須）

1. PRのビルドを使い、既存のContext Labアプリと`wails dev`を完全終了する。
2. `task app:bridge:versions`が成功することを確認してから、`task dev`またはビルド済みアプリを起動する。
3. 実験ブリーフを採用して実験を作成し、実験準備画面で編集フォームが表示されることを確認する。
4. 開発サーバーを停止した状態またはbridge未注入fixtureで、原因分類と「DB消去不要・完全再起動」の案内が表示されることを確認する。
5. 担当者以外の人間が、ビルド識別子、確認日時、結果をGitHub Issue #69へコメントする。

このコメントが確認できるまでPRをマージしない。

## AI確認

- frontend実装と生成binding境界を独立レビューする。
- E2E、Biome、Vite build、Storybook build、差分検査を実行する。
