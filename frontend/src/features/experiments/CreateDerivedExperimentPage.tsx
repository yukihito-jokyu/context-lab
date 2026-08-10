import {
  AlertCircle,
  ArrowLeft,
  GitBranchPlus,
  LoaderCircle,
} from "lucide-react";
import { useMemo, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type {
  CreateDerivedExperimentRequest,
  CreateDerivedExperimentService,
} from "./services/create-derived-experiment-service";

type ChangeFields = CreateDerivedExperimentRequest["changes"];

const createRequestID = () => crypto.randomUUID();

function trimmedChanges(fields: ChangeFields): ChangeFields {
  const prompts = fields.prompts
    ?.map((prompt) => prompt.content.trim())
    .filter(Boolean);
  return {
    purpose: fields.purpose?.trim() || undefined,
    hypothesis: fields.hypothesis?.trim() || undefined,
    environmentConditions: fields.environmentConditions?.trim() || undefined,
    initialInput: fields.initialInput?.trim() || undefined,
    prompts: prompts?.length
      ? prompts.map((content, index) => ({ sequenceNo: index + 1, content }))
      : undefined,
    evaluationAxes: fields.evaluationAxes?.trim() || undefined,
  };
}

function hasChanges(changes: ChangeFields) {
  return Object.values(changes).some(
    (value) => Array.isArray(value) || Boolean(value),
  );
}

export function CreateDerivedExperimentPage({
  createDerivedExperiment,
  sourceExperimentId,
}: {
  createDerivedExperiment: CreateDerivedExperimentService;
  sourceExperimentId: string;
}) {
  const [changes, setChanges] = useState<ChangeFields>({});
  const [reason, setReason] = useState("");
  const [requestId, setRequestId] = useState<string>();
  const [error, setError] = useState<{ code: string; message: string }>();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const normalizedChanges = useMemo(() => trimmedChanges(changes), [changes]);
  const normalizedReason = reason.trim();

  const updateChange = (
    field: Exclude<keyof ChangeFields, "prompts">,
    value: string,
  ) => {
    setChanges((current) => ({ ...current, [field]: value }));
    setRequestId(undefined);
    setError(undefined);
  };

  const submit = async () => {
    if (isSubmitting) {
      return;
    }
    if (!hasChanges(normalizedChanges) || !normalizedReason) {
      setError({
        code: "VALIDATION_ERROR",
        message: "差分を1項目以上入力し、派生する理由を入力してください。",
      });
      return;
    }
    const nextRequestId = requestId ?? createRequestID();
    setRequestId(nextRequestId);
    setError(undefined);
    setIsSubmitting(true);
    try {
      const response = await createDerivedExperiment({
        requestId: nextRequestId,
        sourceExperimentId,
        changes: normalizedChanges,
        reason: normalizedReason,
      });
      if (response.data) {
        window.location.assign(
          `/experiments/${encodeURIComponent(response.data.experimentId)}/preparation`,
        );
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "派生実験を作成できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "派生実験を作成できませんでした。",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="min-h-screen bg-background p-4 text-foreground sm:p-8">
      <div className="mx-auto max-w-3xl space-y-6">
        <header className="space-y-3">
          <Button asChild type="button" variant="ghost">
            <a
              href={`/experiments/${encodeURIComponent(sourceExperimentId)}/derivations`}
            >
              <ArrowLeft /> 派生元に戻る
            </a>
          </Button>
          <div>
            <p className="text-sm text-muted-foreground">
              派生元の実験ID: {sourceExperimentId}
            </p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">
              派生実験を作成
            </h1>
            <p className="mt-2 text-muted-foreground">
              変更する項目と、その理由を記録して新しい実験を準備します。
            </p>
          </div>
        </header>

        {error && (
          <Alert
            id="create-derived-experiment-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>派生実験を作成できません</AlertTitle>
            <AlertDescription>{error.message}</AlertDescription>
          </Alert>
        )}

        <Card>
          <CardHeader>
            <CardTitle>変更する差分</CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <p className="text-sm text-muted-foreground">
              変更する項目だけを入力してください。少なくとも1項目必要です。
            </p>
            <div className="grid gap-5 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="derived-purpose">実験目的</Label>
                <Input
                  disabled={isSubmitting}
                  id="derived-purpose"
                  onChange={(event) =>
                    updateChange("purpose", event.target.value)
                  }
                  value={changes.purpose ?? ""}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="derived-hypothesis">仮説</Label>
                <Input
                  disabled={isSubmitting}
                  id="derived-hypothesis"
                  onChange={(event) =>
                    updateChange("hypothesis", event.target.value)
                  }
                  value={changes.hypothesis ?? ""}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="derived-environment">環境条件</Label>
                <Input
                  disabled={isSubmitting}
                  id="derived-environment"
                  onChange={(event) =>
                    updateChange("environmentConditions", event.target.value)
                  }
                  value={changes.environmentConditions ?? ""}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="derived-evaluation-axes">評価軸</Label>
                <Input
                  disabled={isSubmitting}
                  id="derived-evaluation-axes"
                  onChange={(event) =>
                    updateChange("evaluationAxes", event.target.value)
                  }
                  value={changes.evaluationAxes ?? ""}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="derived-initial-input">初期入力</Label>
              <textarea
                className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                disabled={isSubmitting}
                id="derived-initial-input"
                onChange={(event) =>
                  updateChange("initialInput", event.target.value)
                }
                value={changes.initialInput ?? ""}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="derived-prompts">prompt</Label>
              <textarea
                className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                disabled={isSubmitting}
                id="derived-prompts"
                onChange={(event) => {
                  setChanges((current) => ({
                    ...current,
                    prompts: event.target.value
                      .split("\n")
                      .map((content, index) => ({
                        sequenceNo: index + 1,
                        content,
                      })),
                  }));
                  setRequestId(undefined);
                  setError(undefined);
                }}
                placeholder="1行に1件入力"
                value={
                  changes.prompts?.map((prompt) => prompt.content).join("\n") ??
                  ""
                }
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>派生する理由</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <textarea
              className="min-h-28 w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-soft focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              disabled={isSubmitting}
              id="derived-reason"
              onChange={(event) => {
                setReason(event.target.value);
                setRequestId(undefined);
                setError(undefined);
              }}
              required
              value={reason}
            />
            <Button
              disabled={isSubmitting}
              id="create-derived-experiment-button"
              onClick={() => void submit()}
              type="button"
            >
              {isSubmitting ? (
                <LoaderCircle className="animate-spin" />
              ) : (
                <GitBranchPlus />
              )}
              {isSubmitting ? "作成しています…" : "派生実験を作成"}
            </Button>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
