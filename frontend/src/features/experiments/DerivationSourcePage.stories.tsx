import type { Meta, StoryObj } from "@storybook/react-vite";
import { DerivationSourcePage } from "./DerivationSourcePage";

const derivationSource = {
  source: {
    experimentId: "EXP-20",
    purpose: "条件差分を比較する",
    fixedConditions: {
      fixedConditionId: "condition-20",
      purpose: "条件差分を比較する",
      hypothesis: "条件Aが有効である",
      environmentConditions: "Node.js 22",
      initialInput: "入力データ",
      prompts: [
        { sequenceNo: 1, content: "条件Aで実行する" },
        { sequenceNo: 2, content: "結果を評価する" },
      ],
      evaluationAxes: "正確性",
      fixedAt: "2026-08-10T10:00:00Z",
    },
    conclusion: {
      id: "conclusion-20",
      content: "条件Aを採用します。",
      state: "finalized",
      finalizedAt: "2026-08-10T11:00:00Z",
    },
  },
  eligibility: { canCreateDerivedExperiment: true },
};

const briefingServices = {
  startDerivationBriefing: async (
    _requestId: string,
    sourceExperimentId: string,
  ) => ({
    data: {
      briefingSessionId: "derivation-briefing-20",
      operationId: "derivation-operation-20",
      sourceExperimentId,
    },
  }),
  sendDerivationBriefMessage: async () => ({
    data: { operationId: "derivation-message-operation-20" },
  }),
  getDerivationBriefing: async () => ({
    data: {
      state: "started",
      messages: [],
      lastConfirmedAt: "2026-08-10T10:00:00Z",
    },
  }),
  stopDerivationBriefing: async () => ({
    data: { operationId: "derivation-stop-operation-20" },
  }),
};

const meta = {
  component: DerivationSourcePage,
  title: "Features/Experiments/DerivationSourcePage",
  args: {
    experimentId: "EXP-20",
    getDerivationSource: async () => ({ data: derivationSource }),
    briefingServices,
  },
} satisfies Meta<typeof DerivationSourcePage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Eligible: Story = {};

export const Ineligible: Story = {
  args: {
    getDerivationSource: async () => ({
      data: {
        ...derivationSource,
        source: {
          ...derivationSource.source,
          conclusion: undefined,
        },
        eligibility: {
          canCreateDerivedExperiment: false,
          reasonCode: "CONCLUSION_NOT_FINALIZED",
        },
      },
    }),
  },
};

export const Loading: Story = {
  args: {
    getDerivationSource: () => new Promise<never>(() => undefined),
  },
};

export const LoadFailure: Story = {
  args: {
    getDerivationSource: async () => ({
      error: {
        code: "EXPERIMENT_DERIVATION_SOURCE_UNAVAILABLE",
        message: "派生の作成元を取得できませんでした。",
      },
    }),
  },
};

export const BriefingStartFailure: Story = {
  args: {
    briefingServices: {
      ...briefingServices,
      startDerivationBriefing: async () => ({
        error: {
          code: "DERIVATION_BRIEFING_UNAVAILABLE",
          message: "壁打ちを開始できませんでした。",
        },
      }),
    },
  },
};
