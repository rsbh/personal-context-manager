---
name: pcm
description: Load and save the user's personal coding context (code style, review rules, notes) with the pcm CLI. Use at the start of any coding task, before a code review, when the user asks to remember or save a rule or style, and when the user asks to list, show or delete pcm contexts, sections or projects.
allowed-tools: Bash(pcm:*)
---

# pcm: personal context manager

pcm keeps the user's coding context as markdown. Global context applies to all projects. Project context applies to the current git repo and overrides global context with the same section and name.

## 1. Server

Run `pcm health`. If the exit code is 1, run `pcm server start`.

## 2. Load context

Do this at the start of a coding task and before a code review.

1. Run `pcm get contexts -o md` in the project folder.
2. Follow these rules for the whole task. If the user gives a different instruction in the chat, the user instruction wins.
3. For a review, you can run `pcm get contexts -o md -s review` if a `review` section exists.

If the output is `No contexts found.`, continue without context.

## 3. Save context

Do this when the user asks to remember or save a rule, a style or a note.

1. Scope: use the current project by default. Use `-g` when the user says always, everywhere or global.
2. Section: run `pcm get sections` (add `-g` for global). Pick the section that matches. If no section matches, ask the user for a name, then run `pcm create section <name>`.
3. Name: use a short kebab-case name, for example `error-handling`.
4. Run `pcm create context <name> -s <section> -m '<rule>' --description '<one line>'`.
5. If the context already exists, show it with `pcm get context <name> -s <section>` and ask the user what to do.
6. Tell the user the scope, section and name.

## 4. Manage

| Task | Command |
|---|---|
| List projects | `pcm get projects` |
| List sections | `pcm get sections [-p <project> \| -g]` |
| List current contexts (global + project) | `pcm get contexts [-s <section>] [-p <project>]` |
| List global contexts only | `pcm get contexts -g` |
| Show one context | `pcm get context <name> -s <section>` |
| Delete a context | `pcm delete context <name> -s <section>` |
| Delete a section | `pcm delete section <name> [--force]` |
| Sync with git remote | `pcm sync` (status: `pcm sync status`) |

Ask the user before you delete anything.

Scope flags: `-p <project>` or `-g`. With no flag, pcm uses the current git repo.
