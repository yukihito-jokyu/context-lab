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
  args: {
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
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
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};
export const Loading: Story = {
  args: {
    listExperiments: () => new Promise(() => {}),
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};
export const LoadError: Story = {
  args: {
    listExperiments: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "実験一覧を取得できませんでした。",
      },
    }),
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};

export const BriefingPending: Story = {
  args: {
    listExperiments: async () => success,
    startExperimentBriefing: () => new Promise(() => {}),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};

export const BriefingError: Story = {
  args: {
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "実験設計を開始できませんでした。",
      },
    }),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};
