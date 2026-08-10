import { AlertCircle, Play } from "lucide-react";
import { useCallback, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type {
  StartedPreparation,
  StartPreparationService,
} from "./services/start-preparation-service";

type StartPreparationPanelProps = {
  startPreparation: StartPreparationService;
  onStarted: (preparation: StartedPreparation) => void;
};

type StartError = {
  code: string;
  message: string;
};

function createRequestId() {
  return crypto.randomUUID();
}

const pendingRequestStorageKey = "startPreparationPendingRequest";

type PendingRequest = {
  requestId: string;
  scope: string;
};

function findPendingRequest(scope: string): PendingRequest | undefined {
  try {
    const value = window.localStorage.getItem(pendingRequestStorageKey);
    if (!value) return;
    const pending = JSON.parse(value) as PendingRequest;
    if (pending.scope === scope && pending.requestId) return pending;
  } catch {
    return;
  }
}

function savePendingRequest(pending: PendingRequest) {
  window.localStorage.setItem(
    pendingRequestStorageKey,
    JSON.stringify(pending),
  );
}

function clearPendingRequest() {
  window.localStorage.removeItem(pendingRequestStorageKey);
}

function isPendingStart(error: StartError) {
  return error.code === "PREPARATION_START_PENDING";
}

function validateScope(scope: string): StartError | undefined {
  if (scope === "") {
    return {
      code: "PREPARATION_SCOPE_INVALID",
      message: "対象範囲を入力してください。",
    };
  }
  if (scope.length > 512) {
    return {
      code: "PREPARATION_SCOPE_INVALID",
      message: "対象範囲は512文字以内で入力してください。",
    };
  }
  if (
    scope.startsWith("/") ||
    /^[a-zA-Z]:[\\/]/.test(scope) ||
    scope.split(/[\\/]/).includes("..")
  ) {
    return {
      code: "PREPARATION_SCOPE_INVALID",
      message: "対象範囲はワークスペースからの相対パスで入力してください。",
    };
  }
}

export function StartPreparationPanel({
  startPreparation,
  onStarted,
}: StartPreparationPanelProps) {
  const [scope, setScope] = useState(".");
  const [isStarting, setIsStarting] = useState(false);
  const [error, setError] = useState<StartError>();
  const start = useCallback(async () => {
    const trimmedScope = scope.trim();
    const validationError = validateScope(trimmedScope);
    if (validationError) {
      setError(validationError);
      return;
    }

    const pending = findPendingRequest(trimmedScope);
    const currentRequestId = pending?.requestId ?? createRequestId();
    savePendingRequest({ requestId: currentRequestId, scope: trimmedScope });
    setIsStarting(true);
    setError(undefined);

    try {
      const response = await startPreparation({
        requestId: currentRequestId,
        scope: trimmedScope,
      });
      if (!response.data) {
        const startError = response.error ?? {
          code: "UNKNOWN",
          message: "環境準備を開始できませんでした。",
        };
        setError(startError);
        if (response.error && !isPendingStart(startError)) {
          clearPendingRequest();
        }
        return;
      }
      clearPendingRequest();
      onStarted(response.data);
    } catch {
      setError({
        code: "UNKNOWN",
        message: "環境準備を開始できませんでした。",
      });
    } finally {
      setIsStarting(false);
    }
  }, [onStarted, scope, startPreparation]);

  return (
    <Card id="start-preparation-panel">
      <CardHeader className="gap-2">
        <CardTitle>環境準備を開始</CardTitle>
        <p className="text-sm text-muted-foreground">
          対象範囲を指定して、安全な環境候補と診断を取得します。
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert
            id="start-preparation-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>環境準備を開始できません</AlertTitle>
            <AlertDescription>{error.message}</AlertDescription>
          </Alert>
        )}
        <div className="space-y-2">
          <label className="text-sm font-medium" htmlFor="preparation-scope">
            対象範囲（ワークスペースからの相対パス）
          </label>
          <Input
            disabled={isStarting}
            id="preparation-scope"
            maxLength={512}
            onChange={(event) => {
              clearPendingRequest();
              setScope(event.target.value);
            }}
            required
            value={scope}
          />
        </div>
        <Button
          disabled={isStarting}
          id="start-preparation-button"
          onClick={() => void start()}
          type="button"
        >
          <Play />
          {isStarting ? "環境準備を開始しています…" : "環境準備を開始"}
        </Button>
      </CardContent>
    </Card>
  );
}
