import { AlertCircle, CheckCircle2, Lightbulb, RefreshCw } from "lucide-react";
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
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type {
  GetInsightWorkspaceService,
  InsightWorkspace,
} from "./services/get-insight-workspace-service";

function InsightWorkspaceContent({
  workspace,
  initialExperimentId,
}: {
  workspace: InsightWorkspace;
  initialExperimentId?: string;
}) {
  return (
    <div className="space-y-6">
      <section aria-labelledby="insight-evidence-candidates-title">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <h2
            className="text-xl font-semibold"
            id="insight-evidence-candidates-title"
          >
            根拠候補
          </h2>
          <Badge variant="outline">
            {workspace.evidenceCandidates.length}件
          </Badge>
        </div>
        {workspace.evidenceCandidates.length === 0 ? (
          <Empty id="empty-insight-evidence-candidates">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Lightbulb />
              </EmptyMedia>
              <EmptyTitle>根拠候補がありません</EmptyTitle>
              <EmptyDescription>
                比較結果から結論を確定すると、知見の根拠候補として表示されます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid gap-3">
            {workspace.evidenceCandidates.map((candidate) => (
              <Card
                className={
                  candidate.experimentId === initialExperimentId
                    ? "border-primary"
                    : undefined
                }
                id={`insight-evidence-${candidate.experimentId}`}
                key={candidate.conclusionId}
              >
                <CardContent className="space-y-3 pt-6">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge>{candidate.experimentId}</Badge>
                    {candidate.experimentId === initialExperimentId && (
                      <Badge variant="secondary">選択中</Badge>
                    )}
                    <span className="text-sm text-muted-foreground">
                      結論ID: {candidate.conclusionId}
                    </span>
                  </div>
                  <p className="font-medium">{candidate.purpose}</p>
                  <p className="whitespace-pre-wrap break-words text-sm">
                    {candidate.conclusion}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    評価軸: {candidate.evaluationAxes} ・ 確定日時:{" "}
                    {formatExperimentDateTime(candidate.finalizedAt)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>

      <section aria-labelledby="insight-saved-considerations-title">
        <h2
          className="mb-3 text-xl font-semibold"
          id="insight-saved-considerations-title"
        >
          保存済み考察
        </h2>
        {workspace.savedConsiderations.length === 0 ? (
          <Empty id="empty-insight-saved-considerations">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <CheckCircle2 />
              </EmptyMedia>
              <EmptyTitle>保存済み考察がありません</EmptyTitle>
              <EmptyDescription>
                確定した結論がここに表示されます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid gap-3">
            {workspace.savedConsiderations.map((consideration) => (
              <Card key={consideration.conclusionId}>
                <CardContent className="space-y-2 pt-6">
                  <p className="text-sm text-muted-foreground">
                    {consideration.experimentId} / {consideration.conclusionId}
                  </p>
                  <p className="whitespace-pre-wrap break-words text-sm">
                    {consideration.content}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    確定日時:{" "}
                    {formatExperimentDateTime(consideration.finalizedAt)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>

      <section aria-labelledby="insight-list-title">
        <h2 className="mb-3 text-xl font-semibold" id="insight-list-title">
          既存知見
        </h2>
        {workspace.insights.length === 0 ? (
          <Empty id="empty-insight-list">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Lightbulb />
              </EmptyMedia>
              <EmptyTitle>既存知見はありません</EmptyTitle>
              <EmptyDescription>
                知見の記録は次の操作で追加されます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid gap-3">
            {workspace.insights.map((insight) => (
              <Card key={insight.id}>
                <CardContent className="space-y-2 pt-6">
                  <p className="font-medium">{insight.id}</p>
                  <p className="whitespace-pre-wrap break-words text-sm">
                    {insight.statement}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    根拠 {insight.evidenceCount}件 ・{" "}
                    {formatExperimentDateTime(insight.createdAt)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export function InsightWorkspacePage({
  getInsightWorkspace,
  initialExperimentId,
}: {
  getInsightWorkspace: GetInsightWorkspaceService;
  initialExperimentId?: string;
}) {
  const [workspace, setWorkspace] = useState<InsightWorkspace>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await getInsightWorkspace();
      if (response.data) {
        setWorkspace(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "知見のワークスペースを取得できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "知見のワークスペースを取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [getInsightWorkspace]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="flex items-center gap-2 text-3xl font-semibold tracking-tight">
              <Lightbulb className="size-7" /> 知見
            </h1>
            <p className="mt-2 text-muted-foreground">
              確定済みの結論を根拠候補として確認します。
            </p>
          </div>
          {!isLoading && !error && (
            <Button
              id="reload-insight-workspace-button"
              onClick={() => void load()}
              type="button"
              variant="outline"
            >
              <RefreshCw /> 再読込
            </Button>
          )}
        </header>
        {isLoading && (
          <section
            aria-label="知見のワークスペースを読み込んでいます"
            aria-live="polite"
            className="space-y-3"
          >
            <p className="text-muted-foreground">
              知見のワークスペースを確認しています…
            </p>
            <div className="h-52 animate-pulse rounded-lg bg-muted" />
          </section>
        )}
        {!isLoading && error && (
          <Alert
            id="insight-workspace-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>知見のワークスペースを確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="retry-insight-workspace-button"
                onClick={() => void load()}
                type="button"
                variant="outline"
              >
                <RefreshCw /> 再試行
              </Button>
            </AlertDescription>
          </Alert>
        )}
        {!isLoading && workspace && (
          <InsightWorkspaceContent
            initialExperimentId={initialExperimentId}
            workspace={workspace}
          />
        )}
      </div>
    </main>
  );
}
