import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, fn, userEvent, within } from "storybook/test";
import { StartPreparationPanel } from "./StartPreparationPanel";

const meta = {
  component: StartPreparationPanel,
  title: "Features/EnvironmentPreparation/StartPreparationPanel",
} satisfies Meta<typeof StartPreparationPanel>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    onStarted: fn(),
    startPreparation: async () => ({
      data: { preparationId: "PREP-014", state: "completed" },
    }),
  },
};

export const InputError: Story = {
  args: {
    onStarted: fn(),
    startPreparation: fn(),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.clear(
      canvas.getByLabelText("対象範囲（ワークスペースからの相対パス）"),
    );
    await userEvent.click(
      canvas.getByRole("button", { name: "環境準備を開始" }),
    );
    await expect(canvas.getByRole("alert")).toHaveTextContent(
      "対象範囲を入力してください。",
    );
  },
};

export const Starting: Story = {
  args: {
    onStarted: fn(),
    startPreparation: () => new Promise(() => {}),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(
      canvas.getByRole("button", { name: "環境準備を開始" }),
    );
    await expect(
      canvas.getByRole("button", { name: "環境準備を開始しています…" }),
    ).toBeDisabled();
  },
};

export const StartError: Story = {
  args: {
    onStarted: fn(),
    startPreparation: async () => ({
      error: {
        code: "ACP_NOT_READY",
        message: "環境準備を開始できませんでした。",
      },
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(
      canvas.getByRole("button", { name: "環境準備を開始" }),
    );
    await expect(canvas.getByRole("alert")).toHaveTextContent(
      "環境準備を開始できませんでした。",
    );
  },
};
