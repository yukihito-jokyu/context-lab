# 検証選択

| 変更 | 必須確認 |
| --- | --- |
| domain/application | Go unit test |
| SQLite・artifact・adapter | Go integration test |
| Wails DTO/失敗変換 | handler test |
| React状態・入力検証 | component testまたはStorybook |
| 重要な画面操作 | Wails E2E |
| Docker/codex exec/ACP | fake adapterまたは隔離環境で成功・失敗境界を確認 |

commandは同じrequest IDで重複作成しないことを確認する。queryは読込失敗と再読込を確認する。外部依存を使えない場合は、実行しなかった理由と残るリスクを記録する。
