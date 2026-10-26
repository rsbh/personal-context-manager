---
name: pcm
description: Load and save the user's personal coding context (code style, review rules, notes) with the pcm CLI. Use at the start of any coding task, before a code review, when the user asks to remember or save a rule or style, when the user asks to dump or save the session context to pcm, and when the user asks to list, show or delete pcm contexts, sections or projects.
allowed-tools: Bash(pcm:*), Bash(gh pr view:*), Bash(git branch:*), Write
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
| Find contexts by text (fuzzy) | `pcm search '<words>' [-s <section>] [-g \| --all]` |
| Rebuild search index | `pcm index rebuild` |
| Delete a context | `pcm delete context <name> -s <section>` |
| Delete a section | `pcm delete section <name> [--force]` |
| Sync with git remote | `pcm sync` (status: `pcm sync status`) |

Ask the user before you delete anything.

Scope flags: `-p <project>` or `-g`. With no flag, pcm uses the current git repo.

Search: words match in any order, with typos. Use `+word` when the word must match, `-word` when it must not match, and `"words"` for an exact phrase. If search does not find a context that you know exists, run `pcm index rebuild` and search again.

## 5. Dump session

Do this when the user asks to dump or save the current session context to pcm.

1. Collect the rules, decisions and preferences from this session. Write them as short markdown bullets. Do not include a log of the work.
2. Name: run `gh pr view --json title -q .title`. If there is no PR, run `git branch --show-current`. Change the result to kebab-case. If there is no branch (detached HEAD), suggest a short kebab-case name from the session topic and ask the user to approve or change it.
3. Section: run `pcm get sections`. Pick the section that matches. If no section matches, select a short name and run `pcm create section <name>`.
4. Show the user the scope (current project), section, name and content. Wait for approval. Make the changes that the user asks for.
5. Run `pcm get contexts -s <section>`. If the name exists, ask the user: overwrite or create new.
   - Overwrite: add `--force` to the command in step 6.
   - Create new: add `-2` to the name. If that name exists, use `-3`, and continue.
6. Write the content to `<name>.md` in your scratchpad folder. If you have no scratchpad folder, use `/tmp`. Then run:

   ```sh
   pcm create context <name> -s <section> -f <file> --description '<one line>'
   ```

   Do not use a heredoc. Do not delete the file.

7. Tell the user the scope, section and name.
