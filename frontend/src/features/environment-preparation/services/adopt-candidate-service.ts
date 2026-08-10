import { AdoptCandidate } from "@wails/go/wails/PreparationsHandler";

export type AdoptCandidateInput = {
  requestId: string;
  preparationId: string;
  candidateId: string;
};

export type AdoptedCandidate = {
  preparationId: string;
  candidateId: string;
  environmentConditions: string;
};

export type AdoptCandidateResponse = {
  data?: AdoptedCandidate;
  error?: { code: string; message: string };
};

export type AdoptCandidateService = (
  input: AdoptCandidateInput,
) => Promise<AdoptCandidateResponse>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const adoptCandidate: AdoptCandidateService = async (input) => {
  const response = await AdoptCandidate(
    input.requestId,
    input.preparationId,
    input.candidateId,
  );

  if (!response.data) return { error: response.error };

  return {
    data: {
      preparationId: response.data.preparationId,
      candidateId: response.data.candidateId,
      environmentConditions: response.data.environmentConditions,
    },
  };
};
