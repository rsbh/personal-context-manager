import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { json, mockAPI, renderUI, sentBody } from "../test/render";
import { scopeOf, Sidebar } from "./Sidebar";

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

const apiScope = {
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
};

// openMenu opens the "⋯" menu of a sidebar row.
async function openMenu(label: string) {
  await userEvent.click(await screen.findByRole("button", { name: label }));
}

describe("scopeOf", () => {
  it("reads the scope from the URL", () => {
    expect(scopeOf("/contexts/projects/api/review/pr")).toBe("api");
    expect(scopeOf("/edit/projects/api/review/pr")).toBe("api");
    expect(scopeOf("/projects/api")).toBe("api");
    expect(scopeOf("/global")).toBe("");
    expect(scopeOf("/contexts/global/style/go")).toBe("");
    expect(scopeOf("/sync")).toBeUndefined();
  });
});

describe("Sidebar", () => {
  it("lists the global sections with their contexts", async () => {
    base();
    renderUI(<Sidebar />);
    const nav = screen.getByRole("navigation", { name: "Contexts" });
    const style = await within(nav).findByRole("listitem", { name: "style" });
    expect(within(style).getByRole("link", { name: "go" })).toHaveAttribute(
      "href",
      "/contexts/global/style/go",
    );
    expect(within(nav).getByRole("listitem", { name: "empty" })).toBeInTheDocument();
    expect(within(nav).getByRole("link", { name: /Global overview/ })).toHaveAttribute(
      "href",
      "/global",
    );
    expect(within(nav).getByRole("combobox", { name: "Scope" })).toHaveTextContent("Global");
  });

  it("shows the project of the open context", async () => {
    base(apiScope);
    renderUI(<Sidebar />, "/contexts/projects/api/review/pr");
    const nav = screen.getByRole("navigation", { name: "Contexts" });
    expect(await within(nav).findByRole("link", { name: "pr" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(within(nav).queryByRole("link", { name: "go" })).not.toBeInTheDocument();
    expect(within(nav).getByRole("link", { name: /Project overview/ })).toHaveAttribute(
      "href",
      "/projects/api",
    );
  });

  it("shows an empty state", async () => {
    mockAPI({
      "/api/v1/sections?scope=global": [],
      "/api/v1/contexts?scope=global": [],
      "/api/v1/projects": [],
    });
    renderUI(<Sidebar />);
    expect(await screen.findByText("No sections")).toBeInTheDocument();
  });

  it("opens a new context form from the section menu", async () => {
    base(apiScope);
    renderUI(<Sidebar />, "/projects/api");
    await openMenu("Actions for section review");
    await userEvent.click(await screen.findByRole("menuitem", { name: "New context" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent(
      "/new?scope=project&project=api&section=review",
    );
  });

  it("opens the edit form from the context menu", async () => {
    base();
    renderUI(<Sidebar />);
    await openMenu("Actions for go");
    await userEvent.click(await screen.findByRole("menuitem", { name: "Edit" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/edit/global/style/go");
  });

  it("asks again and forces the delete of a section that is not empty", async () => {
    const fetchMock = base({
      "DELETE /api/v1/sections/style?scope=global": json({ error: "not empty" }, 409),
      "DELETE /api/v1/sections/style?scope=global&force=true": new Response(null, {
        status: 204,
      }),
    });
    renderUI(<Sidebar />, "/contexts/global/style/go");
    await openMenu("Actions for section style");
    await userEvent.click(await screen.findByRole("menuitem", { name: "Delete section" }));
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
    // The redirect runs right after the toast: wait for it.
    await waitFor(() => expect(screen.getByLabelText("Location")).toHaveTextContent(/^\/$/));
  });

  it("creates a project from the scope select and opens it", async () => {
    const fetchMock = base({ "POST /api/v1/projects": json({}, 201) });
    renderUI(<Sidebar />);
    await userEvent.click(screen.getByRole("combobox", { name: "Scope" }));
    await userEvent.click(await screen.findByRole("option", { name: "New project…" }));
    const dialog = await screen.findByRole("dialog", { name: "New project" });
    await userEvent.type(within(dialog).getByLabelText("Name"), "web");
    await userEvent.click(within(dialog).getByRole("button", { name: "Create" }));
    expect(await screen.findByText("Project created")).toBeInTheDocument();
    expect(await sentBody(fetchMock, "POST")).toEqual({ name: "web" });
    await waitFor(() =>
      expect(screen.getByLabelText("Location")).toHaveTextContent("/projects/web"),
    );
  });
});
