import { describe, expect, it } from "vitest";
import { parseChange, pendingCount, syncSummary } from "./sync";

describe("sync", () => {
  it("parses porcelain lines", () => {
    expect(parseChange(" M global/style/go.md")).toEqual({ code: "M", path: "global/style/go.md" });
    expect(parseChange("?? projects/api/")).toEqual({ code: "??", path: "projects/api/" });
    expect(parseChange("R  a.md -> b.md")).toEqual({ code: "R", path: "a.md -> b.md" });
  });

  it("counts changes and commits ahead", () => {
    expect(pendingCount(undefined)).toBe(0);
    expect(pendingCount({ changes: ["?? a/"], ahead: 2 })).toBe(3);
  });

  it("summarizes a sync with plurals", () => {
    expect(syncSummary({ committed: 1, pulled: 2, pushed: true })).toBe(
      "Synced: 1 file committed, 2 commits pulled, pushed",
    );
    expect(syncSummary({})).toBe("Synced: 0 files committed, 0 commits pulled, nothing to push");
  });
});
