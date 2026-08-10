import { ListPreparations } from "@wails/go/wails/PreparationsHandler";

export type PreparationListItem = {
  preparationId: string;
  state: string;
  startedAt?: string;
  lastObservedAt?: string;
};

export type ListPreparationsResponse = {
  data?: {
    preparations: PreparationListItem[];
  };
  error?: {
    code: string;
    message: string;
  };
};

export type ListPreparationsService = () => Promise<ListPreparationsResponse>;

export const listPreparations: ListPreparationsService = () =>
  ListPreparations();
