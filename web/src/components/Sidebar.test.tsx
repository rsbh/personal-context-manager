import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { Sidebar } from "./Sidebar";

// Apsara sidebar items are <a href> with role="listitem" (list semantics), so find them by text.
function anchor(el: HTMLElement): HTMLAnchorElement | null {
  return el.closest("a");
}

const globalContexts = [
  { id: "global/git/commits", scope: "global", section: "git", name: "commits" },
  { id: "global/style/go", scope: "global", section: "style", name: "go" },
];

function base(extra: Record<string, unknown> = {}) {
  return mockAPI({
    "/api/v1/sections?scope=global": ["git", "style", "empty"],
    "/api/v1/contexts?scope=global": globalContexts,
    "/api/v1/projects": ["api"],
    ...extra,
  });
}

describe("Sidebar", () => {
  it("lists global sections with their contexts, and projects", async () => {
    base();
    renderUI(<Sidebar />);
    const nav = screen.getByRole("navigation", { name: "Contexts" });
    const global = within(nav).getByRole("listitem", { name: "Global" });
    expect(await within(global).findByText("commits")).toBeInTheDocument();
    expect(within(global).getByText("go")).toBeInTheDocument();
    expect(within(global).getByText("empty")).toBeInTheDocument();
    expect(anchor(within(global).getByText("go"))).toHaveAttribute(
      "href",
      "/contexts/global/style/go",
    );
    expect(await within(nav).findByText("api")).toBeInTheDocument();
  });

  it("loads project sections and contexts when a project opens", async () => {
    base({
      "/api/v1/sections?scope=project&project=api": ["review"],
      "/api/v1/contexts?scope=project&project=api": [
        {
          id: "projects/api/review/pr",
          scope: "project",
          project: "api",
          section: "review",
          name: "pr",
        },
      ],
    });
    renderUI(<Sidebar />);
    const projects = screen.getByRole("listitem", { name: "Projects" });
    await userEvent.click(await within(projects).findByText("api"));
    expect(anchor(await within(projects).findByText("pr"))).toHaveAttribute(
      "href",
      "/contexts/projects/api/review/pr",
    );
  });

  it("shows empty states", async () => {
    mockAPI({
      "/api/v1/sections?scope=global": [],
      "/api/v1/contexts?scope=global": [],
      "/api/v1/projects": [],
    });
    renderUI(<Sidebar />);
    expect(await screen.findByText("No sections")).toBeInTheDocument();
    expect(await screen.findByText("No projects")).toBeInTheDocument();
  });

  it("shows search results instead of the tree while searching", async () => {
    base({
      "/api/v1/search?q=black": [
        {
          id: "global/style/py",
          scope: "global",
          section: "style",
          name: "py",
          snippet: "Use <mark>black</mark>.",
          score: 1,
        },
      ],
    });
    renderUI(<Sidebar />);
    await userEvent.type(screen.getByLabelText("Search contexts"), "black");
    const results = await screen.findByRole("region", { name: "Search results" });
    expect(anchor(await within(results).findByText("py"))).toHaveAttribute(
      "href",
      "/contexts/global/style/py",
    );
    expect(screen.queryByRole("listitem", { name: "Global" })).not.toBeInTheDocument();
  });

  it("shows no results", async () => {
    base({ "/api/v1/search?q=zzz": [] });
    renderUI(<Sidebar />);
    await userEvent.type(screen.getByLabelText("Search contexts"), "zzz");
    expect(await screen.findByText("No results")).toBeInTheDocument();
  });
});
