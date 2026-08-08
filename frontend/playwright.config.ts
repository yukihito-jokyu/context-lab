import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  failOnFlakyTests: Boolean(process.env.CI),
  outputDir: "test-results",
  reporter: [["html", { outputFolder: "playwright-report", open: "never" }]],
  use: {
    trace: "on-first-retry",
    video: "on",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
