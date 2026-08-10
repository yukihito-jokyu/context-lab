import { GetRunDetail } from "@wails/go/wails/ExperimentRunDetailsHandler";

export type RunDetail = {
  run: {
    id: string;
    experimentId: string;
    state: string;
    summary?: string;
    updatedAt: string;
  };
  fixedPrompt: { sequenceNo: number; content: string };
  operation: { id: string; state: string; updatedAt: string };
  observations: Array<{
    sequenceNo: number;
    kind: string;
    occurredAt: string;
    summary: string;
  }>;
  artifacts: {
    status: string;
    items: Array<{ digest: string; label?: string; status: string }>;
    reasonCode?: string;
  };
  failure?: { code: string; occurredAt: string; partialSummary?: string };
  reconciliation: { state: string; lastObservedAt: string };
  lastConfirmedAt: string;
};

export type GetRunDetailResponse = {
  data?: RunDetail;
  error?: { code: string; message: string };
};
export type GetRunDetailService = (
  runId: string,
) => Promise<GetRunDetailResponse>;

// Wails生成bindingを画面から隔離し、SCR-004が使う安全な表示DTOだけを返す。
export const getRunDetail: GetRunDetailService = async (runId) => {
  const response = await GetRunDetail(runId);
  if (!response.data) return { error: response.error };
  const { data } = response;
  return {
    data: {
      run: {
        id: data.run.id,
        experimentId: data.run.experimentId,
        state: data.run.state,
        summary: data.run.summary,
        updatedAt: String(data.run.updatedAt),
      },
      fixedPrompt: {
        sequenceNo: data.fixedPrompt.sequenceNo,
        content: data.fixedPrompt.content,
      },
      operation: {
        id: data.operation.id,
        state: data.operation.state,
        updatedAt: String(data.operation.updatedAt),
      },
      observations: data.observations.map((observation) => ({
        sequenceNo: observation.sequenceNo,
        kind: observation.kind,
        occurredAt: String(observation.occurredAt),
        summary: observation.summary,
      })),
      artifacts: {
        status: data.artifacts.status,
        items: data.artifacts.items.map((artifact) => ({
          digest: artifact.digest,
          label: artifact.label,
          status: artifact.status,
        })),
        reasonCode: data.artifacts.reasonCode,
      },
      failure: data.failure && {
        code: data.failure.code,
        occurredAt: String(data.failure.occurredAt),
        partialSummary: data.failure.partialSummary,
      },
      reconciliation: {
        state: data.reconciliation.state,
        lastObservedAt: String(data.reconciliation.lastObservedAt),
      },
      lastConfirmedAt: String(data.lastConfirmedAt),
    },
  };
};
