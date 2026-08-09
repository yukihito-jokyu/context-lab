import { AlertCircle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useExperimentBriefing } from "../hooks/use-experiment-briefing";
import type { CreateExperimentFromBriefService } from "../services/create-experiment-from-brief-service";
import type { GetExperimentBriefingService } from "../services/get-experiment-briefing-service";
import type { SendExperimentBriefMessageService } from "../services/send-experiment-brief-message-service";
import type { StartExperimentBriefingService } from "../services/start-experiment-briefing-service";
import type { StopExperimentBriefingService } from "../services/stop-experiment-briefing-service";
import { ExperimentBriefCard } from "./ExperimentBriefCard";
import { ExperimentBriefingConversation } from "./ExperimentBriefingConversation";

type ExperimentBriefingDialogProps = {
  getExperimentBriefing: GetExperimentBriefingService;
  createExperimentFromBrief: CreateExperimentFromBriefService;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  sendExperimentBriefMessage: SendExperimentBriefMessageService;
  startExperimentBriefing: StartExperimentBriefingService;
  stopExperimentBriefing: StopExperimentBriefingService;
};

export function ExperimentBriefingDialog({
  getExperimentBriefing,
  createExperimentFromBrief,
  onOpenChange,
  open,
  sendExperimentBriefMessage,
  startExperimentBriefing,
  stopExperimentBriefing,
}: ExperimentBriefingDialogProps) {
  const {
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
    startError,
    stopBriefing,
    stopError,
  } = useExperimentBriefing({
    isOpen: open,
    getExperimentBriefing,
    createExperimentFromBrief,
    sendExperimentBriefMessage,
    startExperimentBriefing,
    stopExperimentBriefing,
  });

  const isCloseBlocked = isStarting || isSending || isCreating || isStopping;

  const closeBriefing = async () => {
    if (await stopBriefing()) {
      invalidateRefresh();
      onOpenChange(false);
    }
  };

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (isCloseBlocked) return;
        if (!nextOpen && briefingStart) {
          void closeBriefing();
          return;
        }
        if (!nextOpen) invalidateRefresh();
        onOpenChange(nextOpen);
      }}
      open={open}
    >
      <DialogContent
        closeDisabled={isCloseBlocked}
        className="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-3xl"
        onEscapeKeyDown={(event) => {
          if (isCloseBlocked) event.preventDefault();
        }}
        onPointerDownOutside={(event) => {
          if (isCloseBlocked) event.preventDefault();
        }}
      >
        <DialogHeader>
          <DialogTitle>実験設計を開始</DialogTitle>
          <DialogDescription>
            実験の目的や評価方法を整理する壁打ちを開始します。
          </DialogDescription>
        </DialogHeader>
        {isStarting && (
          <p id="briefing-start-pending" role="status">
            実験設計を開始しています…
          </p>
        )}
        {startError && (
          <Alert id="briefing-start-error" role="alert" variant="destructive">
            <AlertCircle />
            <AlertTitle>実験設計を開始できません</AlertTitle>
            <AlertDescription className="space-y-4">
              <p>{startError.message}</p>
              <Button
                onClick={() => void beginBriefing()}
                type="button"
                variant="outline"
              >
                もう一度試す
              </Button>
            </AlertDescription>
          </Alert>
        )}
        {briefingStart && (
          <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(16rem,0.8fr)]">
              <ExperimentBriefingConversation
                briefing={briefing}
                error={refreshError}
                isRefreshing={isRefreshing}
                onRefresh={() => void refreshBriefing()}
                isSending={isSending}
                isStopping={isStopping}
                onSend={(message) => sendBriefingMessage(message)}
                sendError={sendError}
              />
              <ExperimentBriefCard
                briefing={briefing}
                hasRefreshError={Boolean(refreshError)}
                isRefreshing={isRefreshing}
                isCreating={isCreating}
                isStopping={isStopping}
                createError={createError}
                onCreate={() => void createExperiment()}
              />
            </div>
            {isStopping && (
              <p id="briefing-stop-pending" role="status">
                壁打ちを終了しています…
              </p>
            )}
            {stopError && (
              <Alert
                id="briefing-stop-error"
                role="alert"
                variant="destructive"
              >
                <AlertCircle />
                <AlertTitle>壁打ちを終了できません</AlertTitle>
                <AlertDescription className="space-y-4">
                  <p>{stopError.message}</p>
                  <Button onClick={() => void closeBriefing()} type="button">
                    もう一度試す
                  </Button>
                </AlertDescription>
              </Alert>
            )}
            <div className="flex justify-end">
              <Button
                disabled={isCloseBlocked}
                id="stop-experiment-briefing-button"
                onClick={() => void closeBriefing()}
                type="button"
                variant="outline"
              >
                壁打ちを終了
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
