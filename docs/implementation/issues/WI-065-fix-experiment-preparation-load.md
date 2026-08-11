# WI-065 FixExperimentPreparationLoad

## 対象

- GitHub Issue: #65
- Wails関数: `ExperimentPreparationsHandler.GetExperimentPreparation(experimentID string)`
- 画面: SCR-002 実験準備

## 原因

- ブリーフ採用後のSQLite保存データは`GetExperimentPreparation`の読込条件を満たしていた。
- 画面はWails queryと、候補採用の一時引継ぎに使う`sessionStorage`読込を同じ`try`で処理していた。
- WebViewでStorage APIが例外を返すと、成功したqueryまで「実験準備を取得できませんでした。」として表示していた。

## 実装

- 一時引継ぎのStorage API失敗は値なしとして扱い、準備画面の表示を継続する。
- 準備queryの失敗だけを取得失敗として表示する例外境界に分離する。
- ブリーフ採用で作成した実験を同じSQLite Storeから直後に読み戻す結合テストを追加する。

## 検証

- Go: `go test ./...`、`go test ./internal/repository/sqlite -run '^TestStoreCreateExperimentFromBrief$' -count=1`
- Frontend: `task frontend:check`、`task frontend:build`、`task frontend:storybook:build`
- E2E: ブリーフ採用後に準備フォームの実表示を確認し、Storage APIが`SecurityError`でも画面表示できることを確認する。
- 静的検査: `task test:style`、`git diff --check`

## レビュー

- 既存E2Eは作成後のURL遷移だけを確認し、準備queryのmock・画面本文・Storage API失敗を検証していなかった。
- 修正後は作成から準備フォームの読込までと、Storage API例外時の表示継続を回帰対象にした。
