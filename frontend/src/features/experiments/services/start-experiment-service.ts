import { StartExperiment } from "@wails/go/wails/ExperimentRunsHandler";

export type StartExperimentInput = {
  requestId: string;
  experimentId: string;
};

export type StartedExperiment = {
  experimentId: string;
  operationId: string;
  runs: Array<{
    id: string;
    state: string;
    summary?: string;
    updatedAt: string;
  }>;
  state: "running";
};

export type StartExperimentResponse = {
  data?: StartedExperiment;
  error?: { code: string; message: string };
};

export type StartExperimentService = (
  input: StartExperimentInput,
) => Promise<StartExperimentResponse>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const startExperiment: StartExperimentService = async (input) => {
  const response = await StartExperiment(input);

  if (!response.data) {
    return { error: response.error };
  }

  return {
    data: {
      experimentId: response.data.experimentId,
      operationId: response.data.operationId,
      runs: response.data.runs.map((run) => ({
        id: run.id,
        state: run.state,
        summary: run.summary,
        updatedAt: String(run.updatedAt),
      })),
      state: "running",
    },
  };
};
