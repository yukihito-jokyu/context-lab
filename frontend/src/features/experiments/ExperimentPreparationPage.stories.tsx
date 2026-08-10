import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, fn, within } from "storybook/test";
import { ExperimentPreparationPage } from "./ExperimentPreparationPage";

const confirmedAt = "2026-08-09T10:25:00+09:00";
const preparation = {
  experimentId: "EXP-014",
  state: "preparing",
  purpose: "顧客問い合わせの要約品質を比較する",
  hypothesis: "制約を明示したpromptは要点の欠落を減らす",
  environmentConditions: "ローカル実行環境で要約だけを出力する",
  initialInput:
    "お問い合わせ: 契約プランを変更したいが、請求日は変わりますか？",
  prompts: [
    {
      sequenceNo: 1,
      content: "問い合わせを三つの箇条書きで要約してください。",
    },
    {
      sequenceNo: 2,
      content: "事実を保持し、対応要否と期限を明記してください。",
    },
  ],
  evaluationAxes: "正確性、要点保持、読みやすさ",
  source: { state: "採用済み", versionId: "brief-v1" },
  requiredFields: {
    purpose: true,
    environmentConditions: true,
    initialInput: true,
    prompts: true,
    evaluationAxes: true,
  },
  lastConfirmedAt: confirmedAt,
};

const meta = {
  component: ExperimentPreparationPage,
  title: "Features/Experiments/ExperimentPreparationPage",
  args: { experimentId: "EXP-014" },
} satisfies Meta<typeof ExperimentPreparationPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    getExperimentPreparation: async () => ({ data: preparation }),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({})),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText("顧客問い合わせの要約品質を比較する"),
    ).toBeVisible();
  },
};

export const Loading: Story = {
  args: {
    getExperimentPreparation: () => new Promise(() => {}),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({})),
  },
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByText("条件を読み込んでいます…"),
    ).toBeVisible();
  },
};

export const Empty: Story = {
  args: {
    getExperimentPreparation: async () => ({}),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({})),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText("準備内容はまだありません"),
    ).toBeVisible();
  },
};

export const LoadError: Story = {
  args: {
    getExperimentPreparation: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "実験準備を取得できませんでした。",
      },
    }),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({})),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText("実験準備を確認できません"),
    ).toBeVisible();
  },
};

export const NotFound: Story = {
  args: {
    getExperimentPreparation: async () => ({
      error: {
        code: "EXPERIMENT_PREPARATION_NOT_FOUND",
        message: "指定した実験は存在しません。",
      },
    }),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({})),
  },
};

export const SaveSuccess: Story = {
  args: {
    getExperimentPreparation: async () => ({ data: preparation }),
    saveExperimentPreparationDraft: fn(async () => ({
      data: {
        ...preparation,
        prompts: preparation.prompts.map((prompt) => prompt.content),
        savedAt: confirmedAt,
      },
    })),
    fixExperimentConditions: fn(async () => ({})),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await canvas.getByRole("button", { name: "下書きを保存" }).click();
    await expect(await canvas.findByText(/保存済み:/)).toBeVisible();
  },
};

export const SaveError: Story = {
  args: {
    getExperimentPreparation: async () => ({ data: preparation }),
    saveExperimentPreparationDraft: fn(async () => ({
      error: { code: "DRAFT_SAVE_FAILED", message: "保存できませんでした。" },
    })),
    fixExperimentConditions: fn(async () => ({})),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await canvas.getByRole("button", { name: "下書きを保存" }).click();
    await expect(
      await canvas.findByText("下書きを保存できません"),
    ).toBeVisible();
  },
};

export const FixSuccess: Story = {
  args: {
    getExperimentPreparation: async () => ({ data: preparation }),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({
      data: {
        experimentId: preparation.experimentId,
        state: "ready",
        fixedConditionId: "fixed-conditions-1",
        operationId: "operation-1",
        fixedAt: confirmedAt,
      },
    })),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await canvas.getByRole("button", { name: "条件を固定" }).click();
    await expect(await canvas.findByText(/条件を固定しました:/)).toBeVisible();
  },
};

export const FixValidationError: Story = {
  args: {
    getExperimentPreparation: async () => ({ data: preparation }),
    saveExperimentPreparationDraft: fn(async () => ({})),
    fixExperimentConditions: fn(async () => ({
      error: {
        code: "CONDITIONS_INVALID",
        message: "固定に必要な入力を確認してください。",
        fieldErrors: { purpose: "実験目的を入力してください。" },
      },
    })),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await canvas.getByRole("button", { name: "条件を固定" }).click();
    await expect(
      await canvas.findByText("実験目的を入力してください。"),
    ).toBeVisible();
  },
};
