import { AlertCircle, CheckCircle2 } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { StartDerivationBriefingService } from "../services/start-derivation-briefing-service";

type BriefingStart = {
  briefingSessionId: string;
  operationId: string;
  sourceExperimentId: string;
};

export function DerivationBriefingDialog({
  open,
  onOpenChange,
  sourceExperimentId,
  startDerivationBriefing,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  sourceExperimentId: string;
  startDerivationBriefing: StartDerivationBriefingService;
}) {
  const [isStarting, setIsStarting] = useState(false);
  const [start, setStart] = useState<BriefingStart>();
  const [error, setError] = useState<{ code: string; message: string }>();
  const requestIDRef = useRef<string>();
  const startedForOpenRef = useRef(false);

  const begin = useCallback(async () => {
    if (isStarting) return;

    const requestId = requestIDRef.current ?? crypto.randomUUID();
    requestIDRef.current = requestId;
    setIsStarting(true);
    setError(undefined);
    try {
      const response = await startDerivationBriefing(
        requestId,
        sourceExperimentId,
      );
      if (response.data) {
        setStart(response.data);
        return;
      }
      setError(
        response.error ?? {
          code: "UNKNOWN",
          message: "派生実験の壁打ちを開始できませんでした。",
        },
      );
    } catch {
      setError({
        code: "UNKNOWN",
        message: "派生実験の壁打ちを開始できませんでした。",
      });
    } finally {
      setIsStarting(false);
    }
  }, [isStarting, sourceExperimentId, startDerivationBriefing]);

  const retry = () => {
    requestIDRef.current = undefined;
    void begin();
  };

  useEffect(() => {
    if (!open) {
      startedForOpenRef.current = false;
      requestIDRef.current = undefined;
      setIsStarting(false);
      setStart(undefined);
      setError(undefined);
      return;
    }
    if (startedForOpenRef.current) return;

    startedForOpenRef.current = true;
    void begin();
  }, [begin, open]);

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <DialogContent closeDisabled={isStarting}>
        <DialogHeader>
          <DialogTitle>派生実験の壁打ちを開始</DialogTitle>
          <DialogDescription>
            派生元の固定条件と確定した結論をもとに、壁打ちの準備を開始します。
          </DialogDescription>
        </DialogHeader>
        {isStarting && (
          <p id="derivation-briefing-pending" role="status">
            壁打ちを開始しています…
          </p>
        )}
        {error && (
          <Alert
            id="derivation-briefing-start-error"
            role="alert"
            variant="destructive"
          >
            <AlertCircle />
            <AlertTitle>壁打ちを開始できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{error.message}</p>
              <Button onClick={retry} type="button" variant="outline">
                もう一度試す
              </Button>
            </AlertDescription>
          </Alert>
        )}
        {start && (
          <Alert id="derivation-briefing-started" role="status">
            <CheckCircle2 />
            <AlertTitle>壁打ちを開始しました</AlertTitle>
            <AlertDescription className="space-y-1">
              <p>派生元: {start.sourceExperimentId}</p>
              <p>壁打ちsession: {start.briefingSessionId}</p>
              <p>操作ID: {start.operationId}</p>
            </AlertDescription>
          </Alert>
        )}
      </DialogContent>
    </Dialog>
  );
}
