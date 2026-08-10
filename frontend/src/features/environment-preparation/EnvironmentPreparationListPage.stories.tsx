import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, fn, userEvent, within } from "storybook/test";
import { EnvironmentPreparationListPage } from "./EnvironmentPreparationListPage";

const data = {
  preparations: [
    {
      preparationId: "PREP-012",
      state: "completed",
      startedAt: "2026-08-10T14:18:00+09:00",
      lastObservedAt: "2026-08-10T14:24:00+09:00",
    },
    {
      preparationId: "PREP-009",
      state: "failed",
      startedAt: "2026-08-10T10:03:00+09:00",
      lastObservedAt: "2026-08-10T10:12:00+09:00",
    },
  ],
};

const meta = {
  component: EnvironmentPreparationListPage,
  title: "Features/EnvironmentPreparation/EnvironmentPreparationListPage",
} satisfies Meta<typeof EnvironmentPreparationListPage>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    listPreparations: async () => ({ data }),
    startPreparation: async () => ({
      data: { preparationId: "PREP-013", state: "completed" },
    }),
  },
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByRole("heading", { name: "PREP-012" }),
    ).toBeVisible();
  },
};

export const Loading: Story = {
  args: {
    listPreparations: () => new Promise(() => {}),
    startPreparation: async () => ({
      data: { preparationId: "PREP-013", state: "completed" },
    }),
  },
};

export const Empty: Story = {
  args: {
    listPreparations: async () => ({
      data: { preparations: [] },
    }),
    startPreparation: async () => ({
      data: { preparationId: "PREP-013", state: "completed" },
    }),
  },
};

export const LoadError: Story = {
  args: {
    listPreparations: async () => ({
      error: {
        code: "UNAVAILABLE",
        message: "準備session一覧を取得できませんでした。",
      },
    }),
    startPreparation: async () => ({
      data: { preparationId: "PREP-013", state: "completed" },
    }),
  },
};

export const Reload: Story = {
  args: {
    listPreparations: fn(async () => ({ data })),
    startPreparation: async () => ({
      data: { preparationId: "PREP-013", state: "completed" },
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await canvas.findByRole("heading", { name: "PREP-012" });
    await userEvent.click(canvas.getByRole("button", { name: "一覧を再読込" }));
  },
};
