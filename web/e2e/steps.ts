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
  await expect(
    nav.getByRole("group", { name: group, exact: true }).getByText(item, { exact: true }),
  ).toBeVisible();
});

Given(
  "a global section {string} with a context {string} and content {string} and description {string}",
  async ({ pcm }, section: string, name: string, content: string, description: string) => {
    await pcm.post("/api/v1/sections", { scope: "global", name: section }, 201, 409);
    await pcm.post("/api/v1/contexts", {
      scope: "global",
      section,
      name,
      content,
      frontmatter: { description },
    });
  },
);

When("I click {string} in the sidebar", async ({ page }, text: string) => {
  await page.getByRole("navigation", { name: "Contexts" }).getByText(text, { exact: true }).click();
});

When("I click {string} in the search results", async ({ page }, text: string) => {
  await page
    .getByRole("region", { name: "Search results" })
    .getByText(text, { exact: true })
    .click();
});

When("I click the {string} link", async ({ page }, name: string) => {
  await page.getByRole("link", { name, exact: true }).click();
});

When("I type {string} in {string}", async ({ page }, text: string, label: string) => {
  await page.getByLabel(label, { exact: true }).fill(text);
});

When("I select {string} in {string}", async ({ page }, option: string, label: string) => {
  await page.getByRole("combobox", { name: label, exact: true }).click();
  await page.getByRole("option", { name: option, exact: true }).click();
});

Then("the URL path is {string}", async ({ page }, path: string) => {
  await expect.poll(() => new URL(page.url()).pathname).toBe(path);
});

Then("I see {string} in bold", async ({ page }, text: string) => {
  await expect(page.locator("strong", { hasText: text }).first()).toBeVisible();
});

Then("I see the raw file {string}", async ({ page }, path: string) => {
  await expect(page.getByRole("region", { name: "Raw file" })).toContainText(path);
});

Then("the sidebar does not show {string}", async ({ page }, text: string) => {
  await expect(
    page.getByRole("navigation", { name: "Contexts" }).getByText(text, { exact: true }),
  ).toHaveCount(0);
});

Then("the search results show {string}", async ({ page }, text: string) => {
  await expect(
    page.getByRole("region", { name: "Search results" }).getByText(text, { exact: true }),
  ).toBeVisible();
});

Then("the search results do not show {string}", async ({ page }, text: string) => {
  await expect(
    page.getByRole("region", { name: "Search results" }).getByText(text, { exact: true }),
  ).toHaveCount(0);
});

// Rows are table body rows whose cells include the exact text.
function rowsWith(page: import("@playwright/test").Page, text: string) {
  return page.getByRole("row").filter({ has: page.getByRole("cell", { name: text, exact: true }) });
}

Then("the table has {int} row(s) with {string}", async ({ page }, n: number, text: string) => {
  await expect(rowsWith(page, text)).toHaveCount(n);
});

Then("the row with {string} contains {string}", async ({ page }, text: string, want: string) => {
  await expect(rowsWith(page, text).first()).toContainText(want);
});
