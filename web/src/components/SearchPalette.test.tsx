import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { SearchPalette } from "./SearchPalette";

const hit = {
  id: "global/style/py",
  scope: "global",
  section: "style",
  name: "py",
  snippet: "Use <mark>black</mark>.",
  score: 1,
};

describe("SearchPalette", () => {
  it("renders nothing when closed", () => {
    mockAPI({});
    renderUI(<SearchPalette open={false} onOpenChange={vi.fn<() => void>()} />);
    expect(screen.queryByRole("dialog", { name: "Search contexts" })).not.toBeInTheDocument();
  });

  it("shows name, section and snippet without tags", async () => {
    mockAPI({ "/api/v1/search?q=black": [hit] });
    renderUI(<SearchPalette open onOpenChange={vi.fn<() => void>()} />);
    await userEvent.type(screen.getByRole("combobox", { name: "Search contexts" }), "black");
    const row = await screen.findByRole("option", { name: /py/ });
    expect(row).toHaveTextContent("style");
    expect(row).toHaveTextContent("Use black.");
    expect(row).not.toHaveTextContent("<mark>");
  });

  it("opens the context and closes on select", async () => {
    mockAPI({ "/api/v1/search?q=black": [hit] });
    const onOpenChange = vi.fn<() => void>();
    renderUI(<SearchPalette open onOpenChange={onOpenChange} />);
    await userEvent.type(screen.getByRole("combobox", { name: "Search contexts" }), "black");
    await userEvent.click(await screen.findByRole("option", { name: /py/ }));
    expect(screen.getByLabelText("Location")).toHaveTextContent("/contexts/global/style/py");
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows no results", async () => {
    mockAPI({ "/api/v1/search?q=zzz": [] });
    renderUI(<SearchPalette open onOpenChange={vi.fn<() => void>()} />);
    await userEvent.type(screen.getByRole("combobox", { name: "Search contexts" }), "zzz");
    expect(await screen.findByText("No results")).toBeInTheDocument();
  });

  it("shows an error", async () => {
    mockAPI({ "/api/v1/search?q=bad": new Response("{}", { status: 500 }) });
    renderUI(<SearchPalette open onOpenChange={vi.fn<() => void>()} />);
    await userEvent.type(screen.getByRole("combobox", { name: "Search contexts" }), "bad");
    expect(await screen.findByText("Search failed")).toBeInTheDocument();
  });

  it("does not search an empty query", async () => {
    const fetchMock = mockAPI({});
    renderUI(<SearchPalette open onOpenChange={vi.fn<() => void>()} />);
    await new Promise((r) => setTimeout(r, 400));
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
