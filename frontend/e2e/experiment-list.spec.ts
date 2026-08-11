import { expect, type Page, test } from "@playwright/test";

type ListExperimentsResponse = Record<string, unknown>;
type StartExperimentBriefingResponse = Record<string, unknown>;
type GetExperimentBriefingResponse = Record<string, unknown>;
type SendExperimentBriefMessageResponse = Record<string, unknown>;
type CreateExperimentFromBriefResponse = Record<string, unknown>;
type StopExperimentBriefingResponse = Record<string, unknown>;
type GetExperimentPreparationResponse = Record<string, unknown> & {
  throwMessage?: string;
};
type SaveExperimentPreparationDraftResponse = Record<string, unknown>;
type FixExperimentConditionsResponse = Record<string, unknown>;
type GetExperimentWorkspaceResponse = Record<string, unknown>;
type StartExperimentResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type StartRunEvaluationResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type GetRunDetailResponse = Record<string, unknown>;
type GetEvaluationDetailResponse = Record<string, unknown>;
type RetryEndedRunResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type GetExperimentComparisonResponse = Record<string, unknown>;
type GetDerivationSourceResponse = Record<string, unknown>;
type StartDerivationBriefingResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type SendDerivationBriefMessageResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type GetDerivationBriefingResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type StopDerivationBriefingResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type CreateDerivedExperimentResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type FinalizeExperimentConclusionResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type GetInsightWorkspaceResponse = Record<string, unknown>;
type CreateInsightResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type ListPreparationsResponse = Record<string, unknown>;
type GetPreparationResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
};
type StartPreparationResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
  throwMessage?: string;
};
type AdoptCandidateResponse = Record<string, unknown> & {
  delayMs?: number;
  result?: Record<string, unknown>;
  throwMessage?: string;
};

declare global {
  interface Window {
    __briefingGetCallCount: number;
    __briefingMessageRequests: Array<{
      requestId: string;
      briefingSessionId: string;
      message: string;
    }>;
    __briefingRequestIds: string[];
    __derivationBriefingGetRequests: string[];
    __briefingStopRequests: Array<{
      requestId: string;
      briefingSessionId: string;
    }>;
    __createExperimentRequests: Array<{
      requestId: string;
      briefingSessionId: string;
      briefVersionId: string;
    }>;
    __createDerivedExperimentRequests: Array<{
      requestId: string;
      sourceExperimentId: string;
      changes: Record<string, unknown>;
      reason: string;
    }>;
    __derivationBriefingRequests: Array<{
      requestId: string;
      sourceExperimentId: string;
    }>;
    __derivationBriefingMessageRequests: Array<{
      requestId: string;
      briefingSessionId: string;
      message: string;
    }>;
    __derivationBriefingStopRequests: Array<{
      requestId: string;
      briefingSessionId: string;
    }>;
    __draftSaveRequests: Array<{
      requestId: string;
      experimentId: string;
      purpose: string;
    }>;
    __fixConditionsRequests: Array<{
      requestId: string;
      experimentId: string;
      purpose: string;
    }>;
    __startExperimentRequests: Array<{
      requestId: string;
      experimentId: string;
    }>;
    __startRunEvaluationRequests: Array<{
      requestId: string;
      runId: string;
    }>;
    __finalizeExperimentConclusionRequests: Array<{
      requestId: string;
      experimentId: string;
      conclusion: string;
    }>;
    __createInsightRequests: Array<{
      requestId: string;
      evidences: Array<{ experimentId: string; conclusionId: string }>;
      statement: string;
      applicabilityConditions: string;
      verificationGaps: string;
    }>;
    __startPreparationRequests: Array<{
      requestId: string;
      scope: string;
    }>;
    __adoptCandidateRequests: Array<{
      requestId: string;
      preparationId: string;
      candidateId: string;
    }>;
  }
}

const confirmedAt = "2026-08-08T10:25:00+09:00";

const emptyResponse = {
  data: {
    experiments: [],
    cancelledExperiments: [],
    resumeSummary: { statusCounts: {} },
    lastConfirmedAt: confirmedAt,
  },
};

const errorResponse = {
  error: {
    code: "UNAVAILABLE",
    message: "実験一覧を取得できませんでした。",
  },
};

const successResponse = {
  data: {
    experiments: [
      {
        id: "EXP-014",
        purpose: "顧客問い合わせ要約の比較",
        state: "running",
        progressSummary: "run 2/4 実行中",
        updatedAt: confirmedAt,
      },
    ],
    cancelledExperiments: [],
    resumeSummary: {
      recommendedExperimentId: "EXP-014",
      statusCounts: { running: 1 },
    },
    lastConfirmedAt: confirmedAt,
  },
};

async function installListExperimentsMock(
  page: Page,
  responses: ListExperimentsResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = { wails: { ExperimentsHandler: { ListExperiments: () => {
        const response = responses[Math.min(callCount, responses.length - 1)];
        callCount += 1;
        return Promise.resolve(response);
      } } } };
    `,
  });
}

async function installExperimentPreparationMock(
  page: Page,
  responses: GetExperimentPreparationResponse[],
  saveResponses: SaveExperimentPreparationDraftResponse[] = [
    {
      error: {
        code: "DRAFT_SAVE_FAILED",
        message: "下書きを保存できませんでした。",
      },
    },
  ],
  fixResponses: FixExperimentConditionsResponse[] = [
    {
      error: {
        code: "FIX_CONDITIONS_SAVE_FAILED",
        message: "条件を固定できませんでした。",
      },
    },
  ],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      const saveResponses = ${JSON.stringify(saveResponses)};
      const fixResponses = ${JSON.stringify(fixResponses)};
      let callCount = 0;
      let saveCallCount = 0;
      let fixCallCount = 0;
      window.go = window.go || { wails: {} };
      window.__draftSaveRequests = [];
      window.__fixConditionsRequests = window.__fixConditionsRequests || [];
      window.go.wails.ExperimentPreparationsHandler = {
        GetExperimentPreparation: () => {
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          if (response.throwMessage) {
            return Promise.reject(new Error(response.throwMessage));
          }
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        SaveExperimentPreparationDraft: (request) => {
          window.__draftSaveRequests.push(request);
          const response = saveResponses[Math.min(saveCallCount, saveResponses.length - 1)];
          saveCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        FixExperimentConditions: (request) => {
          window.__fixConditionsRequests.push(request);
          const response = fixResponses[Math.min(fixCallCount, fixResponses.length - 1)];
          fixCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        }
      };
    `,
  });
}

async function installExperimentWorkspaceMock(
  page: Page,
  responses: GetExperimentWorkspaceResponse[],
  startResponses: StartExperimentResponse[] = [
    {
      error: {
        code: "START_EXPERIMENT_FAILED",
        message: "実験を開始できませんでした。",
      },
    },
  ],
  evaluationResponses: StartRunEvaluationResponse[] = [
    {
      error: {
        code: "RUN_EVALUATION_UNAVAILABLE",
        message: "評価を開始できませんでした。",
      },
    },
  ],
) {
  await page.addInitScript({
    content: `
      const workspaceResponses = ${JSON.stringify(responses)};
      const startResponses = ${JSON.stringify(startResponses)};
      const evaluationResponses = ${JSON.stringify(evaluationResponses)};
      let workspaceCallCount = 0;
      let startCallCount = 0;
      let evaluationCallCount = 0;
      window.go = window.go || { wails: {} };
      window.__startExperimentRequests = [];
      window.__startRunEvaluationRequests = [];
      window.go.wails.ExperimentWorkspacesHandler = {
        GetExperimentWorkspace: () => {
          const response = workspaceResponses[Math.min(workspaceCallCount, workspaceResponses.length - 1)];
          workspaceCallCount += 1;
          return Promise.resolve(response);
        }
      };
      window.go.wails.ExperimentRunsHandler = {
        StartExperiment: (request) => {
          window.__startExperimentRequests.push(request);
          const response = startResponses[Math.min(startCallCount, startResponses.length - 1)];
          startCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
      };
      window.go.wails.ExperimentEvaluationsHandler = {
        StartRunEvaluation: (request) => {
          window.__startRunEvaluationRequests.push(request);
          const response = evaluationResponses[Math.min(evaluationCallCount, evaluationResponses.length - 1)];
          evaluationCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installRunDetailMock(
  page: Page,
  responses: GetRunDetailResponse[],
) {
  await page.addInitScript({
    content: `
      const runDetailResponses = ${JSON.stringify(responses)};
      let runDetailCallCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.ExperimentRunDetailsHandler = {
        GetRunDetail: () => {
          const response = runDetailResponses[Math.min(runDetailCallCount, runDetailResponses.length - 1)];
          runDetailCallCount += 1;
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installEvaluationDetailMock(
  page: Page,
  responses: GetEvaluationDetailResponse[],
) {
  await page.addInitScript({
    content: `
      const evaluationDetailResponses = ${JSON.stringify(responses)};
      let evaluationDetailCallCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.ExperimentEvaluationDetailsHandler = {
        GetEvaluationDetail: () => {
          const response = evaluationDetailResponses[Math.min(evaluationDetailCallCount, evaluationDetailResponses.length - 1)];
          evaluationDetailCallCount += 1;
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installRetryEndedRunMock(
  page: Page,
  responses: RetryEndedRunResponse[],
) {
  await page.addInitScript({
    content: `
      const retryResponses = ${JSON.stringify(responses)};
      let retryCallCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.ExperimentRunRetriesHandler = {
        RetryEndedRun: () => {
          const response = retryResponses[Math.min(retryCallCount, retryResponses.length - 1)];
          retryCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installComparisonMock(
  page: Page,
  responses: GetExperimentComparisonResponse[],
) {
  await page.addInitScript({
    content: `const r=${JSON.stringify(responses)};let i=0;window.go=window.go||{wails:{}};window.go.wails.ExperimentComparisonsHandler={GetExperimentComparison:()=>Promise.resolve(r[Math.min(i++,r.length-1)])};`,
  });
}

async function installInsightWorkspaceMock(
  page: Page,
  responses: GetInsightWorkspaceResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.InsightsHandler = {
        GetInsightWorkspace: () => {
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installCreateInsightMock(
  page: Page,
  responses: CreateInsightResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.__createInsightRequests = [];
      window.go.wails.InsightsHandler = {
        ...(window.go.wails.InsightsHandler || {}),
        CreateInsight: (request) => {
          window.__createInsightRequests.push(request);
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => window.setTimeout(() => resolve(response.result), response.delayMs));
          }
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installDerivationSourceMock(
  page: Page,
  responses: GetDerivationSourceResponse[],
) {
  await page.addInitScript({
    content: `const r=${JSON.stringify(responses)};let i=0;window.go=window.go||{wails:{}};window.go.wails.ExperimentDerivationSourcesHandler={GetDerivationSource:()=>Promise.resolve(r[Math.min(i++,r.length-1)])};`,
  });
}

async function installCreateDerivedExperimentMock(
  page: Page,
  responses: CreateDerivedExperimentResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.__createDerivedExperimentRequests = [];
      window.go.wails.CreateDerivedExperimentsHandler = {
        CreateDerivedExperiment: (request) => {
          window.__createDerivedExperimentRequests.push(request);
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installDerivationBriefingMock(
  page: Page,
  responses: StartDerivationBriefingResponse[],
  messageResponses: SendDerivationBriefMessageResponse[] = [],
  getResponses: GetDerivationBriefingResponse[] = [
    {
      data: {
        state: "started",
        messages: [],
        lastConfirmedAt: confirmedAt,
      },
    },
  ],
  stopResponses: StopDerivationBriefingResponse[] = [],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      const messageResponses = ${JSON.stringify(messageResponses)};
      const getResponses = ${JSON.stringify(getResponses)};
      const stopResponses = ${JSON.stringify(stopResponses)};
      let callCount = 0;
      let messageCallCount = 0;
      let getCallCount = 0;
      let stopCallCount = 0;
      window.go = window.go || { wails: {} };
      window.__derivationBriefingRequests = [];
      window.__derivationBriefingMessageRequests = [];
      window.__derivationBriefingGetRequests = [];
      window.__derivationBriefingStopRequests = [];
      window.go.wails.DerivationBriefingsHandler = {
        StartDerivationBriefing: (requestId, sourceExperimentId) => {
          window.__derivationBriefingRequests.push({ requestId, sourceExperimentId });
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        SendDerivationBriefMessage: (requestId, briefingSessionId, message) => {
          window.__derivationBriefingMessageRequests.push({ requestId, briefingSessionId, message });
          const response = messageResponses[Math.min(messageCallCount, messageResponses.length - 1)];
          messageCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        GetDerivationBriefing: (briefingSessionId) => {
          window.__derivationBriefingGetRequests.push(briefingSessionId);
          const response = getResponses[Math.min(getCallCount, getResponses.length - 1)];
          getCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        StopDerivationBriefing: (requestId, briefingSessionId) => {
          window.__derivationBriefingStopRequests.push({ requestId, briefingSessionId });
          const response = stopResponses[Math.min(stopCallCount, stopResponses.length - 1)];
          stopCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installFinalizeExperimentConclusionMock(
  page: Page,
  responses: FinalizeExperimentConclusionResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.__finalizeExperimentConclusionRequests = [];
      window.go.wails.FinalizeExperimentConclusionsHandler = {
        FinalizeExperimentConclusion: (request) => {
          window.__finalizeExperimentConclusionRequests.push(request);
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
      };
    `,
  });
}

async function installListPreparationsMock(
  page: Page,
  responses: ListPreparationsResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.PreparationsHandler = {
        ListPreparations: () => {
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          return Promise.resolve(response);
        }
      };
    `,
  });
}

async function installGetPreparationMock(
  page: Page,
  responses: GetPreparationResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.PreparationsHandler = window.go.wails.PreparationsHandler || {};
      window.go.wails.PreparationsHandler.GetPreparation = () => {
        const response = responses[Math.min(callCount, responses.length - 1)];
        callCount += 1;
        if (response.throwMessage) {
          return Promise.reject(new Error(response.throwMessage));
        }
        if (response.delayMs) {
          return new Promise((resolve) => {
            window.setTimeout(() => resolve(response.result), response.delayMs);
          });
        }
        return Promise.resolve(response);
      };
    `,
  });
}

async function installStartPreparationMock(
  page: Page,
  responses: StartPreparationResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.__startPreparationRequests = JSON.parse(
        window.localStorage.getItem("startPreparationRequests") || "[]",
      );
      window.go = window.go || { wails: {} };
      window.go.wails.PreparationsHandler = window.go.wails.PreparationsHandler || {};
      window.go.wails.PreparationsHandler.StartPreparation = (requestId, scope) => {
        window.__startPreparationRequests.push({ requestId, scope });
        window.localStorage.setItem(
          "startPreparationRequests",
          JSON.stringify(window.__startPreparationRequests),
        );
        const response = responses[Math.min(callCount, responses.length - 1)];
        callCount += 1;
        if (response.delayMs) {
          return new Promise((resolve) => {
            window.setTimeout(() => resolve(response.result), response.delayMs);
          });
        }
        return Promise.resolve(response);
      };
    `,
  });
}

async function installAdoptCandidateMock(
  page: Page,
  responses: AdoptCandidateResponse[],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.__adoptCandidateRequests = [];
      window.go = window.go || { wails: {} };
      window.go.wails.PreparationsHandler = window.go.wails.PreparationsHandler || {};
      window.go.wails.PreparationsHandler.AdoptCandidate = (requestId, preparationId, candidateId) => {
        window.__adoptCandidateRequests.push({ requestId, preparationId, candidateId });
        const response = responses[Math.min(callCount, responses.length - 1)];
        callCount += 1;
        if (response.throwMessage) return Promise.reject(new Error(response.throwMessage));
        if (response.delayMs) {
          return new Promise((resolve) => {
            window.setTimeout(() => resolve(response.result), response.delayMs);
          });
        }
        return Promise.resolve(response);
      };
    `,
  });
}

async function installExperimentBriefingMock(
  page: Page,
  responses: StartExperimentBriefingResponse[],
  briefingResponses: GetExperimentBriefingResponse[] = [],
  messageResponses: SendExperimentBriefMessageResponse[] = [],
  createResponses: CreateExperimentFromBriefResponse[] = [],
  stopResponses: StopExperimentBriefingResponse[] = [],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      const briefingResponses = ${JSON.stringify(briefingResponses)};
      const messageResponses = ${JSON.stringify(messageResponses)};
      const createResponses = ${JSON.stringify(createResponses)};
      const stopResponses = ${JSON.stringify(stopResponses)};
      let callCount = 0;
      let briefingCallCount = 0;
      let messageCallCount = 0;
      let createCallCount = 0;
      let stopCallCount = 0;
      window.go = window.go || { wails: {} };
      window.__briefingRequestIds = [];
      window.__briefingMessageRequests = [];
      window.__briefingGetCallCount = 0;
      window.__briefingStopRequests = [];
      window.__createExperimentRequests = [];
      window.go.wails.ExperimentBriefingsHandler = {
        StartExperimentBriefing: (requestId) => {
          window.__briefingRequestIds.push(requestId);
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          return Promise.resolve(response);
        },
        GetExperimentBriefing: () => {
          window.__briefingGetCallCount += 1;
          const response = briefingResponses[Math.min(briefingCallCount, briefingResponses.length - 1)];
          briefingCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        SendExperimentBriefMessage: (requestId, briefingSessionId, message) => {
          window.__briefingMessageRequests.push({ requestId, briefingSessionId, message });
          const response = messageResponses[Math.min(messageCallCount, messageResponses.length - 1)];
          messageCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        CreateExperimentFromBrief: (requestId, briefingSessionId, briefVersionId) => {
          window.__createExperimentRequests.push({ requestId, briefingSessionId, briefVersionId });
          const response = createResponses[Math.min(createCallCount, createResponses.length - 1)];
          createCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        },
        StopExperimentBriefing: (requestId, briefingSessionId) => {
          window.__briefingStopRequests.push({ requestId, briefingSessionId });
          const response = stopResponses[Math.min(stopCallCount, stopResponses.length - 1)];
          stopCallCount += 1;
          if (response.delayMs) {
            return new Promise((resolve) => {
              window.setTimeout(() => resolve(response.result), response.delayMs);
            });
          }
          return Promise.resolve(response);
        }
      };
    `,
  });
}

test("一覧を表示する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "実験一覧" })).toBeVisible();
  await expect(page.getByText("顧客問い合わせ要約の比較")).toBeVisible();
});

test("準備中の実験を一覧から再開できる", async ({ page }) => {
  await installListExperimentsMock(page, [
    {
      data: {
        experiments: [
          {
            id: "EXP-015",
            purpose: "問い合わせ要約の品質を比較する",
            state: "preparing",
            progressSummary: "条件を準備中",
            updatedAt: confirmedAt,
          },
        ],
        cancelledExperiments: [],
        resumeSummary: {
          recommendedExperimentId: "EXP-015",
          statusCounts: { preparing: 1 },
        },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await installExperimentPreparationMock(page, [
    completeExperimentPreparationResponse,
  ]);
  await page.goto("/");

  await page.locator("#open-experiment-EXP-015").click();
  await expect(page).toHaveURL("/experiments/EXP-015/preparation");
  await expect(
    page.getByRole("heading", { name: "実験の条件を準備する" }),
  ).toBeVisible();
});

test("空状態を表示する", async ({ page }) => {
  await installListExperimentsMock(page, [emptyResponse]);
  await page.goto("/");
  await expect(
    page.getByText("実験はまだありません", { exact: true }),
  ).toBeVisible();
});

test("取得失敗から再読込する", async ({ page }) => {
  await installListExperimentsMock(page, [errorResponse, successResponse]);
  await page.goto("/");
  await expect(
    page.getByText("実験一覧を確認できません", { exact: true }),
  ).toBeVisible();

  await page.getByRole("button", { name: "再読込" }).click();
  await expect(page.getByRole("heading", { name: "実験一覧" })).toBeVisible();
});

test("新規実験から壁打ちを開始する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [
      {
        data: {
          state: "active",
          messages: [
            {
              role: "assistant",
              content: "成功基準を教えてください。",
              sequenceNo: 1,
              createdAt: confirmedAt,
            },
          ],
          latestBrief: {
            versionId: "brief-v1",
            decision: "要約品質を比較する",
            successCriteria: "正確性を確認する",
            requiredConditions: "同じ入力を使う",
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.locator("#briefing-chat-log")).toContainText(
    "成功基準を教えてください。",
  );
  await expect(page.getByText("要約品質を比較する")).toBeVisible();
  await expect(page.getByText("briefing-1", { exact: false })).toHaveCount(0);
});

test("開始失敗後に新しいrequest IDで再試行する", async ({ page }) => {
  await installListExperimentsMock(page, [emptyResponse]);
  await installExperimentBriefingMock(
    page,
    [
      {
        error: {
          code: "UNAVAILABLE",
          message: "実験設計を開始できませんでした。",
        },
      },
      { data: { briefingSessionId: "briefing-2", operationId: "operation-2" } },
    ],
    [{ data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } }],
  );
  await page.goto("/");

  await page.locator("#empty-create-experiment-button").click();
  await expect(page.getByRole("alert")).toContainText(
    "実験設計を開始できませんでした。",
  );
  await page.getByRole("button", { name: "もう一度試す" }).click();
  await expect(page.locator("#briefing-chat-log")).toContainText(
    "会話はまだありません。",
  );
  const requestIds = await page.evaluate(
    () =>
      (window as typeof window & { __briefingRequestIds: string[] })
        .__briefingRequestIds,
  );
  expect(requestIds).toHaveLength(2);
  expect(requestIds[0]).not.toBe(requestIds[1]);
});

test("ブリーフ取得の失敗から再読込する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [
      {
        error: {
          code: "UNAVAILABLE",
          message: "最新状態を取得できませんでした。",
        },
      },
      { data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await expect(page.locator("#briefing-refresh-error")).toContainText(
    "最新状態を取得できませんでした。",
  );
  await page.locator("#reload-experiment-briefing-button").click();
  await expect(page.locator("#briefing-chat-log")).toContainText(
    "会話はまだありません。",
  );
});

test("壁打ちメッセージを送信して最新の会話を再取得する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [
      { data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } },
      {
        data: {
          state: "active",
          messages: [
            {
              role: "assistant",
              content: "評価者の人数も決めましょう。",
              sequenceNo: 2,
              createdAt: confirmedAt,
            },
          ],
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    [{ data: { operationId: "operation-2" } }],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  const messageInput = page.locator("#briefing-message-input");
  await messageInput.fill("成功条件を先に決めたいです。");
  await page.locator("#send-briefing-message-button").click();

  await expect(messageInput).toHaveValue("");
  await expect(page.locator("#briefing-chat-log")).toContainText(
    "評価者の人数も決めましょう。",
  );
  const requests = await page.evaluate(
    () =>
      (
        window as typeof window & {
          __briefingMessageRequests: Array<{
            requestId: string;
            briefingSessionId: string;
            message: string;
          }>;
        }
      ).__briefingMessageRequests,
  );
  expect(requests).toEqual([
    {
      requestId: expect.any(String),
      briefingSessionId: "briefing-1",
      message: "成功条件を先に決めたいです。",
    },
  ]);
  await expect
    .poll(() => page.evaluate(() => window.__briefingGetCallCount))
    .toBe(2);
});

test("空白の壁打ちメッセージを送信しない", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [{ data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } }],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page.locator("#send-briefing-message-button").click();
  await expect(page.locator("#briefing-error")).toContainText(
    "AIへ送る内容を入力してください。",
  );
  await expect
    .poll(() => page.evaluate(() => window.__briefingMessageRequests.length))
    .toBe(0);
});

test("壁打ちメッセージ送信中は入力と送信を無効にする", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [{ data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } }],
    [
      {
        delayMs: 300,
        result: { data: { operationId: "operation-2" } },
      },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page
    .locator("#briefing-message-input")
    .fill("成功条件を先に決めたいです。");
  await page.locator("#send-briefing-message-button").click();
  await expect(page.locator("#briefing-pending")).toBeVisible();
  await expect(page.locator("#briefing-message-input")).toBeDisabled();
  await expect(page.locator("#send-briefing-message-button")).toBeDisabled();
});

test("壁打ちメッセージの失敗後に同じ内容を再送する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [
      { data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } },
      { data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } },
    ],
    [
      {
        error: {
          code: "UNAVAILABLE",
          message: "壁打ちを続けられませんでした。もう一度お試しください。",
        },
      },
      { data: { operationId: "operation-2" } },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  const messageInput = page.locator("#briefing-message-input");
  await messageInput.fill("成功条件を先に決めたいです。");
  await page.locator("#send-briefing-message-button").click();
  await expect(page.locator("#briefing-command-error")).toContainText(
    "壁打ちを続けられませんでした。",
  );
  await expect(messageInput).toHaveValue("成功条件を先に決めたいです。");

  await page.locator("#send-briefing-message-button").click();
  await expect(messageInput).toHaveValue("");
  const requests = await page.evaluate(() => window.__briefingMessageRequests);
  expect(requests).toHaveLength(2);
  expect(requests[0].message).toBe("成功条件を先に決めたいです。");
  expect(requests[1].message).toBe("成功条件を先に決めたいです。");
});

test("前のセッションの遅延応答で再開始後の会話を上書きしない", async ({
  page,
}) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [
      {
        data: {
          briefingSessionId: "briefing-old",
          operationId: "operation-old",
        },
      },
      {
        data: {
          briefingSessionId: "briefing-new",
          operationId: "operation-new",
        },
      },
    ],
    [
      {
        delayMs: 250,
        result: {
          data: {
            state: "active",
            messages: [
              {
                role: "assistant",
                content: "古い会話",
                sequenceNo: 1,
                createdAt: confirmedAt,
              },
            ],
            lastConfirmedAt: confirmedAt,
          },
        },
      },
      {
        data: {
          state: "active",
          messages: [
            {
              role: "assistant",
              content: "新しい会話",
              sequenceNo: 1,
              createdAt: confirmedAt,
            },
          ],
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    [],
    [],
    [{ data: { operationId: "operation-stop-old" } }],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await expect(page.locator("#stop-experiment-briefing-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect
    .poll(() => page.evaluate(() => window.__briefingStopRequests))
    .toEqual([
      {
        requestId: expect.any(String),
        briefingSessionId: "briefing-old",
      },
    ]);
  await expect(page.getByRole("dialog")).not.toBeVisible();
  await page.locator("#new-experiment-button").click();
  await expect(page.locator("#briefing-chat-log")).toContainText("新しい会話");
  await page.waitForTimeout(300);
  await expect(page.locator("#briefing-chat-log")).not.toContainText(
    "古い会話",
  );
});

test("壁打ち停止に失敗するとモーダルと下書きを維持する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [{ data: { state: "active", messages: [], lastConfirmedAt: confirmedAt } }],
    [],
    [],
    [
      {
        error: {
          code: "UNAVAILABLE",
          message: "壁打ちを終了できませんでした。もう一度お試しください。",
        },
      },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  const messageInput = page.locator("#briefing-message-input");
  await messageInput.fill("停止前の下書きです。");
  await page.keyboard.press("Escape");

  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.locator("#briefing-stop-error")).toContainText(
    "壁打ちを終了できませんでした。",
  );
  await expect(messageInput).toHaveValue("停止前の下書きです。");
  await expect
    .poll(() => page.evaluate(() => window.__briefingStopRequests.length))
    .toBe(1);
});

const completeBriefResponse = {
  data: {
    state: "active",
    messages: [
      {
        role: "assistant",
        content: "このブリーフを確認してください。",
        sequenceNo: 1,
        createdAt: confirmedAt,
      },
    ],
    latestBrief: {
      versionId: "brief-complete-v1",
      purpose: "問い合わせ要約の品質を比較する",
      candidatePrompts: ["短く要約する", "根拠を保って要約する"],
      evaluationAxes: "正確性、要点保持",
      environmentConditions: "同じ入力と評価手順を用いる",
    },
    lastConfirmedAt: confirmedAt,
  },
};

const completeExperimentPreparationResponse = {
  data: {
    experimentId: "EXP-015",
    state: "preparing",
    purpose: "問い合わせ要約の品質を比較する",
    hypothesis: "根拠を保つpromptが正確性を高める",
    environmentConditions: "同じ入力と評価手順を用いる",
    initialInput: "顧客問い合わせ本文",
    prompts: [
      { sequenceNo: 1, content: "短く要約する" },
      { sequenceNo: 2, content: "根拠を保って要約する" },
    ],
    evaluationAxes: "正確性、要点保持",
    source: { state: "adopted", versionId: "brief-complete-v1" },
    requiredFields: {
      purpose: true,
      environmentConditions: true,
      initialInput: true,
      prompts: true,
      evaluationAxes: true,
    },
    lastConfirmedAt: confirmedAt,
  },
};

test("完全なブリーフを採用して準備画面へ遷移する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [completeBriefResponse],
    [],
    [{ data: { experimentId: "EXP-015", state: "preparing" } }],
  );
  await installExperimentPreparationMock(page, [
    completeExperimentPreparationResponse,
  ]);
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page.locator("#submit-create-experiment-button").click();
  await expect(page).toHaveURL(/\/experiments\/EXP-015\/preparation$/);
  await expect(
    page.getByRole("heading", { name: "実験の条件を準備する" }),
  ).toBeVisible();
  await expect(page.getByRole("textbox", { name: "候補prompt 2" })).toHaveValue(
    "根拠を保って要約する",
  );
});

test("Storageが使えなくても実験準備を表示する", async ({ page }) => {
  await installExperimentPreparationMock(page, [
    completeExperimentPreparationResponse,
  ]);
  await page.addInitScript(() => {
    Object.defineProperty(Storage.prototype, "getItem", {
      configurable: true,
      value: () => {
        throw new DOMException("storage unavailable", "SecurityError");
      },
    });
  });
  await page.goto("/experiments/EXP-015/preparation");

  await expect(
    page.getByRole("heading", { name: "実験の条件を準備する" }),
  ).toBeVisible();
  await expect(page.getByText("実験準備を取得できませんでした。")).toHaveCount(
    0,
  );
});

test("Wailsブリッジ例外を安全なエラーとして表示する", async ({ page }) => {
  await installExperimentPreparationMock(page, [
    { throwMessage: "native bridge details must not be displayed" },
  ]);
  await page.goto("/experiments/EXP-015/preparation");

  await expect(page.locator("#preparation-load-error")).toContainText(
    "アプリとの通信を開始できませんでした。",
  );
  await expect(page.locator("#preparation-load-error-code")).toHaveText(
    "WAILS_BRIDGE_UNAVAILABLE",
  );
  await expect(page.locator("#preparation-load-error")).not.toContainText(
    "native bridge details",
  );
});

test("採用失敗時はブリーフと会話を維持して再試行できる", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [completeBriefResponse],
    [],
    [
      {
        error: { code: "UNAVAILABLE", message: "実験を作成できませんでした。" },
      },
      {
        error: { code: "UNAVAILABLE", message: "実験を作成できませんでした。" },
      },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page.locator("#submit-create-experiment-button").click();
  await expect(page.locator("#create-experiment-command-error")).toContainText(
    "実験を作成できませんでした。",
  );
  await expect(page.locator("#briefing-chat-log")).toContainText(
    "このブリーフを確認してください。",
  );
  await expect(page.getByText("問い合わせ要約の品質を比較する")).toBeVisible();
  await page.locator("#submit-create-experiment-button").click();
  await expect
    .poll(() => page.evaluate(() => window.__createExperimentRequests.length))
    .toBe(2);
});

test("採用操作の二重実行を抑止する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [completeBriefResponse],
    [],
    [
      {
        delayMs: 200,
        result: {
          error: {
            code: "UNAVAILABLE",
            message: "実験を作成できませんでした。",
          },
        },
      },
    ],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page.locator("#submit-create-experiment-button").dblclick();
  await expect(page.locator("#create-experiment-pending")).toBeVisible();
  await expect(page.locator("#submit-create-experiment-button")).toBeDisabled();
  await expect
    .poll(() => page.evaluate(() => window.__createExperimentRequests.length))
    .toBe(1);
});

test("実験準備の入力内容を表示する", async ({ page }) => {
  await installExperimentPreparationMock(page, [
    {
      data: {
        experimentId: "EXP-015",
        state: "preparing",
        purpose: "問い合わせ要約の品質を比較する",
        hypothesis: "根拠を保つpromptが正確性を高める",
        environmentConditions: "同じ入力と評価手順を用いる",
        initialInput: "顧客問い合わせ本文",
        prompts: [
          { sequenceNo: 1, content: "短く要約する" },
          { sequenceNo: 2, content: "根拠を保って要約する" },
        ],
        evaluationAxes: "正確性、要点保持",
        source: { state: "adopted", versionId: "brief-v1" },
        requiredFields: {
          purpose: true,
          environmentConditions: true,
          initialInput: true,
          prompts: true,
          evaluationAxes: true,
        },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-015/preparation");

  await expect(
    page.getByRole("heading", { name: "実験の条件を準備する" }),
  ).toBeVisible();
  await expect(page.getByRole("textbox", { name: "候補prompt 2" })).toHaveValue(
    "根拠を保って要約する",
  );
  await expect(page.getByText("入力済み")).toHaveCount(5);
});

test("実験準備の下書き保存は成功時に保存時刻を表示する", async ({ page }) => {
  await installExperimentPreparationMock(
    page,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [{ sequenceNo: 1, content: "短く要約する" }],
          evaluationAxes: "正確性、要点保持",
          source: { state: "adopted", versionId: "brief-v1" },
          requiredFields: {
            purpose: true,
            environmentConditions: true,
            initialInput: true,
            prompts: false,
            evaluationAxes: true,
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "更新した目的",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [{ sequenceNo: 1, content: "短く要約する" }],
          evaluationAxes: "正確性、要点保持",
          savedAt: confirmedAt,
        },
      },
    ],
  );
  await page.goto("/experiments/EXP-015/preparation");

  await page.getByLabel("実験目的").fill("更新した目的");
  await page.getByRole("button", { name: "下書きを保存" }).click();

  await expect(page.locator("#preparation-save-success")).toBeVisible();
  await expect
    .poll(() => page.evaluate(() => window.__draftSaveRequests.length))
    .toBe(1);
  await expect
    .poll(() => page.evaluate(() => window.__draftSaveRequests[0].requestId))
    .not.toBe("");
});

test("実験準備の下書き保存中は編集を無効化する", async ({ page }) => {
  await installExperimentPreparationMock(
    page,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [],
          evaluationAxes: "正確性、要点保持",
          source: { state: "adopted", versionId: "brief-v1" },
          requiredFields: {
            purpose: true,
            environmentConditions: true,
            initialInput: true,
            prompts: false,
            evaluationAxes: true,
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    [{ delayMs: 300, result: {} }],
  );
  await page.goto("/experiments/EXP-015/preparation");

  await page.getByRole("button", { name: "下書きを保存" }).click();

  await expect(page.getByRole("button", { name: "保存中…" })).toBeDisabled();
  await expect(page.getByLabel("実験目的")).toBeDisabled();
});

test("実験準備の保存失敗では入力を保持し、再試行は別request IDを送る", async ({
  page,
}) => {
  await installExperimentPreparationMock(
    page,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [],
          evaluationAxes: "正確性、要点保持",
          source: { state: "adopted", versionId: "brief-v1" },
          requiredFields: {
            purpose: true,
            environmentConditions: true,
            initialInput: true,
            prompts: false,
            evaluationAxes: true,
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    [
      {
        error: {
          code: "DRAFT_SAVE_FAILED",
          message: "保存に失敗しました。",
        },
      },
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "失敗後も保持する目的",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [],
          evaluationAxes: "正確性、要点保持",
          savedAt: confirmedAt,
        },
      },
    ],
  );
  await page.goto("/experiments/EXP-015/preparation");

  await page.getByLabel("実験目的").fill("失敗後も保持する目的");
  await page.getByRole("button", { name: "下書きを保存" }).click();
  await expect(page.locator("#preparation-save-error")).toBeVisible();
  await expect(page.getByLabel("実験目的")).toHaveValue("失敗後も保持する目的");

  await page.getByRole("button", { name: "下書きを保存" }).click();
  await expect(page.locator("#preparation-save-success")).toBeVisible();
  await expect
    .poll(() => page.evaluate(() => window.__draftSaveRequests.length))
    .toBe(2);
  const requestIds = await page.evaluate(() =>
    window.__draftSaveRequests.map((request) => request.requestId),
  );
  expect(requestIds[0]).not.toBe(requestIds[1]);
});

test("実験準備の条件固定は成功後にワークスペースへ遷移する", async ({
  page,
}) => {
  await installExperimentWorkspaceMock(page, [
    {
      data: {
        experimentId: "EXP-015",
        state: "ready",
        fixedConditions: {
          fixedConditionId: "fixed-1",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [{ sequenceNo: 1, content: "短く要約する" }],
          evaluationAxes: "正確性、要点保持",
          fixedAt: confirmedAt,
        },
        conditionFixOperation: { operationId: "operation-1" },
        runs: [],
        evaluations: [],
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await installExperimentPreparationMock(
    page,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [{ sequenceNo: 1, content: "短く要約する" }],
          evaluationAxes: "正確性、要点保持",
          source: { state: "adopted", versionId: "brief-v1" },
          requiredFields: {
            purpose: true,
            environmentConditions: true,
            initialInput: true,
            prompts: false,
            evaluationAxes: true,
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    undefined,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "ready",
          fixedConditionId: "fixed-1",
          operationId: "operation-1",
          fixedAt: confirmedAt,
        },
      },
    ],
  );
  await page.goto("/experiments/EXP-015/preparation");
  await page.locator("#fix-conditions-button").click();

  await expect(page).toHaveURL(
    /\/experiments\/EXP-015\/workspace\?operationId=operation-1$/,
  );
  await expect(
    page.getByRole("heading", { name: "実験ワークスペース" }),
  ).toBeVisible();
  await expect(page.locator("#experiment-workspace-operation")).toContainText(
    "条件固定操作ID: operation-1",
  );
  await expect(
    page.locator("#experiment-workspace-fixed-conditions"),
  ).toContainText("問い合わせ要約の品質を比較する");
});

test("実験ワークスペースは固定条件と正本のrun・evaluationを表示する", async ({
  page,
}) => {
  await installExperimentWorkspaceMock(page, [
    {
      data: {
        experimentId: "EXP-016",
        state: "evaluating",
        fixedConditions: {
          fixedConditionId: "fixed-2",
          purpose: "回答の正確性を比較する",
          hypothesis: "根拠を指定すると正確性が上がる",
          environmentConditions: "同一のローカル環境を使用する",
          initialInput: "問い合わせ本文",
          prompts: [{ sequenceNo: 1, content: "根拠を添えて回答する" }],
          evaluationAxes: "正確性",
          fixedAt: confirmedAt,
        },
        conditionFixOperation: { operationId: "operation-2" },
        runs: [
          {
            id: "run-2",
            state: "completed",
            summary: "2件のpromptを実行しました。",
            updatedAt: confirmedAt,
          },
        ],
        evaluations: [
          {
            id: "evaluation-1",
            state: "running",
            summary: "正確性を評価中です。",
            updatedAt: confirmedAt,
          },
        ],
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-016/workspace");

  await expect(page.getByText("回答の正確性を比較する")).toBeVisible();
  await expect(page.getByText("run-2")).toBeVisible();
  await expect(page.getByText("completed")).toBeVisible();
  await expect(page.getByText("2件のpromptを実行しました。")).toBeVisible();
  await expect(page.getByText("evaluation-1")).toBeVisible();
  await expect(page.getByText("running")).toBeVisible();
  await expect(page.getByText("正確性を評価中です。")).toBeVisible();
});

test("実験ワークスペースは取得失敗後に再読込できる", async ({ page }) => {
  await installExperimentWorkspaceMock(page, [
    {
      error: {
        code: "EXPERIMENT_WORKSPACE_UNAVAILABLE",
        message: "取得に失敗しました。",
      },
    },
    {
      data: {
        experimentId: "EXP-017",
        state: "ready",
        fixedConditions: {
          fixedConditionId: "fixed-3",
          purpose: "再読込後の固定条件",
          environmentConditions: "同一環境",
          initialInput: "入力",
          prompts: [{ sequenceNo: 1, content: "prompt" }],
          evaluationAxes: "正確性",
          fixedAt: confirmedAt,
        },
        conditionFixOperation: { operationId: "operation-3" },
        runs: [],
        evaluations: [],
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-017/workspace");

  await expect(page.locator("#experiment-workspace-error")).toContainText(
    "取得に失敗しました。",
  );
  await page.locator("#reload-experiment-workspace-button").click();
  await expect(page.getByText("再読込後の固定条件")).toBeVisible();
});

test("実験ワークスペースから全promptの実験を開始し、正本を再読込する", async ({
  page,
}) => {
  const readyWorkspace = {
    experimentId: "EXP-018",
    state: "ready",
    fixedConditions: {
      fixedConditionId: "fixed-4",
      purpose: "全promptを実行する",
      environmentConditions: "同一環境",
      initialInput: "入力",
      prompts: [
        { sequenceNo: 1, content: "prompt 1" },
        { sequenceNo: 2, content: "prompt 2" },
      ],
      evaluationAxes: "正確性",
      fixedAt: confirmedAt,
    },
    conditionFixOperation: { operationId: "operation-4" },
    runs: [],
    evaluations: [],
    lastConfirmedAt: confirmedAt,
  };
  await installExperimentWorkspaceMock(
    page,
    [
      { data: readyWorkspace },
      { data: readyWorkspace },
      {
        data: {
          ...readyWorkspace,
          state: "running",
          runs: [
            { id: "run-1", state: "running", updatedAt: confirmedAt },
            { id: "run-2", state: "queued", updatedAt: confirmedAt },
          ],
        },
      },
    ],
    [
      {
        data: {
          experimentId: "EXP-018",
          operationId: "start-operation-1",
          runs: [
            { id: "run-1", state: "running", updatedAt: confirmedAt },
            { id: "run-2", state: "queued", updatedAt: confirmedAt },
          ],
          state: "running",
        },
      },
    ],
  );
  await page.goto("/experiments/EXP-018/workspace");

  await page.locator("#start-experiment-button").click();

  await expect(page.locator("#experiment-start-success")).toContainText(
    "操作ID: start-operation-1（2件のrun）",
  );
  await expect(page.getByText("run-1")).toBeVisible();
  await expect(page.getByText("run-2")).toBeVisible();
  await expect(page.locator("#start-experiment-button")).toBeDisabled();
  await expect
    .poll(() => page.evaluate(() => window.__startExperimentRequests.length))
    .toBe(1);
});

test("実験開始の失敗は同じrequest IDで終端再現し、開始中は二重送信しない", async ({
  page,
}) => {
  const workspace = {
    experimentId: "EXP-019",
    state: "ready",
    fixedConditions: {
      fixedConditionId: "fixed-5",
      purpose: "再試行する",
      environmentConditions: "同一環境",
      initialInput: "入力",
      prompts: [{ sequenceNo: 1, content: "prompt" }],
      evaluationAxes: "正確性",
      fixedAt: confirmedAt,
    },
    conditionFixOperation: { operationId: "operation-5" },
    runs: [],
    evaluations: [],
    lastConfirmedAt: confirmedAt,
  };
  await installExperimentWorkspaceMock(
    page,
    [
      { data: workspace },
      { data: workspace },
      { data: { ...workspace, state: "running" } },
    ],
    [
      {
        error: {
          code: "START_EXPERIMENT_FAILED",
          message: "一時的に開始できません。",
        },
      },
      {
        error: {
          code: "START_EXPERIMENT_FAILED",
          message: "一時的に開始できません。",
        },
      },
    ],
  );
  await page.goto("/experiments/EXP-019/workspace");

  await page.locator("#start-experiment-button").click();
  await expect(page.locator("#experiment-start-error")).toContainText(
    "一時的に開始できません。",
  );
  await page.locator("#start-experiment-button").click();
  await expect(page.locator("#experiment-start-error")).toContainText(
    "一時的に開始できません。",
  );
  await expect(page.locator("#experiment-start-success")).not.toBeVisible();

  const requests = await page.evaluate(() => window.__startExperimentRequests);
  expect(requests).toHaveLength(2);
  expect(requests[0].requestId).toBe(requests[1].requestId);
  expect(requests[0].experimentId).toBe("EXP-019");
});

test("完了したrunから評価を開始し、評価到達画面へ遷移する", async ({
  page,
}) => {
  const workspace = {
    experimentId: "EXP-020",
    state: "running",
    fixedConditions: {
      fixedConditionId: "fixed-6",
      purpose: "runを評価する",
      environmentConditions: "同一環境",
      initialInput: "入力",
      prompts: [{ sequenceNo: 1, content: "prompt" }],
      evaluationAxes: "正確性",
      fixedAt: confirmedAt,
    },
    conditionFixOperation: { operationId: "operation-6" },
    runs: [{ id: "run-20", state: "completed", updatedAt: confirmedAt }],
    evaluations: [],
    lastConfirmedAt: confirmedAt,
  };
  await installExperimentWorkspaceMock(page, [{ data: workspace }], undefined, [
    {
      data: {
        runId: "run-20",
        evaluationId: "evaluation-20",
        operationId: "evaluation-operation-20",
        state: "evaluating",
      },
    },
  ]);
  await installEvaluationDetailMock(page, [
    {
      data: {
        evaluation: {
          id: "evaluation-20",
          experimentId: "EXP-020",
          runId: "run-20",
          state: "starting",
          updatedAt: confirmedAt,
        },
        operation: {
          id: "evaluation-operation-20",
          state: "starting",
          updatedAt: confirmedAt,
        },
        evidence: { runSummary: "", evaluationAxes: "正確性" },
        result: { status: "notRecorded" },
        reconciliation: { state: "starting", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-020/workspace");

  await page.locator("#start-run-evaluation-button-run-20").click();

  await expect(page).toHaveURL(
    /\/evaluations\/evaluation-20\?operationId=evaluation-operation-20$/,
  );
  await expect(page).toHaveURL(
    /\/evaluations\/evaluation-20\?operationId=evaluation-operation-20$/,
  );
  await expect(page.locator("main")).toContainText("評価ID: evaluation-20");
  await expect(page.locator("main")).toContainText(
    "操作ID: evaluation-operation-20",
  );
  await expect(page.locator("#evaluation-detail-progress")).toContainText(
    "照合しています",
  );
});

test("評価詳細は根拠と確定した評価結果を表示する", async ({ page }) => {
  await installEvaluationDetailMock(page, [
    {
      data: {
        evaluation: {
          id: "evaluation-25",
          experimentId: "EXP-025",
          runId: "run-25",
          state: "completed",
          updatedAt: confirmedAt,
        },
        operation: {
          id: "evaluation-operation-25",
          state: "completed",
          updatedAt: confirmedAt,
        },
        evidence: {
          runSummary: "請求日を含めた回答を生成しました。",
          evaluationAxes: "正確性、要点保持",
        },
        result: { status: "complete", summary: "評価軸を満たしています。" },
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto(
    "/evaluations/evaluation-25?operationId=evaluation-operation-25",
  );

  await expect(page.locator("#evaluation-detail-evidence")).toContainText(
    "請求日を含めた回答を生成しました。",
  );
  await expect(page.locator("#evaluation-detail-result")).toContainText(
    "評価軸を満たしています。",
  );
});

test("実験比較は根拠、空、照合、失敗再読込と詳細導線を表示する", async ({
  page,
}) => {
  const base = {
    experiment: { id: "EXP-C", purpose: "比較目的", evaluationAxes: "正確性" },
    lastConfirmedAt: confirmedAt,
  };
  await installComparisonMock(page, [
    {
      data: {
        ...base,
        evaluations: [
          {
            evaluationId: "eval-c",
            runId: "run-c",
            state: "completed",
            runSummary: "実行根拠",
            result: { status: "complete", summary: "比較結果" },
            reconciliation: {
              state: "reconciling",
              lastObservedAt: confirmedAt,
            },
            updatedAt: confirmedAt,
          },
        ],
      },
    },
  ]);
  await page.goto("/experiments/EXP-C/comparison");
  await expect(page.getByText("根拠（実行要約）: 実行根拠")).toBeVisible();
  await expect(page.getByText("評価結果を照合しています")).toBeVisible();
  await expect(page.getByRole("button", { name: "run詳細" })).toBeVisible();
  await expect(page.getByRole("button", { name: "評価詳細" })).toBeVisible();
  await page.getByRole("button", { name: "run詳細" }).click();
  await expect(page).toHaveURL(/\/experiments\/EXP-C\/runs\/run-c$/);

  await page.goto("/experiments/EXP-C/comparison");
  await page.getByRole("button", { name: "評価詳細" }).click();
  await expect(page).toHaveURL(/\/evaluations\/eval-c$/);

  await installComparisonMock(page, [
    { error: { code: "UNAVAILABLE", message: "失敗" } },
    { data: { ...base, evaluations: [] } },
  ]);
  await page.goto("/experiments/EXP-C/comparison");
  await expect(page.locator("#comparison-error")).toContainText("失敗");
  await page.locator("#reload-comparison-button").click();
  await expect(page.locator("#empty-comparison")).toBeVisible();
});

test("比較結果から結論を確定し、失敗再試行と永続済み結論を確認する", async ({
  page,
}) => {
  const base = {
    experiment: {
      id: "EXP-19",
      purpose: "結論の比較",
      evaluationAxes: "正確性",
    },
    evaluations: [
      {
        evaluationId: "eval-19",
        runId: "run-19",
        state: "completed",
        runSummary: "比較の根拠",
        result: { status: "complete", summary: "比較結果" },
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
        updatedAt: confirmedAt,
      },
    ],
    lastConfirmedAt: confirmedAt,
  };
  await installComparisonMock(page, [
    { data: base },
    { data: base },
    {
      data: {
        ...base,
        conclusion: {
          id: "conclusion-19",
          content: "永続済みの結論",
          state: "finalized",
          finalizedAt: confirmedAt,
        },
      },
    },
  ]);
  await installFinalizeExperimentConclusionMock(page, [
    { error: { code: "UNAVAILABLE", message: "一時的に確定できません" } },
    {
      delayMs: 100,
      result: {
        data: {
          requestId: "ignored-by-ui",
          experimentId: "EXP-19",
          conclusionId: "conclusion-19",
          conclusion: "画面から確定した結論",
          state: "finalized",
          finalizedAt: confirmedAt,
        },
      },
    },
  ]);

  await page.goto("/experiments/EXP-19/comparison");
  const finalize = page.locator("#finalize-experiment-conclusion-button");
  await expect(finalize).toBeDisabled();
  await page.locator("#experiment-conclusion").fill("画面から確定した結論");
  await finalize.click();
  await expect(page.locator("#experiment-conclusion-error")).toContainText(
    "一時的に確定できません",
  );
  await finalize.click();
  await expect(finalize).toBeDisabled();
  await expect(page.locator("#experiment-conclusion-finalized")).toContainText(
    "永続済みの結論",
  );
  const requestIDs = await page.evaluate(() =>
    window.__finalizeExperimentConclusionRequests.map(
      (request) => request.requestId,
    ),
  );
  expect(requestIDs).toHaveLength(2);
  expect(requestIDs[0]).toBe(requestIDs[1]);
  await expect(page.getByRole("link", { name: "派生を確認" })).toHaveAttribute(
    "href",
    "/experiments/EXP-19/derivations",
  );
  await expect(page.getByRole("link", { name: "知見を確認" })).toHaveAttribute(
    "href",
    "/experiments/EXP-19/insights",
  );
  await page.locator("#reload-finalized-experiment-conclusion-button").click();
  await expect(page.locator("#experiment-conclusion-finalized")).toContainText(
    "永続済みの結論",
  );
  await installInsightWorkspaceMock(page, [
    {
      data: {
        evidenceCandidates: [
          {
            experimentId: "EXP-19",
            purpose: "結論の比較",
            evaluationAxes: "正確性",
            conclusionId: "conclusion-19",
            conclusion: "画面から確定した結論",
            finalizedAt: confirmedAt,
          },
          {
            experimentId: "EXP-20",
            purpose: "別条件の比較",
            evaluationAxes: "再現性",
            conclusionId: "conclusion-20",
            conclusion: "別条件の結論",
            finalizedAt: confirmedAt,
          },
        ],
        savedConsiderations: [
          {
            experimentId: "EXP-19",
            conclusionId: "conclusion-19",
            content: "画面から確定した結論",
            finalizedAt: confirmedAt,
          },
        ],
        insights: [],
        lastConfirmedAt: confirmedAt,
      },
    },
    {
      error: {
        code: "INSIGHT_WORKSPACE_UNAVAILABLE",
        message: "一時的に知見を取得できません",
      },
    },
    {
      data: {
        evidenceCandidates: [],
        savedConsiderations: [],
        insights: [],
      },
    },
  ]);
  await page.getByRole("link", { name: "知見を確認" }).click();
  await expect(page.locator("#insight-evidence-EXP-19")).toContainText(
    "画面から確定した結論",
  );
  await expect(page.locator("#insight-evidence-EXP-19")).toContainText(
    "選択中",
  );
  await page.locator("#reload-insight-workspace-button").click();
  await expect(page.locator("#insight-workspace-error")).toContainText(
    "一時的に知見を取得できません",
  );
  await page.locator("#retry-insight-workspace-button").click();
  await expect(
    page.locator("#empty-insight-evidence-candidates"),
  ).toContainText("根拠候補がありません");
});

test("派生の作成元は固定条件・結論・可否を表示し、失敗後に再読込する", async ({
  page,
}) => {
  const eligible = {
    source: {
      experimentId: "EXP-20",
      purpose: "派生の比較",
      fixedConditions: {
        fixedConditionId: "condition-20",
        purpose: "派生の比較",
        hypothesis: "条件Aが有効",
        environmentConditions: "Node.js 22",
        initialInput: "入力データ",
        prompts: [{ sequenceNo: 1, content: "条件Aで実行" }],
        evaluationAxes: "正確性",
        fixedAt: confirmedAt,
      },
      conclusion: {
        id: "conclusion-20",
        content: "条件Aを採用します。",
        state: "finalized",
        finalizedAt: confirmedAt,
      },
    },
    eligibility: { canCreateDerivedExperiment: true },
  };
  const ineligible = {
    source: {
      experimentId: "EXP-20",
      purpose: "派生の比較",
    },
    eligibility: {
      canCreateDerivedExperiment: false,
      reasonCode: "CONDITIONS_NOT_FIXED",
    },
  };
  await installDerivationSourceMock(page, [
    { data: eligible },
    { data: eligible },
  ]);
  await page.goto("/experiments/EXP-20/derivations");
  await expect(
    page.locator("#derivation-source-fixed-conditions"),
  ).toContainText("Node.js 22");
  await expect(page.locator("#derivation-source-conclusion")).toContainText(
    "条件Aを採用します。",
  );
  await expect(
    page.getByRole("link", { name: "派生実験を作成" }),
  ).toHaveAttribute("href", "/experiments/EXP-20/derivations/create");
  await expect(
    page.getByRole("button", { name: "壁打ちを開始" }),
  ).toBeEnabled();
  await expect(page.locator("#derivation-source-eligibility")).toContainText(
    "派生元の条件と結論をもとに、相談を開始できます",
  );

  await installDerivationSourceMock(page, [{ data: ineligible }]);
  await page.goto("/experiments/EXP-20/derivations");
  await expect(page.locator("#derivation-source-eligibility")).toContainText(
    "CONDITIONS_NOT_FIXED",
  );
  await expect(
    page.getByRole("button", { name: "派生実験を作成" }),
  ).toBeDisabled();
  await expect(
    page.getByRole("button", { name: "壁打ちを開始" }),
  ).toBeDisabled();

  await installDerivationSourceMock(page, [
    {
      error: {
        code: "EXPERIMENT_DERIVATION_SOURCE_UNAVAILABLE",
        message: "取得失敗",
      },
    },
    {
      error: {
        code: "EXPERIMENT_DERIVATION_SOURCE_UNAVAILABLE",
        message: "取得失敗",
      },
    },
    { data: eligible },
  ]);
  await page.goto("/experiments/EXP-20/derivations");
  await expect(page.locator("#derivation-source-error")).toContainText(
    "取得失敗",
  );
  await page.locator("#reload-derivation-source-button").click();
  await expect(page.locator("#derivation-source-conclusion")).toContainText(
    "条件Aを採用します。",
  );
});

test("派生元から壁打ちを開始し、失敗後は新しい依頼IDで再試行する", async ({
  page,
}) => {
  await installDerivationSourceMock(page, [
    {
      data: {
        source: {
          experimentId: "EXP-22",
          purpose: "派生の比較",
          fixedConditions: {
            fixedConditionId: "condition-22",
            purpose: "派生の比較",
            environmentConditions: "Node.js 22",
            initialInput: "入力データ",
            prompts: [{ sequenceNo: 1, content: "条件Aで実行" }],
            evaluationAxes: "正確性",
            fixedAt: confirmedAt,
          },
          conclusion: {
            id: "conclusion-22",
            content: "条件Aを採用します。",
            state: "finalized",
            finalizedAt: confirmedAt,
          },
        },
        eligibility: { canCreateDerivedExperiment: true },
      },
    },
  ]);
  await installDerivationBriefingMock(page, [
    {
      error: {
        code: "DERIVATION_BRIEFING_START_FAILED",
        message: "壁打ちを開始できませんでした。",
      },
    },
    {
      data: {
        briefingSessionId: "derivation-briefing-22",
        operationId: "derivation-operation-22",
        sourceExperimentId: "EXP-22",
      },
    },
  ]);
  await page.goto("/experiments/EXP-22/derivations");
  await page.locator("#start-derivation-briefing-button").click();
  await expect(page.locator("#derivation-briefing-start-error")).toContainText(
    "壁打ちを開始できませんでした。",
  );
  await page.getByRole("button", { name: "もう一度試す" }).click();
  await expect(page.locator("#derivation-briefing-started")).toContainText(
    "EXP-22",
  );
  const requests = await page.evaluate(
    () => window.__derivationBriefingRequests,
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].sourceExperimentId).toBe("EXP-22");
  expect(requests[1].sourceExperimentId).toBe("EXP-22");
  expect(requests[0].requestId).not.toBe(requests[1].requestId);
});

test("派生実験の壁打ちメッセージを送信し、入力検証・送信中・失敗再試行を扱う", async ({
  page,
}) => {
  await installDerivationSourceMock(page, [
    {
      data: {
        source: {
          experimentId: "EXP-23",
          purpose: "派生の比較",
          fixedConditions: {
            fixedConditionId: "condition-23",
            purpose: "派生の比較",
            environmentConditions: "Node.js 22",
            initialInput: "入力データ",
            prompts: [{ sequenceNo: 1, content: "条件Aで実行" }],
            evaluationAxes: "正確性",
            fixedAt: confirmedAt,
          },
          conclusion: {
            id: "conclusion-23",
            content: "条件Aを採用します。",
            state: "finalized",
            finalizedAt: confirmedAt,
          },
        },
        eligibility: { canCreateDerivedExperiment: true },
      },
    },
  ]);
  await installDerivationBriefingMock(
    page,
    [
      {
        data: {
          briefingSessionId: "derivation-briefing-23",
          operationId: "derivation-operation-23",
          sourceExperimentId: "EXP-23",
        },
      },
    ],
    [
      {
        error: {
          code: "DERIVATION_BRIEFING_MESSAGE_UNAVAILABLE",
          message: "壁打ちメッセージを送信できませんでした。",
        },
      },
      {
        delayMs: 200,
        result: { data: { operationId: "derivation-message-operation-23" } },
      },
    ],
  );

  await page.goto("/experiments/EXP-23/derivations");
  await page.locator("#start-derivation-briefing-button").click();
  const messageInput = page.locator("#derivation-briefing-message-input");
  await page.locator("#send-derivation-briefing-message-button").click();
  await expect(
    page.locator("#derivation-briefing-message-error"),
  ).toContainText("壁打ちへ送る内容を入力してください。");

  await messageInput.fill("比較条件を一つ増やしたいです。");
  await page.locator("#send-derivation-briefing-message-button").click();
  await expect(
    page.locator("#derivation-briefing-message-command-error"),
  ).toContainText("壁打ちメッセージを送信できませんでした。");
  await expect(messageInput).toHaveValue("比較条件を一つ増やしたいです。");

  await page.locator("#send-derivation-briefing-message-button").click();
  await expect(
    page.locator("#derivation-briefing-message-pending"),
  ).toBeVisible();
  await expect(messageInput).toBeDisabled();
  await expect(
    page.locator("#send-derivation-briefing-message-button"),
  ).toBeDisabled();
  await expect(messageInput).toHaveValue("比較条件を一つ増やしたいです。");
  await expect(page.getByRole("button", { name: "閉じる" })).toBeDisabled();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(messageInput).toHaveValue("");
  await expect(page.locator("#derivation-briefing-message-sent")).toContainText(
    "送信を受け付けました。操作ID: derivation-message-operation-23",
  );
  const requests = await page.evaluate(
    () => window.__derivationBriefingMessageRequests,
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].briefingSessionId).toBe("derivation-briefing-23");
  expect(requests[0].message).toBe("比較条件を一つ増やしたいです。");
  expect(requests[1].requestId).not.toBe(requests[0].requestId);
});

test("派生実験の壁打ち内容を再読込し、会話・差分案・未解決事項を表示する", async ({
  page,
}) => {
  await installDerivationSourceMock(page, [
    {
      data: {
        source: {
          experimentId: "EXP-24",
          purpose: "派生の比較",
          fixedConditions: {
            fixedConditionId: "condition-24",
            purpose: "派生の比較",
            environmentConditions: "Node.js 22",
            initialInput: "入力データ",
            prompts: [{ sequenceNo: 1, content: "条件Aで実行" }],
            evaluationAxes: "正確性",
            fixedAt: confirmedAt,
          },
          conclusion: {
            id: "conclusion-24",
            content: "条件Aを採用します。",
            state: "finalized",
            finalizedAt: confirmedAt,
          },
        },
        eligibility: { canCreateDerivedExperiment: true },
      },
    },
  ]);
  await installDerivationBriefingMock(
    page,
    [
      {
        data: {
          briefingSessionId: "derivation-briefing-24",
          operationId: "derivation-operation-24",
          sourceExperimentId: "EXP-24",
        },
      },
    ],
    [],
    [
      {
        delayMs: 200,
        result: {
          data: {
            state: "started",
            lastConfirmedAt: confirmedAt,
            messages: [
              {
                role: "user",
                content: "比較対象を増やしたいです。",
                sequenceNo: 1,
                createdAt: confirmedAt,
              },
            ],
            latestSuggestion: {
              id: "suggestion-24",
              versionNo: 2,
              purpose: "比較対象を増やす",
              decision: "対象Bを追加",
              candidatePrompts: ["候補B"],
              evaluationCriteria: "正確性",
              environmentConditions: "Node.js 22",
              initialInput: "入力データ",
              successCriteria: "差異が確認できる",
              requiredConditions: "同一環境",
              openQuestion: "対象数を決める必要があります。",
              createdAt: confirmedAt,
            },
          },
        },
      },
      {
        error: {
          code: "DERIVATION_BRIEFING_NOT_FOUND",
          message: "壁打ち内容を取得できませんでした。",
        },
      },
      {
        data: { state: "started", lastConfirmedAt: confirmedAt, messages: [] },
      },
    ],
  );
  await page.goto("/experiments/EXP-24/derivations");
  await page.locator("#start-derivation-briefing-button").click();
  await expect(
    page.locator("#derivation-briefing-refresh-pending"),
  ).toBeVisible();
  await expect(page.locator("#derivation-briefing-messages")).toContainText(
    "比較対象を増やしたいです。",
  );
  await expect(page.locator("#derivation-briefing-suggestion")).toContainText(
    "対象Bを追加",
  );
  await expect(
    page.locator("#derivation-briefing-open-question"),
  ).toContainText("対象数を決める必要があります。");
  await page.locator("#reload-derivation-briefing-button").click();
  await expect(
    page.locator("#derivation-briefing-refresh-error"),
  ).toBeVisible();
  await expect(page.locator("#derivation-briefing-messages")).toContainText(
    "比較対象を増やしたいです。",
  );
  await page.getByRole("button", { name: "もう一度試す" }).last().click();
  await expect(page.getByText("会話はまだありません。")).toBeVisible();
  const requests = await page.evaluate(
    () => window.__derivationBriefingGetRequests,
  );
  expect(requests).toEqual([
    "derivation-briefing-24",
    "derivation-briefing-24",
    "derivation-briefing-24",
  ]);
});

test("派生実験の壁打ちを確認して終了し、停止中・失敗再試行を扱う", async ({
  page,
}) => {
  await installDerivationSourceMock(page, [
    {
      data: {
        source: {
          experimentId: "EXP-25",
          purpose: "派生の比較",
          fixedConditions: {
            fixedConditionId: "condition-25",
            purpose: "派生の比較",
            environmentConditions: "Node.js 22",
            initialInput: "入力データ",
            prompts: [{ sequenceNo: 1, content: "条件Aで実行" }],
            evaluationAxes: "正確性",
            fixedAt: confirmedAt,
          },
          conclusion: {
            id: "conclusion-25",
            content: "条件Aを採用します。",
            state: "finalized",
            finalizedAt: confirmedAt,
          },
        },
        eligibility: { canCreateDerivedExperiment: true },
      },
    },
  ]);
  await installDerivationBriefingMock(
    page,
    [
      {
        data: {
          briefingSessionId: "derivation-briefing-25",
          operationId: "derivation-operation-25",
          sourceExperimentId: "EXP-25",
        },
      },
    ],
    [],
    undefined,
    [
      {
        error: {
          code: "DERIVATION_BRIEFING_STOP_UNAVAILABLE",
          message: "壁打ちを終了できませんでした。",
        },
      },
      {
        delayMs: 200,
        result: { data: { operationId: "derivation-stop-operation-25" } },
      },
    ],
  );

  await page.goto("/experiments/EXP-25/derivations");
  await page.locator("#start-derivation-briefing-button").click();
  await expect(page.locator("#derivation-briefing-started")).toBeVisible();
  await page.locator("#request-stop-derivation-briefing-button").click();
  await expect(
    page.locator("#derivation-briefing-stop-confirmation"),
  ).toContainText("壁打ちを終了しますか？");
  await page.locator("#stop-derivation-briefing-button").click();
  await expect(page.locator("#derivation-briefing-stop-error")).toContainText(
    "壁打ちを終了できませんでした。",
  );
  await page.locator("#stop-derivation-briefing-button").click();
  await expect(page.locator("#derivation-briefing-stop-pending")).toBeVisible();
  await expect(
    page.locator("#send-derivation-briefing-message-button"),
  ).toBeDisabled();
  await expect(page.getByRole("button", { name: "閉じる" })).toBeDisabled();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.getByRole("dialog")).toBeHidden();
  const requests = await page.evaluate(
    () => window.__derivationBriefingStopRequests,
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].briefingSessionId).toBe("derivation-briefing-25");
  expect(requests[1].requestId).not.toBe(requests[0].requestId);
});

test("派生実験は差分と理由を検証し、同じ依頼IDで再試行して準備へ遷移する", async ({
  page,
}) => {
  await installCreateDerivedExperimentMock(page, [
    {
      error: {
        code: "DERIVED_EXPERIMENT_UNAVAILABLE",
        message: "派生実験を作成できませんでした。",
      },
    },
    {
      delayMs: 200,
      result: {
        data: {
          requestId: "server-request-id",
          experimentId: "EXP-21-derived",
          sourceExperimentId: "EXP-21",
          state: "preparing",
          createdAt: confirmedAt,
        },
      },
    },
  ]);
  await installExperimentPreparationMock(page, [
    {
      data: {
        experimentId: "EXP-21-derived",
        state: "preparing",
        purpose: "比較対象の条件を変更する",
        environmentConditions: "Node.js 22",
        initialInput: "入力データ",
        prompts: [{ sequenceNo: 1, content: "比較対象で実行する" }],
        evaluationAxes: "正確性",
        source: { state: "derived", versionId: "source-21" },
        requiredFields: {
          purpose: true,
          environmentConditions: true,
          initialInput: true,
          prompts: true,
          evaluationAxes: true,
        },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-21/derivations/create");

  await page.locator("#create-derived-experiment-button").click();
  await expect(page.locator("#create-derived-experiment-error")).toContainText(
    "差分を1項目以上入力し、派生する理由を入力してください。",
  );

  await page.locator("#derived-purpose").fill("比較対象の条件を変更する");
  await page.locator("#derived-reason").fill("評価結果を踏まえて比較するため");
  await page.locator("#create-derived-experiment-button").click();
  await expect(page.locator("#create-derived-experiment-error")).toContainText(
    "派生実験を作成できませんでした。",
  );

  await page.locator("#create-derived-experiment-button").click();
  await expect(
    page.locator("#create-derived-experiment-button"),
  ).toBeDisabled();
  await expect(page.locator("#derived-purpose")).toBeDisabled();
  await expect(page.locator("#derived-reason")).toBeDisabled();
  const requests = await page.evaluate(
    () => window.__createDerivedExperimentRequests,
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].requestId).toBe(requests[1].requestId);
  expect(requests[0]).toMatchObject({
    sourceExperimentId: "EXP-21",
    changes: { Purpose: "比較対象の条件を変更する" },
    reason: "評価結果を踏まえて比較するため",
  });
  await expect(page).toHaveURL("/experiments/EXP-21-derived/preparation");
  await expect(
    page.getByRole("heading", { name: "実験の条件を準備する" }),
  ).toBeVisible();
});

test("評価詳細は不能理由を表示し、取得失敗後に再読込できる", async ({
  page,
}) => {
  await installEvaluationDetailMock(page, [
    {
      data: {
        evaluation: {
          id: "evaluation-26",
          experimentId: "EXP-026",
          runId: "run-26",
          state: "failed",
          updatedAt: confirmedAt,
        },
        operation: {
          id: "evaluation-operation-26",
          state: "failed",
          updatedAt: confirmedAt,
        },
        evidence: { runSummary: "取得済みの根拠", evaluationAxes: "正確性" },
        result: { status: "notRecorded", reasonCode: "EVALUATION_UNAVAILABLE" },
        failure: { code: "EVALUATION_UNAVAILABLE", occurredAt: confirmedAt },
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/evaluations/evaluation-26");
  await expect(page.locator("#evaluation-detail-unavailable")).toContainText(
    "EVALUATION_UNAVAILABLE",
  );

  await installEvaluationDetailMock(page, [
    {
      error: {
        code: "EVALUATION_DETAIL_UNAVAILABLE",
        message: "一時的に取得できません。",
      },
    },
    {
      data: {
        evaluation: {
          id: "evaluation-27",
          experimentId: "EXP-027",
          runId: "run-27",
          state: "completed",
          updatedAt: confirmedAt,
        },
        operation: {
          id: "evaluation-operation-27",
          state: "completed",
          updatedAt: confirmedAt,
        },
        evidence: { runSummary: "再読込後の根拠", evaluationAxes: "正確性" },
        result: { status: "complete", summary: "再読込後の評価結果" },
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/evaluations/evaluation-27");
  await expect(page.locator("#evaluation-detail-error")).toContainText(
    "一時的に取得できません。",
  );
  await page.locator("#reload-evaluation-detail-button").click();
  await expect(page.getByText("再読込後の評価結果")).toBeVisible();
});

test("失敗したrunは固定条件を確認して再実行用runを作成できる", async ({
  page,
}) => {
  const workspace = {
    experimentId: "EXP-028",
    state: "running",
    fixedConditions: {
      fixedConditionId: "fixed-retry-1",
      purpose: "失敗runを再実行する",
      environmentConditions: "同一環境",
      initialInput: "入力",
      prompts: [{ sequenceNo: 1, content: "固定prompt" }],
      evaluationAxes: "正確性",
      fixedAt: confirmedAt,
    },
    conditionFixOperation: { operationId: "operation-retry-1" },
    runs: [
      {
        id: "run-failed-28",
        state: "failed",
        summary: "実行に失敗しました",
        updatedAt: confirmedAt,
      },
    ],
    evaluations: [],
    lastConfirmedAt: confirmedAt,
  };
  await installExperimentWorkspaceMock(page, [{ data: workspace }]);
  await installRetryEndedRunMock(page, [
    {
      delayMs: 300,
      result: {
        error: {
          code: "RUN_RETRY_UNAVAILABLE",
          message: "一時的に作成できません。",
        },
      },
    },
    {
      data: {
        sourceRunId: "run-failed-28",
        experimentId: "EXP-028",
        retryRunId: "run-retry-28",
        operationId: "retry-operation-28",
        state: "queued",
        createdAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-028/workspace");

  await page.locator("#retry-ended-run-button-run-failed-28").click();
  await expect(
    page.locator("#retry-ended-run-dialog-run-failed-28"),
  ).toContainText("元run「run-failed-28」");
  await expect(
    page.locator("#retry-ended-run-dialog-run-failed-28"),
  ).toContainText("fixed-retry-1");
  await page.getByRole("button", { name: "新runを作成" }).click();
  await expect(
    page.getByRole("button", { name: "再実行用runを作成しています…" }),
  ).toBeDisabled();
  await expect(
    page.locator("#retry-ended-run-error-run-failed-28"),
  ).toContainText("一時的に作成できません。");

  await page.locator("#retry-ended-run-button-run-failed-28").click();
  await page.getByRole("button", { name: "新runを作成" }).click();
  await expect(
    page.locator("#retry-ended-run-success-run-failed-28"),
  ).toContainText("新run: run-retry-28（queued）");
});

test("run詳細は観測、差分、照合中と部分取得を表示する", async ({ page }) => {
  await installRunDetailMock(page, [
    {
      data: {
        run: {
          id: "run-22",
          experimentId: "EXP-022",
          state: "running",
          updatedAt: confirmedAt,
        },
        fixedPrompt: { sequenceNo: 1, content: "根拠を添えて要約する" },
        operation: {
          id: "operation-22",
          state: "running",
          updatedAt: confirmedAt,
        },
        observations: [
          {
            sequenceNo: 1,
            kind: "output",
            occurredAt: confirmedAt,
            summary: "利用可能な観測結果",
          },
        ],
        artifacts: {
          status: "partial",
          items: [],
          reasonCode: "ARTIFACT_PENDING",
        },
        reconciliation: { state: "reconciling", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-022/runs/run-22");

  await expect(page.locator("#run-detail-progress")).toContainText(
    "照合しています",
  );
  await expect(page.locator("#run-detail-observation")).toContainText(
    "利用可能な観測結果",
  );
  await expect(page.getByRole("heading", { name: "差分" })).toBeVisible();
  await expect(page.locator("#run-detail-artifacts")).toContainText(
    "partial（ARTIFACT_PENDING）",
  );
  await expect(
    page.getByText("runの完了後に評価を開始できます。"),
  ).toBeVisible();
});

test("run詳細は失敗理由を表示し、失敗後に再読込できる", async ({ page }) => {
  await installRunDetailMock(page, [
    {
      data: {
        run: {
          id: "run-23",
          experimentId: "EXP-023",
          state: "failed",
          updatedAt: confirmedAt,
        },
        fixedPrompt: { sequenceNo: 1, content: "prompt" },
        operation: {
          id: "operation-23",
          state: "failed",
          updatedAt: confirmedAt,
        },
        observations: [],
        artifacts: {
          status: "partial",
          items: [],
          reasonCode: "RUN_EXECUTION_FAILED",
        },
        failure: {
          code: "RUN_EXECUTION_FAILED",
          occurredAt: confirmedAt,
          partialSummary: "安全に表示できる失敗の要約です。",
        },
        reconciliation: { state: "settled", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-023/runs/run-23");
  await expect(page.locator("#run-detail-failure")).toContainText(
    "RUN_EXECUTION_FAILED",
  );

  await installRunDetailMock(page, [
    {
      error: {
        code: "RUN_DETAIL_UNAVAILABLE",
        message: "一時的に取得できません。",
      },
    },
    {
      data: {
        run: {
          id: "run-24",
          experimentId: "EXP-024",
          state: "completed",
          updatedAt: confirmedAt,
        },
        fixedPrompt: { sequenceNo: 1, content: "prompt" },
        operation: {
          id: "operation-24",
          state: "completed",
          updatedAt: confirmedAt,
        },
        observations: [
          {
            sequenceNo: 1,
            kind: "output",
            occurredAt: confirmedAt,
            summary: "再読込後の観測",
          },
        ],
        artifacts: { status: "complete", items: [] },
        reconciliation: { state: "settled", lastObservedAt: confirmedAt },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-024/runs/run-24");
  await expect(page.locator("#run-detail-error")).toContainText(
    "一時的に取得できません。",
  );
  await page.locator("#reload-run-detail-button").click();
  await expect(page.getByText("再読込後の観測")).toBeVisible();
});

test("評価開始失敗は同じrequest IDで再試行し、開始中は二重送信しない", async ({
  page,
}) => {
  const workspace = {
    experimentId: "EXP-021",
    state: "running",
    fixedConditions: {
      fixedConditionId: "fixed-7",
      purpose: "再試行する",
      environmentConditions: "同一環境",
      initialInput: "入力",
      prompts: [{ sequenceNo: 1, content: "prompt" }],
      evaluationAxes: "正確性",
      fixedAt: confirmedAt,
    },
    conditionFixOperation: { operationId: "operation-7" },
    runs: [{ id: "run-21", state: "completed", updatedAt: confirmedAt }],
    evaluations: [],
    lastConfirmedAt: confirmedAt,
  };
  await installExperimentWorkspaceMock(page, [{ data: workspace }], undefined, [
    {
      delayMs: 300,
      result: {
        error: {
          code: "RUN_EVALUATION_PENDING",
          message: "評価を開始中です。",
        },
      },
    },
    {
      error: {
        code: "RUN_EVALUATION_UNAVAILABLE",
        message: "一時的に評価できません。",
      },
    },
    {
      error: {
        code: "RUN_EVALUATION_UNAVAILABLE",
        message: "一時的に評価できません。",
      },
    },
  ]);
  await page.goto("/experiments/EXP-021/workspace");

  const button = page.locator("#start-run-evaluation-button-run-21");
  await button.click();
  await expect(button).toBeDisabled();
  await button.click({ force: true }).catch(() => undefined);
  await expect(page.locator("#run-evaluation-error-run-21")).toContainText(
    "評価を開始中です。",
  );
  await button.click();
  await expect(page.locator("#run-evaluation-error-run-21")).toContainText(
    "一時的に評価できません。",
  );
  await button.click();

  const requests = await page.evaluate(
    () => window.__startRunEvaluationRequests,
  );
  expect(requests).toHaveLength(3);
  expect(requests[0].requestId).toBe(requests[1].requestId);
  expect(requests[1].requestId).toBe(requests[2].requestId);
  expect(requests[0].runId).toBe("run-21");
});

test("実験準備の条件固定中はフォームと下書き保存を無効化する", async ({
  page,
}) => {
  await installExperimentPreparationMock(
    page,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [],
          evaluationAxes: "正確性、要点保持",
          source: { state: "adopted", versionId: "brief-v1" },
          requiredFields: {
            purpose: true,
            environmentConditions: true,
            initialInput: true,
            prompts: false,
            evaluationAxes: true,
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    undefined,
    [{ delayMs: 300, result: {} }],
  );
  await page.goto("/experiments/EXP-015/preparation");
  await page.locator("#fix-conditions-button").click();

  await expect(page.getByRole("button", { name: "固定中…" })).toBeDisabled();
  await expect(page.locator("#save-preparation-draft-button")).toBeDisabled();
  await expect(page.getByLabel("実験目的")).toBeDisabled();
});

test("実験準備の条件固定失敗は入力エラーを関連付け、同じrequest IDで再試行する", async ({
  page,
}) => {
  await installExperimentPreparationMock(
    page,
    [
      {
        data: {
          experimentId: "EXP-015",
          state: "preparing",
          purpose: "問い合わせ要約の品質を比較する",
          environmentConditions: "同じ入力と評価手順を用いる",
          initialInput: "顧客問い合わせ本文",
          prompts: [],
          evaluationAxes: "正確性、要点保持",
          source: { state: "adopted", versionId: "brief-v1" },
          requiredFields: {
            purpose: true,
            environmentConditions: true,
            initialInput: true,
            prompts: false,
            evaluationAxes: true,
          },
          lastConfirmedAt: confirmedAt,
        },
      },
    ],
    undefined,
    [
      {
        error: {
          code: "CONDITIONS_INVALID",
          message: "入力を確認してください。",
          fieldErrors: { purpose: "実験目的を入力してください。" },
        },
      },
      {
        error: {
          code: "FIX_CONDITIONS_SAVE_FAILED",
          message: "一時的に固定できませんでした。",
        },
      },
    ],
  );
  await page.goto("/experiments/EXP-015/preparation");
  await page.locator("#fix-conditions-button").click();
  await expect(page.locator("#preparation-purpose-error")).toHaveText(
    "実験目的を入力してください。",
  );
  await expect(page.getByLabel("実験目的")).toHaveAttribute(
    "aria-describedby",
    "preparation-purpose-error",
  );
  await page.locator("#fix-conditions-button").click();

  const requestIds = await page.evaluate(() =>
    window.__fixConditionsRequests.map((request) => request.requestId),
  );
  expect(requestIds).toHaveLength(2);
  expect(requestIds[0]).toBe(requestIds[1]);
});

test("実験準備の読込中と空状態を表示する", async ({ page }) => {
  await installExperimentPreparationMock(page, [
    {
      delayMs: 200,
      result: {},
    },
  ]);
  await page.goto("/experiments/EXP-015/preparation");

  await expect(page.locator("#preparation-loading")).toBeVisible();
  await expect(page.locator("#preparation-empty")).toBeVisible();
});

test("実験準備の取得失敗から再読込する", async ({ page }) => {
  await installExperimentPreparationMock(page, [
    {
      error: {
        code: "EXPERIMENT_PREPARATION_NOT_FOUND",
        message: "実験準備が見つかりません",
      },
    },
    {
      data: {
        experimentId: "EXP-015",
        state: "preparing",
        purpose: "問い合わせ要約の品質を比較する",
        environmentConditions: "同じ入力と評価手順を用いる",
        initialInput: "顧客問い合わせ本文",
        prompts: [],
        evaluationAxes: "正確性、要点保持",
        source: { state: "adopted", versionId: "brief-v1" },
        requiredFields: {
          purpose: true,
          environmentConditions: true,
          initialInput: true,
          prompts: false,
          evaluationAxes: true,
        },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-015/preparation");

  await expect(page.getByText("対象の実験は見つかりません")).toBeVisible();
  await page.getByRole("button", { name: "再読込" }).click();
  await expect(page.getByText("問い合わせ要約の品質を比較する")).toBeVisible();
});

test("環境準備session一覧を表示する", async ({ page }) => {
  await installListPreparationsMock(page, [
    {
      data: {
        preparations: [
          {
            preparationId: "PREP-001",
            state: "running",
            startedAt: confirmedAt,
            lastObservedAt: confirmedAt,
          },
        ],
      },
    },
  ]);
  await page.goto("/preparations");

  await expect(page.getByRole("heading", { name: "環境準備" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "PREP-001" })).toBeVisible();
  await expect(page.getByText("running")).toBeVisible();
});

test("環境準備sessionの空状態を表示する", async ({ page }) => {
  await installListPreparationsMock(page, [{ data: { preparations: [] } }]);
  await page.goto("/preparations");

  await expect(page.locator("#preparation-list-empty")).toBeVisible();
  await expect(page.getByRole("button", { name: "再読込" })).toBeVisible();
});

test("環境準備の対象範囲を入力検証する", async ({ page }) => {
  await installListPreparationsMock(page, [{ data: { preparations: [] } }]);
  await page.goto("/preparations");

  const scope = page.getByLabel("対象範囲（ワークスペースからの相対パス）");
  await scope.fill("   ");
  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page.locator("#start-preparation-error")).toContainText(
    "対象範囲を入力してください。",
  );

  await scope.fill("/tmp");
  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page.locator("#start-preparation-error")).toContainText(
    "ワークスペースからの相対パス",
  );
});

test("環境準備を開始中に重複操作を防ぎ、成功後に詳細を表示する", async ({
  page,
}) => {
  await installListPreparationsMock(page, [{ data: { preparations: [] } }]);
  await installStartPreparationMock(page, [
    {
      delayMs: 200,
      result: {
        data: { preparationId: "PREP-START-001", state: "completed" },
      },
    },
  ]);
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-START-001",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await page.goto("/preparations");

  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(
    page.getByRole("button", { name: "環境準備を開始しています…" }),
  ).toBeDisabled();
  await expect(
    page.getByLabel("対象範囲（ワークスペースからの相対パス）"),
  ).toBeDisabled();
  await expect(page).toHaveURL("/preparations/PREP-START-001");
  await expect(page.getByText("PREP-START-001", { exact: true })).toBeVisible();
});

test("環境準備の開始終端失敗は新しいrequest IDで再試行する", async ({
  page,
}) => {
  await installListPreparationsMock(page, [{ data: { preparations: [] } }]);
  await installStartPreparationMock(page, [
    {
      error: {
        code: "ACP_NOT_READY",
        message: "環境準備の接続を確認してください。",
      },
    },
    { data: { preparationId: "PREP-START-002", state: "completed" } },
  ]);
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-START-002",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await page.goto("/preparations");

  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page.locator("#start-preparation-error")).toContainText(
    "環境準備の接続を確認してください。",
  );
  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page).toHaveURL("/preparations/PREP-START-002");
  const requests = await page.evaluate(() =>
    JSON.parse(window.localStorage.getItem("startPreparationRequests") || "[]"),
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].requestId).not.toBe(requests[1].requestId);
  expect(requests[0].scope).toBe(".");
});

test("環境準備の通信失敗は同じrequest IDで再送する", async ({ page }) => {
  await installListPreparationsMock(page, [{ data: { preparations: [] } }]);
  await installStartPreparationMock(page, [
    { throwMessage: "connection lost" },
    { data: { preparationId: "PREP-START-003", state: "completed" } },
  ]);
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-START-003",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await page.goto("/preparations");

  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page.locator("#start-preparation-error")).toBeVisible();
  await expect
    .poll(() =>
      page.evaluate(() =>
        window.localStorage.getItem("startPreparationPendingRequest"),
      ),
    )
    .not.toBeNull();
  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page).toHaveURL("/preparations/PREP-START-003");
  const requests = await page.evaluate(() =>
    JSON.parse(window.localStorage.getItem("startPreparationRequests") || "[]"),
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].requestId).toBe(requests[1].requestId);
});

test("環境準備の対象範囲を変更すると新しいrequest IDを使う", async ({
  page,
}) => {
  await installListPreparationsMock(page, [{ data: { preparations: [] } }]);
  await installStartPreparationMock(page, [
    { throwMessage: "connection lost" },
    { data: { preparationId: "PREP-START-004", state: "completed" } },
  ]);
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-START-004",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await page.goto("/preparations");

  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page.locator("#start-preparation-error")).toBeVisible();
  await page
    .getByLabel("対象範囲（ワークスペースからの相対パス）")
    .fill("frontend");
  await page.getByRole("button", { name: "環境準備を開始" }).click();
  await expect(page).toHaveURL("/preparations/PREP-START-004");
  const requests = await page.evaluate(() =>
    JSON.parse(window.localStorage.getItem("startPreparationRequests") || "[]"),
  );
  expect(requests).toHaveLength(2);
  expect(requests[0].requestId).not.toBe(requests[1].requestId);
  expect(requests[1].scope).toBe("frontend");
});

test("環境準備sessionを選択して詳細を表示する", async ({ page }) => {
  await installListPreparationsMock(page, [
    {
      data: {
        preparations: [
          {
            preparationId: "PREP-001",
            state: "completed",
            startedAt: confirmedAt,
            lastObservedAt: confirmedAt,
          },
        ],
      },
    },
  ]);
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-001",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [
          {
            id: "CAND-001",
            environmentConditions: "Node.js 20 / Linux",
            summary: "隔離環境で実行可能です。",
            createdAt: confirmedAt,
          },
        ],
        diagnostics: [
          {
            id: "DIA-001",
            code: "DEPENDENCY_NOTICE",
            summary: "依存関係を確認しました。",
            occurredAt: confirmedAt,
          },
        ],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await page.goto("/preparations");

  await page.getByRole("link", { name: "詳細を確認" }).click();
  await expect(page).toHaveURL("/preparations/PREP-001");
  await expect(
    page.getByRole("heading", { name: "環境準備の詳細" }),
  ).toBeVisible();
  await expect(page.getByText("隔離環境で実行可能です。")).toBeVisible();
  await expect(page.getByText("DEPENDENCY_NOTICE")).toBeVisible();
});

test("完了済みの環境候補を確認して新規実験へ引き継ぐ", async ({ page }) => {
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-ADOPT-001",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [
          {
            id: "CAND-ADOPT-001",
            environmentConditions: "Node.js 20 / Linux",
            summary: "隔離環境で実行可能です。",
            createdAt: confirmedAt,
          },
        ],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await installAdoptCandidateMock(page, [
    {
      delayMs: 200,
      result: {
        data: {
          preparationId: "PREP-ADOPT-001",
          candidateId: "CAND-ADOPT-001",
          environmentConditions: "Node.js 20 / Linux",
        },
      },
    },
  ]);
  await installListExperimentsMock(page, [emptyResponse]);
  await page.goto("/preparations/PREP-ADOPT-001");

  await page.locator("#adopt-candidate-button-CAND-ADOPT-001").click();
  await expect(page.locator("#adopt-candidate-dialog")).toBeVisible();
  await page.getByRole("button", { name: "採用して新規実験へ" }).click();
  await expect(page.locator("#adopt-candidate-pending")).toBeVisible();
  await expect(page.getByRole("button", { name: "採用中…" })).toBeDisabled();
  await expect(page.getByRole("button", { name: "戻る" })).toBeDisabled();
  await expect(page).toHaveURL("/");
  await expect
    .poll(() =>
      page.evaluate(() =>
        window.sessionStorage.getItem("adoptedEnvironmentConditions"),
      ),
    )
    .toBe("Node.js 20 / Linux");
  const requests = await page.evaluate(() => window.__adoptCandidateRequests);
  expect(requests).toEqual([
    {
      requestId: expect.any(String),
      preparationId: "PREP-ADOPT-001",
      candidateId: "CAND-ADOPT-001",
    },
  ]);
});

test("環境候補の採用失敗後は同じrequest IDで再試行できる", async ({ page }) => {
  await installGetPreparationMock(page, [
    {
      data: {
        preparationId: "PREP-ADOPT-002",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [
          {
            id: "CAND-ADOPT-002",
            environmentConditions: "Node.js 20 / Linux",
            summary: "隔離環境で実行可能です。",
            createdAt: confirmedAt,
          },
        ],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await installAdoptCandidateMock(page, [
    {
      error: {
        code: "CANDIDATE_ADOPTION_UNAVAILABLE",
        message: "候補を採用できませんでした。",
      },
    },
    {
      data: {
        preparationId: "PREP-ADOPT-002",
        candidateId: "CAND-ADOPT-002",
        environmentConditions: "Node.js 20 / Linux",
      },
    },
  ]);
  await installListExperimentsMock(page, [emptyResponse]);
  await page.goto("/preparations/PREP-ADOPT-002");

  await page.locator("#adopt-candidate-button-CAND-ADOPT-002").click();
  await page.getByRole("button", { name: "採用して新規実験へ" }).click();
  await expect(page.locator("#adopt-candidate-error")).toBeVisible();
  await page.getByRole("button", { name: "採用して新規実験へ" }).click();
  await expect(page).toHaveURL("/");
  const requests = await page.evaluate(() => window.__adoptCandidateRequests);
  expect(requests).toHaveLength(2);
  expect(requests[0].requestId).toBe(requests[1].requestId);
});

test("採用した環境条件を新規実験の準備フォームへ一度だけ引き継ぐ", async ({
  page,
}) => {
  await page.addInitScript(() => {
    window.sessionStorage.setItem(
      "adoptedEnvironmentConditions",
      "Node.js 20 / Linux",
    );
  });
  await installExperimentPreparationMock(page, [
    {
      data: {
        experimentId: "EXP-ADOPT-001",
        state: "preparing",
        purpose: "候補の環境を確認する",
        environmentConditions: "ブリーフからの環境条件",
        initialInput: "入力",
        prompts: [
          { sequenceNo: 1, content: "prompt 1" },
          { sequenceNo: 2, content: "prompt 2" },
        ],
        evaluationAxes: "正確性",
        source: { state: "adopted", versionId: "brief-v1" },
        requiredFields: {
          purpose: true,
          environmentConditions: true,
          initialInput: true,
          prompts: true,
          evaluationAxes: true,
        },
        lastConfirmedAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-ADOPT-001/preparation");

  await expect(page.locator("#preparation-environment")).toHaveValue(
    "Node.js 20 / Linux",
  );
  await expect
    .poll(() =>
      page.evaluate(() =>
        window.sessionStorage.getItem("adoptedEnvironmentConditions"),
      ),
    )
    .toBeNull();
});

test("環境準備session詳細のloading、候補なし、照合中、失敗を表示する", async ({
  page,
}) => {
  await installGetPreparationMock(page, [
    {
      delayMs: 200,
      result: {
        data: {
          preparationId: "PREP-002",
          state: "failed",
          startedAt: confirmedAt,
          lastObservedAt: confirmedAt,
          candidates: [],
          diagnostics: [],
          failure: { code: "PREPARATION_TIMEOUT", occurredAt: confirmedAt },
          reconciliation: { state: "reconciling", lastObservedAt: confirmedAt },
        },
      },
    },
  ]);
  await page.goto("/preparations/PREP-002");

  await expect(page.locator("#preparation-detail-loading")).toBeVisible();
  await expect(page.locator("#preparation-candidates-empty")).toBeVisible();
  await expect(page.locator("#preparation-reconciling")).toBeVisible();
  await expect(page.locator("#preparation-failure")).toContainText(
    "PREPARATION_TIMEOUT",
  );
});

test("環境準備session詳細の再読込失敗から回復する", async ({ page }) => {
  await installGetPreparationMock(page, [
    {
      error: {
        code: "PREPARATION_UNAVAILABLE",
        message: "一時的に取得できません。",
      },
    },
    {
      data: {
        preparationId: "PREP-003",
        state: "completed",
        startedAt: confirmedAt,
        lastObservedAt: confirmedAt,
        candidates: [],
        diagnostics: [],
        reconciliation: { state: "confirmed", lastObservedAt: confirmedAt },
      },
    },
  ]);
  await page.goto("/preparations/PREP-003");

  await expect(page.locator("#preparation-detail-error")).toBeVisible();
  await page.getByRole("button", { name: "再読込" }).click();
  await expect(page.getByText("PREP-003", { exact: true })).toBeVisible();
});

test("知見を根拠2件で記録し、失敗後に同じrequest IDで再試行する", async ({
  page,
}) => {
  const workspace = {
    data: {
      evidenceCandidates: [
        {
          experimentId: "EXP-27-A",
          purpose: "条件Aの比較",
          evaluationAxes: "正確性",
          conclusionId: "CON-27-A",
          conclusion: "条件Aは検証可能性を高める。",
          finalizedAt: confirmedAt,
        },
        {
          experimentId: "EXP-27-B",
          purpose: "条件Bの比較",
          evaluationAxes: "再現性",
          conclusionId: "CON-27-B",
          conclusion: "条件Bには追加検証が必要。",
          finalizedAt: confirmedAt,
        },
      ],
      savedConsiderations: [],
      insights: [],
      lastConfirmedAt: confirmedAt,
    },
  };
  await installInsightWorkspaceMock(page, [workspace]);
  await installCreateInsightMock(page, [
    {
      error: {
        code: "INSIGHT_CREATE_UNAVAILABLE",
        message: "一時的に記録できません。",
      },
    },
    {
      data: {
        requestId: "ignored-by-ui",
        insightId: "INS-27",
        evidences: [
          { experimentId: "EXP-27-A", conclusionId: "CON-27-A" },
          { experimentId: "EXP-27-B", conclusionId: "CON-27-B" },
        ],
        statement: "条件を明示すると検証可能性が高まる。",
        applicabilityConditions: "同一環境で比較する。",
        verificationGaps: "異なる環境で再検証する。",
        createdAt: confirmedAt,
      },
    },
  ]);
  await page.goto("/experiments/EXP-27-A/insights");
  await page
    .getByRole("group", { name: "根拠を選択" })
    .getByRole("checkbox")
    .nth(1)
    .check();
  await page
    .locator("#insight-statement")
    .fill("条件を明示すると検証可能性が高まる。");
  await page
    .locator("#insight-applicability-conditions")
    .fill("同一環境で比較する。");
  await page
    .locator("#insight-verification-gaps")
    .fill("異なる環境で再検証する。");
  await page.locator("#open-create-insight-dialog-button").click();
  const dialog = page.locator("#create-insight-dialog");
  await dialog.getByRole("button", { name: "知見を記録" }).click();
  await expect(dialog.getByRole("alert")).toContainText(
    "一時的に記録できません。",
  );
  await dialog.getByRole("button", { name: "知見を記録" }).click();
  await expect(page.locator("#insight-list-title")).toContainText("既存知見");
  await expect(page.getByText("INS-27")).toBeVisible();
  const requestIds = await page.evaluate(() =>
    window.__createInsightRequests.map((request) => request.requestId),
  );
  expect(requestIds).toHaveLength(2);
  expect(requestIds[0]).toBe(requestIds[1]);
});

test("環境準備session一覧の失敗から再読込する", async ({ page }) => {
  await installListPreparationsMock(page, [
    {
      error: {
        code: "PREPARATIONS_UNAVAILABLE",
        message: "準備session一覧を取得できませんでした",
      },
    },
    {
      data: {
        preparations: [
          {
            preparationId: "PREP-002",
            state: "completed",
            startedAt: confirmedAt,
            lastObservedAt: confirmedAt,
          },
        ],
      },
    },
  ]);
  await page.goto("/preparations");

  await expect(page.locator("#preparation-query-error")).toBeVisible();
  await page.getByRole("button", { name: "再読込" }).click();
  await expect(page.getByRole("heading", { name: "PREP-002" })).toBeVisible();
});
