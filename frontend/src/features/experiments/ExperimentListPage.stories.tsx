import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, fn, userEvent, within } from "storybook/test";
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
const sendBriefingMessage = async () => ({
  data: { operationId: "operation-2" },
});
const createExperimentFromBrief = async () => ({
  data: { experimentId: "EXP-015", state: "preparing" },
});
const stopExperimentBriefing = async () => ({
  data: { operationId: "operation-3" },
});

const meta = {
  component: ExperimentListPage,
  title: "Features/Experiments/ExperimentListPage",
  args: { stopExperimentBriefing },
} satisfies Meta<typeof ExperimentListPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};
export const Empty: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => ({
      data: {
        experiments: [],
        cancelledExperiments: [],
        resumeSummary: { statusCounts: {} },
        lastConfirmedAt: confirmedAt,
      },
    }),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};
export const Loading: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: () => new Promise(() => {}),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};
export const LoadError: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "実験一覧を取得できませんでした。",
      },
    }),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
};

export const BriefingPending: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: () => new Promise(() => {}),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};

export const BriefingError: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
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
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    sendExperimentBriefMessage: sendBriefingMessage,
    getExperimentBriefing: () => new Promise(() => {}),
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
  },
};

export const BriefingRefreshError: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    sendExperimentBriefMessage: sendBriefingMessage,
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

export const BriefingMessagePending: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: () => new Promise(() => {}),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    const messageInput = within(dialog).getByLabelText("AIへの回答");
    await userEvent.type(messageInput, "成功条件を先に決めたいです。");
    await userEvent.click(within(dialog).getByRole("button", { name: "送信" }));
    await expect(within(dialog).getByRole("status")).toHaveTextContent(
      "AIの次の質問とブリーフ案を確認しています…",
    );
    await expect(messageInput).toBeDisabled();
    await expect(
      within(dialog).getByRole("button", { name: "送信" }),
    ).toBeDisabled();
  },
};

export const BriefingMessageCloseBlocked: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: () => new Promise(() => {}),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    const messageInput = within(dialog).getByLabelText("AIへの回答");
    await userEvent.type(messageInput, "成功条件を先に決めたいです。");
    await userEvent.click(within(dialog).getByRole("button", { name: "送信" }));
    await userEvent.keyboard("{Escape}");
    await expect(
      body.getByRole("dialog", { name: "実験設計を開始" }),
    ).toBeVisible();
  },
};

export const BriefingStopPending: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    stopExperimentBriefing: () => new Promise(() => {}),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    await userEvent.click(
      within(dialog).getByRole("button", { name: "壁打ちを終了" }),
    );
    await expect(within(dialog).getByRole("status")).toHaveTextContent(
      "壁打ちを終了しています…",
    );
    await expect(
      within(dialog).getByRole("button", { name: "送信" }),
    ).toBeDisabled();
    await userEvent.keyboard("{Escape}");
    await expect(
      body.getByRole("dialog", { name: "実験設計を開始" }),
    ).toBeVisible();
  },
};

export const BriefingStopError: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    stopExperimentBriefing: fn(async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "壁打ちを終了できませんでした。もう一度お試しください。",
      },
    })),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    const messageInput = within(dialog).getByLabelText("AIへの回答");
    await userEvent.type(messageInput, "停止前の下書きです。");
    await userEvent.click(
      within(dialog).getByRole("button", { name: "壁打ちを終了" }),
    );
    await expect(within(dialog).getByRole("alert")).toHaveTextContent(
      "壁打ちを終了できません",
    );
    await expect(messageInput).toHaveValue("停止前の下書きです。");
    await userEvent.click(
      within(dialog).getByRole("button", { name: "もう一度試す" }),
    );
    await expect(args.stopExperimentBriefing).toHaveBeenCalledTimes(2);
  },
};

export const BriefingMessageValidationError: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    await userEvent.click(within(dialog).getByRole("button", { name: "送信" }));
    await expect(within(dialog).getByRole("alert")).toHaveTextContent(
      "AIへ送る内容を入力してください。",
    );
  },
};

export const BriefingMessageError: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: fn(async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "壁打ちを続けられませんでした。もう一度お試しください。",
      },
    })),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    const messageInput = within(dialog).getByLabelText("AIへの回答");
    await userEvent.type(messageInput, "成功条件を先に決めたいです。");
    await userEvent.click(within(dialog).getByRole("button", { name: "送信" }));
    await expect(within(dialog).getByRole("alert")).toHaveTextContent(
      "壁打ちを続けられません",
    );
    await expect(messageInput).toHaveValue("成功条件を先に決めたいです。");
    await expect(
      within(dialog).getByRole("button", { name: "送信" }),
    ).toBeEnabled();
    await userEvent.click(within(dialog).getByRole("button", { name: "送信" }));
    await expect(args.sendExperimentBriefMessage).toHaveBeenCalledTimes(2);
  },
};

export const BriefingMessageSuccess: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
    getExperimentBriefing: fn(async () => ({ data: briefing })),
    sendExperimentBriefMessage: fn(async () => ({
      data: { operationId: "operation-2" },
    })),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    const messageInput = within(dialog).getByLabelText("AIへの回答");
    await userEvent.type(messageInput, "成功条件を先に決めたいです。");
    await userEvent.click(within(dialog).getByRole("button", { name: "送信" }));
    await expect(messageInput).toHaveValue("");
    await expect(args.sendExperimentBriefMessage).toHaveBeenCalledWith(
      expect.any(String),
      "briefing-1",
      "成功条件を先に決めたいです。",
    );
    await expect(args.getExperimentBriefing).toHaveBeenCalledTimes(2);
  },
};

const completeBrief = {
  state: "active",
  messages: briefing.messages,
  latestBrief: {
    versionId: "brief-complete-v1",
    purpose: "問い合わせ要約の品質を比較する",
    hypothesis: "制約を明示すると正確性が向上する",
    candidatePrompts: ["短く要約する", "根拠を保って要約する"],
    evaluationAxes: "正確性、要点保持",
    environmentConditions: "同じ入力と評価手順を用いる",
  },
  lastConfirmedAt: confirmedAt,
};

export const BriefIncomplete: Story = {
  args: {
    createExperimentFromBrief,
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: briefing }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    await expect(
      within(dialog).getByRole("button", {
        name: "このブリーフを採用して作成",
      }),
    ).toBeDisabled();
  },
};

export const CreateExperimentPending: Story = {
  args: {
    createExperimentFromBrief: () => new Promise(() => {}),
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: completeBrief }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    await userEvent.click(
      within(dialog).getByRole("button", {
        name: "このブリーフを採用して作成",
      }),
    );
    await expect(within(dialog).getByRole("status")).toHaveTextContent(
      "実験を作成しています…",
    );
  },
};

export const CreateExperimentFailure: Story = {
  args: {
    createExperimentFromBrief: async () => ({
      error: { code: "UNAVAILABLE", message: "実験を作成できませんでした。" },
    }),
    listExperiments: async () => success,
    getExperimentBriefing: async () => ({ data: completeBrief }),
    sendExperimentBriefMessage: sendBriefingMessage,
    startExperimentBriefing: async () => ({
      data: { briefingSessionId: "briefing-1", operationId: "operation-1" },
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(canvas.getByRole("button", { name: "新規実験" }));
    const dialog = await body.findByRole("dialog", { name: "実験設計を開始" });
    await userEvent.click(
      within(dialog).getByRole("button", {
        name: "このブリーフを採用して作成",
      }),
    );
    await expect(within(dialog).getByRole("alert")).toHaveTextContent(
      "実験を作成できませんでした。",
    );
  },
};
