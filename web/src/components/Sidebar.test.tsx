import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { Sidebar } from "./Sidebar";

describe("Sidebar", () => {
  it("lists global sections and projects in groups", async () => {
    mockAPI({
      "/api/v1/sections?scope=global": ["git", "style"],
      "/api/v1/projects": ["api", "web"],
    });
    renderUI(<Sidebar />);

    const nav = screen.getByRole("navigation", { name: "Contexts" });
    const global = within(nav).getByRole("group", { name: "Global" });
    const projects = within(nav).getByRole("group", { name: "Projects" });
    expect(await within(global).findByText("git")).toBeInTheDocument();
    expect(within(global).getByText("style")).toBeInTheDocument();
    expect(await within(projects).findByText("api")).toBeInTheDocument();
    expect(within(projects).getByText("web")).toBeInTheDocument();
  });

  it("shows empty states", async () => {
    mockAPI({ "/api/v1/sections?scope=global": [], "/api/v1/projects": [] });
    renderUI(<Sidebar />);
    expect(await screen.findByText("No sections")).toBeInTheDocument();
    expect(await screen.findByText("No projects")).toBeInTheDocument();
  });
});
