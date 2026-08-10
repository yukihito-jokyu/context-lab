import { FixExperimentConditions } from "@wails/go/wails/ExperimentPreparationsHandler";

export type FixExperimentConditionsInput = {
  requestId: string;
  experimentId: string;
  purpose: string;
  hypothesis?: string;
  environmentConditions: string;
  initialInput: string;
  prompts: string[];
  evaluationAxes: string;
};

export type FixedExperimentConditions = {
  experimentId: string;
  state: "ready";
  fixedConditionId: string;
  operationId: string;
  fixedAt: string;
};

export type FixExperimentConditionsResponse = {
  data?: FixedExperimentConditions;
  error?: {
    code: string;
    message: string;
    fieldErrors?: Record<string, string>;
  };
};

export type FixExperimentConditionsService = (
  input: FixExperimentConditionsInput,
) => Promise<FixExperimentConditionsResponse>;

export const fixExperimentConditions: FixExperimentConditionsService = async (
  input,
) => {
  const response = await FixExperimentConditions(input);

  if (!response.data) {
    return { error: response.error };
  }

  return {
    data: {
      experimentId: response.data.experimentId,
      state: "ready",
      fixedConditionId: response.data.fixedConditionId,
      operationId: response.data.operationId,
      fixedAt: String(response.data.fixedAt),
    },
  };
};
