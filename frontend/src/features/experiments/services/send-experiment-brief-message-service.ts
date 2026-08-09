import { SendExperimentBriefMessage } from "@wails/go/wails/ExperimentBriefingsHandler";

export type SendExperimentBriefMessageResponse = {
  data?: { operationId: string };
  error?: { code: string; message: string };
};

export type SendExperimentBriefMessageService = (
  requestId: string,
  briefingSessionId: string,
  message: string,
) => Promise<SendExperimentBriefMessageResponse>;

export const sendExperimentBriefMessage: SendExperimentBriefMessageService = (
  requestId,
  briefingSessionId,
  message,
) => SendExperimentBriefMessage(requestId, briefingSessionId, message);
