import { AlertCircle, ArrowLeft, ClipboardList, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
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
import { formatExperimentDateTime } from "../experiments/lib/format-experiment-date-time";
import { saveAdoptedEnvironmentConditions } from "./adopted-environment-conditions";
import type { AdoptCandidateService } from "./services/adopt-candidate-service";
import type {
  GetPreparationResponse,
  GetPreparationService,
  PreparationDetail,
} from "./services/get-preparation-service";

type EnvironmentPreparationDetailPageProps = {
  preparationId: string;
  getPreparation: GetPreparationService;
  adoptCandidate: AdoptCandidateService;
  onBackToList: () => void;
  onBeginExperiment?: () => void;
};

const stateBadgeVariant = (state: string) => {
  if (state === "failed") return "destructive" as const;
  if (state === "cancelled") return "secondary" as const;
  return "outline" as const;
};

function SessionOverview({ data }: { data: PreparationDetail }) {
  return (
    <Card>
      <CardHeader className="gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <CardTitle className="break-words">{data.preparationId}</CardTitle>
          <p className="mt-1 text-sm text-muted-foreground">
            準備sessionの状態
          </p>
        </div>
        <Badge variant={stateBadgeVariant(data.state)}>{data.state}</Badge>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-x-6 gap-y-3 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-muted-foreground">開始</dt>
            <dd>{formatExperimentDateTime(data.startedAt)}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">最終観測</dt>
            <dd>{formatExperimentDateTime(data.lastObservedAt)}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">照合状態</dt>
            <dd>{data.reconciliation.state}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">照合の最終観測</dt>
            <dd>
              {formatExperimentDateTime(data.reconciliation.lastObservedAt)}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  );
}

function Candidates({
  data,
  adoptCandidate,
  onBeginExperiment,
}: {
  data: PreparationDetail;
  adoptCandidate: AdoptCandidateService;
  onBeginExperiment?: () => void;
}) {
  const [candidateId, setCandidateId] = useState<string>();
  const [requestId, setRequestId] = useState<string>();
  const [isAdopting, setIsAdopting] = useState(false);
  const [adoptError, setAdoptError] = useState<{
    code: string;
    message: string;
  }>();
  const selectedCandidate = data.candidates.find(
    (candidate) => candidate.id === candidateId,
  );
  const canAdopt =
    data.state === "completed" &&
    data.reconciliation.state !== "reconciling" &&
    !data.failure;

  const beginAdoption = (nextCandidateId: string) => {
    setCandidateId(nextCandidateId);
    setRequestId(crypto.randomUUID());
    setAdoptError(undefined);
  };

  const adopt = async () => {
    if (!selectedCandidate || !requestId) return;

    setIsAdopting(true);
    setAdoptError(undefined);
    try {
      const response = await adoptCandidate({
        requestId,
        preparationId: data.preparationId,
        candidateId: selectedCandidate.id,
      });
      if (response.data) {
        saveAdoptedEnvironmentConditions(response.data.environmentConditions);
        onBeginExperiment?.();
        return;
      }
      setAdoptError(
        response.error ?? {
          code: "CANDIDATE_ADOPTION_FAILED",
          message: "環境候補を採用できませんでした。",
        },
      );
    } catch {
      setAdoptError({
        code: "CANDIDATE_ADOPTION_FAILED",
        message: "環境候補を採用できませんでした。",
      });
    } finally {
      setIsAdopting(false);
    }
  };

  if (data.candidates.length === 0) {
    return (
      <Empty id="preparation-candidates-empty">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ClipboardList />
          </EmptyMedia>
          <EmptyTitle>環境候補はまだありません</EmptyTitle>
          <EmptyDescription>
            候補が観測されると、ここに安全な要約が表示されます。
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <section aria-labelledby="preparation-candidates-title">
      <h2
        className="mb-3 text-xl font-semibold"
        id="preparation-candidates-title"
      >
        環境候補
      </h2>
      <ul className="space-y-3">
        {data.candidates.map((candidate) => (
          <li key={candidate.id}>
            <Card>
              <CardContent className="space-y-3 p-5">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <h3 className="break-words font-semibold">{candidate.id}</h3>
                  <span className="text-sm text-muted-foreground">
                    {formatExperimentDateTime(candidate.createdAt)}
                  </span>
                </div>
                <dl className="grid gap-3 text-sm">
                  <div>
                    <dt className="text-muted-foreground">環境条件</dt>
                    <dd className="whitespace-pre-wrap break-words">
                      {candidate.environmentConditions}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">要約</dt>
                    <dd className="whitespace-pre-wrap break-words">
                      {candidate.summary}
                    </dd>
                  </div>
                </dl>
                <Button
                  disabled={!canAdopt}
                  id={`adopt-candidate-button-${candidate.id}`}
                  onClick={() => beginAdoption(candidate.id)}
                  type="button"
                >
                  この候補を採用して実験を開始
                </Button>
              </CardContent>
            </Card>
          </li>
        ))}
      </ul>
      {!canAdopt && (
        <p className="mt-3 text-sm text-muted-foreground">
          準備が完了し、照合が確認されると候補を採用できます。
        </p>
      )}
      <AlertDialog
        onOpenChange={(open) => {
          if (!open && !isAdopting) setCandidateId(undefined);
        }}
        open={Boolean(selectedCandidate)}
      >
        <AlertDialogContent id="adopt-candidate-dialog">
          <AlertDialogHeader>
            <AlertDialogTitle>環境候補を採用しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              採用した環境条件を、新しい実験の準備画面へ引き継ぎます。既存の候補や実験は変更しません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {selectedCandidate && (
            <p className="whitespace-pre-wrap break-words rounded-md bg-muted p-3 text-sm">
              {selectedCandidate.environmentConditions}
            </p>
          )}
          {isAdopting && (
            <p id="adopt-candidate-pending" role="status">
              環境候補を採用しています…
            </p>
          )}
          {adoptError && (
            <Alert
              id="adopt-candidate-error"
              role="alert"
              variant="destructive"
            >
              <AlertCircle />
              <AlertTitle>環境候補を採用できません</AlertTitle>
              <AlertDescription>{adoptError.message}</AlertDescription>
            </Alert>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isAdopting}>戻る</AlertDialogCancel>
            <AlertDialogAction
              disabled={isAdopting}
              onClick={(event) => {
                event.preventDefault();
                void adopt();
              }}
            >
              {isAdopting ? "採用中…" : "採用して新規実験へ"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </section>
  );
}

function Diagnostics({ data }: { data: PreparationDetail }) {
  if (data.diagnostics.length === 0) return null;

  return (
    <section aria-labelledby="preparation-diagnostics-title">
      <h2
        className="mb-3 text-xl font-semibold"
        id="preparation-diagnostics-title"
      >
        診断
      </h2>
      <ul className="space-y-3">
        {data.diagnostics.map((diagnostic) => (
          <li key={diagnostic.id}>
            <Card>
              <CardContent className="space-y-2 p-5">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h3 className="break-words font-semibold">
                      {diagnostic.code}
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      {diagnostic.id}
                    </p>
                  </div>
                  <span className="text-sm text-muted-foreground">
                    {formatExperimentDateTime(diagnostic.occurredAt)}
                  </span>
                </div>
                <p className="whitespace-pre-wrap break-words text-sm">
                  {diagnostic.summary}
                </p>
              </CardContent>
            </Card>
          </li>
        ))}
      </ul>
    </section>
  );
}

export function EnvironmentPreparationDetailPage({
  preparationId,
  getPreparation,
  adoptCandidate,
  onBackToList,
  onBeginExperiment = () => window.location.assign("/"),
}: EnvironmentPreparationDetailPageProps) {
  const [data, setData] = useState<PreparationDetail>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<GetPreparationResponse["error"]>();

  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await getPreparation(preparationId);
      if (response.data) {
        setData(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "準備sessionを取得できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "準備sessionを取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [getPreparation, preparationId]);

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
              環境準備の詳細
            </h1>
          </div>
          <Button onClick={onBackToList} type="button" variant="outline">
            <ArrowLeft /> 一覧へ戻る
          </Button>
        </header>

        {isLoading && (
          <section aria-live="polite" id="preparation-detail-loading">
            <p className="mb-3 text-muted-foreground">
              準備sessionを読み込んでいます…
            </p>
            <div className="h-56 animate-pulse rounded-lg bg-muted" />
          </section>
        )}

        {!isLoading && error && (
          <Alert
            id="preparation-detail-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>準備sessionを確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-preparation-detail-button"
                onClick={() => void load()}
                type="button"
                variant="outline"
              >
                <RefreshCw /> 再読込
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {!isLoading && !error && data && (
          <>
            <SessionOverview data={data} />
            {data.reconciliation.state === "reconciling" && (
              <Alert id="preparation-reconciling" aria-live="polite">
                <AlertTitle>照合中です</AlertTitle>
                <AlertDescription>
                  最終観測:{" "}
                  {formatExperimentDateTime(data.reconciliation.lastObservedAt)}
                </AlertDescription>
              </Alert>
            )}
            {data.failure && (
              <Alert
                id="preparation-failure"
                role="alert"
                variant="destructive"
              >
                <AlertCircle />
                <AlertTitle>準備sessionで失敗が観測されました</AlertTitle>
                <AlertDescription>
                  <p>コード: {data.failure.code}</p>
                  <p>
                    発生: {formatExperimentDateTime(data.failure.occurredAt)}
                  </p>
                </AlertDescription>
              </Alert>
            )}
            <Candidates
              adoptCandidate={adoptCandidate}
              data={data}
              onBeginExperiment={onBeginExperiment}
            />
            <Diagnostics data={data} />
            <Button
              id="reload-preparation-detail-button"
              onClick={() => void load()}
              type="button"
              variant="outline"
            >
              <RefreshCw /> 再読込
            </Button>
          </>
        )}
      </div>
    </main>
  );
}
