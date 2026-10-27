import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mockAPI, renderUI } from "../test/render";
import { Header } from "./Header";

describe("Header", () => {
  it("shows server ok when /health is ok", async () => {
    mockAPI({ "/health": { status: "ok" } });
    renderUI(<Header />);
    expect(await screen.findByText("server ok")).toBeInTheDocument();
    expect(screen.getByText("pcm")).toBeInTheDocument();
  });

  it("shows server down when /health fails", async () => {
    mockAPI({ "/health": new Response("", { status: 500 }) });
    renderUI(<Header />);
    expect(await screen.findByText("server down")).toBeInTheDocument();
  });
});

describe("Header nav", () => {
  it("links to current contexts", async () => {
    mockAPI({ "/health": { status: "ok" } });
    renderUI(<Header />);
    expect(screen.getByRole("link", { name: "Current contexts" })).toHaveAttribute(
      "href",
      "/current",
    );
  });
});

describe("Header browse link", () => {
  it("links to browse", async () => {
    mockAPI({ "/health": { status: "ok" } });
    renderUI(<Header />);
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
    renderUI(<Header />);
    const link = screen.getByRole("link", { name: /^Sync/ });
    expect(link).toHaveAttribute("href", "/sync");
    expect(await within(link).findByText("3")).toBeInTheDocument();
  });

  it("has no badge when there is nothing to sync", async () => {
    mockAPI({
      "/health": { status: "ok" },
      "/api/v1/git/status": { branch: "main", remote: "", changes: [], ahead: 0, behind: 0 },
    });
    renderUI(<Header />);
    await screen.findByText("server ok");
    expect(screen.getByRole("link", { name: /^Sync/ })).toHaveTextContent(/^Sync$/);
  });
});
