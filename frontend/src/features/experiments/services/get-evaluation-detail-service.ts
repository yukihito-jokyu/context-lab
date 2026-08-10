import { GetEvaluationDetail } from "@wails/go/wails/ExperimentEvaluationDetailsHandler";

export type EvaluationDetail = {
  evaluation: {
    id: string;
    experimentId: string;
    runId: string;
    state: string;
    summary?: string;
    updatedAt: string;
  };
  operation: { id: string; state: string; updatedAt: string };
  evidence: { runSummary: string; evaluationAxes: string };
  result: { status: string; summary?: string; reasonCode?: string };
  failure?: { code: string; occurredAt: string };
  reconciliation: { state: string; lastObservedAt: string };
  lastConfirmedAt: string;
};

export type GetEvaluationDetailResponse = {
  data?: EvaluationDetail;
  error?: { code: string; message: string };
};

export type GetEvaluationDetailService = (
  evaluationId: string,
) => Promise<GetEvaluationDetailResponse>;

// Wails生成bindingを画面から隔離し、SCR-004で表示する安全なDTOだけを返す。
export const getEvaluationDetail: GetEvaluationDetailService = async (
  evaluationId,
) => {
  const response = await GetEvaluationDetail(evaluationId);
  if (!response.data) {
    return { error: response.error };
  }

  const { data } = response;
  return {
    data: {
      evaluation: {
        id: data.evaluation.id,
        experimentId: data.evaluation.experimentId,
        runId: data.evaluation.runId,
        state: data.evaluation.state,
        summary: data.evaluation.summary,
        updatedAt: String(data.evaluation.updatedAt),
      },
      operation: {
        id: data.operation.id,
        state: data.operation.state,
        updatedAt: String(data.operation.updatedAt),
      },
      evidence: {
        runSummary: data.evidence.runSummary,
        evaluationAxes: data.evidence.evaluationAxes,
      },
      result: {
        status: data.result.status,
        summary: data.result.summary,
        reasonCode: data.result.reasonCode,
      },
      failure: data.failure && {
        code: data.failure.code,
        occurredAt: String(data.failure.occurredAt),
      },
      reconciliation: {
        state: data.reconciliation.state,
        lastObservedAt: String(data.reconciliation.lastObservedAt),
      },
      lastConfirmedAt: String(data.lastConfirmedAt),
    },
  };
};
