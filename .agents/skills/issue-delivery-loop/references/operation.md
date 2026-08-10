# Issue Delivery Loop 運用手順

## 目次

1. 状態モデル
2. プリフライト
3. Issue選定
4. ブランチと再開
5. 実装からPR作成
6. PRゲートとマージ
7. 失敗処理
8. 完了報告

## 状態モデル

一件のIssueを次の順で進める。

`discovered -> eligible -> branched -> implemented -> reviewed -> pushed -> pr-open -> checks-passed -> merged -> synced`

状態を飛ばさない。再開時はGitとGitHubを読み取り、最後に確認できた状態から続ける。ローカルの推測だけで`merged`や`closed`と判断しない。

## プリフライト

最初に以下を確認する。

```bash
rtk git status --porcelain=v1 --branch
rtk git remote -v
rtk gh auth status
rtk git fetch origin main
rtk gh pr status
```

- remoteは期待するcontext-labリポジトリか確認する。
- detached HEADでは開始しない。
- cleanな`main`なら新規選定へ進む。
- `feature/<番号>`上ならIssue、差分、open PRを照合し、同じ仕事だと確認できる場合だけ再開する。
- 無関係な未コミット変更があれば停止する。自動stashや破棄はしない。
- GitHub APIが使えなければ、ローカル実装だけへ暗黙に縮小せず停止する。AIは`gh auth login`を実行または利用者へ依頼せず、Issue・PR取得などのGitHub API呼び出しで利用可否を判定する。

## Issue選定

候補を都度取得する。

```bash
rtk gh issue list --state open --limit 200 --json number,title,body,labels,url
```

候補は次をすべて満たすものに限定する。

1. `docs/tasks/wails-issue-backlog.md`のWIと一意に対応する。
2. Issueが公開Wails関数一件を対象にしている。
3. バックログとIssue本文に記載された依存Issueがclosedである。
4. 対応する設計とSCRが`impl-o`の開始ゲートを満たす。
5. すでにmerged PRまたはcompleted実装記録で完了していない。
6. `duplicate`、`invalid`、`wontfix`、`question`ではない。

`type:*`、`status:*`、`area:*`、`priority:*`が存在する場合は`github-issue-labeling`に従い、`status:todo`だけを新規候補にする。ただし現在のリポジトリに軸ラベルがなければ、ラベルを自動作成せずバックログと依存関係で判定する。

優先順位はバックログの「実装順」、依存グラフ、Issue番号の順とする。同順位を恣意的に選ばない。候補がない場合は、単に「全Issue完了」とせず、残るOpen Issueを対象外・依存待ち・仕様待ちへ分類する。

## ブランチと再開

新規作業では以下を基本とする。

```bash
rtk git switch main
rtk git pull --ff-only origin main
rtk git switch -c feature/<issue番号>
```

- `feature/<issue番号>`は既存履歴から確認できる規約である。規約が変更されていれば最新履歴を優先する。
- 同名branchがlocalまたはoriginにあれば新規作成しない。対応IssueとPRを確認してcheckoutまたはtrackする。
- open PRがある場合は`rtk gh pr view`でhead、base、state、checksを確認して再開する。
- 公開済みbranchをrebaseしない。`main`追従が必要なら`origin/main`を通常mergeし、再検証してpushする。
- force-push、branch削除、reset、checkoutによる変更破棄を行わない。

## 実装からPR作成

1. `impl-o`へIssue番号、WI、Wails関数、対象SCRを渡す。
2. `impl-o`の完了ゲートと独立レビューをすべて満たす。
3. `git diff`、`git diff --cached`、`git status`でIssue外変更がないことを確認する。
4. 機密情報、Docker ID、sidecar PID、内部推論、不要なartifactを検査する。
5. 履歴のコミット形式に合わせてcommitする。既存コミットは書き換えない。
6. `rtk git push -u origin feature/<issue番号>`でpushする。
7. `.github/pull_request_template.md`を使い、チケット欄へ`close #<issue番号>`を記載してPRを作る。

PR本文には最低限、概要、実装したWails関数、画面確認、検証、未実施項目を記載する。未実施の必須検証がある場合はPRを作成しても自動マージしない。

## PRゲートとマージ

次をGitHubから確認する。

```bash
rtk gh pr checks <PR番号> --watch
rtk gh pr view <PR番号> --json state,isDraft,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,headRefName,baseRefName,url
```

すべて満たす場合だけmergeする。

- `state`がOPEN、`isDraft`がfalse、baseが`main`である。
- headが対象Issueのbranchである。
- 必須checkが完了して成功または許容されたskipである。
- `mergeable`がMERGEABLEで、競合がない。
- `reviewDecision`がCHANGES_REQUESTEDではない。
- branch protectionが要求する承認を満たす。
- `impl-o`の必須指摘が0件で、PR作成後の差分もレビュー済みである。
- Issueが依然openで、内容が実装契約から重大変更されていない。

既存履歴に合わせてmerge commitを使う。

```bash
rtk gh pr merge <PR番号> --merge
rtk gh pr view <PR番号> --json state,mergedAt,mergeCommit
rtk gh issue view <issue番号> --json state,url
rtk git switch main
rtk git pull --ff-only origin main
```

`close #<番号>`でIssueが閉じなかった場合は、PRのmergeと受け入れ基準達成を再確認する。利用者が本スキルへIssue完了までを明示委任している場合に限り、merge済みPRを示してIssueをcloseする。

## 失敗処理

- ローカル検証またはCI失敗はログを取得し、Issue範囲内の原因だけを修正する。
- 修正後は該当レビューと検証を再実行し、追加commitとしてpushする。
- 同一原因が3回続いたらブロックとして記録し、無限再試行しない。
- merge conflictは内容を読んでIssue範囲内で安全に解決できる場合だけ通常mergeで解消する。設計判断や他Issueの変更を伴う場合は停止する。
- GitHub権限、必須の第三者承認、外部サービス、資格情報が必要なら承認を偽装・迂回しない。
- ブロックしたIssueに依存しない候補は続行できる。依存する候補は依存待ちへ分類する。

## 完了報告

各Issueについて次を一行で記録する。

| Issue | WI / Wails関数 | Branch | PR | 結果 | 検証 | Merge commit |
| --- | --- | --- | --- | --- | --- | --- |

最後に以下を報告する。

- merged件数
- 実装可能候補が0件であること
- 依存待ちIssue
- 仕様・承認待ちIssue
- 対象外Open Issue
- ローカルの現在branchと`main`同期状態
