import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { Sidebar } from "./Sidebar";

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
    const global = within(nav).getByRole("group", { name: "Global" });
    expect(await within(global).findByText("commits")).toBeInTheDocument();
    expect(within(global).getByText("go")).toBeInTheDocument();
    expect(within(global).getByText("empty")).toBeInTheDocument();
    expect(within(global).getByRole("link", { name: "go" })).toHaveAttribute("href", "/contexts/global/style/go");
    expect(await within(nav).findByText("api")).toBeInTheDocument();
  });

  it("loads project sections and contexts when a project opens", async () => {
    base({
      "/api/v1/sections?scope=project&project=api": ["review"],
      "/api/v1/contexts?scope=project&project=api": [
        { id: "projects/api/review/pr", scope: "project", project: "api", section: "review", name: "pr" },
      ],
    });
    renderUI(<Sidebar />);
    const projects = screen.getByRole("group", { name: "Projects" });
    await userEvent.click(await within(projects).findByText("api"));
    const link = await within(projects).findByRole("link", { name: "pr" });
    expect(link).toHaveAttribute("href", "/contexts/projects/api/review/pr");
  });

  it("shows empty states", async () => {
    mockAPI({ "/api/v1/sections?scope=global": [], "/api/v1/contexts?scope=global": [], "/api/v1/projects": [] });
    renderUI(<Sidebar />);
    expect(await screen.findByText("No sections")).toBeInTheDocument();
    expect(await screen.findByText("No projects")).toBeInTheDocument();
  });

  it("shows search results instead of the tree while searching", async () => {
    base({
      "/api/v1/search?q=black": [
        { id: "global/style/py", scope: "global", section: "style", name: "py", snippet: "Use <mark>black</mark>.", score: 1 },
      ],
    });
    renderUI(<Sidebar />);
    await userEvent.type(screen.getByLabelText("Search contexts"), "black");
    const results = await screen.findByRole("region", { name: "Search results" });
    expect(await within(results).findByRole("link", { name: /py/ })).toHaveAttribute("href", "/contexts/global/style/py");
    expect(screen.queryByRole("group", { name: "Global" })).not.toBeInTheDocument();
  });

  it("shows no results", async () => {
    base({ "/api/v1/search?q=zzz": [] });
    renderUI(<Sidebar />);
    await userEvent.type(screen.getByLabelText("Search contexts"), "zzz");
    expect(await screen.findByText("No results")).toBeInTheDocument();
  });
});
