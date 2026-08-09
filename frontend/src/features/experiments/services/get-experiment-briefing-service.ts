import { GetExperimentBriefing } from "@wails/go/wails/ExperimentBriefingsHandler";

export type ExperimentBriefingMessage = {
  role: string;
  content: string;
  sequenceNo: number;
  createdAt: string;
};

export type ExperimentBriefing = {
  state: string;
  messages: ExperimentBriefingMessage[];
  latestBrief?: {
    versionId: string;
    decision: string;
    hypothesis?: string;
    successCriteria: string;
    requiredConditions: string;
    openQuestion?: string;
  };
  lastConfirmedAt: string;
};

export type GetExperimentBriefingResponse = {
  data?: ExperimentBriefing;
  error?: { code: string; message: string };
};

export type GetExperimentBriefingService = (
  briefingSessionId: string,
) => Promise<GetExperimentBriefingResponse>;

export const getExperimentBriefing: GetExperimentBriefingService = (
  briefingSessionId,
) => GetExperimentBriefing(briefingSessionId);
