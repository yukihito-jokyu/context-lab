import { AlertCircle, FlaskConical, RefreshCw, Sparkles } from "lucide-react";
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
import { DerivationBriefingDialog } from "./components/DerivationBriefingDialog";
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type {
  DerivationSource,
  GetDerivationSourceService,
} from "./services/get-derivation-source-service";
import type { SendDerivationBriefMessageService } from "./services/send-derivation-brief-message-service";
import type { StartDerivationBriefingService } from "./services/start-derivation-briefing-service";

type DerivationBriefingServices = {
  startDerivationBriefing: StartDerivationBriefingService;
  sendDerivationBriefMessage: SendDerivationBriefMessageService;
};

function DerivationSourceContent({
  source,
  briefingServices,
}: {
  source: DerivationSource;
  briefingServices?: DerivationBriefingServices;
}) {
  const fixedConditions = source.source.fixedConditions;
  const conclusion = source.source.conclusion;
  const eligibility = source.eligibility;
  const reasonLabels = {
    CONDITIONS_NOT_FIXED: "固定条件が未確定です",
    CONCLUSION_NOT_FINALIZED: "結論が未確定です",
  };
  const reason = eligibility.reasonCode
    ? `派生を開始できない理由: ${reasonLabels[eligibility.reasonCode]}（${eligibility.reasonCode}）`
    : undefined;
  const [isBriefingOpen, setIsBriefingOpen] = useState(false);

  return (
    <div className="space-y-6">
      <p className="text-muted-foreground">
        元の実験目的: {source.source.purpose}
      </p>
      <Card id="derivation-source-eligibility">
        <CardHeader className="gap-2">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <CardTitle>派生の可否</CardTitle>
            <Badge
              variant={
                eligibility.canCreateDerivedExperiment ? "default" : "outline"
              }
            >
              {eligibility.canCreateDerivedExperiment ? "派生可能" : "派生不可"}
            </Badge>
          </div>
          {reason && <p className="text-sm text-muted-foreground">{reason}</p>}
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-2">
            {eligibility.canCreateDerivedExperiment ? (
              <Button asChild type="button">
                <a
                  href={`/experiments/${encodeURIComponent(source.source.experimentId)}/derivations/create`}
                >
                  <Sparkles /> 派生実験を作成
                </a>
              </Button>
            ) : (
              <Button disabled type="button">
                <Sparkles /> 派生実験を作成
              </Button>
            )}
            <Button
              disabled={
                !eligibility.canCreateDerivedExperiment || !briefingServices
              }
              id="start-derivation-briefing-button"
              onClick={() => setIsBriefingOpen(true)}
              type="button"
              variant="outline"
            >
              <FlaskConical /> 壁打ちを開始
            </Button>
          </div>
          {eligibility.canCreateDerivedExperiment && (
            <p className="text-sm text-muted-foreground">
              派生元の条件と結論をもとに、相談を開始できます。
            </p>
          )}
        </CardContent>
      </Card>
      {briefingServices && (
        <DerivationBriefingDialog
          onOpenChange={setIsBriefingOpen}
          open={isBriefingOpen}
          sourceExperimentId={source.source.experimentId}
          startDerivationBriefing={briefingServices.startDerivationBriefing}
          sendDerivationBriefMessage={
            briefingServices.sendDerivationBriefMessage
          }
        />
      )}

      <section aria-labelledby="derivation-fixed-conditions-title">
        <h2
          className="mb-3 text-xl font-semibold"
          id="derivation-fixed-conditions-title"
        >
          固定条件
        </h2>
        {fixedConditions ? (
          <Card id="derivation-source-fixed-conditions">
            <CardContent className="pt-6">
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
                <div>
                  <dt className="font-medium">固定日時</dt>
                  <dd className="mt-1 text-muted-foreground">
                    {formatExperimentDateTime(fixedConditions.fixedAt)}
                  </dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        ) : (
          <Empty id="empty-derivation-fixed-conditions">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <FlaskConical />
              </EmptyMedia>
              <EmptyTitle>固定条件がありません</EmptyTitle>
              <EmptyDescription>
                条件を固定すると、派生の作成元として利用できます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </section>

      <section aria-labelledby="derivation-conclusion-title">
        <h2
          className="mb-3 text-xl font-semibold"
          id="derivation-conclusion-title"
        >
          確定した結論
        </h2>
        {conclusion ? (
          <Card id="derivation-source-conclusion">
            <CardContent className="space-y-2 pt-6">
              <p className="whitespace-pre-wrap break-words text-sm">
                {conclusion.content}
              </p>
              <p className="text-sm text-muted-foreground">
                結論ID: {conclusion.id}
              </p>
              <p className="text-sm text-muted-foreground">
                確定日時: {formatExperimentDateTime(conclusion.finalizedAt)}
              </p>
            </CardContent>
          </Card>
        ) : (
          <Empty id="empty-derivation-conclusion">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Sparkles />
              </EmptyMedia>
              <EmptyTitle>確定した結論がありません</EmptyTitle>
              <EmptyDescription>
                比較結果から結論を確定すると、派生の根拠として利用できます。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </section>
    </div>
  );
}

export function DerivationSourcePage({
  experimentId,
  getDerivationSource,
  briefingServices,
}: {
  experimentId: string;
  getDerivationSource: GetDerivationSourceService;
  briefingServices?: DerivationBriefingServices;
}) {
  const [source, setSource] = useState<DerivationSource>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const load = useCallback(async () => {
    setIsLoading(true);
    setError(undefined);
    try {
      const response = await getDerivationSource(experimentId);
      if (response.data) {
        setSource(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "派生の作成元を取得できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "派生の作成元を取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [experimentId, getDerivationSource]);

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
            派生実験
          </h1>
        </header>
        {isLoading && (
          <section
            aria-label="派生の作成元を読み込んでいます"
            aria-live="polite"
            className="space-y-3"
          >
            <p className="text-muted-foreground">
              派生の作成元を確認しています…
            </p>
            <div className="h-52 animate-pulse rounded-lg bg-muted" />
          </section>
        )}
        {!isLoading && error && (
          <Alert
            id="derivation-source-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>派生の作成元を確認できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-derivation-source-button"
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
        {!isLoading && !error && !source && (
          <Empty id="empty-derivation-source">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <FlaskConical />
              </EmptyMedia>
              <EmptyTitle>派生の作成元がありません</EmptyTitle>
              <EmptyDescription>
                実験を指定して、固定条件と結論を確認してください。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {!isLoading && !error && source && (
          <DerivationSourceContent
            briefingServices={briefingServices}
            source={source}
          />
        )}
      </div>
    </main>
  );
}
