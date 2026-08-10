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
  args: {
    experimentId: "EXP-015",
    startExperiment: async () => ({
      data: {
        experimentId: "EXP-015",
        operationId: "start-operation-1",
        runs: [
          {
            id: "run-1",
            state: "running",
            summary: "要約を実行中です",
            updatedAt: "2026-08-10T10:25:00+09:00",
          },
        ],
        state: "running",
      },
    }),
  },
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

export const Starting: Story = {
  args: {
    getExperimentWorkspace: async () => ({ data: workspace }),
    startExperiment: () => new Promise(() => {}),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(
      await canvas.findByRole("button", { name: "実験を開始" }),
    );
    await expect(
      await canvas.findByRole("button", { name: "実験を開始しています…" }),
    ).toBeDisabled();
  },
};

export const StartError: Story = {
  args: {
    getExperimentWorkspace: async () => ({ data: workspace }),
    startExperiment: async () => ({
      error: {
        code: "START_EXPERIMENT_FAILED",
        message: "開始に失敗しました。",
      },
    }),
  },
};
