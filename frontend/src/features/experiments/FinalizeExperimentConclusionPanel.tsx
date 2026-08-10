import {
  AlertCircle,
  CheckCircle2,
  Lightbulb,
  RefreshCw,
  Sparkles,
} from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type {
  FinalizedExperimentConclusion,
  FinalizeExperimentConclusionService,
} from "./services/finalize-experiment-conclusion-service";

type FinalizeExperimentConclusionPanelProps = {
  experimentId: string;
  finalizeExperimentConclusion: FinalizeExperimentConclusionService;
  existingConclusion?: Omit<FinalizedExperimentConclusion, "requestId">;
  onReload: () => Promise<void>;
};

function createRequestId() {
  return crypto.randomUUID();
}

function countRunes(value: string) {
  return [...value].length;
}

export function FinalizeExperimentConclusionPanel({
  experimentId,
  finalizeExperimentConclusion,
  existingConclusion,
  onReload,
}: FinalizeExperimentConclusionPanelProps) {
  const [conclusion, setConclusion] = useState("");
  const [finalized, setFinalized] =
    useState<Omit<FinalizedExperimentConclusion, "requestId">>(
      existingConclusion,
    );
  const [isFinalizing, setIsFinalizing] = useState(false);
  const [error, setError] = useState<{ code: string; message: string }>();
  const requestId = useRef<string>();
  const trimmedConclusion = conclusion.trim();
  const conclusionLength = countRunes(trimmedConclusion);
  const validationError =
    conclusionLength === 0
      ? "結論を入力してください。"
      : conclusionLength > 8000
        ? "結論は8,000文字以内で入力してください。"
        : undefined;

  useEffect(() => {
    if (existingConclusion) {
      setFinalized(existingConclusion);
    }
  }, [existingConclusion]);

  const finalize = useCallback(async () => {
    if (validationError || finalized) {
      return;
    }
    const currentRequestId = requestId.current ?? createRequestId();
    requestId.current = currentRequestId;
    setError(undefined);
    setIsFinalizing(true);
    try {
      const response = await finalizeExperimentConclusion({
        requestId: currentRequestId,
        experimentId,
        conclusion: trimmedConclusion,
      });
      if (response.data) {
        setFinalized(response.data);
        await onReload();
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "結論を確定できませんでした。",
        },
      );
    } catch {
      setError({ code: "UNKNOWN", message: "結論を確定できませんでした。" });
    } finally {
      setIsFinalizing(false);
    }
  }, [
    experimentId,
    finalizeExperimentConclusion,
    finalized,
    onReload,
    trimmedConclusion,
    validationError,
  ]);

  if (finalized) {
    return (
      <Card id="experiment-conclusion-finalized">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CheckCircle2 className="size-5" /> 結論を確定しました
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="whitespace-pre-wrap break-words text-sm">
            {finalized.conclusion}
          </p>
          <p className="text-sm text-muted-foreground">
            結論ID: {finalized.conclusionId}
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              id="reload-finalized-experiment-conclusion-button"
              onClick={onReload}
              type="button"
              variant="outline"
            >
              <RefreshCw /> 再読込
            </Button>
            <Button asChild type="button" variant="outline">
              <a
                href={`/experiments/${encodeURIComponent(experimentId)}/derivations`}
              >
                <Sparkles /> 派生を確認
              </a>
            </Button>
            <Button asChild type="button" variant="outline">
              <a
                href={`/experiments/${encodeURIComponent(experimentId)}/insights`}
              >
                <Lightbulb /> 知見を確認
              </a>
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card id="experiment-conclusion-form">
      <CardHeader>
        <CardTitle>結論を確定</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert
            id="experiment-conclusion-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>結論を確定できません</AlertTitle>
            <AlertDescription>{error.message}</AlertDescription>
          </Alert>
        )}
        <div className="space-y-2">
          <label
            className="text-sm font-medium"
            htmlFor="experiment-conclusion"
          >
            根拠を確認した結論
          </label>
          <textarea
            className="min-h-36 w-full rounded-md border bg-background px-3 py-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
            disabled={isFinalizing}
            id="experiment-conclusion"
            maxLength={8000}
            onChange={(event) => {
              requestId.current = undefined;
              setConclusion(event.target.value);
              setError(undefined);
            }}
            value={conclusion}
          />
          <p className="text-sm text-muted-foreground">
            {countRunes(conclusion)} / 8,000文字
          </p>
          {validationError && conclusion.length > 0 && (
            <p className="text-sm text-destructive">{validationError}</p>
          )}
        </div>
        <Button
          disabled={Boolean(validationError) || isFinalizing}
          id="finalize-experiment-conclusion-button"
          onClick={() => void finalize()}
          type="button"
        >
          {isFinalizing ? "結論を確定しています…" : "結論を確定"}
        </Button>
      </CardContent>
    </Card>
  );
}
