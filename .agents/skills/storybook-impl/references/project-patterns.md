# Context Lab Storybook パターン

- screen componentはPropsで表示用DTOと操作コールバックを受け取る。
- page componentだけがWails serviceを選択する。
- storyではWails serviceをfakeへ差し替え、実Docker・ACP・SQLiteへ接続しない。
- Story名は「一覧あり」「候補なし」「実行中」「再試行可能な失敗」のように利用者が見る状態を表す。
