# AGENTS.md

## 応答と作業

- 日本語で簡潔かつ丁寧に回答する。
- shellコマンドは`rtk`を先頭に付ける。
- 実装はGitHub IssueのWails関数一件を単位にし、対応SCR画面で実際に確認できる状態まで縦割りで完了する。

## アーキテクチャ

- Wails v2依存は起動部分と`internal/handler/wails`に閉じ込める。
- `internal/domain`はUI・Wails・永続化へ依存しない。
- `internal/usecase`はアプリケーション処理とportを所有する。
- `internal/repository`はSQLite、artifact store、Docker、ACPなど外部I/Oのadapterを置く。
- frontendは生成bindingをservice境界からのみ呼び、`frontend/wailsjs`を編集しない。

## 実行境界

- ACPは実験前の環境準備・壁打ちだけに使う。
- 実験・評価はDockerで隔離して`codex exec`で実行する。
- AIの内部推論、資格情報、Docker ID、sidecar PIDを画面、DTO、成果物へ出さない。

## 検証

- 定型コマンドは`Taskfile.yml`の`task`を優先する。
- Go、frontend、Storybook、Wails E2Eの必要な検証をIssueごとに実行する。
