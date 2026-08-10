# WI-021 派生実験を作成する

## 公開契約

`CreateDerivedExperimentsHandler.CreateDerivedExperiment` は、確定済みの派生元実験と利用者が指定した差分・理由から、新しい `preparing` 実験を作成する。

- 入力は request ID、派生元実験 ID、少なくとも一つの実質差分、理由である。
- 同一 request ID と同一 canonical payload は同じ作成結果を返す。
- request ID の派生元または payload が異なる場合は安全な競合エラーを返す。
- 派生元の固定条件と finalized の結論を transaction 内で再確認する。

## 永続化と安全性

- 子の編集用下書きは派生元の mutable な下書きではなく、固定条件と固定 prompt を正本として複写する。
- 差分と理由だけを新規実験へ保存し、派生元の条件・結論・結果は変更しない。
- `experiment_derived_operations` に request ID と canonical payload を保存し、再送を収束させる。
- handler は安全な code/message だけを返し、内部エラーや実行環境情報を公開しない。

## 画面

SCR-006 では差分と理由を入力し、送信中は編集と二重送信を無効化する。失敗後は同じ request ID を再利用し、成功後は新しい実験の準備画面へ遷移する。
