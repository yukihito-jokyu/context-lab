import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, userEvent, within } from "storybook/test";
import { FinalizeExperimentConclusionPanel } from "./FinalizeExperimentConclusionPanel";

const meta = {
  component: FinalizeExperimentConclusionPanel,
  title: "Features/Experiments/FinalizeExperimentConclusionPanel",
  args: {
    experimentId: "EXP-19",
    onReload: async () => undefined,
    finalizeExperimentConclusion: async (input) => ({
      data: {
        requestId: input.requestId,
        experimentId: input.experimentId,
        conclusionId: "conclusion-19",
        conclusion: input.conclusion,
        state: "finalized",
        finalizedAt: "2026-08-10T10:00:00Z",
      },
    }),
  },
} satisfies Meta<typeof FinalizeExperimentConclusionPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Initial: Story = {};

export const Finalized: Story = {
  args: {
    existingConclusion: {
      conclusionId: "conclusion-19",
      conclusion: "比較結果から条件Aを採用します。",
      state: "finalized",
      finalizedAt: "2026-08-10T10:00:00Z",
    },
  },
};

export const SubmissionFailure: Story = {
  args: {
    finalizeExperimentConclusion: async () => ({
      error: { code: "UNAVAILABLE", message: "結論を確定できませんでした。" },
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(
      canvas.getByLabelText("根拠を確認した結論"),
      "結論を記録します。",
    );
    await userEvent.click(canvas.getByRole("button", { name: "結論を確定" }));
    await expect(
      await canvas.findByText("結論を確定できませんでした。"),
    ).toBeVisible();
  },
};

export const SubmissionPending: Story = {
  args: {
    finalizeExperimentConclusion: () => new Promise<never>(() => undefined),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(
      canvas.getByLabelText("根拠を確認した結論"),
      "結論を記録します。",
    );
    await userEvent.click(canvas.getByRole("button", { name: "結論を確定" }));
    await expect(
      canvas.getByRole("button", { name: "結論を確定しています…" }),
    ).toBeDisabled();
  },
};
