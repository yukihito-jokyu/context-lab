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
import type { GetExperimentBriefingService } from "../services/get-experiment-briefing-service";
import type { SendExperimentBriefMessageService } from "../services/send-experiment-brief-message-service";
import type { StartExperimentBriefingService } from "../services/start-experiment-briefing-service";
import { ExperimentBriefCard } from "./ExperimentBriefCard";
import { ExperimentBriefingConversation } from "./ExperimentBriefingConversation";

type ExperimentBriefingDialogProps = {
  getExperimentBriefing: GetExperimentBriefingService;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  sendExperimentBriefMessage: SendExperimentBriefMessageService;
  startExperimentBriefing: StartExperimentBriefingService;
};

export function ExperimentBriefingDialog({
  getExperimentBriefing,
  onOpenChange,
  open,
  sendExperimentBriefMessage,
  startExperimentBriefing,
}: ExperimentBriefingDialogProps) {
  const {
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
    startError,
  } = useExperimentBriefing({
    isOpen: open,
    getExperimentBriefing,
    sendExperimentBriefMessage,
    startExperimentBriefing,
  });

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (isStarting || isSending) return;
        if (!nextOpen) invalidateRefresh();
        onOpenChange(nextOpen);
      }}
      open={open}
    >
      <DialogContent className="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-3xl">
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
          <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(16rem,0.8fr)]">
            <ExperimentBriefingConversation
              briefing={briefing}
              error={refreshError}
              isRefreshing={isRefreshing}
              onRefresh={() => void refreshBriefing()}
              isSending={isSending}
              onSend={(message) => sendBriefingMessage(message)}
              sendError={sendError}
            />
            <ExperimentBriefCard
              briefing={briefing}
              hasRefreshError={Boolean(refreshError)}
              isRefreshing={isRefreshing}
            />
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
