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
