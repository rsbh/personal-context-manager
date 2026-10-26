import { defineConfig } from "@playwright/test";
import { defineBddConfig } from "playwright-bdd";

// UI specs live with the other specs; only @ui features run here.
const testDir = defineBddConfig({
  features: "../e2e/features/ui/*.feature",
  featuresRoot: "../e2e/features/ui",
  steps: "e2e/*.ts",
});

export default defineConfig({
  testDir,
  // Each scenario has its own HOME, port and pcm-server, so scenarios can run in parallel.
  fullyParallel: true,
  // CI runner has 4 vCPUs: one worker per vCPU. Local: Playwright default.
  workers: process.env.CI ? 4 : undefined,
  globalSetup: "./e2e/global-setup.ts",
  reporter: process.env.CI ? "list" : "line",
  use: { browserName: "chromium", trace: "retain-on-failure" },
});
