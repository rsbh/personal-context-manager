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

When("I press {string}", async ({ page }, key: string) => {
  await page.keyboard.press(key);
});

// The palette is a dialog named "Search contexts".
function palette(page: import("@playwright/test").Page) {
  return page.getByRole("dialog", { name: "Search contexts" });
}

When("I click {string} in the palette", async ({ page }, text: string) => {
  await palette(page).getByRole("option", { name: text }).click();
});

When("I click the {string} link", async ({ page }, name: string) => {
  await page.getByRole("link", { name, exact: true }).click();
});

When("I type {string} in {string}", async ({ page }, text: string, label: string) => {
  // Only fields: a dialog can have the same label as its input.
  await page
    .getByLabel(label, { exact: true })
    .and(page.locator("input, textarea, [contenteditable]"))
    .fill(text);
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

Then("the palette shows {string}", async ({ page }, text: string) => {
  await expect(palette(page).getByText(text, { exact: true })).toBeVisible();
});

Then("the palette does not show {string}", async ({ page }, text: string) => {
  await expect(palette(page).getByText(text, { exact: true })).toHaveCount(0);
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

Then("I do not see the dialog {string}", async ({ page }, name: string) => {
  await expect(dialog(page, name)).toHaveCount(0);
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

// The Sync link name includes the badge count ("Sync 1"), so match by prefix.
When("I click the Sync link", async ({ page }) => {
  await syncLink(page).click();
});

// Sidebar tree steps.

// heading is the collapse button of a sidebar section.
function heading(page: import("@playwright/test").Page, name: string) {
  return page
    .getByRole("navigation", { name: "Contexts" })
    .getByRole("button", { name, exact: true });
}

When("I click the {string} section heading", async ({ page }, name: string) => {
  await heading(page, name).click();
});

async function fontSize(loc: import("@playwright/test").Locator): Promise<number> {
  return loc.evaluate((el) => parseFloat(getComputedStyle(el).fontSize));
}

Then(
  "the {string} heading font size is larger than the {string} item",
  async ({ page }, head: string, item: string) => {
    const nav = page.getByRole("navigation", { name: "Contexts" });
    const h = await fontSize(heading(page, head));
    const i = await fontSize(nav.locator('a[href^="/contexts/"]').filter({ hasText: item }));
    expect(h).toBeGreaterThan(i);
  },
);

// rows returns the bounding boxes of the item links, top to bottom.
async function rows(page: import("@playwright/test").Page) {
  const links = page.getByRole("navigation", { name: "Contexts" }).locator('a[href^="/contexts/"]');
  const boxes = await links.evaluateAll((els) =>
    els.map((el) => {
      const r = el.getBoundingClientRect();
      return { top: r.top, height: r.height };
    }),
  );
  return boxes.toSorted((a, b) => a.top - b.top);
}

Then("all item rows have the same height", async ({ page }) => {
  const heights = (await rows(page)).map((r) => Math.round(r.height));
  expect(heights.length).toBeGreaterThan(1);
  expect(new Set(heights).size).toBe(1);
});

Then("the gaps between rows are equal", async ({ page }) => {
  const r = await rows(page);
  const gaps = r.slice(1).map((b, i) => Math.round(b.top - (r[i].top + r[i].height)));
  expect(gaps.length).toBeGreaterThan(0);
  expect(new Set(gaps).size).toBe(1);
});

Then("the {string} button is at least {int}px high", async ({ page }, name: string, px: number) => {
  const box = await page.getByRole("button", { name, exact: true }).boundingBox();
  expect(Math.round(box?.height ?? 0)).toBeGreaterThanOrEqual(px);
});

Then("the {string} button is at least {int}px wide", async ({ page }, name: string, px: number) => {
  const box = await page.getByRole("button", { name, exact: true }).boundingBox();
  expect(box?.width ?? 0).toBeGreaterThanOrEqual(px);
});

Then(
  "the gap below the {string} heading equals the gaps between rows",
  async ({ page }, name: string) => {
    const head = await heading(page, name).boundingBox();
    const r = await rows(page);
    expect(head).not.toBeNull();
    const below = Math.round(r[0].top - ((head?.y ?? 0) + (head?.height ?? 0)));
    const gap = Math.round(r[1].top - (r[0].top + r[0].height));
    expect(below).toBe(gap);
  },
);

Given("the window is {int}px wide", async ({ page }, px: number) => {
  await page.setViewportSize({ width: px, height: 800 });
});

Then("the page does not scroll sideways", async ({ page }) => {
  await expect(page.getByRole("table")).toBeVisible();
  const widths = await page.evaluate(() => [document.documentElement.scrollWidth, innerWidth]);
  expect(widths[0]).toBeLessThanOrEqual(widths[1]);
});

// Theme: the Apsara theme element carries the settings as data attributes.
const themeRoot = (page: import("@playwright/test").Page) =>
  page.locator("[data-accent-color]").first();

// resolved returns the computed value of a CSS property on a probe element,
// so tokens compare as real colours in light and dark mode.
async function resolved(
  page: import("@playwright/test").Page,
  prop: "color" | "backgroundColor",
  value: string,
) {
  return themeRoot(page).evaluate(
    (root, [p, v]) => {
      const probe = document.createElement("span");
      probe.style[p as "color"] = v;
      root.append(probe);
      const out = getComputedStyle(probe)[p as "color"];
      probe.remove();
      return out;
    },
    [prop, value],
  );
}

Then("the theme accent color is {string}", async ({ page }, value: string) => {
  await expect(themeRoot(page)).toHaveAttribute("data-accent-color", value);
});

Then("the theme gray color is {string}", async ({ page }, value: string) => {
  await expect(themeRoot(page)).toHaveAttribute("data-gray-color", value);
});

Then("the body text size is {int}px", async ({ page }, px: number) => {
  await expect(themeRoot(page)).toHaveCSS("font-size", `${px}px`);
});

Then("the main text color is gray step {int}", async ({ page }, step: number) => {
  await expect(themeRoot(page)).toHaveCSS(
    "color",
    await resolved(page, "color", `var(--rs-gray-${step})`),
  );
});

Then("the page background is gray step {int}", async ({ page }, step: number) => {
  await expect(themeRoot(page)).toHaveCSS(
    "background-color",
    await resolved(page, "backgroundColor", `var(--rs-gray-${step})`),
  );
});

Then("the secondary text color is gray step {int}", async ({ page }, step: number) => {
  const secondary = await resolved(page, "color", "var(--rs-color-foreground-base-secondary)");
  expect(secondary).toBe(await resolved(page, "color", `var(--rs-gray-${step})`));
});

// Header: the main navigation is a nav landmark named "Main".
const mainNav = (page: import("@playwright/test").Page) =>
  page.getByRole("navigation", { name: "Main" });

async function navBoxes(page: import("@playwright/test").Page) {
  const links = await mainNav(page).getByRole("link").all();
  const boxes = await Promise.all(links.map((l) => l.boundingBox()));
  return boxes.filter((b) => b !== null);
}

Then("the header is {int}px high", async ({ page }, px: number) => {
  // clientHeight leaves out the 0.5px bottom border: screens draw it as 0.5px or 1px.
  const height = await page.getByRole("banner").evaluate((el) => el.clientHeight);
  expect(height).toBe(px);
});

Then("each main navigation link is at least {int}px high", async ({ page }, px: number) => {
  const boxes = await navBoxes(page);
  expect(boxes.length).toBeGreaterThan(1);
  for (const b of boxes) expect(Math.round(b.height)).toBeGreaterThanOrEqual(px);
});

Then("the gap between main navigation links is at least {int}px", async ({ page }, px: number) => {
  const boxes = await navBoxes(page);
  const gaps = boxes.slice(1).map((b, i) => Math.round(b.x - (boxes[i].x + boxes[i].width)));
  expect(gaps.length).toBeGreaterThan(0);
  for (const g of gaps) expect(g).toBeGreaterThanOrEqual(px);
});

Then("the {string} navigation link is the current page", async ({ page }, name: string) => {
  await expect(mainNav(page).getByRole("link", { name })).toHaveAttribute("aria-current", "page");
});

Then("the {string} navigation link is not the current page", async ({ page }, name: string) => {
  await expect(mainNav(page).getByRole("link", { name })).not.toHaveAttribute("aria-current");
});

Then("the page body has no margin", async ({ page }) => {
  await expect(page.locator("body")).toHaveCSS("margin", "0px");
});

// Scope overview pages: steps look only inside main, the sidebar has the same names.
const overview = (page: import("@playwright/test").Page) => page.getByRole("main");

Then("I see the heading {string}", async ({ page }, name: string) => {
  await expect(page.getByRole("heading", { name, exact: true })).toBeVisible();
});

Then(
  "the overview section {string} lists {string}",
  async ({ page }, section: string, name: string) => {
    const card = overview(page).getByRole("region", { name: section, exact: true });
    await expect(card.getByRole("link", { name, exact: true })).toBeVisible();
  },
);

When("I click {string} in the overview", async ({ page }, name: string) => {
  await overview(page).getByRole("link", { name, exact: true }).click();
});

When("I click the {string} button in the overview", async ({ page }, name: string) => {
  await overview(page).getByRole("button", { name, exact: true }).click();
});

Then("the overview shows the section {string}", async ({ page }, name: string) => {
  await expect(overview(page).getByRole("region", { name, exact: true })).toBeVisible();
});

Then("the overview does not show the section {string}", async ({ page }, name: string) => {
  await expect(overview(page).getByRole("region", { name, exact: true })).toHaveCount(0);
});

Then("the overview does not show {string}", async ({ page }, name: string) => {
  await expect(overview(page).getByRole("link", { name, exact: true })).toHaveCount(0);
});

for (const tone of ["success", "danger"]) {
  Then(`the {string} button in the overview is ${tone} colour`, async ({ page }, name: string) => {
    const button = overview(page).getByRole("button", { name, exact: true });
    await expect(button).toHaveCSS(
      "color",
      await resolved(page, "color", `var(--rs-color-foreground-${tone}-primary)`),
    );
  });
}
