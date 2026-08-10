import { StartPreparation } from "@wails/go/wails/PreparationsHandler";

export type StartPreparationInput = {
  requestId: string;
  scope: string;
};

export type StartedPreparation = {
  preparationId: string;
  state: string;
};

export type StartPreparationResponse = {
  data?: StartedPreparation;
  error?: {
    code: string;
    message: string;
  };
};

export type StartPreparationService = (
  input: StartPreparationInput,
) => Promise<StartPreparationResponse>;

// Wails生成bindingは画面へ漏らさず、このserviceで画面用DTOに限定する。
export const startPreparation: StartPreparationService = async (input) => {
  const response = await StartPreparation(input.requestId, input.scope);

  if (!response.data) {
    return { error: response.error };
  }

  return {
    data: {
      preparationId: response.data.preparationId,
      state: response.data.state,
    },
  };
};
