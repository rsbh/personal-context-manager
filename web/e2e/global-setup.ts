import { execFileSync } from "node:child_process";
import { existsSync, mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = fileURLToPath(new URL("../..", import.meta.url));

// Builds pcm and pcm-server once. They embed web/dist/app, so build the UI first.
export default function globalSetup() {
  if (!existsSync(join(repoRoot, "web/dist/app/index.html"))) {
    throw new Error("UI not built: run `npm run build` in web/ (or make e2e-ui)");
  }
  const binDir = mkdtempSync(join(tmpdir(), "pcm-e2e-ui-bin-"));
  execFileSync("go", ["build", "-o", binDir + "/", "./cmd/..."], { cwd: repoRoot, stdio: "inherit" });
  process.env.PCM_BIN_DIR = binDir;
}
