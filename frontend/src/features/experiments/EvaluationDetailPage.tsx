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
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type {
  EvaluationDetail,
  GetEvaluationDetailService,
} from "./services/get-evaluation-detail-service";

type EvaluationDetailPageProps = {
  evaluationId: string;
  operationId?: string;
  getEvaluationDetail: GetEvaluationDetailService;
};

function EvaluationDetailContent({ detail }: { detail: EvaluationDetail }) {
  const isReconciling = ["starting", "reconciling"].includes(
    detail.reconciliation.state,
  );
  const isPartial = detail.result.status !== "complete";
  const reasonCode = detail.failure?.code || detail.result.reasonCode;

  return (
    <div className="space-y-6">
      {(isReconciling || isPartial) && (
        <Alert id="evaluation-detail-progress" aria-live="polite">
          <Scale />
          <AlertTitle>
            {isReconciling
              ? "評価結果を照合しています"
              : "評価結果を一部取得しました"}
          </AlertTitle>
          <AlertDescription>
            {isReconciling
              ? "保存済みの評価結果と最新状態を照合しています。"
              : "評価結果が確定するまで、取得済みの情報を表示しています。"}
          </AlertDescription>
        </Alert>
      )}
      {reasonCode && (
        <Alert
          id="evaluation-detail-unavailable"
          role="alert"
          variant="destructive"
        >
          <AlertCircle />
          <AlertTitle>評価結果を確定できません</AlertTitle>
          <AlertDescription>理由: {reasonCode}</AlertDescription>
        </Alert>
      )}
      <section aria-labelledby="evaluation-evidence-title">
        <h2
          className="mb-3 text-xl font-semibold"
          id="evaluation-evidence-title"
        >
          評価の根拠
        </h2>
        <div className="grid gap-4 lg:grid-cols-2">
          <Card id="evaluation-detail-evidence">
            <CardHeader>
              <CardTitle className="text-base">runの要約</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="whitespace-pre-wrap break-words text-sm text-muted-foreground">
                {detail.evidence.runSummary ||
                  "記録されたrunの要約はありません。"}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">評価軸</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="whitespace-pre-wrap break-words text-sm text-muted-foreground">
                {detail.evidence.evaluationAxes ||
                  "評価軸は記録されていません。"}
              </p>
            </CardContent>
          </Card>
        </div>
      </section>
      <section aria-labelledby="evaluation-result-title">
        <h2 className="mb-3 text-xl font-semibold" id="evaluation-result-title">
          評価結果
        </h2>
        <Card id="evaluation-detail-result">
          <CardContent className="space-y-2 p-5">
            <p className="text-sm text-muted-foreground">
              結果状態: {detail.result.status}
            </p>
            {detail.result.summary ? (
              <p className="whitespace-pre-wrap break-words text-sm">
                {detail.result.summary}
              </p>
            ) : (
              <p className="text-sm text-muted-foreground">
                評価結果はまだ記録されていません。
              </p>
            )}
          </CardContent>
        </Card>
      </section>
    </div>
  );
}

export function EvaluationDetailPage({
  evaluationId,
  operationId,
  getEvaluationDetail,
}: EvaluationDetailPageProps) {
  const [detail, setDetail] = useState<EvaluationDetail>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await getEvaluationDetail(evaluationId);
      if (response.data) {
        setDetail(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "評価詳細を取得できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "評価詳細を取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [evaluationId, getEvaluationDetail]);

  useEffect(() => {
    if (!evaluationId) {
      setIsLoading(false);
      return;
    }
    void load();
  }, [evaluationId, load]);

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm text-muted-foreground">
              評価ID: {evaluationId}
            </p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">
              評価詳細
            </h1>
            {operationId && (
              <p className="mt-2 text-sm text-muted-foreground">
                操作ID: {operationId}
              </p>
            )}
          </div>
          {detail && (
            <div className="flex items-center gap-3">
              <Badge variant="outline">{detail.evaluation.state}</Badge>
              <span className="text-sm text-muted-foreground">
                最終確認: {formatExperimentDateTime(detail.lastConfirmedAt)}
              </span>
            </div>
          )}
        </header>
        {!evaluationId && (
          <Empty id="empty-evaluation-detail">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Scale />
              </EmptyMedia>
              <EmptyTitle>評価を指定してください</EmptyTitle>
              <EmptyDescription>
                評価開始後に表示される評価IDから詳細を確認してください。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {evaluationId && isLoading && (
          <section
            aria-label="評価詳細を読み込んでいます"
            aria-live="polite"
            className="space-y-3"
          >
            <p className="text-muted-foreground">評価詳細を確認しています…</p>
            <div className="h-52 animate-pulse rounded-lg bg-muted" />
          </section>
        )}
        {evaluationId && !isLoading && error && (
          <Alert
            id="evaluation-detail-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>評価詳細を確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-evaluation-detail-button"
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
        {evaluationId && !isLoading && !error && !detail && (
          <Empty id="empty-evaluation-detail">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Scale />
              </EmptyMedia>
              <EmptyTitle>評価詳細はまだありません</EmptyTitle>
              <EmptyDescription>
                評価が開始されると、根拠と評価結果がここに表示されます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {evaluationId && !isLoading && !error && detail && (
          <EvaluationDetailContent detail={detail} />
        )}
      </div>
    </main>
  );
}
