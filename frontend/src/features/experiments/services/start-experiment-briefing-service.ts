import type { wails } from "@wails/go/models";
import { StartExperimentBriefing } from "@wails/go/wails/ExperimentBriefingsHandler";

export type StartExperimentBriefingService = (
  requestId: string,
) => Promise<wails.StartExperimentBriefingResponse>;

export const startExperimentBriefing: StartExperimentBriefingService = (
  requestId,
) => StartExperimentBriefing(requestId);
