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
import type { SendDerivationBriefMessageService } from "../services/send-derivation-brief-message-service";
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
  sendDerivationBriefMessage,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  sourceExperimentId: string;
  startDerivationBriefing: StartDerivationBriefingService;
  sendDerivationBriefMessage: SendDerivationBriefMessageService;
}) {
  const [isStarting, setIsStarting] = useState(false);
  const [start, setStart] = useState<BriefingStart>();
  const [error, setError] = useState<{ code: string; message: string }>();
  const [message, setMessage] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [sendError, setSendError] = useState<{
    code: string;
    message: string;
  }>();
  const [isMessageInvalid, setIsMessageInvalid] = useState(false);
  const [isMessageSent, setIsMessageSent] = useState(false);
  const [sendOperationID, setSendOperationID] = useState<string>();
  const messageInputRef = useRef<HTMLTextAreaElement>(null);
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

  const send = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!start || isSending) return;

    const trimmedMessage = message.trim();
    if (!trimmedMessage) {
      setIsMessageInvalid(true);
      messageInputRef.current?.focus();
      return;
    }

    setIsSending(true);
    setIsMessageSent(false);
    setSendOperationID(undefined);
    setSendError(undefined);
    try {
      const response = await sendDerivationBriefMessage(
        crypto.randomUUID(),
        start.briefingSessionId,
        trimmedMessage,
      );
      if (response.data) {
        setMessage("");
        setSendOperationID(response.data.operationId);
        setIsMessageSent(true);
        return;
      }
      setSendError(
        response.error ?? {
          code: "UNKNOWN",
          message: "壁打ちメッセージを送信できませんでした。",
        },
      );
    } catch {
      setSendError({
        code: "UNKNOWN",
        message: "壁打ちメッセージを送信できませんでした。",
      });
    } finally {
      setIsSending(false);
    }
  };

  useEffect(() => {
    if (!open) {
      startedForOpenRef.current = false;
      requestIDRef.current = undefined;
      setIsStarting(false);
      setStart(undefined);
      setError(undefined);
      setMessage("");
      setIsSending(false);
      setSendError(undefined);
      setIsMessageInvalid(false);
      setIsMessageSent(false);
      setSendOperationID(undefined);
      return;
    }
    if (startedForOpenRef.current) return;

    startedForOpenRef.current = true;
    void begin();
  }, [begin, open]);

  const isCloseBlocked = isStarting || isSending;

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!nextOpen && isCloseBlocked) return;
        onOpenChange(nextOpen);
      }}
      open={open}
    >
      <DialogContent
        closeDisabled={isCloseBlocked}
        onEscapeKeyDown={(event) => {
          if (isCloseBlocked) event.preventDefault();
        }}
        onPointerDownOutside={(event) => {
          if (isCloseBlocked) event.preventDefault();
        }}
      >
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
          <div className="space-y-4">
            <Alert id="derivation-briefing-started" role="status">
              <CheckCircle2 />
              <AlertTitle>壁打ちを開始しました</AlertTitle>
              <AlertDescription>
                <p>派生元: {start.sourceExperimentId}</p>
                <p>壁打ちsession: {start.briefingSessionId}</p>
                <p>開始操作ID: {start.operationId}</p>
              </AlertDescription>
            </Alert>
            <form
              className="space-y-2"
              id="derivation-briefing-message-form"
              onSubmit={send}
            >
              <label
                className="sr-only"
                htmlFor="derivation-briefing-message-input"
              >
                壁打ちへのメッセージ
              </label>
              <textarea
                aria-describedby={
                  isMessageInvalid
                    ? "derivation-briefing-message-error"
                    : undefined
                }
                aria-invalid={isMessageInvalid}
                className="min-h-24 w-full rounded-md border bg-background p-3 text-sm"
                disabled={isSending}
                id="derivation-briefing-message-input"
                onChange={(event) => {
                  setMessage(event.target.value);
                  setIsMessageInvalid(false);
                  setIsMessageSent(false);
                }}
                placeholder="派生させたい観点、判断基準、制約などを入力"
                ref={messageInputRef}
                value={message}
              />
              {isMessageInvalid && (
                <p id="derivation-briefing-message-error" role="alert">
                  壁打ちへ送る内容を入力してください。
                </p>
              )}
              {isSending && (
                <p id="derivation-briefing-message-pending" role="status">
                  メッセージを送信しています…
                </p>
              )}
              {sendError && (
                <Alert
                  id="derivation-briefing-message-command-error"
                  role="alert"
                  variant="destructive"
                >
                  <AlertCircle />
                  <AlertTitle>壁打ちメッセージを送信できません</AlertTitle>
                  <AlertDescription>{sendError.message}</AlertDescription>
                </Alert>
              )}
              {isMessageSent && (
                <p id="derivation-briefing-message-sent" role="status">
                  送信を受け付けました。操作ID: {sendOperationID}
                </p>
              )}
              <Button
                disabled={isSending}
                id="send-derivation-briefing-message-button"
                type="submit"
              >
                送信
              </Button>
            </form>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
