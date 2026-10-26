# API and CLI

## Concepts

- **Scope**: `global` or `project/<name>`.
- **Project**: name = basename of git repo root of cwd. No `.pcm` file in project.
- **Section**: user defined (e.g. `style`, `review`). One folder per section.
- **Context**: one md file. A section can have many contexts (e.g. `backend.md`, `ui.md`). Many contexts can apply to one project.
- **Binaries**: `pcm-server` (API server, runs in background) and `pcm` (CLI client, calls API).
- **Home**: `~/.pcm/` holds `config.yaml`, `server.pid`, `server.log`, `index.db` and `repo/`. Only `repo/` is synced.
- **Repo**: separate git repo. Default `~/.pcm/repo`, user can select other path with `pcm config set path <dir>`.
- **First run**: no init command. If repo folder is missing, pcm creates it and runs `git init`. If `remote` is not set, CLI prints info:
  ```
  info: remote not set. Contexts are local only.
    pcm config set remote <git-url>
    pcm sync
  ```
- **Layout**: `<repo>/<scope>/<section>/<name>.md`
- **SQLite**: search index only. md files are source of truth.
- **Sync**: manual only. CLI `pcm sync` or UI sync button.
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
| GET | `/resolve?project=&section=&name=` | Merged global + project contexts |
| GET | `/git/status` | Repo status |
| POST | `/git/sync` | Pull, commit, push |
| GET | `/git/log` | Recent commits |
| GET | `/config` | Read config |
| PUT | `/config` | Update config (remote URL, author) |
| GET | `/*` | Embedded React app |

## CLI (`pcm`)

k8s style: `pcm <verb> <resource> [name] [flags]`. Output is a table. No short names.

| Command | Purpose |
|---|---|
| `pcm add [text] -s <section> [-p <project> \| -g] [-n <name>]` | Save text (arg or stdin) |
| `pcm list [-s <section>] [-p <project>] [-g]` | List contexts |
| `pcm show <id>` | Print context |
| `pcm edit <id>` | Open in `$EDITOR` |
| `pcm rm <id>` | Delete context |
| `pcm search <query>` | Search contexts |
| `pcm context [-s <section>] [-n <name>]` | Print merged global + current project contexts (for skill) |
| `pcm get project(s) [name]` | List projects or show one |
| `pcm create project <name>` | Create project |
| `pcm delete project <name> [--force]` | Delete project (`--force` if not empty) |
| `pcm get section(s) [-p <project> \| -g]` | List sections in a scope |
| `pcm create section <name> [-p <project> \| -g]` | Create section |
| `pcm delete section <name> [-p <project> \| -g] [--force]` | Delete section (`--force` if not empty) |
| `pcm sync` | Git pull, commit, push |
| `pcm status` | Git status |
| `pcm health` | Check server is reachable (text output, exit 1 if not) |
| `pcm server start` | Start `pcm-server` in background (pid/log in `~/.pcm/`) |
| `pcm server stop` | Stop background server |
| `pcm server status` | Show server health and pid |
| `pcm config get\|set <key> [value]` | Read/write config. Keys: `path`, `remote` |

Scope flags (like kubectl `-n`): `-p <project>` or `-g` (global). Both: error.
No flag: current project = git repo root folder name (`git rev-parse --show-toplevel`), normalized (`My.Repo` → `my-repo`).
- Not in git repo: error, hint `use -p <project> or -g`.
- `create` auto-creates the current project. `get` / `delete` do not.
- `-p <project>` never auto-creates (typo safety).

## Claude Code skill

Skill runs `pcm context` and `pcm add`. Skill file ships in repo.
