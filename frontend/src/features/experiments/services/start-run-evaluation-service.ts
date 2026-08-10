import { StartRunEvaluation } from "@wails/go/wails/ExperimentEvaluationsHandler";

export type StartRunEvaluationInput = {
  requestId: string;
  runId: string;
};

export type StartedRunEvaluation = {
  evaluationId: string;
  operationId: string;
  runId: string;
  state: string;
};

export type StartRunEvaluationResponse = {
  data?: StartedRunEvaluation;
  error?: { code: string; message: string };
};

export type StartRunEvaluationService = (
  input: StartRunEvaluationInput,
) => Promise<StartRunEvaluationResponse>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const startRunEvaluation: StartRunEvaluationService = async (input) => {
  const response = await StartRunEvaluation(input);

  if (!response.data) {
    return { error: response.error };
  }

  return {
    data: {
      evaluationId: response.data.evaluationId,
      operationId: response.data.operationId,
      runId: response.data.runId,
      state: response.data.state,
    },
  };
};
