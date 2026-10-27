import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { json, mockAPI, renderUI } from "../test/render";
import { SyncPage } from "./SyncPage";

const status = {
  branch: "main",
  remote: "git@github.com:me/contexts.git",
  changes: [" M global/style/go.md", "?? projects/api/"],
  ahead: 1,
  behind: 0,
};

const commits = [
  { hash: "fcc281a", date: "2026-10-26 19:40", message: "pcm sync: 5 files changed" },
];

function base(extra: Record<string, unknown> = {}) {
  return mockAPI({
    "/api/v1/git/status": status,
    "/api/v1/git/log?n=10": commits,
    ...extra,
  });
}

function card(label: string) {
  return screen.getByRole("group", { name: label });
}

describe("SyncPage", () => {
  it("shows remote, branch, status cards, changed files and commits", async () => {
    base();
    renderUI(<SyncPage />);
    expect(await screen.findByText("git@github.com:me/contexts.git · main")).toBeInTheDocument();
    expect(card("Changes")).toHaveTextContent("2");
    expect(card("Ahead")).toHaveTextContent("1");
    expect(card("Behind")).toHaveTextContent("0");
    expect(screen.queryByRole("group", { name: "Last sync" })).not.toBeInTheDocument();

    const files = screen.getByRole("region", { name: "Changed files" });
    expect(within(files).getByText("M")).toBeInTheDocument();
    expect(within(files).getByText("global/style/go.md")).toBeInTheDocument();
    expect(within(files).getByText("??")).toBeInTheDocument();
    expect(within(files).getByText("projects/api/")).toBeInTheDocument();

    const log = await screen.findByRole("region", { name: "Recent commits" });
    expect(within(log).getByText("fcc281a")).toBeInTheDocument();
    expect(within(log).getByText("pcm sync: 5 files changed")).toBeInTheDocument();
  });

  it("shows no changes", async () => {
    base({ "/api/v1/git/status": { ...status, changes: [], ahead: 0 } });
    renderUI(<SyncPage />);
    expect(await screen.findByText("No changes")).toBeInTheDocument();
  });

  it("syncs and shows the result", async () => {
    base({
      "POST /api/v1/git/sync": json({ committed: 2, pulled: 1, pushed: true, remote: "x" }),
    });
    renderUI(<SyncPage />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    expect(
      await screen.findByText("Synced: 2 files committed, 1 commit pulled, pushed"),
    ).toBeInTheDocument();
  });

  it("says nothing to push", async () => {
    base({
      "POST /api/v1/git/sync": json({ committed: 0, pulled: 0, pushed: false, remote: "x" }),
    });
    renderUI(<SyncPage />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    expect(
      await screen.findByText("Synced: 0 files committed, 0 commits pulled, nothing to push"),
    ).toBeInTheDocument();
  });

  it("shows a conflict until the next successful sync", async () => {
    const conflict = json(
      { error: "sync conflict, rebase aborted", files: ["global/style/go.md"] },
      409,
    );
    const fetchMock = base({ "POST /api/v1/git/sync": conflict });
    renderUI(<SyncPage />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    const callout = await screen.findByRole("region", { name: "Conflict" });
    expect(callout).toHaveTextContent("Sync conflict, rebase aborted");
    expect(callout).toHaveTextContent("global/style/go.md");
    expect(callout).toHaveTextContent("~/.pcm/repo");

    fetchMock.mockImplementation(async (input) => {
      const req = input as Request;
      if (req.method === "POST") {
        return json({ committed: 1, pulled: 0, pushed: true, remote: "x" });
      }
      const url = new URL(req.url).pathname;
      return json(url.endsWith("/log") ? commits : status);
    });
    await userEvent.click(screen.getByRole("button", { name: "Sync now" }));
    expect(await screen.findByText(/^Synced:/)).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Conflict" })).not.toBeInTheDocument();
  });

  it("shows a server error in a toast", async () => {
    base({ "POST /api/v1/git/sync": json({ error: "git not found" }, 501) });
    renderUI(<SyncPage />);
    await userEvent.click(await screen.findByRole("button", { name: "Sync now" }));
    expect(await screen.findByText("git not found")).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Conflict" })).not.toBeInTheDocument();
  });

  it("disables sync without a remote and shows how to set it", async () => {
    base({ "/api/v1/git/status": { ...status, remote: "" } });
    renderUI(<SyncPage />);
    expect(await screen.findByText("Remote not set")).toBeInTheDocument();
    expect(screen.getByText("pcm config set remote <git-url>")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sync now" })).toBeDisabled();
  });
});
