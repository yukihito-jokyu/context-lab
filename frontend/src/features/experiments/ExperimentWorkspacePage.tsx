import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type ExperimentWorkspacePageProps = {
  experimentId: string;
  operationId?: string;
};

export function ExperimentWorkspacePage({
  experimentId,
  operationId,
}: ExperimentWorkspacePageProps) {
  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-4xl space-y-6">
        <header>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>{experimentId}</span>
            <Badge variant="outline">ready</Badge>
          </div>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight">
            実験ワークスペース
          </h1>
        </header>
        <Card id="experiment-workspace-ready">
          <CardHeader>
            <CardTitle>条件を固定しました</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>固定済みの条件を使って、実験の実行と評価を続けられます。</p>
            {operationId && <p>操作ID: {operationId}</p>}
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
