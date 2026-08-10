import { GetExperimentComparison } from "@wails/go/wails/ExperimentComparisonsHandler";
export type ExperimentComparison = {
  experiment: { id: string; purpose: string; evaluationAxes: string };
  conclusion?: {
    conclusionId: string;
    conclusion: string;
    state: string;
    finalizedAt: string;
  };
  evaluations: Array<{
    evaluationId: string;
    runId: string;
    state: string;
    runSummary?: string;
    result: { status: string; summary?: string; reasonCode?: string };
    reconciliation: { state: string; lastObservedAt: string };
    updatedAt: string;
  }>;
  lastConfirmedAt: string;
};
export type GetExperimentComparisonService = (id: string) => Promise<{
  data?: ExperimentComparison;
  error?: { code: string; message: string };
}>;
export const getExperimentComparison: GetExperimentComparisonService = async (
  id,
) => {
  const response = await GetExperimentComparison(id);
  if (!response.data) return { error: response.error };
  const d = response.data;
  return {
    data: {
      experiment: {
        id: d.experiment.id,
        purpose: d.experiment.purpose,
        evaluationAxes: d.experiment.evaluationAxes,
      },
      conclusion: d.conclusion
        ? {
            conclusionId: d.conclusion.id,
            conclusion: d.conclusion.content,
            state: d.conclusion.state,
            finalizedAt: String(d.conclusion.finalizedAt),
          }
        : undefined,
      evaluations: d.evaluations.map((e) => ({
        evaluationId: e.evaluationId,
        runId: e.runId,
        state: e.state,
        runSummary: e.runSummary,
        result: {
          status: e.result.status,
          summary: e.result.summary,
          reasonCode: e.result.reasonCode,
        },
        reconciliation: {
          state: e.reconciliation.state,
          lastObservedAt: String(e.reconciliation.lastObservedAt),
        },
        updatedAt: String(e.updatedAt),
      })),
      lastConfirmedAt: String(d.lastConfirmedAt),
    },
  };
};
