import {
  AlertCircle,
  ClipboardList,
  FlaskConical,
  Play,
  RefreshCw,
} from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

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
  ExperimentWorkspace,
  GetExperimentWorkspaceService,
} from "./services/get-experiment-workspace-service";
import type {
  StartExperimentService,
  StartedExperiment,
} from "./services/start-experiment-service";

type ExperimentWorkspacePageProps = {
  experimentId: string;
  getExperimentWorkspace: GetExperimentWorkspaceService;
  startExperiment: StartExperimentService;
};

function createRequestId() {
  return crypto.randomUUID();
}

function WorkspaceLoading() {
  return (
    <section
      aria-label="実験ワークスペースを読み込んでいます"
      aria-live="polite"
      className="space-y-3"
    >
      <p className="text-muted-foreground">
        実験ワークスペースを確認しています…
      </p>
      {[1, 2, 3].map((item) => (
        <div className="h-28 animate-pulse rounded-lg bg-muted" key={item} />
      ))}
    </section>
  );
}

function FixedConditions({ workspace }: { workspace: ExperimentWorkspace }) {
  const { fixedConditions } = workspace;
  return (
    <Card id="experiment-workspace-fixed-conditions">
      <CardHeader className="gap-2">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <CardTitle>固定条件</CardTitle>
          <Badge variant="outline">固定済み</Badge>
        </div>
        <p className="text-sm text-muted-foreground">
          固定日時: {formatExperimentDateTime(fixedConditions.fixedAt)}
        </p>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-5 text-sm sm:grid-cols-2">
          <div className="min-w-0 sm:col-span-2">
            <dt className="font-medium">実験目的</dt>
            <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
              {fixedConditions.purpose}
            </dd>
          </div>
          {fixedConditions.hypothesis && (
            <div className="min-w-0 sm:col-span-2">
              <dt className="font-medium">仮説</dt>
              <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
                {fixedConditions.hypothesis}
              </dd>
            </div>
          )}
          <div className="min-w-0">
            <dt className="font-medium">環境条件</dt>
            <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
              {fixedConditions.environmentConditions}
            </dd>
          </div>
          <div className="min-w-0">
            <dt className="font-medium">評価軸</dt>
            <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
              {fixedConditions.evaluationAxes}
            </dd>
          </div>
          <div className="min-w-0 sm:col-span-2">
            <dt className="font-medium">初期入力</dt>
            <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
              {fixedConditions.initialInput}
            </dd>
          </div>
          <div className="min-w-0 sm:col-span-2">
            <dt className="font-medium">prompt</dt>
            <dd className="mt-2 space-y-2">
              {fixedConditions.prompts.map((prompt) => (
                <p
                  className="rounded-md border bg-muted/30 p-3 whitespace-pre-wrap break-words text-muted-foreground"
                  key={prompt.sequenceNo}
                >
                  <span className="mr-2 font-medium text-foreground">
                    {prompt.sequenceNo}.
                  </span>
                  {prompt.content}
                </p>
              ))}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  );
}

function WorkList({
  title,
  items,
  emptyTitle,
  emptyDescription,
}: {
  title: string;
  items: Array<{
    id: string;
    state: string;
    summary?: string;
    updatedAt: string;
  }>;
  emptyTitle: string;
  emptyDescription: string;
}) {
  return (
    <section aria-label={title}>
      <h2 className="mb-3 text-xl font-semibold">{title}</h2>
      {items.length === 0 ? (
        <Empty className="min-h-44" id={`empty-${title}`}>
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <ClipboardList />
            </EmptyMedia>
            <EmptyTitle>{emptyTitle}</EmptyTitle>
            <EmptyDescription>{emptyDescription}</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <ul className="space-y-3">
          {items.map((item) => (
            <li key={item.id}>
              <Card>
                <CardContent className="flex flex-wrap items-center justify-between gap-3 p-4">
                  <div className="min-w-0">
                    <p className="font-medium">{item.id}</p>
                    {item.summary && (
                      <p className="mt-1 break-words text-sm text-muted-foreground">
                        {item.summary}
                      </p>
                    )}
                    <p className="mt-1 text-xs text-muted-foreground">
                      更新: {formatExperimentDateTime(item.updatedAt)}
                    </p>
                  </div>
                  <Badge variant="outline">{item.state}</Badge>
                </CardContent>
              </Card>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export function ExperimentWorkspacePage({
  experimentId,
  getExperimentWorkspace,
  startExperiment,
}: ExperimentWorkspacePageProps) {
  const [workspace, setWorkspace] = useState<ExperimentWorkspace>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const [startError, setStartError] = useState<{
    code: string;
    message: string;
  }>();
  const [startedExperiment, setStartedExperiment] =
    useState<StartedExperiment>();
  const [isStarting, setIsStarting] = useState(false);
  const startRequestId = useRef<string | undefined>(undefined);

  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await getExperimentWorkspace(experimentId);
      if (response.data) {
        setWorkspace(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "実験ワークスペースを取得できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "実験ワークスペースを取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [experimentId, getExperimentWorkspace]);

  const start = useCallback(async () => {
    const requestId = startRequestId.current ?? createRequestId();
    startRequestId.current = requestId;
    setIsStarting(true);
    setStartError(undefined);

    try {
      const response = await startExperiment({ experimentId, requestId });
      if (!response.data) {
        setStartError(
          response.error ?? {
            code: "UNKNOWN",
            message: "実験を開始できませんでした。",
          },
        );
        return;
      }

      setStartedExperiment(response.data);
      await load();
    } catch {
      setStartError({
        code: "UNKNOWN",
        message: "実験を開始できませんでした。",
      });
    } finally {
      setIsStarting(false);
    }
  }, [experimentId, load, startExperiment]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>{experimentId}</span>
              {workspace && <Badge variant="outline">{workspace.state}</Badge>}
            </div>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">
              実験ワークスペース
            </h1>
          </div>
          {workspace && (
            <p className="text-sm text-muted-foreground">
              最終確認: {formatExperimentDateTime(workspace.lastConfirmedAt)}
            </p>
          )}
        </header>

        {isLoading && <WorkspaceLoading />}

        {!isLoading && error && (
          <Alert
            id="experiment-workspace-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>実験ワークスペースを確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-experiment-workspace-button"
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

        {!isLoading && !error && !workspace && (
          <Empty id="empty-experiment-workspace">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <FlaskConical />
              </EmptyMedia>
              <EmptyTitle>実験ワークスペースはまだありません</EmptyTitle>
              <EmptyDescription>
                条件を固定すると、ここで実行と評価の進行状況を確認できます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}

        {!isLoading && !error && workspace && (
          <>
            <Card id="experiment-workspace-start">
              <CardHeader className="gap-2">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <CardTitle>実験を開始</CardTitle>
                  <Badge variant="outline">{workspace.state}</Badge>
                </div>
                <p className="text-sm text-muted-foreground">
                  固定済みの全promptを同じ条件で実行します。
                </p>
              </CardHeader>
              <CardContent className="space-y-4">
                {startError && (
                  <Alert
                    id="experiment-start-error"
                    role="alert"
                    variant="destructive"
                  >
                    <AlertCircle />
                    <AlertTitle>実験を開始できません</AlertTitle>
                    <AlertDescription>{startError.message}</AlertDescription>
                  </Alert>
                )}
                {startedExperiment && (
                  <Alert id="experiment-start-success">
                    <AlertTitle>実験を開始しました</AlertTitle>
                    <AlertDescription>
                      操作ID: {startedExperiment.operationId}（
                      {startedExperiment.runs.length}件のrun）
                    </AlertDescription>
                  </Alert>
                )}
                <Button
                  disabled={isStarting || workspace.state !== "ready"}
                  id="start-experiment-button"
                  onClick={() => void start()}
                  type="button"
                >
                  <Play />
                  {isStarting ? "実験を開始しています…" : "実験を開始"}
                </Button>
                {workspace.state !== "ready" && !startedExperiment && (
                  <p className="text-sm text-muted-foreground">
                    実験開始後の状態です。runの進行状況を確認してください。
                  </p>
                )}
              </CardContent>
            </Card>
            <FixedConditions workspace={workspace} />
            <Card id="experiment-workspace-operation">
              <CardContent className="p-4 text-sm text-muted-foreground">
                条件固定操作ID: {workspace.conditionFixOperation.operationId}
              </CardContent>
            </Card>
            <div className="grid gap-6 lg:grid-cols-2">
              <WorkList
                emptyDescription="実験を開始すると、実行状況がここに表示されます。"
                emptyTitle="runはまだありません"
                items={workspace.runs}
                title="実行"
              />
              <WorkList
                emptyDescription="評価を実行すると、結果がここに表示されます。"
                emptyTitle="evaluationはまだありません"
                items={workspace.evaluations}
                title="評価"
              />
            </div>
          </>
        )}
      </div>
    </main>
  );
}
