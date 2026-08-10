import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type EvaluationOperationPageProps = {
  evaluationId: string;
  operationId?: string;
};

// 評価詳細・比較が未実装の間も、評価開始後に識別子を確認できる到達画面を提供する。
export function EvaluationOperationPage({
  evaluationId,
  operationId,
}: EvaluationOperationPageProps) {
  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <Card className="mx-auto max-w-2xl" id="evaluation-operation-page">
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <CardTitle>評価を開始しました</CardTitle>
            <Badge variant="outline">evaluating</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>評価ID: {evaluationId}</p>
          {operationId && <p>操作ID: {operationId}</p>}
          <p>評価の進行状況と比較は、この評価IDを使って確認できます。</p>
        </CardContent>
      </Card>
    </main>
  );
}
