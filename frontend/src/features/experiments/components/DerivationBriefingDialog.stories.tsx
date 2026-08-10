import type { Meta, StoryObj } from "@storybook/react-vite";
import { DerivationBriefingDialog } from "./DerivationBriefingDialog";

const meta = {
  component: DerivationBriefingDialog,
  title: "Features/Experiments/DerivationBriefingDialog",
  args: {
    onOpenChange: () => undefined,
    open: true,
    sourceExperimentId: "EXP-22",
    startDerivationBriefing: async (_requestId, sourceExperimentId) => ({
      data: {
        briefingSessionId: "derivation-briefing-22",
        operationId: "derivation-operation-22",
        sourceExperimentId,
      },
    }),
    sendDerivationBriefMessage: async () => ({
      data: { operationId: "derivation-message-operation-22" },
    }),
    getDerivationBriefing: async () => ({
      data: {
        state: "started",
        lastConfirmedAt: "2026-08-10T09:00:00Z",
        messages: [
          {
            role: "user",
            content: "比較条件を変えたいです。",
            sequenceNo: 1,
            createdAt: "2026-08-10T09:00:00Z",
          },
          {
            role: "assistant",
            content: "評価軸を保ったまま条件を調整しましょう。",
            sequenceNo: 2,
            createdAt: "2026-08-10T09:00:02Z",
          },
        ],
        latestSuggestion: {
          id: "suggestion-22",
          versionNo: 1,
          purpose: "比較条件の変更を確認する",
          decision: "比較対象を一つ増やす",
          hypothesis: "対象を増やすと差異を確認できる",
          candidatePrompts: ["候補A", "候補B"],
          evaluationCriteria: "正確性",
          environmentConditions: "同じ環境",
          initialInput: "既存入力",
          successCriteria: "差異が確認できる",
          requiredConditions: "条件を固定する",
          openQuestion: "対象数を決める必要があります。",
          createdAt: "2026-08-10T09:00:02Z",
        },
      },
    }),
  },
} satisfies Meta<typeof DerivationBriefingDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Started: Story = {};

export const Pending: Story = {
  args: {
    startDerivationBriefing: () => new Promise<never>(() => undefined),
  },
};

export const StartFailure: Story = {
  args: {
    startDerivationBriefing: async () => ({
      error: {
        code: "DERIVATION_BRIEFING_START_FAILED",
        message: "壁打ちを開始できませんでした。",
      },
    }),
  },
};

export const SendFailure: Story = {
  args: {
    sendDerivationBriefMessage: async () => ({
      error: {
        code: "DERIVATION_BRIEFING_MESSAGE_UNAVAILABLE",
        message: "壁打ちメッセージを送信できませんでした。",
      },
    }),
  },
};

export const SendPending: Story = {
  args: {
    sendDerivationBriefMessage: () => new Promise<never>(() => undefined),
  },
};

export const RefreshFailure: Story = {
  args: {
    getDerivationBriefing: async () => ({
      error: {
        code: "DERIVATION_BRIEFING_NOT_FOUND",
        message: "壁打ち内容を取得できませんでした。",
      },
    }),
  },
};

export const Empty: Story = {
  args: {
    getDerivationBriefing: async () => ({
      data: {
        state: "started",
        messages: [],
        lastConfirmedAt: "2026-08-10T09:00:00Z",
      },
    }),
  },
};
