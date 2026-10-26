import { spawnSync } from "node:child_process";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test as base } from "playwright-bdd";

export interface RunResult {
  code: number;
  output: string;
}

// PcmWorld is one scenario's pcm: its own HOME, port and server.
export interface PcmWorld {
  home: string;
  port: number;
  url(path: string): string;
  run(...args: string[]): RunResult;
  post(path: string, body: unknown, ...okStatus: number[]): Promise<void>;
}

function freePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = createServer();
    srv.listen(0, "127.0.0.1", () => {
      const addr = srv.address();
      srv.close(() => (typeof addr === "object" && addr ? resolve(addr.port) : reject(new Error("no port"))));
    });
  });
}

export const test = base.extend<{ pcm: PcmWorld }>({
  pcm: async ({}, use) => {
    const binDir = process.env.PCM_BIN_DIR;
    if (!binDir) throw new Error("PCM_BIN_DIR not set: global setup did not run");

    const home = mkdtempSync(join(tmpdir(), "pcm-e2e-ui-home-"));
    const port = await freePort();
    mkdirSync(join(home, ".pcm"));
    writeFileSync(join(home, ".pcm", "config.yaml"), `port: ${port}\n`);
    const env = {
      ...process.env,
      HOME: home,
      PCM_NO_BROWSER: "1",
      GIT_CONFIG_NOSYSTEM: "1",
      GIT_AUTHOR_NAME: "pcm-test",
      GIT_AUTHOR_EMAIL: "pcm-test@example.com",
      GIT_COMMITTER_NAME: "pcm-test",
      GIT_COMMITTER_EMAIL: "pcm-test@example.com",
    };

    const world: PcmWorld = {
      home,
      port,
      url: (path) => `http://127.0.0.1:${port}${path}`,
      run: (...args) => {
        const r = spawnSync(join(binDir, "pcm"), args, { env, cwd: home, encoding: "utf8" });
        return { code: r.status ?? -1, output: (r.stdout ?? "") + (r.stderr ?? "") };
      },
      post: async (path, body, ...okStatus) => {
        const res = await fetch(world.url(path), {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });
        const ok = okStatus.length > 0 ? okStatus : [201];
        if (!ok.includes(res.status)) {
          throw new Error(`POST ${path}: status ${res.status}: ${await res.text()}`);
        }
      },
    };

    await use(world);

    world.run("server", "stop");
    rmSync(home, { recursive: true, force: true });
  },
});
