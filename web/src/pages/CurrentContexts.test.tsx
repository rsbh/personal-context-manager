import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { CurrentContexts } from "./CurrentContexts";

describe("CurrentContexts", () => {
  it("shows the merged contexts for the project in the URL", async () => {
    mockAPI({
      "/api/v1/projects": ["api", "web"],
      "/api/v1/sections?scope=global": ["style"],
      "/api/v1/resolve?project=api": [
        { id: "global/style/py", scope: "global", section: "style", name: "py", frontmatter: { title: "Python" } },
        { id: "projects/api/style/go", scope: "project", project: "api", section: "style", name: "go", frontmatter: {} },
      ],
    });
    renderUI(<CurrentContexts />, "/current?project=api");

    const table = await screen.findByRole("table");
    const rows = await within(table).findAllByRole("row");
    expect(rows).toHaveLength(3); // header + 2
    expect(within(rows[1]).getByText("global")).toBeInTheDocument();
    expect(within(rows[1]).getByRole("link", { name: "py" })).toHaveAttribute("href", "/contexts/global/style/py");
    expect(within(rows[2]).getByText("project")).toBeInTheDocument();
  });

  it("shows global only without a project", async () => {
    mockAPI({
      "/api/v1/projects": ["api"],
      "/api/v1/sections?scope=global": [],
      "/api/v1/resolve": [],
    });
    renderUI(<CurrentContexts />, "/current");
    expect(await screen.findByText("No contexts found.")).toBeInTheDocument();
  });
});
