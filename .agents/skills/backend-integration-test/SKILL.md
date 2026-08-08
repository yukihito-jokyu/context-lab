---
name: backend-integration-test
description: Context LabのGo結合テストを設計・実装する。SQLite永続化、artifact保存、Docker上のcodex exec実験、ACP環境準備の境界を安全かつ決定的に検証するときに使用する。
---

# Context Lab バックエンド結合テスト

## 前提

- 一時ディレクトリ配下のSQLiteとartifact保存先をテストごとに作成し、終了時に片付ける。
- 実験・評価の実行境界はDocker + `codex exec` adapterである。通常の結合テストではfake adapterを使う。
- ACPは実験前の環境準備とAI壁打ち専用である。通常はfake ACP adapterを使い、実接続は明示承認された隔離試験だけで扱う。
- 外部RDBMS、DSN、開発者の資格情報、実Docker・実ACPへの依存は持ち込まない。

## 手順

1. Issue、受け入れ基準、処理設計（AP）、詳細設計（DO）、画面仕様（SCR）を読む。
2. handlerを通さず、UseCase → repository / adapterを結合する最小のシナリオを決める。
3. temp SQLite、artifact root、clock、ID生成、Docker/ACP fakeを注入するfixtureを用意する。
4. 正常系、状態遷移、永続化後の再読込、失敗時の安全なError DTOを検証する。
5. Docker/ACPの呼出しは入力DTO・起動要求・停止/失敗のマッピングまでをfakeで検証し、秘密情報・内部推論・Docker ID・sidecar PIDを記録しない。
6. `task test`、必要に応じて`task test:integration`を実行する。

## 観点

| 境界 | 確認すること |
| --- | --- |
| SQLite repository | 保存・再読込・制約違反時のドメインエラー変換 |
| artifact store | 実験記録の保存先分離、欠損時の安全な扱い |
| Docker adapter | `codex exec`の開始要求、終了結果、失敗・取消の変換 |
| ACP adapter | 環境候補取得・壁打ち結果の変換。実験実行へ使わない |
| UseCase | 状態遷移、入力検証、表示可能な失敗理由 |

## 完了条件

- テストはネットワーク、ローカルDocker、個人設定に依存せず再現可能である。
- テスト名は利用者の結果またはドメイン状態を表す。
- Issueが必要とする永続化・実行・失敗の境界を網羅する。
- `task test`が成功する。
