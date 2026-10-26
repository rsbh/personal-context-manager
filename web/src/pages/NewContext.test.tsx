import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI, sentBody } from "../test/render";
import { NewContext } from "./NewContext";

const created = {
  id: "global/style/py",
  scope: "global",
  section: "style",
  name: "py",
  frontmatter: {},
  content: "",
};

function base(extra: Record<string, unknown> = {}) {
  return mockAPI({
    "/api/v1/projects": ["api"],
    "/api/v1/sections?scope=global": ["git", "style"],
    "/api/v1/sections?scope=project&project=api": ["review"],
    ...extra,
  });
}

describe("NewContext", () => {
  it("fills scope and section from the query", async () => {
    base();
    renderUI(<NewContext />, "/new?scope=global&section=style", "/new");
    expect(await screen.findByRole("combobox", { name: "Scope" })).toHaveTextContent("Global");
    expect(await screen.findByRole("combobox", { name: "Section" })).toHaveTextContent("style");
  });

  it("fills a project scope from the query", async () => {
    base();
    renderUI(<NewContext />, "/new?scope=project&project=api&section=review", "/new");
    expect(await screen.findByRole("combobox", { name: "Scope" })).toHaveTextContent("api");
    expect(await screen.findByRole("combobox", { name: "Section" })).toHaveTextContent("review");
  });

  it("creates with POST and opens the new context", async () => {
    const fetchMock = base({
      "POST /api/v1/contexts": new Response(JSON.stringify(created), { status: 201 }),
    });
    renderUI(<NewContext />, "/new?scope=global&section=style", "/new");
    await userEvent.type(await screen.findByLabelText("Name"), "py");
    await userEvent.type(screen.getByLabelText("Title"), "Python");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByLabelText("Location")).toHaveTextContent("/contexts/global/style/py");
    expect(await sentBody(fetchMock, "POST")).toEqual({
      scope: "global",
      section: "style",
      name: "py",
      frontmatter: { title: "Python", description: "", tags: [] },
      content: "",
    });
    expect(await screen.findByText("Saved")).toBeInTheDocument();
  });

  it("sends the project for a project scope", async () => {
    const fetchMock = base({
      "POST /api/v1/contexts": new Response(
        JSON.stringify({ ...created, id: "projects/api/review/pr" }),
        { status: 201 },
      ),
    });
    renderUI(<NewContext />, "/new?scope=project&project=api&section=review", "/new");
    await userEvent.type(await screen.findByLabelText("Name"), "pr");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await sentBody(fetchMock, "POST")).toMatchObject({
      scope: "project",
      project: "api",
      section: "review",
      name: "pr",
    });
  });

  it("disables Save without a section or name", async () => {
    base();
    renderUI(<NewContext />, "/new", "/new");
    expect(await screen.findByRole("button", { name: "Save" })).toBeDisabled();
  });

  it("shows a conflict in a toast and keeps the values", async () => {
    base({
      "POST /api/v1/contexts": new Response(
        JSON.stringify({ error: 'context "global/style/go": already exists' }),
        { status: 409 },
      ),
    });
    renderUI(<NewContext />, "/new?scope=global&section=style", "/new");
    await userEvent.type(await screen.findByLabelText("Name"), "go");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText(/already exists/)).toBeInTheDocument();
    expect(screen.getByLabelText("Location")).toHaveTextContent("/new");
    expect(screen.getByLabelText("Name")).toHaveValue("go");
  });
});
