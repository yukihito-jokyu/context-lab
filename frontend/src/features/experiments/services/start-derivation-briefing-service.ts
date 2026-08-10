import { StartDerivationBriefing } from "@wails/go/wails/DerivationBriefingsHandler";

export type StartDerivationBriefingResponse = {
  data?: {
    briefingSessionId: string;
    operationId: string;
    sourceExperimentId: string;
  };
  error?: { code: string; message: string };
};

export type StartDerivationBriefingService = (
  requestId: string,
  sourceExperimentId: string,
) => Promise<StartDerivationBriefingResponse>;

export const startDerivationBriefing: StartDerivationBriefingService = (
  requestId,
  sourceExperimentId,
) => StartDerivationBriefing(requestId, sourceExperimentId);
