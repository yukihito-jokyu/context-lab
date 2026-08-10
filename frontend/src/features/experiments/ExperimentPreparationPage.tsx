import { AlertCircle, FileQuestion, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type {
  ExperimentPreparation,
  GetExperimentPreparationService,
} from "./services/get-experiment-preparation-service";

type ExperimentPreparationPageProps = {
  experimentId: string;
  getExperimentPreparation: GetExperimentPreparationService;
  onBackToExperimentList?: () => void;
};

function displaySourceState(state: string) {
  if (state === "adopted") {
    return "採用済み";
  }

  return state || "未設定";
}

function PreparationDetails({ data }: { data: ExperimentPreparation }) {
  const requiredFields = [
    ["実験目的", data.requiredFields.purpose],
    ["環境", data.requiredFields.environmentConditions],
    ["初期入力", data.requiredFields.initialInput],
    ["候補prompt（2件以上）", data.requiredFields.prompts],
    ["評価軸", data.requiredFields.evaluationAxes],
  ] as const;

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
      <section aria-labelledby="preparation-conditions-title">
        <Card>
          <CardHeader>
            <CardTitle id="preparation-conditions-title">比較条件</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <dl className="space-y-5">
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  実験目的
                </dt>
                <dd className="mt-1 whitespace-pre-wrap break-words">
                  {data.purpose}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  仮説
                </dt>
                <dd className="mt-1 whitespace-pre-wrap break-words">
                  {data.hypothesis || "未設定"}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  環境
                </dt>
                <dd className="mt-1 whitespace-pre-wrap break-words">
                  {data.environmentConditions || "未設定"}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  初期入力
                </dt>
                <dd className="mt-1 whitespace-pre-wrap break-words">
                  {data.initialInput || "未設定"}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  候補prompt
                </dt>
                <dd>
                  {data.prompts.length > 0 ? (
                    <ol className="mt-2 space-y-3">
                      {data.prompts.map((prompt) => (
                        <li
                          className="rounded-md border p-3"
                          key={prompt.sequenceNo}
                        >
                          <p className="text-sm text-muted-foreground">
                            Prompt {prompt.sequenceNo}
                          </p>
                          <p className="mt-1 whitespace-pre-wrap break-words">
                            {prompt.content}
                          </p>
                        </li>
                      ))}
                    </ol>
                  ) : (
                    <p className="mt-1 text-muted-foreground">未設定</p>
                  )}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  評価軸
                </dt>
                <dd className="mt-1 whitespace-pre-wrap break-words">
                  {data.evaluationAxes || "未設定"}
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>
      </section>

      <aside className="space-y-4" aria-label="実験準備の補助情報">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">入力の確認</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              {requiredFields.map(([field, complete]) => (
                <li
                  className="flex items-center justify-between gap-3 rounded-md bg-muted px-3 py-2"
                  key={field}
                >
                  <span>{field}</span>
                  <Badge variant={complete ? "outline" : "destructive"}>
                    {complete ? "入力済み" : "未入力"}
                  </Badge>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">採用元</CardTitle>
          </CardHeader>
          <CardContent className="break-words text-sm text-muted-foreground">
            <p>状態: {displaySourceState(data.source.state)}</p>
            <p className="mt-1">
              ブリーフ版: {data.source.versionId || "未設定"}
            </p>
          </CardContent>
        </Card>
      </aside>
    </div>
  );
}

export function ExperimentPreparationPage({
  experimentId,
  getExperimentPreparation,
  onBackToExperimentList,
}: ExperimentPreparationPageProps) {
  const [data, setData] = useState<ExperimentPreparation>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<{ code: string; message: string }>();
  const [isEmpty, setIsEmpty] = useState(false);

  const load = useCallback(async () => {
    setIsLoading(true);
    setData(undefined);
    setError(undefined);
    setIsEmpty(false);
    try {
      const response = await getExperimentPreparation(experimentId);
      if (response.data) {
        setData(response.data);
        return;
      }
      if (!response.error) {
        setData(undefined);
        setIsEmpty(true);
        return;
      }
      setError(response.error);
    } catch {
      setError({
        code: "UNKNOWN",
        message: "実験準備を取得できませんでした。",
      });
    } finally {
      setIsLoading(false);
    }
  }, [experimentId, getExperimentPreparation]);

  useEffect(() => {
    void load();
  }, [load]); // 初期表示と再読込は同じqueryを使う。

  const notFound = error?.code === "EXPERIMENT_PREPARATION_NOT_FOUND";

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <header className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
              <span>{data?.experimentId ?? experimentId}</span>
              {data?.state && <Badge variant="outline">{data.state}</Badge>}
            </div>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">
              実験の条件を準備する
            </h1>
            <p className="mt-2 text-muted-foreground">
              固定後は条件を変更できません。比較できる前提を確認してください。
            </p>
          </div>
          {onBackToExperimentList && (
            <Button
              id="back-to-experiment-list-button"
              onClick={onBackToExperimentList}
              type="button"
              variant="outline"
            >
              一覧へ戻る
            </Button>
          )}
        </header>

        {isLoading && (
          <section aria-live="polite" id="preparation-loading">
            <p className="text-muted-foreground">条件を読み込んでいます…</p>
            <div className="mt-4 h-96 animate-pulse rounded-lg bg-muted" />
          </section>
        )}

        {!isLoading && error && (
          <Alert id="preparation-load-error" role="alert" variant="destructive">
            {notFound ? <FileQuestion /> : <AlertCircle />}
            <AlertTitle>
              {notFound
                ? "対象の実験は見つかりません"
                : "実験準備を確認できません"}
            </AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button
                id="reload-preparation-button"
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

        {!isLoading && isEmpty && (
          <section
            aria-live="polite"
            className="rounded-lg border border-dashed p-8 text-center"
            id="preparation-empty"
          >
            <h2 className="text-lg font-semibold">準備内容はまだありません</h2>
            <p className="mt-2 text-muted-foreground">
              実験の条件を作成すると、ここで内容を確認できます。
            </p>
            <Button
              className="mt-4"
              id="reload-preparation-button"
              onClick={() => void load()}
              type="button"
              variant="outline"
            >
              <RefreshCw />
              再読込
            </Button>
          </section>
        )}

        {!isLoading && !error && data && <PreparationDetails data={data} />}

        {!isLoading && (
          <p className="text-sm text-muted-foreground">
            最終確認: {formatExperimentDateTime(data?.lastConfirmedAt)}
          </p>
        )}
      </div>
    </main>
  );
}
