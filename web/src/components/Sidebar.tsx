import { IconButton, Menu, Select, Spinner, Text } from "@raystack/apsara";
import { useQueries } from "@tanstack/react-query";
import { Ellipsis, LayoutGrid, Pencil, Plus, Trash2 } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import {
  listContextsOptions,
  listProjectsOptions,
  listSectionsOptions,
} from "../client/@tanstack/react-query.gen";
import type { StoreContext } from "../client/types.gen";
import { contextPath } from "../lib/format";
import { DeleteContextDialog } from "./DeleteContextDialog";
import { useCreateProject, useScopeDelete } from "./SidebarActions";
import classes from "./Sidebar.module.css";

// GLOBAL and NEW_PROJECT are select values. Project names cannot start with "_".
const GLOBAL = "_global";
const NEW_PROJECT = "_new";

// scopeOf reads the project of the page from the URL: "" for global, undefined when the page
// has no scope (home, current, sync).
export function scopeOf(pathname: string): string | undefined {
  const m = pathname.match(/^\/(?:contexts\/|edit\/)?projects\/([^/]+)/);
  if (m) return decodeURIComponent(m[1]);
  if (/^\/(?:contexts\/|edit\/)?global(?:\/|$)/.test(pathname)) return "";
  return undefined;
}

// overviewPath is the scope page of a project, or of global for "".
function overviewPath(project: string): string {
  return project ? `/projects/${encodeURIComponent(project)}` : "/global";
}

// newContextPath opens the new context form with the scope and section filled in.
function newContextPath(section: string, project: string): string {
  const q = new URLSearchParams(
    project ? { scope: "project", project, section } : { scope: "global", section },
  );
  return `/new?${q}`;
}

const editIcon = <Pencil size={16} />;
const deleteIcon = <Trash2 size={16} />;
const newIcon = <Plus size={16} />;

interface MoreButtonProps {
  label: string;
}

// MoreButton is the "⋯" menu trigger. It shows on row hover or focus.
function MoreButton({ label }: MoreButtonProps) {
  const trigger = useMemo(
    () => <IconButton size={3} aria-label={label} className={classes.more} />,
    [label],
  );
  return (
    <Menu.Trigger render={trigger}>
      <Ellipsis size={16} />
    </Menu.Trigger>
  );
}

interface ContextRowProps {
  context: StoreContext;
}

function ContextRow({ context }: ContextRowProps) {
  const [open, setOpen] = useState(false);
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const id = context.id ?? "";
  const to = contextPath(id);
  const onEdit = useCallback(() => navigate(`/edit/${id}`), [navigate, id]);
  const show = useCallback(() => setOpen(true), []);
  const onDeleted = useCallback(() => {
    if (pathname === to || pathname === `/edit/${id}`) navigate("/");
  }, [pathname, to, id, navigate]);
  return (
    <li className={classes.row}>
      <Link to={to} aria-current={pathname === to ? "page" : undefined} className={classes.item}>
        {context.name}
      </Link>
      <Menu>
        <MoreButton label={`Actions for ${context.name}`} />
        <Menu.Content align="start">
          <Menu.Item leadingIcon={editIcon} onClick={onEdit}>
            Edit
          </Menu.Item>
          <Menu.Item leadingIcon={deleteIcon} className={classes.danger} onClick={show}>
            Delete
          </Menu.Item>
        </Menu.Content>
      </Menu>
      <DeleteContextDialog id={id} open={open} onOpenChange={setOpen} onDeleted={onDeleted} />
    </li>
  );
}

interface SectionGroupProps {
  section: string;
  project: string;
  contexts: StoreContext[];
}

function SectionGroup({ section, project, contexts }: SectionGroupProps) {
  const navigate = useNavigate();
  const remove = useScopeDelete({ kind: "section", name: section, project: project || undefined });
  const onNew = useCallback(
    () => navigate(newContextPath(section, project)),
    [navigate, section, project],
  );
  return (
    <li aria-label={section} className={classes.group}>
      <div className={classes.row}>
        <span className={classes.sectionLabel}>{section}</span>
        <Menu>
          <MoreButton label={`Actions for section ${section}`} />
          <Menu.Content align="start">
            <Menu.Item leadingIcon={newIcon} onClick={onNew}>
              New context
            </Menu.Item>
            <Menu.Item leadingIcon={deleteIcon} className={classes.danger} onClick={remove.show}>
              Delete section
            </Menu.Item>
          </Menu.Content>
        </Menu>
      </div>
      <ul className={classes.list}>
        {contexts.map((c) => (
          <ContextRow key={c.id} context={c} />
        ))}
      </ul>
      {remove.dialog}
    </li>
  );
}

// Sidebar shows one scope at a time: a scope select, the scope overview link, then the
// sections of the scope with their contexts.
export function Sidebar() {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const project = scopeOf(pathname) ?? "";
  const scope = project ? { scope: "project", project } : { scope: "global" };
  const [projects, sections, contexts] = useQueries({
    queries: [
      listProjectsOptions(),
      listSectionsOptions({ query: scope }),
      listContextsOptions({ query: scope }),
    ],
  });
  const created = useCallback((name: string) => navigate(overviewPath(name)), [navigate]);
  const newProject = useCreateProject(created);
  const onScope = useCallback(
    (v: string) => {
      if (v === NEW_PROJECT) newProject.show();
      else navigate(overviewPath(v === GLOBAL ? "" : v));
    },
    [navigate, newProject],
  );
  const groups = useMemo(
    () =>
      (sections.data ?? []).map(
        (s) => [s, (contexts.data ?? []).filter((c) => c.section === s)] as const,
      ),
    [sections.data, contexts.data],
  );
  const overview = overviewPath(project);

  return (
    <aside className={classes.sidebar}>
      <nav aria-label="Contexts" className={classes.main}>
        <Select value={project || GLOBAL} onValueChange={onScope}>
          <Select.Trigger aria-label="Scope" className={classes.select}>
            <Select.Value />
          </Select.Trigger>
          <Select.Content>
            <Select.Item value={GLOBAL}>Global</Select.Item>
            {projects.data?.map((p) => (
              <Select.Item key={p} value={p}>
                {p}
              </Select.Item>
            ))}
            <Select.Item value={NEW_PROJECT}>New project…</Select.Item>
          </Select.Content>
        </Select>
        <Link
          to={overview}
          aria-current={pathname === overview ? "page" : undefined}
          className={`${classes.item} ${classes.overview}`}
        >
          <LayoutGrid size={16} />
          {project ? "Project overview" : "Global overview"}
        </Link>
        {(sections.isPending || contexts.isPending) && <Spinner size={2} />}
        {sections.data?.length === 0 && (
          <Text size="small" variant="tertiary" className={classes.empty}>
            No sections
          </Text>
        )}
        <ul className={classes.list}>
          {groups.map(([s, cs]) => (
            <SectionGroup key={s} section={s} project={project} contexts={cs} />
          ))}
        </ul>
      </nav>
      {newProject.dialog}
    </aside>
  );
}
