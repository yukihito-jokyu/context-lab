import { AlertCircle, Play } from "lucide-react";
import { useCallback, useRef, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import type {
  StartedRunEvaluation,
  StartRunEvaluationService,
} from "./services/start-run-evaluation-service";

type RunEvaluationStartPanelProps = {
  runId: string;
  runState: string;
  startRunEvaluation: StartRunEvaluationService;
  onStarted?: (evaluation: StartedRunEvaluation) => void;
};

function createRequestId() {
  return crypto.randomUUID();
}

export function RunEvaluationStartPanel({
  runId,
  runState,
  startRunEvaluation,
  onStarted,
}: RunEvaluationStartPanelProps) {
  const [isStarting, setIsStarting] = useState(false);
  const [error, setError] = useState<{ code: string; message: string }>();
  const requestId = useRef<string | undefined>(undefined);

  const start = useCallback(async () => {
    const currentRequestId = requestId.current ?? createRequestId();
    requestId.current = currentRequestId;
    setIsStarting(true);
    setError(undefined);

    try {
      const response = await startRunEvaluation({
        requestId: currentRequestId,
        runId,
      });
      if (response.data) {
        onStarted?.(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "評価を開始できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "評価を開始できませんでした。",
      });
    } finally {
      setIsStarting(false);
    }
  }, [onStarted, runId, startRunEvaluation]);

  if (runState !== "completed") {
    return null;
  }

  return (
    <div className="mt-3 space-y-3 border-t pt-3">
      {error && (
        <Alert
          id={`run-evaluation-error-${runId}`}
          role="alert"
          variant="destructive"
        >
          <AlertCircle />
          <AlertTitle>評価を開始できません</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
        </Alert>
      )}
      <Button
        disabled={isStarting}
        id={`start-run-evaluation-button-${runId}`}
        onClick={() => void start()}
        size="sm"
        type="button"
      >
        <Play />
        {isStarting ? "評価を開始しています…" : "評価を開始"}
      </Button>
    </div>
  );
}
