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
    purpose?: string;
    hypothesis?: string;
    candidatePrompts?: string[];
    evaluationAxes?: string;
    environmentConditions?: string;
    /** @deprecated WI-005より前の保存済み下書きとの表示互換用。 */
    decision?: string;
    /** @deprecated WI-005より前の保存済み下書きとの表示互換用。 */
    successCriteria?: string;
    /** @deprecated WI-005より前の保存済み下書きとの表示互換用。 */
    requiredConditions?: string;
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
