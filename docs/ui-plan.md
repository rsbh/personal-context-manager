# UI plan

Web UI for pcm, served by `pcm-server`. Design: Claude Design canvas "pcm UI design" (4 screens: Browse, Edit, Current contexts, Sync).

## Decisions

| Topic | Choice |
|---|---|
| Build | Vite + React + TypeScript in `web/`, npm |
| Components | Mantine |
| Routing / data | react-router, TanStack Query |
| Editor | CodeMirror 6 (`@uiw/react-codemirror`) + preview (react-markdown, remark-gfm) |
| Layout | Sidebar tree (Global / Projects > sections > contexts) + main panel. Table "Current contexts" view per project |
| Saving | Save button + Discard, unsaved warning. No autosave |
| Embed | `web/dist` embedded with `go:embed`. Not in git. No build: placeholder page "UI not built: run make web" |
| Serving | One port: UI at `/` (SPA fallback to `index.html`), API at `/api/v1`, `/health`. Unknown `/api/*`: JSON 404 |
| Bind | `127.0.0.1` only |
| `pcm ui` | Opens browser at `http://127.0.0.1:<config port>`. Server down: error + hint `pcm server start`. `PCM_NO_BROWSER=1`: print URL only |
| Tests | Vitest + Testing Library (components). godog + playwright-go browser specs (`@ui` features, build tag `e2e_ui`, own CI job) |

## Phases (one PR each, specs first)

1. **Skeleton**: `web/` scaffold, embed + placeholder, SPA route, bind 127.0.0.1, `pcm ui`, Makefile, CI. Header (server status) + sidebar (projects, global sections), read-only.
2. **Browse**: full tree (sections > contexts), filter, context view (rendered md + raw file), Current contexts table with project/section filter.
3. **Edit**: create/edit/delete contexts (fields, CodeMirror, preview, save/discard, unsaved warning); create/delete sections and projects (confirm, force).
4. **Sync**: header badge + Sync button, Sync page (status, changed files, log, conflict).
