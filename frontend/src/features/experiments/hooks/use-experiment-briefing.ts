import { useCallback, useEffect, useRef, useState } from "react";

import type {
  ExperimentBriefing,
  GetExperimentBriefingService,
} from "../services/get-experiment-briefing-service";
import type { SendExperimentBriefMessageService } from "../services/send-experiment-brief-message-service";
import type { StartExperimentBriefingService } from "../services/start-experiment-briefing-service";

type BriefingError = { code: string; message: string };

type UseExperimentBriefingOptions = {
  isOpen: boolean;
  getExperimentBriefing: GetExperimentBriefingService;
  sendExperimentBriefMessage: SendExperimentBriefMessageService;
  startExperimentBriefing: StartExperimentBriefingService;
};

export function useExperimentBriefing({
  isOpen,
  getExperimentBriefing,
  sendExperimentBriefMessage,
  startExperimentBriefing,
}: UseExperimentBriefingOptions) {
  const [isStarting, setIsStarting] = useState(false);
  const [startError, setStartError] = useState<BriefingError>();
  const [briefingStart, setBriefingStart] = useState<{
    briefingSessionId: string;
    operationId: string;
  }>();
  const [briefing, setBriefing] = useState<ExperimentBriefing>();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [refreshError, setRefreshError] = useState<BriefingError>();
  const [isSending, setIsSending] = useState(false);
  const [sendOperationId, setSendOperationId] = useState<string>();
  const [sendError, setSendError] = useState<BriefingError>();
  const hasStartedForOpenRef = useRef(false);
  const refreshGenerationRef = useRef(0);

  const beginBriefing = useCallback(async () => {
    setIsStarting(true);
    setStartError(undefined);
    setBriefingStart(undefined);
    try {
      const response = await startExperimentBriefing(crypto.randomUUID());
      if (response.data) {
        setBriefingStart(response.data);
        return;
      }
      setStartError(
        response.error ?? {
          code: "UNKNOWN",
          message: "実験設計の開始を受け付けられませんでした。",
        },
      );
    } catch {
      setStartError({
        code: "UNKNOWN",
        message: "実験設計の開始を受け付けられませんでした。",
      });
    } finally {
      setIsStarting(false);
    }
  }, [startExperimentBriefing]);

  const refreshBriefing = useCallback(async () => {
    if (!briefingStart) return;
    const refreshGeneration = ++refreshGenerationRef.current;
    setIsRefreshing(true);
    setRefreshError(undefined);
    try {
      const response = await getExperimentBriefing(
        briefingStart.briefingSessionId,
      );
      if (refreshGeneration !== refreshGenerationRef.current) return;
      if (response.data) {
        setBriefing(response.data);
        return;
      }
      setRefreshError(
        response.error ?? {
          code: "UNKNOWN",
          message: "最新状態を取得できませんでした。",
        },
      );
    } catch {
      if (refreshGeneration !== refreshGenerationRef.current) return;
      setRefreshError({
        code: "UNKNOWN",
        message: "最新状態を取得できませんでした。",
      });
    } finally {
      if (refreshGeneration === refreshGenerationRef.current) {
        setIsRefreshing(false);
      }
    }
  }, [briefingStart, getExperimentBriefing]);

  useEffect(() => {
    if (!isOpen) {
      hasStartedForOpenRef.current = false;
      refreshGenerationRef.current += 1;
      setBriefing(undefined);
      setBriefingStart(undefined);
      setStartError(undefined);
      setRefreshError(undefined);
      setSendOperationId(undefined);
      setIsRefreshing(false);
      return;
    }
    if (hasStartedForOpenRef.current) return;

    hasStartedForOpenRef.current = true;
    void beginBriefing();
  }, [beginBriefing, isOpen]);

  useEffect(() => {
    if (briefingStart) void refreshBriefing();
  }, [briefingStart, refreshBriefing]);

  const invalidateRefresh = useCallback(() => {
    refreshGenerationRef.current += 1;
    setIsRefreshing(false);
  }, []);

  const sendBriefingMessage = useCallback(
    async (message: string) => {
      if (!briefingStart || isSending) return false;

      setIsSending(true);
      setSendError(undefined);
      setSendOperationId(undefined);
      try {
        const response = await sendExperimentBriefMessage(
          crypto.randomUUID(),
          briefingStart.briefingSessionId,
          message,
        );
        if (!response.data) {
          setSendError(
            response.error ?? {
              code: "UNKNOWN",
              message: "壁打ちを続けられませんでした。もう一度お試しください。",
            },
          );
          return false;
        }

        setSendOperationId(response.data.operationId);
        await refreshBriefing();
        return true;
      } catch {
        setSendError({
          code: "UNKNOWN",
          message: "壁打ちを続けられませんでした。もう一度お試しください。",
        });
        return false;
      } finally {
        setIsSending(false);
      }
    },
    [briefingStart, isSending, refreshBriefing, sendExperimentBriefMessage],
  );

  return {
    beginBriefing,
    briefing,
    briefingStart,
    invalidateRefresh,
    isRefreshing,
    isSending,
    isStarting,
    refreshBriefing,
    refreshError,
    sendBriefingMessage,
    sendError,
    sendOperationId,
    startError,
  };
}
