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

export const getExperimentPreparation: GetExperimentPreparationService = async (
  experimentId,
) => {
  try {
    return await GetExperimentPreparation(experimentId);
  } catch {
    // Wails bridgeの初期化失敗など、handlerに到達しない例外は内部情報を出さない。
    return {
      error: {
        code: "WAILS_BRIDGE_UNAVAILABLE",
        message:
          "アプリとの通信を開始できませんでした。アプリを再起動してから再試行してください。",
      },
    };
  }
};
