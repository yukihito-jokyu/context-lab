export type DerivationBriefingMessage = {
  role: string;
  content: string;
  sequenceNo: number;
  createdAt: string;
};

export type DerivationBriefingSuggestion = {
  id: string;
  versionNo: number;
  purpose: string;
  decision: string;
  hypothesis?: string;
  candidatePrompts: string[];
  evaluationCriteria: string;
  environmentConditions: string;
  initialInput: string;
  successCriteria: string;
  requiredConditions: string;
  openQuestion?: string;
  createdAt: string;
};

export type DerivationBriefing = {
  state: string;
  messages: DerivationBriefingMessage[];
  latestSuggestion?: DerivationBriefingSuggestion;
  lastConfirmedAt: string;
};

export type GetDerivationBriefingResponse = {
  data?: DerivationBriefing;
  error?: { code: string; message: string };
};

export type GetDerivationBriefingService = (
  briefingSessionId: string,
) => Promise<GetDerivationBriefingResponse>;

export const getDerivationBriefing: GetDerivationBriefingService = (
  briefingSessionId,
) => GetDerivationBriefing(briefingSessionId);

import { GetDerivationBriefing } from "@wails/go/wails/DerivationBriefingsHandler";
