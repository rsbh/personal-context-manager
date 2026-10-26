import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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

  it("has an Edit button that opens the editor", async () => {
    mockAPI({ "/api/v1/contexts/global%2Fstyle%2Fgo": ctx });
    renderUI(<ContextView id="global/style/go" />);
    await userEvent.click(await screen.findByRole("button", { name: "Edit" }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/edit/global/style/go");
  });

  it("deletes after confirm and goes home", async () => {
    const fetchMock = mockAPI({
      "/api/v1/contexts/global%2Fstyle%2Fgo": ctx,
      "DELETE /api/v1/contexts/global%2Fstyle%2Fgo": new Response(null, { status: 204 }),
    });
    renderUI(<ContextView id="global/style/go" />, "/contexts/global/style/go", "/contexts/*");
    await userEvent.click(await screen.findByRole("button", { name: "Delete" }));
    const dialog = await screen.findByRole("alertdialog", { name: "Delete context?" });
    expect(dialog).toHaveTextContent("global/style/go");
    await userEvent.click(within(dialog).getByRole("button", { name: "Delete" }));

    expect(await screen.findByText("Deleted")).toBeInTheDocument();
    expect(screen.getByLabelText("Location")).toHaveTextContent(/^\/$/);
    expect(fetchMock.mock.calls.some(([r]) => r instanceof Request && r.method === "DELETE")).toBe(
      true,
    );
  });

  it("keeps the context when delete is cancelled", async () => {
    const fetchMock = mockAPI({ "/api/v1/contexts/global%2Fstyle%2Fgo": ctx });
    renderUI(<ContextView id="global/style/go" />);
    await userEvent.click(await screen.findByRole("button", { name: "Delete" }));
    await userEvent.click(await screen.findByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([r]) => r instanceof Request && r.method === "DELETE")).toBe(
      false,
    );
  });
});
