import { GetInsightWorkspace } from "@wails/go/wails/InsightsHandler";

export type InsightEvidenceCandidate = {
  experimentId: string;
  purpose: string;
  evaluationAxes: string;
  conclusionId: string;
  conclusion: string;
  finalizedAt: string;
};

export type InsightSavedConsideration = {
  experimentId: string;
  conclusionId: string;
  content: string;
  finalizedAt: string;
};

export type InsightSummary = {
  id: string;
  statement: string;
  evidenceCount: number;
  createdAt: string;
};

export type InsightWorkspace = {
  evidenceCandidates: InsightEvidenceCandidate[];
  savedConsiderations: InsightSavedConsideration[];
  insights: InsightSummary[];
  lastConfirmedAt?: string;
};

export type GetInsightWorkspaceService = () => Promise<{
  data?: InsightWorkspace;
  error?: { code: string; message: string };
}>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const getInsightWorkspace: GetInsightWorkspaceService = async () => {
  const response = await GetInsightWorkspace();
  if (!response.data) {
    return { error: response.error };
  }
  return {
    data: {
      evidenceCandidates: response.data.evidenceCandidates.map((candidate) => ({
        experimentId: candidate.experimentId,
        purpose: candidate.purpose,
        evaluationAxes: candidate.evaluationAxes,
        conclusionId: candidate.conclusionId,
        conclusion: candidate.conclusion,
        finalizedAt: String(candidate.finalizedAt),
      })),
      savedConsiderations: response.data.savedConsiderations.map(
        (consideration) => ({
          experimentId: consideration.experimentId,
          conclusionId: consideration.conclusionId,
          content: consideration.content,
          finalizedAt: String(consideration.finalizedAt),
        }),
      ),
      insights: response.data.insights.map((insight) => ({
        id: insight.id,
        statement: insight.statement,
        evidenceCount: insight.evidenceCount,
        createdAt: String(insight.createdAt),
      })),
      lastConfirmedAt: response.data.lastConfirmedAt
        ? String(response.data.lastConfirmedAt)
        : undefined,
    },
  };
};
