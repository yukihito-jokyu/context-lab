import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, fn, within } from "storybook/test";
import { EnvironmentPreparationDetailPage } from "./EnvironmentPreparationDetailPage";

const detail = {
  preparationId: "PREP-012",
  state: "completed",
  startedAt: "2026-08-10T14:18:00+09:00",
  lastObservedAt: "2026-08-10T14:24:00+09:00",
  candidates: [
    {
      id: "CAND-001",
      environmentConditions: "Node.js 20 / Linux",
      summary: "隔離環境で実行可能です。",
      createdAt: "2026-08-10T14:20:00+09:00",
    },
  ],
  diagnostics: [
    {
      id: "DIA-001",
      code: "DEPENDENCY_NOTICE",
      summary: "依存関係の確認が必要です。",
      occurredAt: "2026-08-10T14:21:00+09:00",
    },
  ],
  reconciliation: {
    state: "confirmed",
    lastObservedAt: "2026-08-10T14:24:00+09:00",
  },
};

const meta = {
  component: EnvironmentPreparationDetailPage,
  title: "Features/EnvironmentPreparation/EnvironmentPreparationDetailPage",
} satisfies Meta<typeof EnvironmentPreparationDetailPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    getPreparation: async () => ({ data: detail }),
    onBackToList: fn(),
    preparationId: detail.preparationId,
  },
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByRole("heading", { name: "環境候補" }),
    ).toBeVisible();
  },
};

export const Loading: Story = {
  args: {
    getPreparation: () => new Promise(() => {}),
    onBackToList: fn(),
    preparationId: detail.preparationId,
  },
};

export const EmptyAndReconciling: Story = {
  args: {
    getPreparation: async () => ({
      data: {
        ...detail,
        candidates: [],
        diagnostics: [],
        reconciliation: {
          state: "reconciling",
          lastObservedAt: detail.lastObservedAt,
        },
      },
    }),
    onBackToList: fn(),
    preparationId: detail.preparationId,
  },
};

export const Failure: Story = {
  args: {
    getPreparation: async () => ({
      data: {
        ...detail,
        failure: {
          code: "PREPARATION_TIMEOUT",
          occurredAt: detail.lastObservedAt,
        },
      },
    }),
    onBackToList: fn(),
    preparationId: detail.preparationId,
  },
};

export const LoadError: Story = {
  args: {
    getPreparation: async () => ({
      error: {
        code: "PREPARATION_UNAVAILABLE",
        message: "準備sessionを取得できませんでした。",
      },
    }),
    onBackToList: fn(),
    preparationId: detail.preparationId,
  },
};
