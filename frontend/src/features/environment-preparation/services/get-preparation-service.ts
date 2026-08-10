import { GetPreparation } from "@wails/go/wails/PreparationsHandler";

export type PreparationCandidate = {
  id: string;
  environmentConditions: string;
  summary: string;
  createdAt: string;
};

export type PreparationDiagnostic = {
  id: string;
  code: string;
  summary: string;
  occurredAt: string;
};

export type PreparationFailure = {
  code: string;
  occurredAt: string;
};

export type PreparationReconciliation = {
  state: string;
  lastObservedAt: string;
};

export type PreparationDetail = {
  preparationId: string;
  state: string;
  startedAt: string;
  lastObservedAt: string;
  candidates: PreparationCandidate[];
  diagnostics: PreparationDiagnostic[];
  failure?: PreparationFailure;
  reconciliation: PreparationReconciliation;
};

export type GetPreparationResponse = {
  data?: PreparationDetail;
  error?: {
    code: string;
    message: string;
  };
};

export type GetPreparationService = (
  preparationId: string,
) => Promise<GetPreparationResponse>;

export const getPreparation: GetPreparationService = (preparationId) =>
  GetPreparation(preparationId);
