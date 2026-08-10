---
document_type: impl-o
issue: "#19"
work_item: WI-019
status: completed
---

# WI-019 FinalizeExperimentConclusion

## 実装した契約

`FinalizeExperimentConclusionsHandler.FinalizeExperimentConclusion` は、request ID・実験 ID・結論を受け取り、評価済みの実験に不変の結論を確定する。成功時は結論 ID、結論本文、確定時刻を返し、同じ request ID は保存済み snapshot を再生する。比較 query は確定済み結論を任意項目として再表示する。

## 変更範囲

- SQLite の結論・操作 snapshot 正本と migration
- domain、usecase、Wails handler、生成 binding
- SCR-005 比較画面の結論確定フォームと確定済み表示
- Storybook と Playwright E2E

## 画面での確認操作

比較画面で評価根拠を確認し結論を入力して確定する。確定中は入力と操作を無効化し、失敗時は同じ request ID で再試行できる。成功後と再読込後は、正本から取得した確定済み結論、派生・知見への導線を表示する。

## 実行した検証

- Go 全体テスト、style、変更関数の関数別 coverage
- frontend check、build、Storybook build
- Playwright E2E
- SQLite 2 pool による同一 request ID の並行確定

## 未実施検証と理由

なし。

## 設計との差分

なし。
