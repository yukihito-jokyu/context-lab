import { CreateInsight } from "@wails/go/wails/InsightsHandler";

export type InsightEvidence = {
  experimentId: string;
  conclusionId: string;
};

export type CreateInsightRequest = {
  requestId: string;
  evidences: InsightEvidence[];
  statement: string;
  applicabilityConditions: string;
  verificationGaps: string;
};

export type CreatedInsight = CreateInsightRequest & {
  insightId: string;
  createdAt: string;
};

export type CreateInsightService = (request: CreateInsightRequest) => Promise<{
  data?: CreatedInsight;
  error?: { code: string; message: string };
}>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const createInsight: CreateInsightService = async (request) => {
  const response = await CreateInsight(request);
  if (!response.data) {
    return { error: response.error };
  }
  return {
    data: {
      requestId: response.data.requestId,
      insightId: response.data.insightId,
      evidences: response.data.evidences.map((evidence) => ({
        experimentId: evidence.experimentId,
        conclusionId: evidence.conclusionId,
      })),
      statement: response.data.statement,
      applicabilityConditions: response.data.applicabilityConditions,
      verificationGaps: response.data.verificationGaps,
      createdAt: String(response.data.createdAt),
    },
  };
};
