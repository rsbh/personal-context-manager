import { describe, expect, it } from "vitest";
import { emptyDraft, fromContext, sameDraft, toFrontmatter } from "./draft";

describe("draft", () => {
  it("makes a draft from a context with empty defaults", () => {
    expect(fromContext({ id: "global/a/b" })).toEqual(emptyDraft);
    expect(
      fromContext({ frontmatter: { title: "T", description: "D", tags: ["x"] }, content: "c" }),
    ).toEqual({ title: "T", description: "D", tags: ["x"], content: "c" });
  });

  it("makes frontmatter from a draft", () => {
    expect(toFrontmatter({ ...emptyDraft, title: "T", tags: ["x"] })).toEqual({
      title: "T",
      description: "",
      tags: ["x"],
    });
  });

  it("compares drafts by value", () => {
    expect(sameDraft(emptyDraft, { ...emptyDraft, tags: [] })).toBe(true);
    expect(sameDraft(emptyDraft, { ...emptyDraft, tags: ["x"] })).toBe(false);
    expect(sameDraft(emptyDraft, { ...emptyDraft, content: "x" })).toBe(false);
  });
});
