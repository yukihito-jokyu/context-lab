import { expect, type Page, test } from "@playwright/test";

type ListExperimentsResponse = Record<string, unknown>;
type StartExperimentBriefingResponse = Record<string, unknown>;
type GetExperimentBriefingResponse = Record<string, unknown>;

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

async function installExperimentBriefingMock(
  page: Page,
  responses: StartExperimentBriefingResponse[],
  briefingResponses: GetExperimentBriefingResponse[] = [],
) {
  await page.addInitScript({
    content: `
      const responses = ${JSON.stringify(responses)};
      const briefingResponses = ${JSON.stringify(briefingResponses)};
      let callCount = 0;
      let briefingCallCount = 0;
      window.go = window.go || { wails: {} };
      window.__briefingRequestIds = [];
      window.go.wails.ExperimentBriefingsHandler = {
        StartExperimentBriefing: (requestId) => {
          window.__briefingRequestIds.push(requestId);
          const response = responses[Math.min(callCount, responses.length - 1)];
          callCount += 1;
          return Promise.resolve(response);
        },
        GetExperimentBriefing: () => {
          const response = briefingResponses[Math.min(briefingCallCount, briefingResponses.length - 1)];
          briefingCallCount += 1;
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
  );
  await page.goto("/");

  await page.locator("#new-experiment-button").click();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).not.toBeVisible();
  await page.locator("#new-experiment-button").click();
  await expect(page.locator("#briefing-chat-log")).toContainText("新しい会話");
  await page.waitForTimeout(300);
  await expect(page.locator("#briefing-chat-log")).not.toContainText(
    "古い会話",
  );
});
