import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, within } from "storybook/test";
import { ExperimentWorkspacePage } from "./ExperimentWorkspacePage";

const workspace = {
  experimentId: "EXP-015",
  state: "ready",
  fixedConditions: {
    fixedConditionId: "fixed-1",
    purpose: "問い合わせ要約の品質を比較する",
    hypothesis: "制約を明示すると要点の欠落を減らせる",
    environmentConditions: "同じ入力と評価手順を用いる",
    initialInput: "顧客問い合わせ本文",
    prompts: [{ sequenceNo: 1, content: "短く要約する" }],
    evaluationAxes: "正確性、要点保持",
    fixedAt: "2026-08-10T10:25:00+09:00",
  },
  conditionFixOperation: { operationId: "operation-1" },
  runs: [
    {
      id: "run-1",
      state: "completed",
      summary: "要約を保存しました",
      updatedAt: "2026-08-10T10:25:00+09:00",
    },
  ],
  evaluations: [
    {
      id: "evaluation-1",
      state: "completed",
      summary: "評価を完了しました",
      updatedAt: "2026-08-10T10:25:00+09:00",
    },
  ],
  lastConfirmedAt: "2026-08-10T10:25:00+09:00",
};

const meta = {
  component: ExperimentWorkspacePage,
  title: "Features/Experiments/ExperimentWorkspacePage",
  args: { experimentId: "EXP-015" },
} satisfies Meta<typeof ExperimentWorkspacePage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { getExperimentWorkspace: async () => ({ data: workspace }) },
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByText("固定条件"),
    ).toBeVisible();
  },
};

export const Loading: Story = {
  args: { getExperimentWorkspace: () => new Promise(() => {}) },
};

export const Empty: Story = {
  args: { getExperimentWorkspace: async () => ({}) },
};

export const LoadError: Story = {
  args: {
    getExperimentWorkspace: async () => ({
      error: {
        code: "EXPERIMENT_WORKSPACE_UNAVAILABLE",
        message: "実験ワークスペースを取得できませんでした。",
      },
    }),
  },
};
