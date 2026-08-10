import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, fn, userEvent, within } from "storybook/test";
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
    adoptCandidate: fn(async () => ({
      data: {
        preparationId: detail.preparationId,
        candidateId: detail.candidates[0].id,
        environmentConditions: detail.candidates[0].environmentConditions,
      },
    })),
    getPreparation: async () => ({ data: detail }),
    onBackToList: fn(),
    onBeginExperiment: fn(),
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
    adoptCandidate: fn(async () => ({})),
    getPreparation: () => new Promise(() => {}),
    onBackToList: fn(),
    preparationId: detail.preparationId,
  },
};

export const CandidateAdoptionPending: Story = {
  args: {
    adoptCandidate: fn(() => new Promise(() => {})),
    getPreparation: async () => ({ data: detail }),
    onBackToList: fn(),
    onBeginExperiment: fn(),
    preparationId: detail.preparationId,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(
      await canvas.findByRole("button", {
        name: "この候補を採用して実験を開始",
      }),
    );
    const dialog = await body.findByRole("alertdialog", {
      name: "環境候補を採用しますか？",
    });
    await userEvent.click(
      within(dialog).getByRole("button", { name: "採用して新規実験へ" }),
    );
    await expect(within(dialog).getByRole("status")).toHaveTextContent(
      "環境候補を採用しています…",
    );
    await expect(
      within(dialog).getByRole("button", { name: "戻る" }),
    ).toBeDisabled();
  },
};

const retryAdoption = fn(async () => ({
  error: {
    code: "CANDIDATE_ADOPTION_UNAVAILABLE",
    message: "候補を採用できませんでした。",
  },
}));

export const CandidateAdoptionFailureAndRetry: Story = {
  args: {
    adoptCandidate: retryAdoption,
    getPreparation: async () => ({ data: detail }),
    onBackToList: fn(),
    onBeginExperiment: fn(),
    preparationId: detail.preparationId,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(
      await canvas.findByRole("button", {
        name: "この候補を採用して実験を開始",
      }),
    );
    const dialog = await body.findByRole("alertdialog", {
      name: "環境候補を採用しますか？",
    });
    const action = within(dialog).getByRole("button", {
      name: "採用して新規実験へ",
    });
    await userEvent.click(action);
    await expect(within(dialog).getByRole("alert")).toHaveTextContent(
      "候補を採用できませんでした。",
    );
    await userEvent.click(action);
    await expect(retryAdoption).toHaveBeenCalledTimes(2);
  },
};

export const EmptyAndReconciling: Story = {
  args: {
    adoptCandidate: fn(async () => ({})),
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
    adoptCandidate: fn(async () => ({})),
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
    adoptCandidate: fn(async () => ({})),
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
