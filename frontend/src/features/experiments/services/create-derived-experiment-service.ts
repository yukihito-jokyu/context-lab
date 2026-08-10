export type CreateDerivedExperimentRequest = {
  requestId: string;
  sourceExperimentId: string;
  changes: {
    purpose?: string;
    hypothesis?: string;
    environmentConditions?: string;
    initialInput?: string;
    prompts?: Array<{ sequenceNo: number; content: string }>;
    evaluationAxes?: string;
  };
  reason: string;
};

export type CreateDerivedExperimentService = (
  request: CreateDerivedExperimentRequest,
) => Promise<{
  data?: {
    requestId: string;
    experimentId: string;
    sourceExperimentId: string;
    state: "preparing";
    createdAt: string;
  };
  error?: { code: string; message: string };
}>;

export const createDerivedExperiment: CreateDerivedExperimentService = async (
  request,
) => {
  const changes: domain.DerivedExperimentChanges = {
    Purpose: request.changes.purpose,
    Hypothesis: request.changes.hypothesis,
    EnvironmentConditions: request.changes.environmentConditions,
    InitialInput: request.changes.initialInput,
    Prompts: request.changes.prompts?.map((prompt) => ({
      SequenceNo: prompt.sequenceNo,
      Content: prompt.content,
    })),
    EvaluationAxes: request.changes.evaluationAxes,
  };
  const response = await CreateDerivedExperiment({
    requestId: request.requestId,
    sourceExperimentId: request.sourceExperimentId,
    changes,
    reason: request.reason,
  });
  if (!response.data) {
    return { error: response.error };
  }
  return {
    data: {
      requestId: response.data.requestId,
      experimentId: response.data.experimentId,
      sourceExperimentId: response.data.sourceExperimentId,
      state: response.data.state as "preparing",
      createdAt: response.data.createdAt,
    },
  };
};

import type { domain } from "@wails/go/models";
import { CreateDerivedExperiment } from "@wails/go/wails/CreateDerivedExperimentsHandler";
