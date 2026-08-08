# 縦割り実装

| 層 | 責務 |
| --- | --- |
| `internal/handler/wails` | Request/Response、失敗DTO、Wails公開関数 |
| application/usecase | 操作手順、トランザクション、port呼出し |
| domain | 状態・不変条件・ドメイン操作 |
| repository/adapter | SQLite、artifact store、Docker、ACPなどの外部I/O |
| frontend service | 生成bindingを呼ぶ薄い境界 |
| React/shadcn | 表示状態、入力、確認、結果反映 |

Docker + `codex exec` の実験実行はapplicationからadapterへ依頼する。ACPは実験前の準備・壁打ちでだけ使い、実験・評価の実行経路へ混在させない。
