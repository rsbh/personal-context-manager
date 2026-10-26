import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { ContextView } from "./ContextView";

const ctx = {
  id: "global/style/go",
  scope: "global",
  section: "style",
  name: "go",
  frontmatter: { title: "Go", description: "Go style" },
  content: "Use **gofmt**.",
  updated_at: new Date().toISOString(),
};

describe("ContextView", () => {
  it("shows title, description, rendered markdown and raw file", async () => {
    mockAPI({ "/api/v1/contexts/global%2Fstyle%2Fgo": ctx });
    renderUI(<ContextView id="global/style/go" />);

    expect(await screen.findByRole("heading", { name: "Go" })).toBeInTheDocument();
    expect(screen.getByText("Go style")).toBeInTheDocument();
    expect(screen.getByText("gofmt").tagName).toBe("STRONG");
    const raw = screen.getByRole("region", { name: "Raw file" });
    expect(raw).toHaveTextContent("global/style/go.md");
    expect(raw).toHaveTextContent("title: Go");
    expect(screen.getByText("global")).toBeInTheDocument();
  });

  it("uses the name as heading without title", async () => {
    mockAPI({ "/api/v1/contexts/global%2Fstyle%2Fgo": { ...ctx, frontmatter: {} } });
    renderUI(<ContextView id="global/style/go" />);
    expect(await screen.findByRole("heading", { name: "go" })).toBeInTheDocument();
  });

  it("shows not found for 404", async () => {
    mockAPI({});
    renderUI(<ContextView id="global/style/none" />);
    expect(await screen.findByText("Context not found")).toBeInTheDocument();
  });
});
