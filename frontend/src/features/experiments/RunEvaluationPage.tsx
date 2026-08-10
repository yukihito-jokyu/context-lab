import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { RunEvaluationStartPanel } from "./RunEvaluationStartPanel";
import type {
  StartedRunEvaluation,
  StartRunEvaluationService,
} from "./services/start-run-evaluation-service";

type RunEvaluationPageProps = {
  experimentId: string;
  runId: string;
  title: "run詳細" | "実験比較";
  startRunEvaluation: StartRunEvaluationService;
};

// #15/#18の正本詳細取得前に、評価開始とその成功導線を保つ最小到達画面。
export function RunEvaluationPage({
  experimentId,
  runId,
  title,
  startRunEvaluation,
}: RunEvaluationPageProps) {
  const onStarted = (evaluation: StartedRunEvaluation) => {
    window.location.assign(
      `/evaluations/${encodeURIComponent(evaluation.evaluationId)}?operationId=${encodeURIComponent(evaluation.operationId)}`,
    );
  };

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <Card className="mx-auto max-w-2xl" id="run-evaluation-page">
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">
            実験ID: {experimentId}
          </p>
          {runId ? (
            <>
              <p className="text-sm text-muted-foreground">run ID: {runId}</p>
              <RunEvaluationStartPanel
                onStarted={onStarted}
                runId={runId}
                runState="completed"
                startRunEvaluation={startRunEvaluation}
              />
            </>
          ) : (
            <p className="text-sm text-muted-foreground">
              評価を開始するrunを指定してください。
            </p>
          )}
        </CardContent>
      </Card>
    </main>
  );
}
