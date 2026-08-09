import type { Meta, StoryObj } from "@storybook/react-vite";
import { ExperimentListPage } from "./ExperimentListPage";

const confirmedAt = "2026-08-08T10:25:00+09:00";
const briefing = {
  state: "active",
  messages: [
    {
      role: "assistant" as const,
      content: "比較したい対象と成功基準を教えてください。",
      sequenceNo: 1,
      createdAt: confirmedAt,
    },
  ],
  latestBrief: {
    versionId: "brief-v1",
    decision: "問い合わせ要約の品質を比較する",
    hypothesis: "要約の制約を明示すると正確性が向上する",
    successCriteria: "正確性と要点保持を評価する",
    requiredConditions: "同じ入力と評価手順を用いる",
    openQuestion: "評価者の人数を確定する",
  },
  lastConfirmedAt: confirmedAt,
};
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
    getExperimentBriefing: async () => ({ data: briefing }),
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
    getExperimentBriefing: async () => ({ data: briefing }),
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};
export const Loading: Story = {
  args: {
    listExperiments: () => new Promise(() => {}),
    getExperimentBriefing: async () => ({ data: briefing }),
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
    getExperimentBriefing: async () => ({ data: briefing }),
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};

export const BriefingPending: Story = {
  args: {
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    startExperimentBriefing: () => new Promise(() => {}),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};

export const BriefingError: Story = {
  args: {
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
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

export const BriefingRefreshPending: Story = {
  args: {
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: () => new Promise(() => {}),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};

export const BriefingRefreshError: Story = {
  args: {
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "最新状態を取得できませんでした。",
      },
    }),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};
