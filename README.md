# pcm: personal context manager

Keep reusable context for coding agents (code style, review rules, notes) as md files in a separate git repo. Global context applies everywhere. Project context applies to one repo.

- `pcm-server`: API server. Runs in the background.
- `pcm`: CLI client. kubectl style: `pcm <verb> <resource> [name] [flags]`.

Design: [docs/api-and-cli.md](docs/api-and-cli.md). Requirements: [docs/requirements.md](docs/requirements.md).

## Install

Needs Go (version in `go.mod`) and git. macOS or Linux.

```sh
make install          # pcm + pcm-server into $(go env GOPATH)/bin, skill into ~/.claude/skills/pcm
```

`$(go env GOPATH)/bin` must be in your `PATH`.

## Claude Code

`make install` also installs the pcm skill. In Claude Code, the skill loads your contexts at the start of a coding task and before a review, and saves a rule when you ask Claude to remember it. Source: `internal/skill/SKILL.md`.

## Quick start

```sh
pcm server start                    # start background server
pcm health                          # server: ok

pcm create section style -g         # global section
cd ~/code/my-repo
pcm create section review           # section in current project (auto-created)
pcm get sections
pcm create context go -s review -m 'Check error handling.' --title 'Go review'
pcm get contexts
pcm get context go -s review
pcm get projects

pcm server stop
```

Data lives in `~/.pcm/`:

| Path | Purpose |
|---|---|
| `config.yaml` | Optional. `path`, `remote`, `port` (default 8080) |
| `repo/` | Context git repo. Created with `git init` on first run |
| `server.pid`, `server.log` | Background server |

After a code change: `make install && make restart`.

## Commands

| Command | Purpose |
|---|---|
| `pcm health` | Check server is reachable |
| `pcm server start\|stop\|status` | Manage background server |
| `pcm get projects [name]` | List projects or show one |
| `pcm create project <name>` | Create project |
| `pcm delete project <name> [--force]` | Delete project |
| `pcm get sections [-p <project> \| -g]` | List sections |
| `pcm create section <name> [-p <project> \| -g]` | Create section |
| `pcm delete section <name> [-p <project> \| -g] [--force]` | Delete section |
| `pcm get contexts [-s <section>] [-p <project> \| -g]` | List contexts |
| `pcm get context <name> -s <section>` | Print context md |
| `pcm get contexts -o md [-s <section>]` | Merged global + project contexts as markdown |
| `pcm create context <name> -s <section> -m <text> [--title] [--description] [--tag ...]` | Create context |
| `pcm delete context <name> -s <section>` | Delete context |
| `pcm skill install\|uninstall\|show` | Claude Code skill in `~/.claude/skills/pcm/` |

Scope: `-p <project>` or `-g` (global). No flag: current git repo root name (`My.Repo` → `my-repo`).

## Development

| Target | Purpose |
|---|---|
| `make build` | Build both binaries into `bin/` |
| `make install` | Install both binaries and the Claude Code skill |
| `make restart` | Restart server with installed binary |
| `make test` | Unit tests |
| `make e2e` | BDD specs (godog) with real binaries |
| `make fmt` | gofmt |
| `make vet` | go vet, unit + e2e |

Flow: BDD + TDD. For each feature:

1. Write specs in `e2e/features/*.feature`. Commit them first.
2. Run `make e2e`. Specs fail.
3. For each layer (server, client, CLI): write unit tests, see them fail, write code.
4. Run `make test` and `make e2e`. All pass.

Layout:

```
cmd/pcm/          CLI entry
cmd/pcm-server/   server entry
internal/cli/     cobra commands
internal/client/  API client
internal/server/  HTTP handlers (chi)
internal/store/   md files + frontmatter
internal/config/  ~/.pcm/config.yaml
e2e/              godog specs + steps
```

CI (`.github/workflows/test.yml`): `unit` and `e2e` jobs run in parallel on each PR to `main`.
