---
name: storybook-impl
description: Context LabのReact/shadcn画面をStorybookで状態カタログとして実装・検証する。Wails bindingの成功・空・失敗・処理中をfake serviceで表現するときに使用する。
---

# Context Lab Storybook 実装

## 目的

画面仕様（SCR）とHTMLプロトタイプにある状態を、Wails実行環境なしで確認できる状態カタログにする。Storybookは表示・操作・アクセシビリティ確認に使い、実Docker、ACP、SQLite、Wails bindingの実呼出しは行わない。

## 手順

1. 該当Issue、SCR、HTML、Wails binding一覧を読む。
2. ページから表示領域を分離し、Propsまたは画面用view modelで状態を注入できるようにする。
3. 通常、空、読み込み中、操作中、入力エラー、再試行可能な失敗をstoryとして作る。
4. fake Wails serviceでDTOを返し、画面が公開しない情報（内部推論、資格情報、Docker ID、sidecar PID）を表示しないことを確認する。
5. `task frontend:storybook:build`を実行する。画面操作がある場合はplay関数またはE2Eで補う。

## 状態の選び方

| 画面 | 最低限の状態 |
| --- | --- |
| 実験一覧 | 一覧あり、空、取得失敗 |
| 実験準備 | 入力途中、検証エラー、環境候補取得中、候補なし |
| 実験実行・評価 | 実行中、成功、失敗、取消済み |
| 比較・知見 | 比較対象なし、比較結果あり、保存失敗 |
| 環境準備 | ACP壁打ち中、提案あり、接続失敗 |

## 完了条件

- 画面仕様の主要状態がstoryで再現できる。
- StorybookがWails backendや実験コンテナなしで起動・ビルドできる。
- UI操作は対応するWails bindingのユースケースと矛盾しない。
