import { AlertCircle, CheckCircle2, RefreshCw } from "lucide-react";
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
import type {
  DerivationBriefing,
  GetDerivationBriefingService,
} from "../services/get-derivation-briefing-service";
import type { SendDerivationBriefMessageService } from "../services/send-derivation-brief-message-service";
import type { StartDerivationBriefingService } from "../services/start-derivation-briefing-service";
import type { StopDerivationBriefingService } from "../services/stop-derivation-briefing-service";

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
  getDerivationBriefing,
  stopDerivationBriefing,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  sourceExperimentId: string;
  startDerivationBriefing: StartDerivationBriefingService;
  sendDerivationBriefMessage: SendDerivationBriefMessageService;
  getDerivationBriefing: GetDerivationBriefingService;
  stopDerivationBriefing: StopDerivationBriefingService;
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
  const [briefing, setBriefing] = useState<DerivationBriefing>();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [refreshError, setRefreshError] = useState<{
    code: string;
    message: string;
  }>();
  const [isStopConfirmationOpen, setIsStopConfirmationOpen] = useState(false);
  const [isStopping, setIsStopping] = useState(false);
  const [stopError, setStopError] = useState<{
    code: string;
    message: string;
  }>();
  const messageInputRef = useRef<HTMLTextAreaElement>(null);
  const requestIDRef = useRef<string>();
  const startedForOpenRef = useRef(false);
  const refreshGenerationRef = useRef(0);
  const isRefreshingRef = useRef(false);

  const refreshBriefing = useCallback(async () => {
    if (!start || isRefreshingRef.current) return;

    const generation = ++refreshGenerationRef.current;
    isRefreshingRef.current = true;
    setIsRefreshing(true);
    setRefreshError(undefined);
    try {
      const response = await getDerivationBriefing(start.briefingSessionId);
      if (generation !== refreshGenerationRef.current) return;
      if (response.data) {
        setBriefing(response.data);
        return;
      }
      setRefreshError(
        response.error ?? {
          code: "UNKNOWN",
          message: "壁打ち内容を取得できませんでした。",
        },
      );
    } catch {
      if (generation !== refreshGenerationRef.current) return;
      setRefreshError({
        code: "UNKNOWN",
        message: "壁打ち内容を取得できませんでした。",
      });
    } finally {
      if (generation === refreshGenerationRef.current) {
        isRefreshingRef.current = false;
        setIsRefreshing(false);
      }
    }
  }, [getDerivationBriefing, start]);

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

  useEffect(() => {
    if (start) void refreshBriefing();
  }, [refreshBriefing, start]);

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
        await refreshBriefing();
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

  const requestClose = () => {
    if (!start) {
      onOpenChange(false);
      return;
    }
    setStopError(undefined);
    setIsStopConfirmationOpen(true);
  };

  const stop = async () => {
    if (!start || isStopping) return;

    setIsStopping(true);
    setStopError(undefined);
    try {
      const response = await stopDerivationBriefing(
        crypto.randomUUID(),
        start.briefingSessionId,
      );
      if (response.data) {
        setIsStopConfirmationOpen(false);
        onOpenChange(false);
        return;
      }
      setStopError(
        response.error ?? {
          code: "UNKNOWN",
          message: "壁打ちを終了できませんでした。",
        },
      );
    } catch {
      setStopError({
        code: "UNKNOWN",
        message: "壁打ちを終了できませんでした。",
      });
    } finally {
      setIsStopping(false);
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
      setBriefing(undefined);
      setIsRefreshing(false);
      isRefreshingRef.current = false;
      setRefreshError(undefined);
      setIsStopConfirmationOpen(false);
      setIsStopping(false);
      setStopError(undefined);
      refreshGenerationRef.current += 1;
      return;
    }
    if (startedForOpenRef.current) return;

    startedForOpenRef.current = true;
    void begin();
  }, [begin, open]);

  const isCloseBlocked = isStarting || isSending || isStopping;

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          if (isCloseBlocked) return;
          requestClose();
          return;
        }
        onOpenChange(true);
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
            <section
              aria-labelledby="derivation-briefing-content-title"
              className="space-y-3"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2
                  className="font-semibold"
                  id="derivation-briefing-content-title"
                >
                  壁打ち内容
                </h2>
                <Button
                  disabled={isRefreshing}
                  id="reload-derivation-briefing-button"
                  onClick={() => void refreshBriefing()}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  <RefreshCw
                    className={isRefreshing ? "animate-spin" : undefined}
                  />
                  再読込
                </Button>
              </div>
              {isRefreshing && (
                <p id="derivation-briefing-refresh-pending" role="status">
                  壁打ち内容を読み込んでいます…
                </p>
              )}
              {refreshError && (
                <Alert
                  id="derivation-briefing-refresh-error"
                  role="alert"
                  variant="destructive"
                >
                  <AlertCircle />
                  <AlertTitle>壁打ち内容を取得できません</AlertTitle>
                  <AlertDescription className="space-y-3">
                    <p>{refreshError.message}</p>
                    {briefing?.lastConfirmedAt && (
                      <p className="text-xs">
                        前回確認: {briefing.lastConfirmedAt}
                      </p>
                    )}
                    <Button
                      onClick={() => void refreshBriefing()}
                      size="sm"
                      type="button"
                      variant="outline"
                    >
                      もう一度試す
                    </Button>
                  </AlertDescription>
                </Alert>
              )}
              {briefing && <DerivationBriefingContent briefing={briefing} />}
              {!briefing && !isRefreshing && !refreshError && (
                <p className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
                  まだ壁打ち内容はありません。メッセージを送ると、ここに会話と提案が表示されます。
                </p>
              )}
            </section>
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
                disabled={isSending || isStopping}
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
                disabled={isSending || isStopping}
                id="send-derivation-briefing-message-button"
                type="submit"
              >
                送信
              </Button>
            </form>
            {isStopConfirmationOpen && (
              <Alert id="derivation-briefing-stop-confirmation" role="alert">
                <AlertCircle />
                <AlertTitle>壁打ちを終了しますか？</AlertTitle>
                <AlertDescription className="space-y-4">
                  <p>終了後は、この壁打ちにメッセージを送信できません。</p>
                  {isStopping && (
                    <p id="derivation-briefing-stop-pending" role="status">
                      壁打ちを終了しています…
                    </p>
                  )}
                  {stopError && (
                    <p id="derivation-briefing-stop-error" role="alert">
                      {stopError.message}
                    </p>
                  )}
                  <div className="flex flex-wrap justify-end gap-2">
                    <Button
                      disabled={isStopping}
                      onClick={() => setIsStopConfirmationOpen(false)}
                      type="button"
                      variant="outline"
                    >
                      続ける
                    </Button>
                    <Button
                      disabled={isStopping}
                      id="stop-derivation-briefing-button"
                      onClick={() => void stop()}
                      type="button"
                      variant="destructive"
                    >
                      壁打ちを終了
                    </Button>
                  </div>
                </AlertDescription>
              </Alert>
            )}
            {!isStopConfirmationOpen && (
              <div className="flex justify-end">
                <Button
                  disabled={isCloseBlocked}
                  id="request-stop-derivation-briefing-button"
                  onClick={requestClose}
                  type="button"
                  variant="outline"
                >
                  壁打ちを終了
                </Button>
              </div>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

function DerivationBriefingContent({
  briefing,
}: {
  briefing: DerivationBriefing;
}) {
  const suggestion = briefing.latestSuggestion;
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <section
        aria-labelledby="derivation-briefing-conversation-title"
        className="space-y-2"
      >
        <h3 className="font-medium" id="derivation-briefing-conversation-title">
          会話
        </h3>
        {briefing.messages.length === 0 ? (
          <p className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
            会話はまだありません。
          </p>
        ) : (
          <ol className="space-y-2" id="derivation-briefing-messages">
            {briefing.messages.map((item) => (
              <li
                className="rounded-md border p-3 text-sm"
                key={`${item.sequenceNo}-${item.createdAt}`}
              >
                <p className="font-medium">
                  {item.role === "user" ? "あなた" : "アシスタント"}
                </p>
                <p className="mt-1 whitespace-pre-wrap break-words">
                  {item.content}
                </p>
                <p className="mt-2 text-xs text-muted-foreground">
                  {item.createdAt}
                </p>
              </li>
            ))}
          </ol>
        )}
      </section>
      <section
        aria-labelledby="derivation-briefing-suggestion-title"
        className="space-y-2"
      >
        <h3 className="font-medium" id="derivation-briefing-suggestion-title">
          差分案
        </h3>
        {!suggestion ? (
          <p className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
            差分案はまだありません。
          </p>
        ) : (
          <div
            className="space-y-3 rounded-md border p-3 text-sm"
            id="derivation-briefing-suggestion"
          >
            <p className="font-medium">{suggestion.purpose}</p>
            <dl className="space-y-2">
              <BriefingField label="判断" value={suggestion.decision} />
              <BriefingField label="仮説" value={suggestion.hypothesis} />
              <BriefingField
                label="評価基準"
                value={suggestion.evaluationCriteria}
              />
              <BriefingField
                label="環境条件"
                value={suggestion.environmentConditions}
              />
              <BriefingField label="初期入力" value={suggestion.initialInput} />
              <BriefingField
                label="成功基準"
                value={suggestion.successCriteria}
              />
              <BriefingField
                label="必須条件"
                value={suggestion.requiredConditions}
              />
            </dl>
            {suggestion.candidatePrompts.length > 0 && (
              <div>
                <p className="font-medium">候補prompt</p>
                <ul className="mt-1 list-disc pl-5">
                  {suggestion.candidatePrompts.map((prompt) => (
                    <li key={prompt}>{prompt}</li>
                  ))}
                </ul>
              </div>
            )}
            <div
              className="rounded-md bg-muted/60 p-3"
              id="derivation-briefing-open-question"
            >
              <p className="font-medium">未解決事項</p>
              <p className="mt-1 whitespace-pre-wrap break-words">
                {suggestion.openQuestion || "未解決事項はありません。"}
              </p>
            </div>
          </div>
        )}
      </section>
    </div>
  );
}

function BriefingField({ label, value }: { label: string; value?: string }) {
  if (!value) return null;
  return (
    <div>
      <dt className="font-medium">{label}</dt>
      <dd className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
        {value}
      </dd>
    </div>
  );
}
