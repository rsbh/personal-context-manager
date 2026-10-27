import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { json, mockAPI, renderUI, sentBody } from "../test/render";
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

  it("opens a new context form from a global section", async () => {
    base();
    renderUI(<Sidebar />);
    await userEvent.click(await screen.findByRole("button", { name: "New context in style" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/new?scope=global&section=style");
  });

  it("opens a new context form from a project section", async () => {
    base({
      "/api/v1/sections?scope=project&project=api": ["review"],
      "/api/v1/contexts?scope=project&project=api": [],
    });
    renderUI(<Sidebar />);
    const projects = screen.getByRole("listitem", { name: "Projects" });
    await userEvent.click(await within(projects).findByText("api"));
    await userEvent.click(await screen.findByRole("button", { name: "New context in review" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent(
      "/new?scope=project&project=api&section=review",
    );
  });

  it("creates a global section", async () => {
    const fetchMock = base({ "POST /api/v1/sections": json({}, 201) });
    renderUI(<Sidebar />);
    await userEvent.click(screen.getByRole("button", { name: "New global section" }));
    const dialog = await screen.findByRole("dialog", { name: "New section" });
    await userEvent.type(within(dialog).getByLabelText("Name"), "review");
    await userEvent.click(within(dialog).getByRole("button", { name: "Create" }));
    expect(await screen.findByText("Section created")).toBeInTheDocument();
    expect(await sentBody(fetchMock, "POST")).toEqual({ scope: "global", name: "review" });
    expect(screen.queryByRole("dialog", { name: "New section" })).not.toBeInTheDocument();
  });

  it("creates a project", async () => {
    const fetchMock = base({ "POST /api/v1/projects": json({}, 201) });
    renderUI(<Sidebar />);
    await userEvent.click(screen.getByRole("button", { name: "New project" }));
    const dialog = await screen.findByRole("dialog", { name: "New project" });
    await userEvent.type(within(dialog).getByLabelText("Name"), "web");
    await userEvent.click(within(dialog).getByRole("button", { name: "Create" }));
    expect(await screen.findByText("Project created")).toBeInTheDocument();
    expect(await sentBody(fetchMock, "POST")).toEqual({ name: "web" });
  });

  it("creates a section in a project", async () => {
    const fetchMock = base({ "POST /api/v1/sections": json({}, 201) });
    renderUI(<Sidebar />);
    await userEvent.click(await screen.findByRole("button", { name: "New section in api" }));
    const dialog = await screen.findByRole("dialog", { name: "New section" });
    await userEvent.type(within(dialog).getByLabelText("Name"), "review");
    await userEvent.click(within(dialog).getByRole("button", { name: "Create" }));
    expect(await sentBody(fetchMock, "POST")).toEqual({
      scope: "project",
      project: "api",
      name: "review",
    });
  });

  it("keeps the create dialog open on an error", async () => {
    base({
      "POST /api/v1/sections": json({ error: 'section "style": already exists' }, 409),
    });
    renderUI(<Sidebar />);
    await userEvent.click(screen.getByRole("button", { name: "New global section" }));
    const dialog = await screen.findByRole("dialog", { name: "New section" });
    await userEvent.type(within(dialog).getByLabelText("Name"), "style");
    await userEvent.click(within(dialog).getByRole("button", { name: "Create" }));
    expect(await screen.findByText(/already exists/)).toBeInTheDocument();
    expect(within(dialog).getByLabelText("Name")).toHaveValue("style");
  });

  it("deletes an empty section", async () => {
    base({ "DELETE /api/v1/sections/empty?scope=global": new Response(null, { status: 204 }) });
    renderUI(<Sidebar />);
    await userEvent.click(await screen.findByRole("button", { name: "Section empty actions" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "Delete" }));
    const dialog = await screen.findByRole("alertdialog", { name: "Delete section?" });
    await userEvent.click(within(dialog).getByRole("button", { name: "Delete" }));
    expect(await screen.findByText("Deleted")).toBeInTheDocument();
  });

  it("asks again and forces the delete of a section that is not empty", async () => {
    const fetchMock = base({
      "DELETE /api/v1/sections/style?scope=global": json({ error: "not empty" }, 409),
      "DELETE /api/v1/sections/style?scope=global&force=true": new Response(null, {
        status: 204,
      }),
    });
    renderUI(<Sidebar />, "/contexts/global/style/go");
    await userEvent.click(await screen.findByRole("button", { name: "Section style actions" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "Delete" }));
    await userEvent.click(
      within(await screen.findByRole("alertdialog")).getByRole("button", { name: "Delete" }),
    );
    const dialog = await screen.findByRole("alertdialog", { name: "Section is not empty" });
    await userEvent.click(within(dialog).getByRole("button", { name: "Delete all" }));

    expect(await screen.findByText("Deleted")).toBeInTheDocument();
    expect(
      fetchMock.mock.calls.some(
        ([r]) => r instanceof Request && r.method === "DELETE" && r.url.endsWith("force=true"),
      ),
    ).toBe(true);
    expect(screen.getByLabelText("Location")).toHaveTextContent(/^\/$/);
  });

  it("deletes a project with force", async () => {
    base({
      "DELETE /api/v1/projects/api": json({ error: "not empty" }, 409),
      "DELETE /api/v1/projects/api?force=true": new Response(null, { status: 204 }),
    });
    renderUI(<Sidebar />);
    await userEvent.click(await screen.findByRole("button", { name: "Project api actions" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "Delete" }));
    await userEvent.click(
      within(await screen.findByRole("alertdialog")).getByRole("button", { name: "Delete" }),
    );
    await userEvent.click(
      within(await screen.findByRole("alertdialog", { name: "Project is not empty" })).getByRole(
        "button",
        { name: "Delete all" },
      ),
    );
    expect(await screen.findByText("Deleted")).toBeInTheDocument();
  });

  it("stays on the page when a different section is deleted", async () => {
    base({ "DELETE /api/v1/sections/empty?scope=global": new Response(null, { status: 204 }) });
    renderUI(<Sidebar />, "/contexts/global/style/go");
    await userEvent.click(await screen.findByRole("button", { name: "Section empty actions" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "Delete" }));
    await userEvent.click(
      within(await screen.findByRole("alertdialog")).getByRole("button", { name: "Delete" }),
    );
    expect(await screen.findByText("Deleted")).toBeInTheDocument();
    expect(screen.getByLabelText("Location")).toHaveTextContent("/contexts/global/style/go");
  });
});
