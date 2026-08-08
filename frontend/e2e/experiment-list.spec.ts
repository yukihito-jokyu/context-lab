import { expect, type Page, test } from "@playwright/test";

type ListExperimentsResponse = Record<string, unknown>;

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
