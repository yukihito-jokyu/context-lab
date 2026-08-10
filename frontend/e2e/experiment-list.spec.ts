import { expect, type Page, test } from "@playwright/test";

type ListExperimentsResponse = Record<string, unknown>;
type StartExperimentBriefingResponse = Record<string, unknown>;
type GetExperimentBriefingResponse = Record<string, unknown>;
type SendExperimentBriefMessageResponse = Record<string, unknown>;
type CreateExperimentFromBriefResponse = Record<string, unknown>;
type StopExperimentBriefingResponse = Record<string, unknown>;
type GetExperimentPreparationResponse = Record<string, unknown>;

declare global {
  interface Window {
    __briefingGetCallCount: number;
    __briefingMessageRequests: Array<{
      requestId: string;
      briefingSessionId: string;
      message: string;
    }>;
    __briefingRequestIds: string[];
    __briefingStopRequests: Array<{
      requestId: string;
      briefingSessionId: string;
    }>;
    __createExperimentRequests: Array<{
      requestId: string;
      briefingSessionId: string;
      briefVersionId: string;
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
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      let callCount = 0;
      window.go = window.go || { wails: {} };
      window.go.wails.ExperimentPreparationsHandler = {
        GetExperimentPreparation: () => {
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
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

test("完全なブリーフを採用して準備画面へ遷移する", async ({ page }) => {
  await installListExperimentsMock(page, [successResponse]);
  await installExperimentBriefingMock(
    page,
    [{ data: { briefingSessionId: "briefing-1", operationId: "operation-1" } }],
    [completeBriefResponse],
    [],
    [{ data: { experimentId: "EXP-015", state: "preparing" } }],
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page.locator("#submit-create-experiment-button").click();
  await expect(page).toHaveURL(/\/experiments\/EXP-015\/preparation$/);
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
  await expect(page.getByText("根拠を保って要約する")).toBeVisible();
  await expect(page.getByText("入力済み")).toHaveCount(5);
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
