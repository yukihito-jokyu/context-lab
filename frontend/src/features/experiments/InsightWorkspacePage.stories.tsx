import type { Meta, StoryObj } from "@storybook/react-vite";
import { InsightWorkspacePage } from "./InsightWorkspacePage";

const workspace = {
  evidenceCandidates: [
    {
      experimentId: "EXP-26-A",
      purpose: "条件差分を比較する",
      evaluationAxes: "正確性",
      conclusionId: "CON-26-A",
      conclusion: "条件Aは検証可能性を高める。",
      finalizedAt: "2026-08-10T10:00:00Z",
    },
    {
      experimentId: "EXP-26-B",
      purpose: "別の条件差分を比較する",
      evaluationAxes: "再現性",
      conclusionId: "CON-26-B",
      conclusion: "条件Bでは追加検証が必要である。",
      finalizedAt: "2026-08-10T11:00:00Z",
    },
  ],
  savedConsiderations: [
    {
      experimentId: "EXP-26-A",
      conclusionId: "CON-26-A",
      content: "条件Aは検証可能性を高める。",
      finalizedAt: "2026-08-10T10:00:00Z",
    },
  ],
  insights: [],
  lastConfirmedAt: "2026-08-10T11:00:00Z",
};

const meta = {
  component: InsightWorkspacePage,
  title: "Features/Experiments/InsightWorkspacePage",
  args: {
    initialExperimentId: "EXP-26-A",
    getInsightWorkspace: async () => ({ data: workspace }),
  },
} satisfies Meta<typeof InsightWorkspacePage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Ready: Story = {};

export const Empty: Story = {
  args: {
    initialExperimentId: undefined,
    getInsightWorkspace: async () => ({
      data: { evidenceCandidates: [], savedConsiderations: [], insights: [] },
    }),
  },
};

export const Loading: Story = {
  args: { getInsightWorkspace: () => new Promise<never>(() => undefined) },
};

export const LoadFailure: Story = {
  args: {
    getInsightWorkspace: async () => ({
      error: {
        code: "INSIGHT_WORKSPACE_UNAVAILABLE",
        message: "知見のワークスペースを取得できませんでした。",
      },
    }),
  },
};
