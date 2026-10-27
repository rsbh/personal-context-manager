import type { GitsyncStatus, GitsyncSyncResult } from "../client/types.gen";

// Change is one line of git status --porcelain: "XY path".
export interface Change {
  code: string;
  path: string;
}

export function parseChange(line: string): Change {
  return { code: line.slice(0, 2).trim(), path: line.slice(3) };
}

// pendingCount is what a sync would send: changed files + commits ahead.
export function pendingCount(s: GitsyncStatus | undefined): number {
  return (s?.changes?.length ?? 0) + (s?.ahead ?? 0);
}

function plural(n: number, word: string): string {
  return `${n} ${word}${n === 1 ? "" : "s"}`;
}

// syncSummary is the toast text for a sync result.
export function syncSummary(r: GitsyncSyncResult): string {
  const committed = r.committed ?? 0;
  const pulled = r.pulled ?? 0;
  return `Synced: ${plural(committed, "file")} committed, ${plural(pulled, "commit")} pulled, ${
    r.pushed ? "pushed" : "nothing to push"
  }`;
}
