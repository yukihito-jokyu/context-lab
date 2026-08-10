import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, within } from "storybook/test";
import { EvaluationDetailPage } from "./EvaluationDetailPage";

const detail = {
  evaluation: {
    id: "evaluation-016",
    experimentId: "EXP-016",
    runId: "run-016",
    state: "completed",
    summary: "評価を完了しました。",
    updatedAt: "2026-08-10T10:25:00+09:00",
  },
  operation: {
    id: "evaluation-operation-016",
    state: "completed",
    updatedAt: "2026-08-10T10:25:00+09:00",
  },
  evidence: {
    runSummary:
      "顧客問い合わせについて、請求日と変更時の注意点を回答しました。",
    evaluationAxes: "正確性、要点保持",
  },
  result: {
    status: "complete",
    summary: "評価軸を満たしています。",
  },
  reconciliation: {
    state: "confirmed",
    lastObservedAt: "2026-08-10T10:25:00+09:00",
  },
  lastConfirmedAt: "2026-08-10T10:25:00+09:00",
};

const meta = {
  component: EvaluationDetailPage,
  title: "Features/Experiments/EvaluationDetailPage",
  args: {
    evaluationId: "evaluation-016",
    operationId: "evaluation-operation-016",
  },
} satisfies Meta<typeof EvaluationDetailPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Completed: Story = {
  args: { getEvaluationDetail: async () => ({ data: detail }) },
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByText("評価軸を満たしています。"),
    ).toBeVisible();
  },
};

export const Reconciling: Story = {
  args: {
    getEvaluationDetail: async () => ({
      data: {
        ...detail,
        evaluation: { ...detail.evaluation, state: "starting" },
        result: { status: "notRecorded" },
        reconciliation: {
          state: "reconciling",
          lastObservedAt: detail.lastConfirmedAt,
        },
      },
    }),
  },
};

export const Unavailable: Story = {
  args: {
    getEvaluationDetail: async () => ({
      data: {
        ...detail,
        evaluation: { ...detail.evaluation, state: "failed" },
        result: { status: "notRecorded", reasonCode: "EVALUATION_UNAVAILABLE" },
        failure: {
          code: "EVALUATION_UNAVAILABLE",
          occurredAt: detail.lastConfirmedAt,
        },
      },
    }),
  },
};

export const Loading: Story = {
  args: { getEvaluationDetail: () => new Promise(() => {}) },
};

export const LoadError: Story = {
  args: {
    getEvaluationDetail: async () => ({
      error: {
        code: "EVALUATION_DETAIL_UNAVAILABLE",
        message: "評価詳細を取得できませんでした。",
      },
    }),
  },
};
