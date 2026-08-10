import { SendDerivationBriefMessage } from "@wails/go/wails/DerivationBriefingsHandler";

export type SendDerivationBriefMessageResponse = {
  data?: { operationId: string };
  error?: { code: string; message: string };
};

export type SendDerivationBriefMessageService = (
  requestId: string,
  briefingSessionId: string,
  message: string,
) => Promise<SendDerivationBriefMessageResponse>;

export const sendDerivationBriefMessage: SendDerivationBriefMessageService = (
  requestId,
  briefingSessionId,
  message,
) => SendDerivationBriefMessage(requestId, briefingSessionId, message);
