import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { Header } from "./Header";

describe("Header", () => {
  it("shows server ok when /health is ok", async () => {
    mockAPI({ "/health": { status: "ok" } });
    renderUI(<Header onSearch={vi.fn()} />);
    expect(await screen.findByText("server ok")).toBeInTheDocument();
    expect(screen.getByText("pcm")).toBeInTheDocument();
  });

  it("shows server down when /health fails", async () => {
    mockAPI({ "/health": new Response("", { status: 500 }) });
    renderUI(<Header onSearch={vi.fn()} />);
    expect(await screen.findByText("server down")).toBeInTheDocument();
  });
});

describe("Header nav", () => {
  it("links to current contexts", async () => {
    mockAPI({ "/health": { status: "ok" } });
    renderUI(<Header onSearch={vi.fn()} />);
    expect(screen.getByRole("link", { name: "Current contexts" })).toHaveAttribute(
      "href",
      "/current",
    );
  });
});

describe("Header browse link", () => {
  it("links to browse", async () => {
    mockAPI({ "/health": { status: "ok" } });
    renderUI(<Header onSearch={vi.fn()} />);
    expect(screen.getByRole("link", { name: "Browse" })).toHaveAttribute("href", "/");
  });
});

describe("Header sync link", () => {
  it("links to sync with a badge of changes + ahead", async () => {
    mockAPI({
      "/health": { status: "ok" },
      "/api/v1/git/status": {
        branch: "main",
        remote: "",
        changes: ["?? a/", " M b.md"],
        ahead: 1,
        behind: 0,
      },
    });
    renderUI(<Header onSearch={vi.fn()} />);
    const link = screen.getByRole("link", { name: /^Sync/ });
    expect(link).toHaveAttribute("href", "/sync");
    expect(await within(link).findByText("3")).toBeInTheDocument();
  });

  it("has no badge when there is nothing to sync", async () => {
    mockAPI({
      "/health": { status: "ok" },
      "/api/v1/git/status": { branch: "main", remote: "", changes: [], ahead: 0, behind: 0 },
    });
    renderUI(<Header onSearch={vi.fn()} />);
    await screen.findByText("server ok");
    expect(screen.getByRole("link", { name: /^Sync/ })).toHaveTextContent(/^Sync$/);
  });
});

describe("Header search", () => {
  it("opens the search palette from the Search button", async () => {
    mockAPI({ "/health": { status: "ok" } });
    const onSearch = vi.fn();
    renderUI(<Header onSearch={onSearch} />);
    await userEvent.click(screen.getByRole("button", { name: "Search" }));
    expect(onSearch).toHaveBeenCalledTimes(1);
  });
});
