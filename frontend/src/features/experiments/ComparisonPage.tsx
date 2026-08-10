import { AlertCircle, RefreshCw, Scale } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { FinalizeExperimentConclusionPanel } from "./FinalizeExperimentConclusionPanel";
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type { FinalizeExperimentConclusionService } from "./services/finalize-experiment-conclusion-service";
import type {
  ExperimentComparison,
  GetExperimentComparisonService,
} from "./services/get-experiment-comparison-service";

export function ComparisonPage({
  experimentId,
  finalizeExperimentConclusion,
  getExperimentComparison,
}: {
  experimentId: string;
  finalizeExperimentConclusion: FinalizeExperimentConclusionService;
  getExperimentComparison: GetExperimentComparisonService;
}) {
  const [comparison, setComparison] = useState<ExperimentComparison>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const response = await getExperimentComparison(experimentId);
      if (response.data) {
        setComparison(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "比較を取得できませんでした。",
        },
      );
    } catch {
      setError({ code: "UNKNOWN", message: "比較を取得できませんでした。" });
    } finally {
      setLoading(false);
    }
  }, [experimentId, getExperimentComparison]);
  useEffect(() => {
    void load();
  }, [load]);
  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header>
          <p className="text-sm text-muted-foreground">
            実験ID: {experimentId}
          </p>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight">
            実験比較
          </h1>
        </header>
        {loading && (
          <section aria-label="比較を読み込んでいます" aria-live="polite">
            <p className="text-muted-foreground">比較を確認しています…</p>
            <div className="mt-3 h-52 animate-pulse rounded-lg bg-muted" />
          </section>
        )}
        {!loading && error && (
          <Alert id="comparison-error" role="alert" variant="destructive">
            <AlertCircle />
            <AlertTitle>比較を確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-comparison-button"
                onClick={() => void load()}
                type="button"
                variant="outline"
              >
                <RefreshCw />
                再読込
              </Button>
            </AlertDescription>
          </Alert>
        )}
        {!loading && !error && comparison && (
          <>
            <Card>
              <CardHeader>
                <CardTitle>{comparison.experiment.purpose}</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                評価軸: {comparison.experiment.evaluationAxes} / 最終確認:{" "}
                {formatExperimentDateTime(comparison.lastConfirmedAt)}
              </CardContent>
            </Card>
            {comparison.evaluations.length === 0 ? (
              <Empty id="empty-comparison">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <Scale />
                  </EmptyMedia>
                  <EmptyTitle>比較できる評価はまだありません</EmptyTitle>
                  <EmptyDescription>
                    評価が完了すると比較結果を確認できます。
                  </EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : (
              <section aria-label="評価比較" className="space-y-3">
                {comparison.evaluations.map((evaluation) => (
                  <Card key={evaluation.evaluationId}>
                    <CardContent className="space-y-3 p-5">
                      <div className="flex flex-wrap items-center justify-between gap-3">
                        <p className="font-medium">{evaluation.evaluationId}</p>
                        <Badge variant="outline">{evaluation.state}</Badge>
                      </div>
                      {evaluation.reconciliation.state === "reconciling" && (
                        <Alert>
                          <AlertTitle>評価結果を照合しています</AlertTitle>
                        </Alert>
                      )}
                      <p className="whitespace-pre-wrap break-words text-sm">
                        {evaluation.result.summary ??
                          evaluation.result.reasonCode ??
                          "評価結果はまだ記録されていません。"}
                      </p>
                      <p className="whitespace-pre-wrap break-words text-sm text-muted-foreground">
                        根拠（実行要約）: {evaluation.runSummary ?? "未記録"}
                      </p>
                      <div className="flex flex-wrap gap-2">
                        <Button
                          onClick={() =>
                            window.location.assign(
                              `/experiments/${encodeURIComponent(experimentId)}/runs/${encodeURIComponent(evaluation.runId)}`,
                            )
                          }
                          type="button"
                          variant="outline"
                        >
                          run詳細
                        </Button>
                        <Button
                          onClick={() =>
                            window.location.assign(
                              `/evaluations/${encodeURIComponent(evaluation.evaluationId)}`,
                            )
                          }
                          type="button"
                          variant="outline"
                        >
                          評価詳細
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </section>
            )}
            {(comparison.evaluations.length > 0 || comparison.conclusion) && (
              <FinalizeExperimentConclusionPanel
                existingConclusion={comparison.conclusion}
                experimentId={experimentId}
                finalizeExperimentConclusion={finalizeExperimentConclusion}
                onReload={load}
              />
            )}
          </>
        )}
      </div>
    </main>
  );
}
