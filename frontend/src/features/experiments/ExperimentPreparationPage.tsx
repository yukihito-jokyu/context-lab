import {
  AlertCircle,
  FileQuestion,
  Plus,
  RefreshCw,
  Save,
  Trash2,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { formatExperimentDateTime } from "./lib/format-experiment-date-time";
import type {
  ExperimentPreparation,
  GetExperimentPreparationService,
} from "./services/get-experiment-preparation-service";
import type {
  SaveExperimentPreparationDraftInput,
  SaveExperimentPreparationDraftService,
} from "./services/save-experiment-preparation-draft-service";

type ExperimentPreparationPageProps = {
  experimentId: string;
  getExperimentPreparation: GetExperimentPreparationService;
  saveExperimentPreparationDraft: SaveExperimentPreparationDraftService;
  onBackToExperimentList?: () => void;
};

type DraftForm = Omit<
  SaveExperimentPreparationDraftInput,
  "requestId" | "experimentId"
>;

const emptyDraft: DraftForm = {
  purpose: "",
  hypothesis: "",
  environmentConditions: "",
  initialInput: "",
  prompts: ["", ""],
  evaluationAxes: "",
};

function toDraft(data: ExperimentPreparation): DraftForm {
  return {
    purpose: data.purpose,
    hypothesis: data.hypothesis ?? "",
    environmentConditions: data.environmentConditions,
    initialInput: data.initialInput,
    prompts: data.prompts.map((prompt) => prompt.content),
    evaluationAxes: data.evaluationAxes,
  };
}

function createRequestId() {
  return crypto.randomUUID();
}

function PreparationSupport({ data }: { data: ExperimentPreparation }) {
  const requiredFields = [
    ["実験目的", data.requiredFields.purpose],
    ["環境", data.requiredFields.environmentConditions],
    ["初期入力", data.requiredFields.initialInput],
    ["候補prompt（2件以上）", data.requiredFields.prompts],
    ["評価軸", data.requiredFields.evaluationAxes],
  ] as const;

  return (
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
          <p>状態: {data.source.state || "未設定"}</p>
          <p className="mt-1">
            ブリーフ版: {data.source.versionId || "未設定"}
          </p>
        </CardContent>
      </Card>
    </aside>
  );
}

function DraftFormFields({
  draft,
  disabled,
  onChange,
}: {
  draft: DraftForm;
  disabled: boolean;
  onChange: (draft: DraftForm) => void;
}) {
  const update = (
    field: Exclude<keyof DraftForm, "prompts">,
    value: string,
  ) => {
    onChange({ ...draft, [field]: value });
  };
  const updatePrompt = (index: number, content: string) => {
    const prompts = [...draft.prompts];
    prompts[index] = content;
    onChange({ ...draft, prompts });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <label className="text-sm font-medium" htmlFor="preparation-purpose">
          実験目的
        </label>
        <textarea
          className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={disabled}
          id="preparation-purpose"
          onChange={(event) => update("purpose", event.target.value)}
          value={draft.purpose}
        />
      </div>
      <div className="space-y-2">
        <label className="text-sm font-medium" htmlFor="preparation-hypothesis">
          仮説（任意）
        </label>
        <textarea
          className="min-h-20 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={disabled}
          id="preparation-hypothesis"
          onChange={(event) => update("hypothesis", event.target.value)}
          value={draft.hypothesis}
        />
      </div>
      <div className="space-y-2">
        <label
          className="text-sm font-medium"
          htmlFor="preparation-environment"
        >
          環境
        </label>
        <textarea
          className="min-h-20 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={disabled}
          id="preparation-environment"
          onChange={(event) =>
            update("environmentConditions", event.target.value)
          }
          value={draft.environmentConditions}
        />
      </div>
      <div className="space-y-2">
        <label
          className="text-sm font-medium"
          htmlFor="preparation-initial-input"
        >
          初期入力
        </label>
        <textarea
          className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={disabled}
          id="preparation-initial-input"
          onChange={(event) => update("initialInput", event.target.value)}
          value={draft.initialInput}
        />
      </div>
      <fieldset className="space-y-3">
        <legend className="text-sm font-medium">候補prompt</legend>
        {draft.prompts.map((prompt, index) => {
          const occurrence = draft.prompts
            .slice(0, index)
            .filter((candidate) => candidate === prompt).length;

          return (
            <div className="flex gap-2" key={`${prompt}-${occurrence}`}>
              <Input
                aria-label={`候補prompt ${index + 1}`}
                disabled={disabled}
                onChange={(event) => updatePrompt(index, event.target.value)}
                value={prompt}
              />
              <Button
                aria-label={`候補prompt ${index + 1}を削除`}
                disabled={disabled || draft.prompts.length <= 1}
                onClick={() =>
                  onChange({
                    ...draft,
                    prompts: draft.prompts.filter(
                      (_, promptIndex) => promptIndex !== index,
                    ),
                  })
                }
                size="icon"
                type="button"
                variant="outline"
              >
                <Trash2 />
              </Button>
            </div>
          );
        })}
        <Button
          disabled={disabled}
          id="add-preparation-prompt-button"
          onClick={() =>
            onChange({ ...draft, prompts: [...draft.prompts, ""] })
          }
          type="button"
          variant="outline"
        >
          <Plus />
          promptを追加
        </Button>
      </fieldset>
      <div className="space-y-2">
        <label
          className="text-sm font-medium"
          htmlFor="preparation-evaluation-axes"
        >
          評価軸
        </label>
        <textarea
          className="min-h-20 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={disabled}
          id="preparation-evaluation-axes"
          onChange={(event) => update("evaluationAxes", event.target.value)}
          value={draft.evaluationAxes}
        />
      </div>
    </div>
  );
}

export function ExperimentPreparationPage({
  experimentId,
  getExperimentPreparation,
  saveExperimentPreparationDraft,
  onBackToExperimentList,
}: ExperimentPreparationPageProps) {
  const [data, setData] = useState<ExperimentPreparation>();
  const [draft, setDraft] = useState<DraftForm>(emptyDraft);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<{ code: string; message: string }>();
  const [saveError, setSaveError] = useState<{
    code: string;
    message: string;
  }>();
  const [isEmpty, setIsEmpty] = useState(false);
  const [savedAt, setSavedAt] = useState<string>();

  const load = useCallback(async () => {
    setIsLoading(true);
    setData(undefined);
    setError(undefined);
    setIsEmpty(false);
    try {
      const response = await getExperimentPreparation(experimentId);
      if (response.data) {
        setData(response.data);
        setDraft(toDraft(response.data));
        setSavedAt(undefined);
        return;
      }
      if (!response.error) {
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
  }, [load]);

  const save = async () => {
    setIsSaving(true);
    setSaveError(undefined);
    try {
      const response = await saveExperimentPreparationDraft({
        requestId: createRequestId(),
        experimentId,
        ...draft,
        hypothesis: draft.hypothesis || undefined,
      });
      if (response.data) {
        setSavedAt(response.data.savedAt);
        setDraft({
          purpose: response.data.purpose,
          hypothesis: response.data.hypothesis ?? "",
          environmentConditions: response.data.environmentConditions,
          initialInput: response.data.initialInput,
          prompts: response.data.prompts,
          evaluationAxes: response.data.evaluationAxes,
        });
        return;
      }
      setSaveError(
        response.error ?? {
          code: "DRAFT_SAVE_FAILED",
          message: "下書きを保存できませんでした。",
        },
      );
    } catch {
      setSaveError({
        code: "DRAFT_SAVE_FAILED",
        message: "下書きを保存できませんでした。",
      });
    } finally {
      setIsSaving(false);
    }
  };

  const notFound = error?.code === "EXPERIMENT_PREPARATION_NOT_FOUND";
  const editable = data?.state === "preparing";

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

        {!isLoading && !error && data && (
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
            <Card>
              <CardHeader>
                <CardTitle>比較条件</CardTitle>
              </CardHeader>
              <CardContent className="space-y-6">
                {!editable && (
                  <Alert>
                    <AlertCircle />
                    <AlertTitle>固定済みの条件です</AlertTitle>
                    <AlertDescription>
                      条件は編集できません。変更する場合は派生実験を作成してください。
                    </AlertDescription>
                  </Alert>
                )}
                {saveError && (
                  <Alert
                    id="preparation-save-error"
                    role="alert"
                    variant="destructive"
                  >
                    <AlertCircle />
                    <AlertTitle>下書きを保存できません</AlertTitle>
                    <AlertDescription>{saveError.message}</AlertDescription>
                  </Alert>
                )}
                {savedAt && (
                  <p
                    aria-live="polite"
                    className="text-sm text-muted-foreground"
                    id="preparation-save-success"
                  >
                    保存済み: {formatExperimentDateTime(savedAt)}
                  </p>
                )}
                <DraftFormFields
                  disabled={!editable || isSaving}
                  draft={draft}
                  onChange={setDraft}
                />
                <Button
                  disabled={!editable || isSaving}
                  id="save-preparation-draft-button"
                  onClick={() => void save()}
                  type="button"
                >
                  <Save />
                  {isSaving ? "保存中…" : "下書きを保存"}
                </Button>
              </CardContent>
            </Card>
            <PreparationSupport data={data} />
          </div>
        )}
        {!isLoading && (
          <p className="text-sm text-muted-foreground">
            最終確認: {formatExperimentDateTime(data?.lastConfirmedAt)}
          </p>
        )}
      </div>
    </main>
  );
}
