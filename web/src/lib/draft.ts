import type { StoreContext, StoreFrontmatter } from "../client/types.gen";

// Draft is the editable part of a context.
export interface Draft {
  title: string;
  description: string;
  tags: string[];
  content: string;
}

export const emptyDraft: Draft = { title: "", description: "", tags: [], content: "" };

// fromContext makes a draft from a context.
export function fromContext(c: StoreContext): Draft {
  return {
    title: c.frontmatter?.title ?? "",
    description: c.frontmatter?.description ?? "",
    tags: c.frontmatter?.tags ?? [],
    content: c.content ?? "",
  };
}

// toFrontmatter is the frontmatter to send for a draft.
export function toFrontmatter(d: Draft): StoreFrontmatter {
  return { title: d.title, description: d.description, tags: d.tags };
}

// sameDraft is true when two drafts have the same values.
export function sameDraft(a: Draft, b: Draft): boolean {
  return (
    a.title === b.title &&
    a.description === b.description &&
    a.content === b.content &&
    a.tags.length === b.tags.length &&
    a.tags.every((t, i) => t === b.tags[i])
  );
}
