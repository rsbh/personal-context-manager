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
  globalSetup: "./e2e/global-setup.ts",
  reporter: process.env.CI ? "list" : "line",
  use: { browserName: "chromium", trace: "retain-on-failure" },
});
