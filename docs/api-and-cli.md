# API and CLI

## Concepts

- **Scope**: `global` or `project/<name>`.
- **Project**: name = basename of git repo root of cwd. No `.pcm` file in project.
- **Section**: user defined (e.g. `style`, `review`). One folder per section.
- **Context**: one md file. A section can have many contexts (e.g. `backend.md`, `ui.md`). Many contexts can apply to one project.
- **Binaries**: `pcm-server` (API server, runs in background) and `pcm` (CLI client, calls API).
- **Home**: `~/.pcm/` holds `config.yaml`, `server.pid`, `server.log`, `index.db` and `repo/`. Only `repo/` is synced.
- **Repo**: separate git repo. Default `~/.pcm/repo`, user can select other path with `pcm config set path <dir>`.
- **First run**: no init command. If repo folder is missing, pcm creates it and runs `git init -b main`. If `remote` is not set, `pcm sync` and `pcm sync status` print:
  ```
  info: remote not set, contexts are local only
  hint: pcm config set remote <git-url> && pcm sync
  ```
- **Layout**: `<repo>/<scope>/<section>/<name>.md`
- **SQLite**: search index only. md files are source of truth.
- **Sync**: manual only. CLI `pcm sync` or UI sync button. Server runs the git CLI (uses user's SSH agent, `~/.ssh/config`, credential helpers). Branch `main`.
  1. `git add -A`, commit `pcm sync: N files changed` if changes.
  2. No remote: stop. Else set `origin` (add or set-url), fetch, rebase on `origin/main`.
  3. Conflict: `git rebase --abort`, 409 with files. Fix by hand in repo, then `pcm sync`.
  4. Push if ahead (`-u origin main`).
  Server reads `remote` from `config.yaml` on each call: no restart needed.
  Timeouts: CLI health 2s, API calls 30s, sync 5m. Server finishes a sync even if the client disconnects. `pcm sync status` before the first push: `ahead` = local commits, hint `run pcm sync to push`.
- **Claude Code**: skill calls CLI. No MCP.

## REST API (`/api/v1`)

`GET /health` (root, not under `/api/v1`): health check.

| Method | Path | Purpose |
|---|---|---|
| GET | `/projects` | List projects |
| POST | `/projects` | Create project |
| GET | `/projects/{project}` | Get project |
| DELETE | `/projects/{project}` | Delete project |
| GET | `/sections` | List sections |
| POST | `/sections` | Create section |
| DELETE | `/sections/{section}` | Delete section |
| GET | `/contexts?scope=&project=&section=` | List contexts with filters |
| GET | `/contexts/{id}` | Read context (md + metadata) |
| POST | `/contexts` | Create context |
| PUT | `/contexts/{id}` | Update context |
| DELETE | `/contexts/{id}` | Delete context |
| GET | `/search?q=&scope=&section=` | Full-text search (SQLite FTS) |
| GET | `/resolve?project=&section=` | Merged global + project contexts with content. Project overrides global on same section/name. No project: global only |
| GET | `/git/status` | branch, remote, changes (porcelain), ahead, behind (from last fetch) |
| POST | `/git/sync` | Commit, pull --rebase, push. 409 `{"error","files"}` on conflict |
| GET | `/git/log?n=10` | Recent commits: hash, date, message |
| GET | `/config` | Read config |
| PUT | `/config` | Update config (remote URL, author) |
| GET | `/*` | Embedded React app |

## CLI (`pcm`)

k8s style: `pcm <verb> <resource> [name] [flags]`. Output is a table. No short names.

| Command | Purpose |
|---|---|
| `pcm get context(s) [-s <section>] [-p <project>]` | Current contexts: `project:` line, then table SCOPE, NAME, SECTION, TITLE, DESCRIPTION, UPDATED. Global first, then project (`-p` or git repo). Project overrides global on same section/name. Not in git repo: global only |
| `pcm get context(s) -g [-s <section>]` | Global contexts only (no SCOPE column) |
| `pcm get context <name> -s <section> [-p <project> \| -g]` | Print md file (frontmatter + content) |
| `pcm create context <name> -s <section> (-m <text> \| -f <file\|->) [--title] [--description] [--tag ...] [--force] [-p <project> \| -g]` | Create context. `--force` replaces existing (PUT) |
| `pcm delete context <name> -s <section> [-p <project> \| -g]` | Delete context |
| `pcm edit context <name> -s <section>` | Open in `$EDITOR` (later) |
| `pcm search <query>` | Search contexts |
| `pcm get contexts -o md [-s <section>] [-p <project> \| -g]` | Merged global + project contexts as markdown (for skill). Not in git repo: global only |
| `pcm get project(s) [name]` | List projects or show one |
| `pcm create project <name>` | Create project |
| `pcm delete project <name> [--force]` | Delete project (`--force` if not empty) |
| `pcm get section(s) [-p <project> \| -g]` | List sections in a scope |
| `pcm create section <name> [-p <project> \| -g]` | Create section |
| `pcm delete section <name> [-p <project> \| -g] [--force]` | Delete section (`--force` if not empty) |
| `pcm sync` | Commit, pull, push. Output: committed, pulled, pushed |
| `pcm sync status` | Branch, remote, changes, ahead, behind, changed files |
| `pcm sync log [-n 10]` | Recent commits (HASH, DATE, MESSAGE) |
| `pcm health` | Check server is reachable (text output, exit 1 if not) |
| `pcm server start` | Start `pcm-server` in background (pid/log in `~/.pcm/`) |
| `pcm server stop` | Stop background server |
| `pcm server status` | Show server health and pid |
| `pcm config get [key]` | Print all keys or one. Keys: `path`, `remote`, `port` |
| `pcm config set <key> <value>` | Write key to `config.yaml`. `path` / `port`: hint to restart server |

Scope flags (like kubectl `-n`): `-p <project>` or `-g` (global). Both: error.
No flag: current project = git repo root folder name (`git rev-parse --show-toplevel`), normalized (`My.Repo` → `my-repo`).
- Not in git repo: error, hint `use -p <project> or -g`.
- `create` auto-creates the current project. `get` / `delete` do not.
- `-p <project>` never auto-creates (typo safety).

## Claude Code skill

Source: `internal/skill/SKILL.md`, embedded in `pcm` with `go:embed`.

| Command | Purpose |
|---|---|
| `pcm skill install` | Write SKILL.md to `~/.claude/skills/pcm/` (overwrite) |
| `pcm skill uninstall` | Remove `~/.claude/skills/pcm/` |
| `pcm skill show` | Print SKILL.md |

Skill: start server if needed, load `pcm get contexts -o md` at start of coding task and before review, save with `pcm create context` when user asks to remember, dump session rules and decisions to a project context named after the PR title (fallback: branch) on request (user approves first; asks overwrite with `--force` or new name with `-2` suffix), manage contexts on request. `allowed-tools: Bash(pcm:*), Bash(gh pr view:*), Bash(git branch:*)`.
