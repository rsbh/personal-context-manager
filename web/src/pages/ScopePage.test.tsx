import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { ScopePage } from "./ScopePage";

describe("ScopePage", () => {
  it("lists the sections of a project with their contexts", async () => {
    mockAPI({
      "/api/v1/projects": ["api"],
      "/api/v1/sections?scope=project&project=api": ["review", "empty"],
      "/api/v1/contexts?scope=project&project=api": [
        {
          id: "projects/api/review/pr",
          scope: "project",
          project: "api",
          section: "review",
          name: "pr",
          frontmatter: { description: "Small PRs" },
        },
      ],
    });
    renderUI(<ScopePage project="api" />, "/projects/api");

    expect(await screen.findByRole("heading", { name: "api" })).toBeInTheDocument();
    expect(screen.getByText("2 sections · 1 context")).toBeInTheDocument();
    const review = screen.getByRole("region", { name: "review" });
    expect(within(review).getByRole("link", { name: "pr" })).toHaveAttribute(
      "href",
      "/contexts/projects/api/review/pr",
    );
    expect(within(review).getByText("Small PRs")).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "empty" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete project" })).toBeInTheDocument();
  });

  it("has no Delete project on the global page", async () => {
    mockAPI({
      "/api/v1/projects": [],
      "/api/v1/sections?scope=global": [],
      "/api/v1/contexts?scope=global": [],
    });
    renderUI(<ScopePage />, "/global");

    expect(await screen.findByRole("heading", { name: "Global" })).toBeInTheDocument();
    expect(screen.getByText("No sections yet.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete project" })).not.toBeInTheDocument();
  });

  it("shows not found for an unknown project", async () => {
    mockAPI({
      "/api/v1/projects": ["api"],
      "/api/v1/sections?scope=project&project=nope": [],
      "/api/v1/contexts?scope=project&project=nope": [],
    });
    renderUI(<ScopePage project="nope" />, "/projects/nope");

    expect(await screen.findByText("Project not found")).toBeInTheDocument();
  });
});
