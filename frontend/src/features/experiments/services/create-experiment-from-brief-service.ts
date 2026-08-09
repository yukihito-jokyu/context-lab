import { CreateExperimentFromBrief } from "@wails/go/wails/ExperimentBriefingsHandler";

export type CreateExperimentFromBriefResponse = {
  data?: { experimentId: string; state: string };
  error?: { code: string; message: string };
};

export type CreateExperimentFromBriefService = (
  requestId: string,
  briefingSessionId: string,
  briefVersionId: string,
) => Promise<CreateExperimentFromBriefResponse>;

export const createExperimentFromBrief: CreateExperimentFromBriefService = (
  requestId,
  briefingSessionId,
  briefVersionId,
) => CreateExperimentFromBrief(requestId, briefingSessionId, briefVersionId);
