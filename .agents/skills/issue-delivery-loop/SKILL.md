---
name: issue-delivery-loop
description: context-labの実装可能なWails関数単位GitHub Issueを依存順に選び、main同期、Issueブランチ作成、impl-oによる実装・独立レビュー・検証、commit・push、PR作成、CI確認、merge、main再同期を一件ずつ反復する。利用者がIssueからPRマージまでをAIへ委任したい、実装Issueがなくなるまで自律的に進めたい、または途中のIssueブランチ・PRから反復処理を再開したい場合に使用する。
---

# Issue Delivery Loop

`impl-o`を一件のIssueを完了させる実装エンジンとして使い、その外側のGit・GitHubライフサイクルを管理する。常に一件ずつ処理し、検証済み変更だけを`main`へ統合する。

長時間の反復では、終了条件を明示して起動する。

```text
/goal $issue-delivery-loopを使い、実装可能なWails Issueが0件になるまで一件ずつ実装・検証・PR・マージする。ブロックしたIssueは記録し、依存しない候補を続行する。
```

## 開始時に読むもの

1. [operation.md](references/operation.md)を最後まで読む。
2. `impl-o`、`github-issue-labeling`と、`impl-o`が指定する関連スキルを読む。
3. `AGENTS.md`、`docs/tasks/wails-issue-backlog.md`、`.github/pull_request_template.md`、CI workflowを読む。

## 実行契約

- 利用者がPR作成とマージを含む反復実行を明示した場合だけ、push、PR作成、Issue close、mergeを行う。
- Open Issueすべてではなく、Wailsバックログに対応し、依存が解決した実装可能Issueを対象にする。
- バックログの実装順を優先し、同順位ではIssue番号の小さいものを選ぶ。
- ラベルは選定の補助情報として読む。ラベルの新設・変更は利用者が明示依頼した場合だけ行う。
- 一つの作業ツリーで一件だけ実装する。別Issueの変更を同じブランチやPRへ混ぜない。
- 既存の未コミット変更、ブランチ、PRを勝手に破棄、stash、reset、rebase、force-pushしない。
- branch protection、必須レビュー、CI、承認フローを迂回しない。

## 反復フロー

### 1. プリフライトまたは再開

Git状態、remote、GitHub認証、現在ブランチ、未コミット変更、既存PRを読み取る。安全に再開できるIssueブランチまたはPRがあれば重複作成せず続行する。無関係な変更や曖昧な状態があれば停止する。

### 2. Issue選定

Open Issueとバックログを再取得し、依存Issue、対応WI、設計状態を照合する。実装可能候補がなければ反復を終了し、未解決・対象外・ブロック中Issueを分けて報告する。

### 3. main同期とブランチ作成

作業ツリーが安全な場合だけ`main`へ切り替え、`origin/main`をfast-forwardで取得する。既存規約に従い、既定では`feature/<issue番号>`を作成する。既存の同名ブランチやPRがあれば状態を確認して再開する。

### 4. impl-o実行

対象Issue一件だけを`impl-o`へ渡す。設計契約の固定、Go/React実装、生成binding統合、画面確認、独立レビュー、修正、再検証、実装記録まで完了させる。必須指摘または必須検証の未解決があればPRをマージしない。

### 5. commit・push・PR

差分がIssueの契約内であること、機密情報や生成物の誤混入がないことを確認する。既存のコミット規約に従ってcommitし、forceなしでpushする。PRテンプレートを使い、本文に`close #<issue番号>`を含める。既存PRがあれば新規作成せず更新する。

### 6. 最終ゲートとmerge

PR差分に対する独立レビュー結果、ローカル検証、必須CI、merge可能状態、レビュー決定を再確認する。すべて成功した場合だけリポジトリ既存方式のmerge commitでマージする。`CHANGES_REQUESTED`、必須承認待ち、競合、失敗・保留中CIがあればマージしない。

### 7. main再同期と次Issue

マージとIssue closeを確認後、`main`へ戻って`origin/main`へfast-forwardする。作業ツリーがcleanであることを確認し、Open Issueを再取得して次の一件へ進む。

## 停止条件

次の場合は安全な情報収集まで行い、勝手に回避せず停止または当該Issueをブロックとして記録する。

- Issue・設計・SCR・Wails関数の対応が一意に決まらない。
- 依存Issueが未完了、または外部状態・資格情報・人間の製品判断が必要。
- 未コミット変更が対象Issueのものか判定できない。
- 同一原因の実装・検証・CI失敗が3回続く。
- GitHub APIへの接続または必要な権限が利用できない。AIは`gh auth login`を実行または利用者へ依頼せず、利用可能なGitHub API呼び出しで状態を確認する。
- merge conflict、branch protection、権限不足を安全な通常操作で解消できない。
- データ破壊、公開API互換性、認証・機密情報に関する判断がIssueで承認されていない。

ブロックしたIssueだけを無限再試行しない。他に依存しない実装可能Issueがあれば進め、最後に完了・対象外・ブロックを一覧化する。

## 完了報告

Issueごとにブランチ、PR、merge commit、実行した検証、Issue状態を報告する。最後に残存Open Issueを「実装可能なし」「依存待ち」「仕様・承認待ち」「対象外」に分類し、対象候補が0件になったことを終了条件とする。
