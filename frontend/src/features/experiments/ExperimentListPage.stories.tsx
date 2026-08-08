import type { Meta, StoryObj } from "@storybook/react-vite";
import { ExperimentListPage } from "./ExperimentListPage";

const confirmedAt = "2026-08-08T10:25:00+09:00";
const success = {
  data: {
    experiments: [
      {
        id: "EXP-014",
        purpose: "顧客問い合わせ要約の比較",
        state: "running",
        progressSummary: "run 2/4 実行中",
        updatedAt: confirmedAt,
      },
    ],
    cancelledExperiments: [
      {
        id: "EXP-007",
        purpose: "初期仮説の見直し",
        state: "cancelled",
        progressSummary: "取消済み",
        updatedAt: confirmedAt,
      },
    ],
    resumeSummary: {
      recommendedExperimentId: "EXP-014",
      statusCounts: { running: 1, cancelled: 1 },
    },
    lastConfirmedAt: confirmedAt,
  },
};

const meta = {
  component: ExperimentListPage,
  title: "Features/Experiments/ExperimentListPage",
} satisfies Meta<typeof ExperimentListPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { listExperiments: async () => success },
};
export const Empty: Story = {
  args: {
    listExperiments: async () => ({
      data: {
        experiments: [],
        cancelledExperiments: [],
        resumeSummary: { statusCounts: {} },
        lastConfirmedAt: confirmedAt,
      },
    }),
  },
};
export const Loading: Story = {
  args: { listExperiments: () => new Promise(() => {}) },
};
export const LoadError: Story = {
  args: {
    listExperiments: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "実験一覧を取得できませんでした。",
      },
    }),
  },
};
