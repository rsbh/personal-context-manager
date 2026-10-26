import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { json, mockAPI, renderUI, sentBody } from "../test/render";
import { EditContext } from "./EditContext";

const ctx = {
  id: "global/style/go",
  scope: "global",
  section: "style",
  name: "go",
  frontmatter: { title: "Go", description: "Go style", tags: ["go"] },
  content: "Use gofmt.",
  updated_at: new Date().toISOString(),
};

const path = "/api/v1/contexts/global%2Fstyle%2Fgo";

function renderEdit() {
  return renderUI(<EditContext />, "/edit/global/style/go", "/edit/*");
}

describe("EditContext", () => {
  it("fills the form from the context", async () => {
    mockAPI({ [path]: ctx });
    renderEdit();
    expect(await screen.findByLabelText("Title")).toHaveValue("Go");
    expect(screen.getByLabelText("Description")).toHaveValue("Go style");
    expect(screen.getByText("go")).toBeInTheDocument();
    expect(await screen.findByRole("textbox", { name: "Markdown" })).toHaveTextContent(
      "Use gofmt.",
    );
    expect(screen.getByRole("region", { name: "Preview" })).toHaveTextContent("Use gofmt.");
    expect(screen.queryByText("Unsaved changes")).not.toBeInTheDocument();
  });

  it("shows unsaved changes and discards them", async () => {
    mockAPI({ [path]: ctx });
    renderEdit();
    const title = await screen.findByLabelText("Title");
    await userEvent.clear(title);
    await userEvent.type(title, "Go rules");
    expect(screen.getByText("Unsaved changes")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Discard" }));
    expect(title).toHaveValue("Go");
    expect(screen.queryByText("Unsaved changes")).not.toBeInTheDocument();
  });

  it("saves with PUT and opens the context view", async () => {
    const fetchMock = mockAPI({ [path]: ctx, [`PUT ${path}`]: ctx });
    renderEdit();
    const title = await screen.findByLabelText("Title");
    await userEvent.clear(title);
    await userEvent.type(title, "Go rules");
    await userEvent.type(screen.getByLabelText("Tags"), "fmt{Enter}");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByLabelText("Location")).toHaveTextContent("/contexts/global/style/go");
    expect(await sentBody(fetchMock, "PUT")).toEqual({
      frontmatter: { title: "Go rules", description: "Go style", tags: ["go", "fmt"] },
      content: "Use gofmt.",
    });
    expect(await screen.findByText("Saved")).toBeInTheDocument();
  });

  it("shows the server error in a toast and stays", async () => {
    mockAPI({
      [path]: ctx,
      [`PUT ${path}`]: json({ error: "invalid tag" }, 400),
    });
    renderEdit();
    await userEvent.type(await screen.findByLabelText("Title"), "!");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText(/invalid tag/)).toBeInTheDocument();
    expect(screen.getByLabelText("Location")).toHaveTextContent("/edit/global/style/go");
    expect(screen.getByLabelText("Title")).toHaveValue("Go!");
  });

  it("shows not found for a missing context", async () => {
    mockAPI({});
    renderEdit();
    expect(await screen.findByText("Context not found")).toBeInTheDocument();
  });
});
