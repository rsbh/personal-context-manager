import { expect } from "@playwright/test";
import { createBdd } from "playwright-bdd";
import { test } from "./fixtures";

const { Given, When, Then } = createBdd(test);

// Setup steps: same text as the godog steps in e2e/steps_test.go.

Given("the server is running", async ({ pcm }) => {
  const r = pcm.run("server", "start");
  expect(r.code, r.output).toBe(0);
});

Given(
  "a global section {string} with a context {string} and content {string}",
  async ({ pcm }, section: string, name: string, content: string) => {
    await pcm.post("/api/v1/sections", { scope: "global", name: section }, 201, 409);
    await pcm.post("/api/v1/contexts", { scope: "global", section, name, content });
  },
);

Given(
  "a project {string} section {string} with a context {string} and content {string}",
  async ({ pcm }, project: string, section: string, name: string, content: string) => {
    await pcm.post("/api/v1/projects", { name: project }, 201, 409);
    await pcm.post("/api/v1/sections", { scope: "project", project, name: section }, 201, 409);
    await pcm.post("/api/v1/contexts", { scope: "project", project, section, name, content });
  },
);

// Browser steps.

When("I open the UI", async ({ page, pcm }) => {
  await page.goto(pcm.url("/"));
});

When("I open the UI at {string}", async ({ page, pcm }, path: string) => {
  await page.goto(pcm.url(path));
});

Then("the page title is {string}", async ({ page }, title: string) => {
  await expect(page).toHaveTitle(title);
});

Then("I see {string}", async ({ page }, text: string) => {
  await expect(page.getByText(text).first()).toBeVisible();
});

// The sidebar is <nav aria-label="Contexts">; each group is role="group" with aria-label = its name.
Then("the sidebar shows {string} under {string}", async ({ page }, item: string, group: string) => {
  const nav = page.getByRole("navigation", { name: "Contexts" });
  await expect(nav.getByRole("group", { name: group, exact: true }).getByText(item, { exact: true })).toBeVisible();
});
