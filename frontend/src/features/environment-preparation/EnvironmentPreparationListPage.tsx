import { AlertCircle, ClipboardList, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { formatExperimentDateTime } from "../experiments/lib/format-experiment-date-time";
import type {
  ListPreparationsResponse,
  ListPreparationsService,
  PreparationListItem,
} from "./services/list-preparations-service";

type EnvironmentPreparationListPageProps = {
  listPreparations: ListPreparationsService;
};

const stateBadgeVariant = (state: string) => {
  if (state === "failed") return "destructive" as const;
  if (state === "cancelled") return "secondary" as const;
  return "outline" as const;
};

function PreparationSessionCard({
  preparation,
}: {
  preparation: PreparationListItem;
}) {
  const detailPath = `/preparations/${encodeURIComponent(preparation.preparationId)}`;

  return (
    <li>
      <Card>
        <CardContent className="space-y-3 p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <h2 className="break-words font-semibold">
              {preparation.preparationId}
            </h2>
            <Badge variant={stateBadgeVariant(preparation.state)}>
              {preparation.state}
            </Badge>
          </div>
          <dl className="grid gap-x-5 gap-y-1 text-sm text-muted-foreground sm:grid-cols-2">
            <div className="flex min-w-0 justify-between gap-3 sm:block">
              <dt>開始</dt>
              <dd>{formatExperimentDateTime(preparation.startedAt)}</dd>
            </div>
            <div className="flex min-w-0 justify-between gap-3 sm:col-span-2 sm:block">
              <dt>最終観測</dt>
              <dd>{formatExperimentDateTime(preparation.lastObservedAt)}</dd>
            </div>
          </dl>
          <Button asChild size="sm" variant="outline">
            <a href={detailPath}>詳細を確認</a>
          </Button>
        </CardContent>
      </Card>
    </li>
  );
}

export function EnvironmentPreparationListPage({
  listPreparations,
}: EnvironmentPreparationListPageProps) {
  const [data, setData] = useState<ListPreparationsResponse["data"]>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<ListPreparationsResponse["error"]>();

  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await listPreparations();
      if (response.data) {
        setData(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "準備session一覧を取得できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "準備session一覧を取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [listPreparations]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-primary">Context Lab</p>
            <h1 className="mt-1 text-3xl font-semibold tracking-tight">
              環境準備
            </h1>
            <p className="mt-2 text-muted-foreground">
              実験を作成する前の準備sessionを確認します。
            </p>
          </div>
        </header>

        <Alert>
          <AlertDescription>
            表示するのは安全に要約された状態と観測です。credential、sidecarの生通信、内部推論は表示しません。
          </AlertDescription>
        </Alert>

        {isLoading && (
          <section aria-live="polite" id="preparation-loading">
            <p className="mb-3 text-muted-foreground">
              準備session一覧を読み込んでいます…
            </p>
            <div className="space-y-3">
              {[1, 2, 3].map((item) => (
                <div
                  className="h-40 animate-pulse rounded-lg bg-muted"
                  key={item}
                />
              ))}
            </div>
          </section>
        )}

        {!isLoading && error && (
          <Alert
            id="preparation-query-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>準備session一覧を確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-preparation-list-button"
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

        {!isLoading && !error && data?.preparations.length === 0 && (
          <Empty id="preparation-list-empty">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ClipboardList />
              </EmptyMedia>
              <EmptyTitle>準備sessionはまだありません</EmptyTitle>
              <EmptyDescription>
                環境準備を開始すると、ここに状態と最終観測が表示されます。
              </EmptyDescription>
            </EmptyHeader>
            <Button
              id="reload-preparation-list-button"
              onClick={() => void load()}
              type="button"
              variant="outline"
            >
              <RefreshCw />
              再読込
            </Button>
          </Empty>
        )}

        {!isLoading && !error && data && data.preparations.length > 0 && (
          <section aria-labelledby="preparation-list-title">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-xl font-semibold" id="preparation-list-title">
                準備session一覧
              </h2>
              <Button
                id="reload-preparation-list-button"
                onClick={() => void load()}
                type="button"
                variant="outline"
              >
                <RefreshCw />
                一覧を再読込
              </Button>
            </div>
            <ul className="space-y-3" id="session-list">
              {data.preparations.map((preparation) => (
                <PreparationSessionCard
                  key={preparation.preparationId}
                  preparation={preparation}
                />
              ))}
            </ul>
          </section>
        )}
      </div>
    </main>
  );
}
