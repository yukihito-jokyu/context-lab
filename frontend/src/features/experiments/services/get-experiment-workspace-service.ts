import { GetExperimentWorkspace } from "@wails/go/wails/ExperimentWorkspacesHandler";

export type WorkspacePrompt = {
  sequenceNo: number;
  content: string;
};

export type ExperimentWorkspace = {
  experimentId: string;
  state: string;
  fixedConditions: {
    fixedConditionId: string;
    purpose: string;
    hypothesis?: string;
    environmentConditions: string;
    initialInput: string;
    prompts: WorkspacePrompt[];
    evaluationAxes: string;
    fixedAt: string;
  };
  conditionFixOperation: {
    operationId: string;
  };
  runs: Array<{
    id: string;
    state: string;
    summary?: string;
    updatedAt: string;
  }>;
  evaluations: Array<{
    id: string;
    state: string;
    summary?: string;
    updatedAt: string;
  }>;
  lastConfirmedAt: string;
};

export type GetExperimentWorkspaceResponse = {
  data?: ExperimentWorkspace;
  error?: { code: string; message: string };
};

export type GetExperimentWorkspaceService = (
  experimentId: string,
) => Promise<GetExperimentWorkspaceResponse>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const getExperimentWorkspace: GetExperimentWorkspaceService = async (
  experimentId,
) => {
  const response = await GetExperimentWorkspace(experimentId);

  if (!response.data) {
    return { error: response.error };
  }

  return {
    data: {
      experimentId: response.data.experimentId,
      state: response.data.state,
      fixedConditions: {
        fixedConditionId: response.data.fixedConditions.fixedConditionId,
        purpose: response.data.fixedConditions.purpose,
        hypothesis: response.data.fixedConditions.hypothesis,
        environmentConditions:
          response.data.fixedConditions.environmentConditions,
        initialInput: response.data.fixedConditions.initialInput,
        prompts: response.data.fixedConditions.prompts.map((prompt) => ({
          sequenceNo: prompt.sequenceNo,
          content: prompt.content,
        })),
        evaluationAxes: response.data.fixedConditions.evaluationAxes,
        fixedAt: String(response.data.fixedConditions.fixedAt),
      },
      conditionFixOperation: {
        operationId: response.data.conditionFixOperation.operationId,
      },
      runs: response.data.runs.map((run) => ({
        id: run.id,
        state: run.state,
        summary: run.summary,
        updatedAt: String(run.updatedAt),
      })),
      evaluations: response.data.evaluations.map((evaluation) => ({
        id: evaluation.id,
        state: evaluation.state,
        summary: evaluation.summary,
        updatedAt: String(evaluation.updatedAt),
      })),
      lastConfirmedAt: String(response.data.lastConfirmedAt),
    },
  };
};
