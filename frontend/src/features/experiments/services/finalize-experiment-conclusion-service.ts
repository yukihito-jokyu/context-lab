export type FinalizeExperimentConclusionInput = {
  requestId: string;
  experimentId: string;
  conclusion: string;
};

export type FinalizedExperimentConclusion = {
  requestId: string;
  experimentId: string;
  conclusionId: string;
  conclusion: string;
  state: "finalized";
  finalizedAt: string;
};

export type FinalizeExperimentConclusionResponse = {
  data?: FinalizedExperimentConclusion;
  error?: { code: string; message: string };
};

export type FinalizeExperimentConclusionService = (
  input: FinalizeExperimentConclusionInput,
) => Promise<FinalizeExperimentConclusionResponse>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const finalizeExperimentConclusion: FinalizeExperimentConclusionService =
  async (input) => {
    const response = await FinalizeExperimentConclusion(input);

    if (!response.data) {
      return { error: response.error };
    }

    return {
      data: {
        requestId: response.data.requestId,
        experimentId: response.data.experimentId,
        conclusionId: response.data.conclusionId,
        conclusion: response.data.conclusion,
        state: "finalized",
        finalizedAt: String(response.data.finalizedAt),
      },
    };
  };

import { FinalizeExperimentConclusion } from "@wails/go/wails/FinalizeExperimentConclusionsHandler";
