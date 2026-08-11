import { AlertCircle, ClipboardList, RefreshCw } from "lucide-react";
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
import { ExperimentBriefingDialog } from "./components/ExperimentBriefingDialog";
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type { CreateExperimentFromBriefService } from "./services/create-experiment-from-brief-service";
import type { GetExperimentBriefingService } from "./services/get-experiment-briefing-service";
import type { ListExperimentsService } from "./services/list-experiments-service";
import type { SendExperimentBriefMessageService } from "./services/send-experiment-brief-message-service";
import type { StartExperimentBriefingService } from "./services/start-experiment-briefing-service";
import type { StopExperimentBriefingService } from "./services/stop-experiment-briefing-service";

type ListData = Awaited<ReturnType<ListExperimentsService>>["data"];

type ExperimentListPageProps = {
  listExperiments: ListExperimentsService;
  startExperimentBriefing: StartExperimentBriefingService;
  getExperimentBriefing: GetExperimentBriefingService;
  createExperimentFromBrief: CreateExperimentFromBriefService;
  sendExperimentBriefMessage: SendExperimentBriefMessageService;
  stopExperimentBriefing: StopExperimentBriefingService;
  onOpenExperiment: (experimentId: string, state: string) => void;
};

const stateBadgeVariant = (state: string) => {
  if (state === "cancelled") return "secondary" as const;
  if (state === "failed") return "destructive" as const;
  return "outline" as const;
};

function ExperimentCard({
  experiment,
  onOpenExperiment,
}: {
  experiment: NonNullable<ListData>["experiments"][number];
  onOpenExperiment: (experimentId: string, state: string) => void;
}) {
  const isPreparing = experiment.state === "preparing";

  return (
    <li>
      <Card>
        <CardContent className="space-y-3 p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-sm text-muted-foreground">{experiment.id}</p>
              <h3 className="mt-1 break-words font-semibold">
                {experiment.purpose}
              </h3>
            </div>
            <Badge variant={stateBadgeVariant(experiment.state)}>
              {experiment.state}
            </Badge>
          </div>
          <p className="break-words text-sm text-muted-foreground">
            {experiment.progressSummary}
          </p>
          <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
            <span>
              最終更新: {formatExperimentDateTime(experiment.updatedAt)}
            </span>
            {experiment.derivedFromExperimentId && (
              <span>派生元: {experiment.derivedFromExperimentId}</span>
            )}
          </div>
          <Button
            id={`open-experiment-${experiment.id}`}
            onClick={() => onOpenExperiment(experiment.id, experiment.state)}
            type="button"
            variant="outline"
          >
            {isPreparing ? "実験準備を開く" : "実験を開く"}
          </Button>
        </CardContent>
      </Card>
    </li>
  );
}

export function ExperimentListPage({
  listExperiments,
  startExperimentBriefing,
  getExperimentBriefing,
  createExperimentFromBrief,
  sendExperimentBriefMessage,
  stopExperimentBriefing,
  onOpenExperiment,
}: ExperimentListPageProps) {
  const [data, setData] = useState<ListData>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const [isBriefingDialogOpen, setIsBriefingDialogOpen] = useState(false);

  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await listExperiments();
      if (response.data) {
        setData(response.data);
        return;
      }
      if (response.error) {
        setError(response.error);
        return;
      }
      setError({
        code: "UNKNOWN",
        message: "実験一覧を取得できませんでした。",
      });
    } catch {
      setError({
        code: "UNKNOWN",
        message: "実験一覧を取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [listExperiments]);

  useEffect(() => {
    void load();
  }, [load]); // 初期表示と再読込は同じqueryを使う。

  const hasExperiments = Boolean(
    data &&
      (data.experiments.length > 0 || data.cancelledExperiments.length > 0),
  );

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-primary">Context Lab</p>
            <h1 className="mt-1 text-3xl font-semibold tracking-tight">実験</h1>
            <p className="mt-2 text-muted-foreground">
              目的と進行状況を確認し、次の作業を再開します。
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <p className="text-sm text-muted-foreground">
              最終確認: {formatExperimentDateTime(data?.lastConfirmedAt)}
            </p>
            <Button
              id="new-experiment-button"
              onClick={() => setIsBriefingDialogOpen(true)}
              type="button"
            >
              新規実験
            </Button>
          </div>
        </header>

        {isLoading && (
          <section
            aria-label="実験一覧を読み込んでいます"
            className="space-y-3"
            aria-live="polite"
          >
            <p className="text-muted-foreground">実験一覧を確認しています…</p>
            {[1, 2, 3].map((item) => (
              <div
                className="h-28 animate-pulse rounded-lg bg-muted"
                key={item}
              />
            ))}
          </section>
        )}

        {!isLoading && error && (
          <Alert role="alert" variant="destructive">
            <AlertCircle />
            <AlertTitle>実験一覧を確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <p>
                最後に確認できた時刻:{" "}
                {formatExperimentDateTime(data?.lastConfirmedAt)}
              </p>
              <Button
                id="reload-experiment-list-button"
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

        {!isLoading && !error && data && !hasExperiments && (
          <Empty id="empty-state">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ClipboardList />
              </EmptyMedia>
              <EmptyTitle>実験はまだありません</EmptyTitle>
              <EmptyDescription>
                実験を作成すると、ここに進行状況が表示されます。
              </EmptyDescription>
              <Button
                id="empty-create-experiment-button"
                onClick={() => setIsBriefingDialogOpen(true)}
                type="button"
              >
                新規実験
              </Button>
            </EmptyHeader>
          </Empty>
        )}

        {!isLoading && !error && data && hasExperiments && (
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
            <section aria-labelledby="experiment-list-title">
              <div className="mb-3 flex items-center justify-between gap-3">
                <h2
                  className="text-xl font-semibold"
                  id="experiment-list-title"
                >
                  実験一覧
                </h2>
                <span className="text-sm text-muted-foreground">
                  {data.experiments.length}件
                </span>
              </div>
              <ul className="space-y-3">
                {data.experiments.map((experiment) => (
                  <ExperimentCard
                    experiment={experiment}
                    key={experiment.id}
                    onOpenExperiment={onOpenExperiment}
                  />
                ))}
              </ul>
            </section>
            <aside className="space-y-4" aria-label="作業の再開要約">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">作業を再開</CardTitle>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground">
                  {data.resumeSummary.recommendedExperimentId
                    ? `推奨: ${data.resumeSummary.recommendedExperimentId}`
                    : "再開する実験はありません。"}
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">進行状況</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="space-y-2 text-sm">
                    {Object.entries(data.resumeSummary.statusCounts).map(
                      ([state, count]) => (
                        <div className="flex justify-between gap-3" key={state}>
                          <dt>{state}</dt>
                          <dd>{count}件</dd>
                        </div>
                      ),
                    )}
                  </dl>
                </CardContent>
              </Card>
            </aside>
          </div>
        )}

        {!isLoading &&
          !error &&
          data &&
          data.cancelledExperiments.length > 0 && (
            <details className="rounded-lg border bg-card p-5">
              <summary className="cursor-pointer font-medium">
                取消済みの実験 {data.cancelledExperiments.length}件
              </summary>
              <ul className="mt-4 space-y-3">
                {data.cancelledExperiments.map((experiment) => (
                  <ExperimentCard
                    experiment={experiment}
                    key={experiment.id}
                    onOpenExperiment={onOpenExperiment}
                  />
                ))}
              </ul>
            </details>
          )}
      </div>
      <ExperimentBriefingDialog
        getExperimentBriefing={getExperimentBriefing}
        createExperimentFromBrief={createExperimentFromBrief}
        onOpenChange={setIsBriefingDialogOpen}
        open={isBriefingDialogOpen}
        sendExperimentBriefMessage={sendExperimentBriefMessage}
        startExperimentBriefing={startExperimentBriefing}
        stopExperimentBriefing={stopExperimentBriefing}
      />
    </main>
  );
}
