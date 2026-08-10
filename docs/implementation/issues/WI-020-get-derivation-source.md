---
document_type: impl-o
issue: "#20"
work_item: WI-020
status: completed
---

# WI-020 GetDerivationSource

## 実装した契約

`ExperimentDerivationSourcesHandler.GetDerivationSource` は実験 ID を受け取り、派生元の固定条件、確定済み結論、派生実験を作成できるかを返す query である。固定条件または確定済み結論が不足する場合は成功応答で可否理由を返し、未確定の結論本文は返さない。

## 変更範囲

- SQLite 正本からの派生元、固定条件、順序付き prompt、確定済み結論の取得
- domain、usecase、Wails handler、生成 binding
- SCR-006 の派生元表示、読込・失敗・再読込状態
- Storybook と Playwright E2E

## 画面での確認操作

比較画面から派生元画面を開き、固定条件、結論、可否理由を確認する。派生作成と壁打ちは後続 Issue で提供するため、可否にかかわらず無効化し、利用可能時期を表示する。取得失敗時は再読込できる。

## 実行した検証

- Go 全体テスト、style、変更関数の関数別 coverage
- golangci-lint
- frontend check、build、Storybook build
- Playwright E2E

## 未実施検証と理由

なし。

## 設計との差分

なし。
