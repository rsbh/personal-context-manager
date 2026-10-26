import { describe, expect, it } from "vitest";
import { age, rawFile } from "./format";

describe("age", () => {
  const now = new Date("2026-10-26T12:00:00Z");
  it.each([
    ["2026-10-26T11:59:50Z", "10s"],
    ["2026-10-26T11:55:00Z", "5m"],
    ["2026-10-26T09:00:00Z", "3h"],
    ["2026-10-24T10:00:00Z", "2d"],
    ["2026-10-26T12:01:00Z", "0s"],
  ])("%s -> %s", (t, want) => {
    expect(age(t, now)).toBe(want);
  });
});

describe("rawFile", () => {
  it("renders frontmatter then content", () => {
    expect(rawFile({ title: "Go", description: "Go style", tags: ["go", "fmt"] }, "Use gofmt.\n")).toBe(
      "---\ntitle: Go\ndescription: Go style\ntags:\n  - go\n  - fmt\n---\nUse gofmt.\n",
    );
  });

  it("has no header without frontmatter", () => {
    expect(rawFile({}, "plain")).toBe("plain");
  });
});
