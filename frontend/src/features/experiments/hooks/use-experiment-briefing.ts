import { useCallback, useEffect, useRef, useState } from "react";
import type { CreateExperimentFromBriefService } from "../services/create-experiment-from-brief-service";
import type {
  ExperimentBriefing,
  GetExperimentBriefingService,
} from "../services/get-experiment-briefing-service";
import type { SendExperimentBriefMessageService } from "../services/send-experiment-brief-message-service";
import type { StartExperimentBriefingService } from "../services/start-experiment-briefing-service";
import type { StopExperimentBriefingService } from "../services/stop-experiment-briefing-service";

type BriefingError = { code: string; message: string };

type UseExperimentBriefingOptions = {
  isOpen: boolean;
  getExperimentBriefing: GetExperimentBriefingService;
  createExperimentFromBrief: CreateExperimentFromBriefService;
  sendExperimentBriefMessage: SendExperimentBriefMessageService;
  startExperimentBriefing: StartExperimentBriefingService;
  stopExperimentBriefing: StopExperimentBriefingService;
};

export function useExperimentBriefing({
  isOpen,
  getExperimentBriefing,
  createExperimentFromBrief,
  sendExperimentBriefMessage,
  startExperimentBriefing,
  stopExperimentBriefing,
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
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<BriefingError>();
  const [isStopping, setIsStopping] = useState(false);
  const [stopError, setStopError] = useState<BriefingError>();
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
      setCreateError(undefined);
      setIsCreating(false);
      setStopError(undefined);
      setIsStopping(false);
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

  const createExperiment = useCallback(async () => {
    const briefVersionId = briefing?.latestBrief?.versionId;
    if (!briefingStart || !briefVersionId || isCreating) return;

    setIsCreating(true);
    setCreateError(undefined);
    try {
      const response = await createExperimentFromBrief(
        crypto.randomUUID(),
        briefingStart.briefingSessionId,
        briefVersionId,
      );
      if (response.data) {
        window.location.assign(
          `/experiments/${encodeURIComponent(response.data.experimentId)}/preparation`,
        );
        return;
      }
      setCreateError(
        response.error ?? {
          code: "UNKNOWN",
          message: "実験を作成できませんでした。もう一度お試しください。",
        },
      );
    } catch {
      setCreateError({
        code: "UNKNOWN",
        message: "実験を作成できませんでした。もう一度お試しください。",
      });
    } finally {
      setIsCreating(false);
    }
  }, [briefing, briefingStart, createExperimentFromBrief, isCreating]);

  const stopBriefing = useCallback(async () => {
    if (!briefingStart || isStopping) return false;

    setIsStopping(true);
    setStopError(undefined);
    try {
      const response = await stopExperimentBriefing(
        crypto.randomUUID(),
        briefingStart.briefingSessionId,
      );
      if (response.data) return true;

      setStopError(
        response.error ?? {
          code: "UNKNOWN",
          message: "壁打ちを終了できませんでした。もう一度お試しください。",
        },
      );
      return false;
    } catch {
      setStopError({
        code: "UNKNOWN",
        message: "壁打ちを終了できませんでした。もう一度お試しください。",
      });
      return false;
    } finally {
      setIsStopping(false);
    }
  }, [briefingStart, isStopping, stopExperimentBriefing]);

  return {
    beginBriefing,
    briefing,
    briefingStart,
    createError,
    createExperiment,
    invalidateRefresh,
    isRefreshing,
    isCreating,
    isSending,
    isStopping,
    isStarting,
    refreshBriefing,
    refreshError,
    sendBriefingMessage,
    sendError,
    sendOperationId,
    startError,
    stopBriefing,
    stopError,
  };
}
