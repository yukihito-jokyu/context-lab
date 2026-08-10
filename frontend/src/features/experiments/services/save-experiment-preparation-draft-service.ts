import { SaveExperimentPreparationDraft } from "@wails/go/wails/ExperimentPreparationsHandler";

export type SaveExperimentPreparationDraftInput = {
  requestId: string;
  experimentId: string;
  purpose: string;
  hypothesis?: string;
  environmentConditions: string;
  initialInput: string;
  prompts: string[];
  evaluationAxes: string;
};

export type SavedExperimentPreparationDraft = {
  experimentId: string;
  state: string;
  purpose: string;
  hypothesis?: string;
  environmentConditions: string;
  initialInput: string;
  prompts: string[];
  evaluationAxes: string;
  savedAt: string;
};

export type SaveExperimentPreparationDraftResponse = {
  data?: SavedExperimentPreparationDraft;
  error?: { code: string; message: string };
};

export type SaveExperimentPreparationDraftService = (
  input: SaveExperimentPreparationDraftInput,
) => Promise<SaveExperimentPreparationDraftResponse>;

export const saveExperimentPreparationDraft: SaveExperimentPreparationDraftService =
  async (input) => {
    const response = await SaveExperimentPreparationDraft(input);

    if (!response.data) {
      return { error: response.error };
    }

    return {
      data: {
        experimentId: response.data.experimentId,
        state: response.data.state,
        purpose: response.data.purpose,
        hypothesis: response.data.hypothesis,
        environmentConditions: response.data.environmentConditions,
        initialInput: response.data.initialInput,
        prompts: response.data.prompts.map((prompt) => prompt.content),
        evaluationAxes: response.data.evaluationAxes,
        savedAt: String(response.data.savedAt),
      },
    };
  };
