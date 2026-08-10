import type { Meta, StoryObj } from "@storybook/react-vite";
import { CreateDerivedExperimentPage } from "./CreateDerivedExperimentPage";

const meta = {
  component: CreateDerivedExperimentPage,
  title: "Features/Experiments/CreateDerivedExperimentPage",
  args: {
    sourceExperimentId: "EXP-21",
    createDerivedExperiment: async () => ({
      data: {
        requestId: "request-21",
        experimentId: "EXP-21-derived",
        sourceExperimentId: "EXP-21",
        state: "preparing",
        createdAt: "2026-08-10T12:00:00Z",
      },
    }),
  },
} satisfies Meta<typeof CreateDerivedExperimentPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Input: Story = {};

export const CreateFailure: Story = {
  args: {
    createDerivedExperiment: async () => ({
      error: {
        code: "DERIVED_EXPERIMENT_UNAVAILABLE",
        message: "派生実験を作成できませんでした。",
      },
    }),
  },
};
