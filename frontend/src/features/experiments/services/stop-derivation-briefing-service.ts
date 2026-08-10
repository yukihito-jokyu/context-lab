import { StopDerivationBriefing } from "@wails/go/wails/DerivationBriefingsHandler";

export type StopDerivationBriefingResponse = {
  data?: { operationId: string };
  error?: { code: string; message: string };
};

export type StopDerivationBriefingService = (
  requestId: string,
  briefingSessionId: string,
) => Promise<StopDerivationBriefingResponse>;

export const stopDerivationBriefing: StopDerivationBriefingService = (
  requestId,
  briefingSessionId,
) => StopDerivationBriefing(requestId, briefingSessionId);
