import type { wails } from "@wails/go/models";
import { ListExperiments } from "@wails/go/wails/ExperimentsHandler";

export type ListExperimentsService =
  () => Promise<wails.ListExperimentsResponse>;

export const listExperiments: ListExperimentsService = () => ListExperiments();
