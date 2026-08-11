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

export const experimentPreparationBridgeUnavailableError: {
  code: string;
  message: string;
} = {
  code: "WAILS_BRIDGE_UNAVAILABLE",
  message:
    "アプリとの通信機能を開始できませんでした。アプリを完全に終了してから再起動してください。実験データを消去する必要はありません。",
};

export const experimentPreparationHandlerUnavailableError: {
  code: string;
  message: string;
} = {
  code: "WAILS_HANDLER_UNAVAILABLE",
  message:
    "この画面とアプリ本体の通信機能が一致していません。開発中は開発サーバーを完全に停止してから起動し直してください。実験データを消去する必要はありません。",
};

export const experimentPreparationBridgeCallFailedError: {
  code: string;
  message: string;
} = {
  code: "WAILS_BRIDGE_CALL_FAILED",
  message:
    "アプリ本体との通信が完了しませんでした。アプリを完全に終了してから再起動し、再読込してください。実験データを消去する必要はありません。",
};

type ExperimentPreparationsBridge = {
  GetExperimentPreparation?: (experimentId: string) => unknown;
};

type WailsBridge = {
  wails?: {
    ExperimentPreparationsHandler?: ExperimentPreparationsBridge;
  };
};

const bridgeInitialisationTimeoutMilliseconds = 1_000;
const bridgeInitialisationPollMilliseconds = 25;

function findExperimentPreparationsBridge():
  | ExperimentPreparationsBridge
  | undefined {
  return (globalThis as typeof globalThis & { go?: WailsBridge }).go?.wails
    ?.ExperimentPreparationsHandler;
}

async function waitForExperimentPreparationsBridge(): Promise<
  ExperimentPreparationsBridge | undefined
> {
  const deadline = Date.now() + bridgeInitialisationTimeoutMilliseconds;
  let bridge = findExperimentPreparationsBridge();
  while (!bridge && Date.now() < deadline) {
    await new Promise((resolve) => {
      window.setTimeout(resolve, bridgeInitialisationPollMilliseconds);
    });
    bridge = findExperimentPreparationsBridge();
  }
  return bridge;
}

export const getExperimentPreparation: GetExperimentPreparationService = async (
  experimentId,
) => {
  const bridge = await waitForExperimentPreparationsBridge();
  if (!bridge) {
    return { error: experimentPreparationBridgeUnavailableError };
  }
  if (typeof bridge.GetExperimentPreparation !== "function") {
    return { error: experimentPreparationHandlerUnavailableError };
  }

  try {
    return await GetExperimentPreparation(experimentId);
  } catch {
    // handler到達前後の例外は内部情報を出さず、回復可能な分類だけを返す。
    return {
      error: experimentPreparationBridgeCallFailedError,
    };
  }
};
