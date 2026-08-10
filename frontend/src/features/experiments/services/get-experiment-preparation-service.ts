import { GetExperimentPreparation } from "@wails/go/wails/ExperimentPreparationsHandler";

export type ExperimentPreparationPrompt = {
  sequenceNo: number;
  content: string;
};

export type ExperimentPreparationSource = {
  state: string;
  versionId: string;
};

export type ExperimentPreparationRequiredFields = {
  purpose: boolean;
  environmentConditions: boolean;
  initialInput: boolean;
  prompts: boolean;
  evaluationAxes: boolean;
};

export type ExperimentPreparation = {
  experimentId: string;
  state: string;
  purpose: string;
  hypothesis?: string;
  environmentConditions: string;
  initialInput: string;
  prompts: ExperimentPreparationPrompt[];
  evaluationAxes: string;
  /** 画面表示用の採用元。session IDなどの内部識別子は含めない。 */
  source: ExperimentPreparationSource;
  requiredFields: ExperimentPreparationRequiredFields;
  lastConfirmedAt: string;
};

export type GetExperimentPreparationResponse = {
  data?: ExperimentPreparation;
  error?: { code: string; message: string };
};

export type GetExperimentPreparationService = (
  experimentId: string,
) => Promise<GetExperimentPreparationResponse>;

export const getExperimentPreparation: GetExperimentPreparationService = (
  experimentId,
) => GetExperimentPreparation(experimentId);
