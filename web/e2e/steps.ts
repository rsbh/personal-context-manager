import { mkdirSync, mkdtempSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
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

Given("a global section {string}", async ({ pcm }, name: string) => {
  await pcm.post("/api/v1/sections", { scope: "global", name });
});

Given(
  "a project {string} with a section {string}",
  async ({ pcm }, project: string, name: string) => {
    await pcm.post("/api/v1/projects", { name: project });
    await pcm.post("/api/v1/sections", { scope: "project", project, name });
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

// The sidebar is <nav aria-label="Contexts">; each group is a listitem with aria-label = its name.
Then("the sidebar shows {string} under {string}", async ({ page }, item: string, group: string) => {
  const nav = page.getByRole("navigation", { name: "Contexts" });
  await expect(
    nav.getByRole("listitem", { name: group, exact: true }).getByText(item, { exact: true }),
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

// Edit steps.

When("I click the {string} button", async ({ page }, name: string) => {
  await page.getByRole("button", { name, exact: true }).click();
});

// dialog is the open dialog: a Dialog (role dialog) or an AlertDialog (role alertdialog).
function dialog(page: import("@playwright/test").Page, name?: string) {
  return page.getByRole("dialog", { name }).or(page.getByRole("alertdialog", { name }));
}

When("I click the {string} button in the dialog", async ({ page }, name: string) => {
  await dialog(page).getByRole("button", { name, exact: true }).click();
});

// The CodeMirror content element is a textbox with aria-label "Markdown".
function editor(page: import("@playwright/test").Page) {
  return page.getByRole("textbox", { name: "Markdown", exact: true });
}

When("I type {string} in the editor", async ({ page }, text: string) => {
  await editor(page).fill(text);
});

Then("the editor contains {string}", async ({ page }, text: string) => {
  await expect(editor(page)).toContainText(text);
});

When("I add the tag {string}", async ({ page }, tag: string) => {
  const input = page.getByLabel("Tags", { exact: true });
  await input.fill(tag);
  await input.press("Enter");
});

Then("I do not see {string}", async ({ page }, text: string) => {
  await expect(page.getByText(text, { exact: true })).toHaveCount(0);
});

Then("I see the dialog {string}", async ({ page }, name: string) => {
  await expect(dialog(page, name)).toBeVisible();
});

Then("I see the toast {string}", async ({ page }, text: string) => {
  await expect(page.getByRole("region", { name: /notifications/i })).toContainText(text);
});

Then("the {string} select shows {string}", async ({ page }, label: string, value: string) => {
  await expect(page.getByRole("combobox", { name: label, exact: true })).toHaveText(value);
});

// Section and project steps.

When("I click {string} in the menu", async ({ page }, name: string) => {
  await page.getByRole("menuitem", { name, exact: true }).click();
});

Then("the {string} field is {string}", async ({ page }, label: string, value: string) => {
  await expect(page.getByLabel(label, { exact: true })).toHaveValue(value);
});

// Git and sync steps: same text as the godog steps in e2e/steps_test.go.

Given("a git remote", async ({ pcm }) => {
  pcm.remote = join(pcm.home, "remote.git");
  pcm.git(pcm.home, "init", "--quiet", "--bare", "-b", "main", pcm.remote);
});

Given("the remote is set", async ({ pcm }) => {
  const r = pcm.run("config", "set", "remote", pcm.remote);
  expect(r.code, r.output).toBe(0);
});

Given("I run {string}", async ({ pcm }, command: string) => {
  const [bin, ...args] = command.split(" ");
  expect(bin).toBe("pcm");
  const r = pcm.run(...args);
  expect(r.code, r.output).toBe(0);
});

Given(
  "another clone pushes {string} with content {string}",
  async ({ pcm }, path: string, content: string) => {
    const dir = mkdtempSync(join(pcm.home, "clone-"));
    pcm.git(pcm.home, "clone", "--quiet", pcm.remote, dir);
    const file = join(dir, path);
    mkdirSync(dirname(file), { recursive: true });
    writeFileSync(file, content);
    pcm.git(dir, "add", "-A");
    pcm.git(dir, "commit", "--quiet", "-m", "change from other clone");
    pcm.git(dir, "push", "--quiet", "origin", "main");
  },
);

Given("the context {string} has content {string}", async ({ pcm }, id: string, content: string) => {
  await pcm.put(`/api/v1/contexts/${id}`, { content });
});

Then("the remote has the file {string}", async ({ pcm }, path: string) => {
  const out = pcm.git(pcm.home, "--git-dir", pcm.remote, "ls-tree", "-r", "--name-only", "main");
  expect(out.split("\n")).toContain(path);
});

// The Sync link is in the header nav; its badge is the number of changes + commits ahead.
function syncLink(page: import("@playwright/test").Page) {
  return page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: /^Sync/ });
}

Then("the Sync badge shows {string}", async ({ page }, count: string) => {
  await expect(syncLink(page)).toHaveText(`Sync${count}`);
});

Then("the Sync badge is not shown", async ({ page }) => {
  await expect(syncLink(page)).toHaveText("Sync");
});

// Status cards are groups named by their label.
Then("the {string} card shows {string}", async ({ page }, label: string, value: string) => {
  await expect(page.getByRole("group", { name: label, exact: true })).toHaveText(
    `${label}${value}`,
  );
});

Then("the changed files show {string}", async ({ page }, path: string) => {
  await expect(page.getByRole("region", { name: "Changed files" })).toContainText(path);
});

Then("the recent commits show {string}", async ({ page }, text: string) => {
  await expect(page.getByRole("region", { name: "Recent commits" })).toContainText(text);
});

Then("the {string} button is disabled", async ({ page }, name: string) => {
  await expect(page.getByRole("button", { name, exact: true })).toBeDisabled();
});
