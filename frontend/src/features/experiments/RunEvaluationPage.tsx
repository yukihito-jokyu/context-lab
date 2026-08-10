import { AlertCircle, RefreshCw, ScanSearch } from "lucide-react";
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
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import { RunEvaluationStartPanel } from "./RunEvaluationStartPanel";
import type {
  GetRunDetailService,
  RunDetail,
} from "./services/get-run-detail-service";
import type {
  StartedRunEvaluation,
  StartRunEvaluationService,
} from "./services/start-run-evaluation-service";

type RunEvaluationPageProps = {
  experimentId: string;
  runId: string;
  title: "run詳細" | "実験比較";
  getRunDetail: GetRunDetailService;
  startRunEvaluation: StartRunEvaluationService;
};

function RunDetailContent({
  detail,
  startRunEvaluation,
}: {
  detail: RunDetail;
  startRunEvaluation: StartRunEvaluationService;
}) {
  const onStarted = (evaluation: StartedRunEvaluation) =>
    window.location.assign(
      `/evaluations/${encodeURIComponent(evaluation.evaluationId)}?operationId=${encodeURIComponent(evaluation.operationId)}`,
    );
  const isReconciling = detail.reconciliation.state === "reconciling";
  const isPartial = detail.artifacts.status !== "complete";
  return (
    <div className="space-y-6">
      {(isReconciling || isPartial) && (
        <Alert id="run-detail-progress" aria-live="polite">
          <ScanSearch />
          <AlertTitle>
            {isReconciling
              ? "実行結果を照合しています"
              : "実行結果を一部取得しました"}
          </AlertTitle>
          <AlertDescription>
            {isReconciling
              ? "保存済みの観測と最新状態を照合しています。"
              : `artifact取得状態: ${detail.artifacts.status}${detail.artifacts.reasonCode ? `（${detail.artifacts.reasonCode}）` : ""}`}
          </AlertDescription>
        </Alert>
      )}
      {detail.failure && (
        <Alert id="run-detail-failure" role="alert" variant="destructive">
          <AlertCircle />
          <AlertTitle>runは完了できませんでした</AlertTitle>
          <AlertDescription className="space-y-1">
            <p>失敗理由: {detail.failure.code}</p>
            {detail.failure.partialSummary && (
              <p>{detail.failure.partialSummary}</p>
            )}
          </AlertDescription>
        </Alert>
      )}
      <section aria-labelledby="run-observation-title">
        <h2 className="mb-3 text-xl font-semibold" id="run-observation-title">
          観測
        </h2>
        <Card id="run-detail-observation">
          <CardContent className="p-5">
            <ol className="space-y-3">
              {detail.observations.length === 0 ? (
                <li className="text-sm text-muted-foreground">
                  まだ観測されていません。
                </li>
              ) : (
                detail.observations.map((observation) => (
                  <li
                    className="rounded-md border p-3"
                    key={observation.sequenceNo}
                  >
                    <div className="flex flex-wrap justify-between gap-2 text-xs text-muted-foreground">
                      <span>{observation.kind}</span>
                      <span>
                        {formatExperimentDateTime(observation.occurredAt)}
                      </span>
                    </div>
                    <p className="mt-2 whitespace-pre-wrap break-words text-sm">
                      {observation.summary}
                    </p>
                  </li>
                ))
              )}
            </ol>
          </CardContent>
        </Card>
      </section>
      <section aria-labelledby="run-diff-title">
        <h2 className="mb-3 text-xl font-semibold" id="run-diff-title">
          差分
        </h2>
        <div className="grid gap-4 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">固定prompt</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="rounded-md border bg-muted/30 p-3 whitespace-pre-wrap break-words text-sm text-muted-foreground">
                {detail.fixedPrompt.content}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">artifact</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2" id="run-detail-artifacts">
              <p className="text-sm text-muted-foreground">
                取得状態: {detail.artifacts.status}
                {detail.artifacts.reasonCode
                  ? `（${detail.artifacts.reasonCode}）`
                  : ""}
              </p>
              {detail.artifacts.items.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  artifactはまだありません。
                </p>
              ) : (
                detail.artifacts.items.map((artifact) => (
                  <div
                    className="rounded-md border p-3 text-sm"
                    key={artifact.digest}
                  >
                    <p className="break-all font-medium">
                      {artifact.label ?? artifact.digest}
                    </p>
                    <p className="mt-1 text-muted-foreground">
                      {artifact.status}
                    </p>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </section>
      <Card id="run-detail-evaluation">
        <CardHeader>
          <CardTitle>評価</CardTitle>
        </CardHeader>
        <CardContent>
          {detail.run.state === "completed" ? (
            <RunEvaluationStartPanel
              onStarted={onStarted}
              runId={detail.run.id}
              runState={detail.run.state}
              startRunEvaluation={startRunEvaluation}
            />
          ) : (
            <p className="text-sm text-muted-foreground">
              runの完了後に評価を開始できます。
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export function RunEvaluationPage({
  experimentId,
  runId,
  title,
  getRunDetail,
  startRunEvaluation,
}: RunEvaluationPageProps) {
  const [detail, setDetail] = useState<RunDetail>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await getRunDetail(runId);
      if (response.data) {
        setDetail(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "run詳細を取得できませんでした。",
        },
      );
    } catch {
      setError({ code: "UNKNOWN", message: "run詳細を取得できませんでした。" });
    } finally {
      setIsLoading(false);
    }
  }, [getRunDetail, runId]);
  useEffect(() => {
    if (!runId) {
      setIsLoading(false);
      return;
    }
    void load();
  }, [load, runId]);
  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm text-muted-foreground">
              実験ID: {experimentId}
            </p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">
              {title}
            </h1>
            {runId && (
              <p className="mt-2 text-sm text-muted-foreground">
                run ID: {runId}
              </p>
            )}
          </div>
          {detail && (
            <div className="flex items-center gap-3">
              <Badge variant="outline">{detail.run.state}</Badge>
              <span className="text-sm text-muted-foreground">
                最終確認: {formatExperimentDateTime(detail.lastConfirmedAt)}
              </span>
            </div>
          )}
        </header>
        {!runId && (
          <Empty id="empty-run-detail">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ScanSearch />
              </EmptyMedia>
              <EmptyTitle>runを指定してください</EmptyTitle>
              <EmptyDescription>
                実験ワークスペースから確認するrunを選択してください。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {runId && isLoading && (
          <section
            aria-label="run詳細を読み込んでいます"
            aria-live="polite"
            className="space-y-3"
          >
            <p className="text-muted-foreground">run詳細を確認しています…</p>
            <div className="h-52 animate-pulse rounded-lg bg-muted" />
          </section>
        )}
        {runId && !isLoading && error && (
          <Alert id="run-detail-error" role="alert" variant="destructive">
            <AlertCircle />
            <AlertTitle>run詳細を確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-run-detail-button"
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
        {runId && !isLoading && !error && !detail && (
          <Empty id="empty-run-detail">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ScanSearch />
              </EmptyMedia>
              <EmptyTitle>run詳細はまだありません</EmptyTitle>
              <EmptyDescription>
                実行を開始すると、観測と差分がここに表示されます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {runId && !isLoading && !error && detail && (
          <RunDetailContent
            detail={detail}
            startRunEvaluation={startRunEvaluation}
          />
        )}
      </div>
    </main>
  );
}
