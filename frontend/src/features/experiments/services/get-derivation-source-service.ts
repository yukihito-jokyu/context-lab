import { GetDerivationSource } from "@wails/go/wails/ExperimentDerivationSourcesHandler";

export type DerivationSource = {
  source: {
    experimentId: string;
    purpose: string;
    fixedConditions?: {
      fixedConditionId: string;
      purpose: string;
      hypothesis?: string;
      environmentConditions: string;
      initialInput: string;
      prompts: Array<{ sequenceNo: number; content: string }>;
      evaluationAxes: string;
      fixedAt: string;
    };
    conclusion?: {
      id: string;
      content: string;
      state: string;
      finalizedAt: string;
    };
  };
  eligibility: {
    canCreateDerivedExperiment: boolean;
    reasonCode?: "CONDITIONS_NOT_FIXED" | "CONCLUSION_NOT_FINALIZED";
  };
};

export type GetDerivationSourceService = (experimentId: string) => Promise<{
  data?: DerivationSource;
  error?: { code: string; message: string };
}>;

export const getDerivationSource: GetDerivationSourceService = async (
  experimentId,
) => {
  const response = await GetDerivationSource(experimentId);
  if (!response.data) {
    return { error: response.error };
  }
  const data = response.data;
  return {
    data: {
      source: {
        experimentId: data.source.experimentId,
        purpose: data.source.purpose,
        fixedConditions: data.source.fixedConditions
          ? {
              fixedConditionId: data.source.fixedConditions.fixedConditionId,
              purpose: data.source.fixedConditions.purpose,
              hypothesis: data.source.fixedConditions.hypothesis,
              environmentConditions:
                data.source.fixedConditions.environmentConditions,
              initialInput: data.source.fixedConditions.initialInput,
              prompts: data.source.fixedConditions.prompts.map((prompt) => ({
                sequenceNo: prompt.sequenceNo,
                content: prompt.content,
              })),
              evaluationAxes: data.source.fixedConditions.evaluationAxes,
              fixedAt: String(data.source.fixedConditions.fixedAt),
            }
          : undefined,
        conclusion: data.source.conclusion
          ? {
              id: data.source.conclusion.id,
              content: data.source.conclusion.content,
              state: data.source.conclusion.state,
              finalizedAt: String(data.source.conclusion.finalizedAt),
            }
          : undefined,
      },
      eligibility: {
        canCreateDerivedExperiment: data.eligibility.canCreateDerivedExperiment,
        reasonCode: data.eligibility.reasonCode as
          | "CONDITIONS_NOT_FIXED"
          | "CONCLUSION_NOT_FINALIZED"
          | undefined,
      },
    },
  };
};
