import type { Meta, StoryObj } from "@storybook/react-vite";
import { ComparisonPage } from "./ComparisonPage";

const data = {
  experiment: { id: "EXP-18", purpose: "比較", evaluationAxes: "正確性" },
  evaluations: [
    {
      evaluationId: "eval-1",
      runId: "run-1",
      state: "completed",
      runSummary: "根拠",
      result: { status: "complete", summary: "結果" },
      reconciliation: {
        state: "confirmed",
        lastObservedAt: "2026-08-10T10:00:00Z",
      },
      updatedAt: "2026-08-10T10:00:00Z",
    },
  ],
  lastConfirmedAt: "2026-08-10T10:00:00Z",
};
const meta = {
  component: ComparisonPage,
  title: "Features/Experiments/ComparisonPage",
  args: { experimentId: "EXP-18" },
} satisfies Meta<typeof ComparisonPage>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Success: Story = {
  args: { getExperimentComparison: async () => ({ data }) },
};
export const Empty: Story = {
  args: {
    getExperimentComparison: async () => ({
      data: { ...data, evaluations: [] },
    }),
  },
};
export const Reconciling: Story = {
  args: {
    getExperimentComparison: async () => ({
      data: {
        ...data,
        evaluations: [
          {
            ...data.evaluations[0],
            reconciliation: {
              state: "reconciling",
              lastObservedAt: data.lastConfirmedAt,
            },
          },
        ],
      },
    }),
  },
};
export const LoadError: Story = {
  args: {
    getExperimentComparison: async () => ({
      error: { code: "COMPARISON_UNAVAILABLE", message: "取得できません" },
    }),
  },
};
