import type { StoreFrontmatter } from "../client/types.gen";

// age formats the time since t like kubectl AGE: 10s, 5m, 3h, 2d.
export function age(t: string | undefined, now = new Date()): string {
  if (!t) return "";
  const s = Math.max(0, Math.floor((now.getTime() - new Date(t).getTime()) / 1000));
  if (s < 60) return `${s}s`;
  if (s < 3600) return `${Math.floor(s / 60)}m`;
  if (s < 86400) return `${Math.floor(s / 3600)}h`;
  return `${Math.floor(s / 86400)}d`;
}

// rawFile renders a context as its md file: frontmatter, then content (same as the server).
export function rawFile(fm: StoreFrontmatter | undefined, content: string | undefined): string {
  const lines: string[] = [];
  if (fm?.title) lines.push(`title: ${fm.title}`);
  if (fm?.description) lines.push(`description: ${fm.description}`);
  if (fm?.tags?.length) lines.push("tags:", ...fm.tags.map((t) => `  - ${t}`));
  const body = content ?? "";
  return lines.length ? `---\n${lines.join("\n")}\n---\n${body}` : body;
}

// contextPath is the UI route of a context id.
export function contextPath(id: string): string {
  return `/contexts/${id}`;
}
