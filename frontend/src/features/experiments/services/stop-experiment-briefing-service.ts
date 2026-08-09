import { StopExperimentBriefing } from "@wails/go/wails/ExperimentBriefingsHandler";

export type StopExperimentBriefingResponse = {
  data?: { operationId: string };
  error?: { code: string; message: string };
};

export type StopExperimentBriefingService = (
  requestId: string,
  briefingSessionId: string,
) => Promise<StopExperimentBriefingResponse>;

export const stopExperimentBriefing: StopExperimentBriefingService = (
  requestId,
  briefingSessionId,
) => StopExperimentBriefing(requestId, briefingSessionId);
