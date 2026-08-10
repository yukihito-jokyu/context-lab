import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, within } from "storybook/test";
import { RunEvaluationPage } from "./RunEvaluationPage";

const detail = {
  run: {
    id: "run-015",
    experimentId: "EXP-015",
    state: "completed",
    summary: "顧客問い合わせを要約しました。",
    updatedAt: "2026-08-10T10:25:00+09:00",
  },
  fixedPrompt: {
    sequenceNo: 1,
    content: "根拠を添えて簡潔に要約してください。",
  },
  operation: {
    id: "operation-015",
    state: "completed",
    updatedAt: "2026-08-10T10:25:00+09:00",
  },
  observations: [
    {
      sequenceNo: 1,
      kind: "output",
      occurredAt: "2026-08-10T10:24:00+09:00",
      summary: "請求日と変更時の注意点を三つの箇条書きで回答しました。",
    },
  ],
  artifacts: {
    status: "complete",
    items: [
      {
        digest: "sha256:answer-015",
        label: "回答",
        status: "stored",
      },
    ],
  },
  reconciliation: {
    state: "settled",
    lastObservedAt: "2026-08-10T10:25:00+09:00",
  },
  lastConfirmedAt: "2026-08-10T10:25:00+09:00",
};

const meta = {
  component: RunEvaluationPage,
  title: "Features/Experiments/RunEvaluationPage",
  args: {
    experimentId: "EXP-015",
    runId: "run-015",
    title: "run詳細",
    startRunEvaluation: async () => ({}),
  },
} satisfies Meta<typeof RunEvaluationPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Completed: Story = {
  args: { getRunDetail: async () => ({ data: detail }) },
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText("観測")).toBeVisible();
  },
};

export const Reconciling: Story = {
  args: {
    getRunDetail: async () => ({
      data: {
        ...detail,
        run: { ...detail.run, state: "running" },
        artifacts: {
          status: "partial",
          items: [],
          reasonCode: "ARTIFACT_PENDING",
        },
        reconciliation: {
          state: "reconciling",
          lastObservedAt: detail.lastConfirmedAt,
        },
      },
    }),
  },
};

export const Failed: Story = {
  args: {
    getRunDetail: async () => ({
      data: {
        ...detail,
        run: { ...detail.run, state: "failed" },
        artifacts: {
          status: "partial",
          items: [],
          reasonCode: "RUN_EXECUTION_FAILED",
        },
        failure: {
          code: "RUN_EXECUTION_FAILED",
          occurredAt: detail.lastConfirmedAt,
          partialSummary: "安全に表示できる実行失敗の要約です。",
        },
      },
    }),
  },
};

export const Loading: Story = {
  args: { getRunDetail: () => new Promise(() => {}) },
};

export const LoadError: Story = {
  args: {
    getRunDetail: async () => ({
      error: {
        code: "RUN_DETAIL_UNAVAILABLE",
        message: "run詳細を取得できませんでした。",
      },
    }),
  },
};
