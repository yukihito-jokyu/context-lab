---
name: impl-o
description: Wails関数1件を単位とするGitHub Issueの実装を、設計確認、Go/React実装、独立レビュー、修正、再検証へ分担・統合するオーケストレーションスキル。サブエージェントを用いたContext Labの実装・レビュー・完了判定に使用する。
---

# 実装オーケストレーション

一つのGitHub Issueで公開するWails関数は一つだけに固定する。オーケストレーターは成果物の統合責任を持ち、担当者へ実装または読み取り専用レビューを委譲する。委譲できない実行環境では同じゲートを順番に自身で実行する。

## 開始ゲート

1. Issue、依存Issue、`docs/tasks/wails-issue-backlog.md`、対応するAP/DO/SCR/HTMLを読む。
2. 関数名、command/query、入力・成功・失敗DTO、対象画面操作、受け入れ基準を一枚の実装契約として固定する。
3. 設計の不足・画面操作との矛盾が実装判断を変える場合は、実装を止めて設計へ戻す。
4. 各担当の所有範囲と成果物を決める。詳細は[delegation-protocol.md](references/delegation-protocol.md)を読む。

## 委譲と統合

以下を必要な範囲だけ委譲する。複数担当が編集する場合は、領域またはworktreeを分離する。同じファイルを並行編集させない。

| 担当 | 主な責務 | 書込み範囲 |
| --- | --- | --- |
| 設計確認 | Issueと設計の整合、DTO・状態・境界の確認 | 原則読み取り専用 |
| Go実装 | handler、application、domain、repository/adapter、テスト | `internal/`、Goテスト |
| React実装 | service、画面、shadcn状態、画面テスト | `frontend/src/` |
| 検証 | テスト選定と実行、失敗再現 | 原則読み取り専用。必要時のみテスト |
| 独立レビュー | 実装・画面・テスト・安全境界の検査 | 読み取り専用 |

Go実装とReact実装は、所有パスが分離される場合に同時に開始する。共有DTO・生成bindingはオーケストレーターだけが統合し、必要な生成はGo側の公開契約確定後に実行する。担当結果を統合した後、バックエンドとフロントエンドの独立レビューを同時に開始する。レビュー後はオーケストレーターが指摘を分類し、該当担当へ修正を委譲して再検証する。[review-gates.md](references/review-gates.md)を読む。

## 常に守る境界

- Wails bindingはIssue対象の一つだけを公開し、handlerで安全なDTOへ変換する。
- React画面はfrontend service経由だけでbindingを呼ぶ。生成済み`frontend/wailsjs`は手編集しない。
- 実験・評価はDocker + `codex exec` adapter、実験前の環境準備と壁打ちはACP adapterだけを使う。
- Docker、SQLite、ACPを画面から直接呼ばない。内部推論、資格情報、Docker ID、sidecar PIDをDTO・画面・ログ・成果物へ出さない。
- queryは読込中・空・失敗・再読込、commandは入力検証・実行中・失敗・成功を対応SCR画面で確認可能にする。

## 完了ゲート

1. Go・React実装を統合し、画面操作からWails関数まで接続する。
2. 成功と代表的失敗を、必要なunit/integration/component/Storybook/E2Eで検証する。
3. 独立レビューの必須指摘をすべて解消し、修正対象を再検証する。
4. `docs/implementation/issues/WI-XXX-<slug>.md`へ契約、委譲結果、レビュー結果、検証、未実施項目を記録する。
5. GitHub Issueへのコメント・closeは利用者が依頼した場合だけ行う。

関連する実装規約は`wails-clean-architecture`、`frontend-implementation`、`backend-integration-test`、`storybook-impl`、`wails-e2e-test`を使う。
