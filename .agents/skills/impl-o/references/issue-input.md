# Issue入力

## 必須の照合

- GitHub Issueの目的、操作、依存Issue、完了条件
- `docs/tasks/wails-issue-backlog.md` のWI IDとWails関数
- 対応する `AP-*`、`DO-*`、`SCR-*`、HTMLプロトタイプ
- `docs/detailed-design/technology-stack.md` と `system-architecture.md`

## 確定する項目

| 項目 | 内容 |
| --- | --- |
| 関数 | 一つだけの公開Wails関数名 |
| 種別 | queryまたはcommand |
| 入力 | UI入力、識別子、commandのrequest ID |
| 成功 | 画面が表示・遷移・更新するDTO |
| 失敗 | 安全なerror codeと画面表示 |
| 画面 | 操作元と結果を確認するSCR |

未定義の項目を推測で追加せず、実装可否に影響する場合は設計へ戻す。
