import { RetryEndedRun } from "@wails/go/wails/ExperimentRunRetriesHandler";

export type RetryEndedRunInput = {
  requestId: string;
  runId: string;
};

export type RetriedEndedRun = {
  sourceRunId: string;
  experimentId: string;
  retryRunId: string;
  operationId: string;
  state: "queued";
  createdAt: string;
};

export type RetryEndedRunResponse = {
  data?: RetriedEndedRun;
  error?: { code: string; message: string };
};

export type RetryEndedRunService = (
  input: RetryEndedRunInput,
) => Promise<RetryEndedRunResponse>;

// Wails生成bindingを画面から隔離し、再実行確認に必要な安全DTOだけを返す。
export const retryEndedRun: RetryEndedRunService = async (input) => {
  const response = await RetryEndedRun(input);
  if (!response.data) {
    return { error: response.error };
  }

  return {
    data: {
      sourceRunId: response.data.sourceRunId,
      experimentId: response.data.experimentId,
      retryRunId: response.data.retryRunId,
      operationId: response.data.operationId,
      state: "queued",
      createdAt: String(response.data.createdAt),
    },
  };
};
