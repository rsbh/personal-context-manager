import { screen } from "@testing-library/react";
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
