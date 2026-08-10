import { AlertCircle, RotateCcw } from "lucide-react";
import { useRef, useState } from "react";
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
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import type {
  RetriedEndedRun,
  RetryEndedRunService,
} from "./services/retry-ended-run-service";

type RetryEndedRunPanelProps = {
  experimentId: string;
  fixedConditionId: string;
  runId: string;
  retryEndedRun: RetryEndedRunService;
};

function createRequestId() {
  return crypto.randomUUID();
}

export function RetryEndedRunPanel({
  experimentId,
  fixedConditionId,
  runId,
  retryEndedRun,
}: RetryEndedRunPanelProps) {
  const requestId = useRef<string | undefined>(undefined);
  const [isRetrying, setIsRetrying] = useState(false);
  const [error, setError] = useState<{ code: string; message: string }>();
  const [retriedRun, setRetriedRun] = useState<RetriedEndedRun>();

  const retry = async () => {
    const nextRequestId = requestId.current ?? createRequestId();
    requestId.current = nextRequestId;
    setIsRetrying(true);
    setError(undefined);
    try {
      const response = await retryEndedRun({
        requestId: nextRequestId,
        runId,
      });
      if (response.data) {
        setRetriedRun(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "再実行用runを作成できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "再実行用runを作成できませんでした。",
      });
    } finally {
      setIsRetrying(false);
    }
  };

  return (
    <div className="space-y-3">
      {error && (
        <Alert
          id={`retry-ended-run-error-${runId}`}
          role="alert"
          variant="destructive"
        >
          <AlertCircle />
          <AlertTitle>再実行用runを作成できません</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
        </Alert>
      )}
      {retriedRun && (
        <Alert id={`retry-ended-run-success-${runId}`}>
          <AlertTitle>再実行用runを作成しました</AlertTitle>
          <AlertDescription className="space-y-2">
            <p>
              元run: {retriedRun.sourceRunId} / 新run: {retriedRun.retryRunId}（
              {retriedRun.state}）
            </p>
            <p>操作ID: {retriedRun.operationId}</p>
            <Button
              onClick={() =>
                window.location.assign(
                  `/experiments/${encodeURIComponent(retriedRun.experimentId)}/runs/${encodeURIComponent(retriedRun.retryRunId)}`,
                )
              }
              type="button"
              variant="outline"
            >
              新runを確認
            </Button>
          </AlertDescription>
        </Alert>
      )}
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            disabled={isRetrying || Boolean(retriedRun)}
            id={`retry-ended-run-button-${runId}`}
            type="button"
            variant="outline"
          >
            <RotateCcw />
            {isRetrying ? "再実行用runを作成しています…" : "このrunを再実行"}
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent id={`retry-ended-run-dialog-${runId}`}>
          <AlertDialogHeader>
            <AlertDialogTitle>終了したrunを再実行しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              元run「{runId}」と同じ固定条件（{fixedConditionId}
              ）で、新しいrun記録を作成します。この操作では実行を開始しません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <dl className="grid gap-2 rounded-md border p-3 text-sm">
            <div>
              <dt className="font-medium">実験</dt>
              <dd className="text-muted-foreground">{experimentId}</dd>
            </div>
            <div>
              <dt className="font-medium">元run</dt>
              <dd className="text-muted-foreground">{runId}</dd>
            </div>
            <div>
              <dt className="font-medium">固定条件</dt>
              <dd className="text-muted-foreground">{fixedConditionId}</dd>
            </div>
          </dl>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRetrying}>
              キャンセル
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={isRetrying}
              onClick={() => void retry()}
            >
              新runを作成
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
