import { AlertCircle, RefreshCw } from "lucide-react";
import { useRef, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { formatExperimentDateTime } from "../lib/format-experiment-date-time";
import type { ExperimentBriefing } from "../services/get-experiment-briefing-service";

type BriefingError = { code: string; message: string };

type ExperimentBriefingConversationProps = {
  briefing?: ExperimentBriefing;
  error?: BriefingError;
  isRefreshing: boolean;
  isSending: boolean;
  isStopping: boolean;
  onRefresh: () => void;
  onSend: (message: string) => Promise<boolean>;
  sendError?: BriefingError;
};

export function ExperimentBriefingConversation({
  briefing,
  error,
  isRefreshing,
  isSending,
  isStopping,
  onRefresh,
  onSend,
  sendError,
}: ExperimentBriefingConversationProps) {
  const [message, setMessage] = useState("");
  const [validationError, setValidationError] = useState(false);
  const messageInputRef = useRef<HTMLTextAreaElement>(null);

  const submitMessage = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (isStopping) return;
    const trimmedMessage = message.trim();
    if (!trimmedMessage) {
      setValidationError(true);
      messageInputRef.current?.focus();
      return;
    }

    const wasSent = await onSend(trimmedMessage);
    if (wasSent) setMessage("");
  };

  return (
    <section aria-labelledby="briefing-chat-title" className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h3 className="font-semibold" id="briefing-chat-title">
          実験設計の壁打ち
        </h3>
        <Button
          disabled={isRefreshing || isStopping}
          id="reload-experiment-briefing-button"
          onClick={onRefresh}
          type="button"
          variant="outline"
        >
          <RefreshCw className={isRefreshing ? "animate-spin" : undefined} />
          会話を再取得
        </Button>
      </div>
      {isRefreshing && (
        <p id="briefing-refresh-pending" role="status">
          会話とブリーフの最新状態を取得しています…
        </p>
      )}
      {error && (
        <Alert id="briefing-refresh-error" role="alert" variant="destructive">
          <AlertCircle />
          <AlertTitle>最新状態を取得できません</AlertTitle>
          <AlertDescription className="space-y-3">
            <p>{error.message}</p>
            <Button
              disabled={isRefreshing || isStopping}
              onClick={onRefresh}
              type="button"
              variant="outline"
            >
              もう一度試す
            </Button>
          </AlertDescription>
        </Alert>
      )}
      <div
        aria-live="polite"
        className="max-h-80 space-y-3 overflow-y-auto rounded-lg border bg-muted/30 p-3"
        id="briefing-chat-log"
      >
        {!isRefreshing &&
          !error &&
          (!briefing || briefing.messages.length === 0) && (
            <p className="text-sm text-muted-foreground">
              会話はまだありません。
            </p>
          )}
        {briefing?.messages.map((message) => (
          <article
            className="rounded-md bg-background p-3 text-sm"
            key={message.sequenceNo}
          >
            <p className="font-medium">
              {message.role === "user" ? "あなた" : "AI"}
            </p>
            <p className="mt-1 whitespace-pre-wrap break-words">
              {message.content}
            </p>
            <time
              className="mt-2 block text-xs text-muted-foreground"
              dateTime={message.createdAt}
            >
              {formatExperimentDateTime(message.createdAt)}
            </time>
          </article>
        ))}
      </div>
      <form
        className="space-y-2"
        id="briefing-message-form"
        onSubmit={submitMessage}
      >
        <label className="sr-only" htmlFor="briefing-message-input">
          AIへの回答
        </label>
        <textarea
          aria-describedby={validationError ? "briefing-error" : undefined}
          aria-invalid={validationError}
          className="min-h-24 w-full rounded-md border bg-background p-3 text-sm"
          disabled={isSending || isStopping}
          id="briefing-message-input"
          onChange={(event) => {
            setMessage(event.target.value);
            setValidationError(false);
          }}
          placeholder="背景、判断基準、対象入力、制約などを入力"
          ref={messageInputRef}
          value={message}
        />
        {validationError && (
          <p id="briefing-error" role="alert">
            AIへ送る内容を入力してください。
          </p>
        )}
        {isSending && (
          <p id="briefing-pending" role="status">
            AIの次の質問とブリーフ案を確認しています…
          </p>
        )}
        {sendError && (
          <Alert id="briefing-command-error" role="alert" variant="destructive">
            <AlertCircle />
            <AlertTitle>壁打ちを続けられません</AlertTitle>
            <AlertDescription>{sendError.message}</AlertDescription>
          </Alert>
        )}
        <Button
          disabled={isSending || isStopping}
          id="send-briefing-message-button"
          type="submit"
        >
          送信
        </Button>
      </form>
    </section>
  );
}
