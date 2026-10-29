import { Callout, Flex, Headline, IconButton, Spinner, Text } from "@raystack/apsara";
import { useQueries } from "@tanstack/react-query";
import { FolderPlus, Pencil, Plus, Trash2 } from "lucide-react";
import { useCallback, useId, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import {
  listContextsOptions,
  listProjectsOptions,
  listSectionsOptions,
} from "../client/@tanstack/react-query.gen";
import type { StoreContext } from "../client/types.gen";
import { DeleteContextDialog } from "../components/DeleteContextDialog";
import { CreateSectionButton, useScopeDelete } from "../components/SidebarActions";
import { age, contextPath } from "../lib/format";
import { renderH1 } from "../lib/render";
import classes from "./ScopePage.module.css";

// count formats "1 section" or "2 sections".
function count(n: number, word: string): string {
  return `${n} ${word}${n === 1 ? "" : "s"}`;
}

interface ContextRowProps {
  context: StoreContext;
}

// ContextRow links to the context and has Edit and Delete icons.
function ContextRow({ context }: ContextRowProps) {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const id = context.id ?? "";
  const onEdit = useCallback(() => navigate(`/edit/${id}`), [navigate, id]);
  const show = useCallback(() => setOpen(true), []);
  return (
    <li className={classes.row}>
      <Link to={contextPath(id)} aria-label={context.name} className={classes.link}>
        <span className={classes.name}>{context.name}</span>
        <span className={classes.desc}>{context.frontmatter?.description}</span>
        <span className={classes.age}>{age(context.updated_at)}</span>
      </Link>
      <Flex gap={1} className={classes.rowActions}>
        <IconButton size={3} aria-label={`Edit ${context.name}`} title="Edit" onClick={onEdit}>
          <Pencil size={16} />
        </IconButton>
        <IconButton
          size={3}
          aria-label={`Delete ${context.name}`}
          title="Delete"
          className={classes.danger}
          onClick={show}
        >
          <Trash2 size={16} />
        </IconButton>
      </Flex>
      <DeleteContextDialog id={id} open={open} onOpenChange={setOpen} />
    </li>
  );
}

interface SectionCardProps {
  section: string;
  project?: string;
  contexts: StoreContext[];
}

// SectionCard lists the contexts of one section, with New context and Delete section.
function SectionCard({ section, project, contexts }: SectionCardProps) {
  const headingId = useId();
  const navigate = useNavigate();
  const remove = useScopeDelete({ kind: "section", name: section, project });
  const onNew = useCallback(() => {
    const q = new URLSearchParams(
      project ? { scope: "project", project, section } : { scope: "global", section },
    );
    navigate(`/new?${q}`);
  }, [navigate, section, project]);
  return (
    <section aria-labelledby={headingId} className={classes.card}>
      <div className={classes.cardHead}>
        <Flex align="baseline" gap={3}>
          <h2 id={headingId} className={classes.cardTitle}>
            {section}
          </h2>
          <Text size="small" variant="secondary">
            {count(contexts.length, "context")}
          </Text>
        </Flex>
        <Flex gap={1}>
          <IconButton
            size={3}
            aria-label={`New context in ${section}`}
            title="New context"
            className={classes.success}
            onClick={onNew}
          >
            <Plus size={16} />
          </IconButton>
          <IconButton
            size={3}
            aria-label={`Delete section ${section}`}
            title="Delete section"
            className={classes.danger}
            onClick={remove.show}
          >
            <Trash2 size={16} />
          </IconButton>
        </Flex>
      </div>
      {contexts.length > 0 && (
        <ul className={classes.rows}>
          {contexts.map((c) => (
            <ContextRow key={c.id} context={c} />
          ))}
        </ul>
      )}
      {remove.dialog}
    </section>
  );
}

interface DeleteProjectButtonProps {
  project: string;
}

function DeleteProjectButton({ project }: DeleteProjectButtonProps) {
  const remove = useScopeDelete({ kind: "project", name: project });
  return (
    <>
      <IconButton
        size={4}
        aria-label="Delete project"
        title="Delete project"
        className={`${classes.headAction} ${classes.danger}`}
        onClick={remove.show}
      >
        <Trash2 size={16} />
      </IconButton>
      {remove.dialog}
    </>
  );
}

const folderPlusIcon = <FolderPlus size={16} />;

interface ScopePageProps {
  // project is empty for the global scope.
  project?: string;
}

// ScopePage is the overview of the global scope or one project: its sections and contexts.
export function ScopePage({ project }: ScopePageProps) {
  const scope = project ? { scope: "project", project } : { scope: "global" };
  const [sections, contexts, projects] = useQueries({
    queries: [
      listSectionsOptions({ query: scope }),
      listContextsOptions({ query: scope }),
      listProjectsOptions(),
    ],
  });
  // groups pairs each section with its contexts, also for empty sections.
  const groups = useMemo(
    () =>
      (sections.data ?? []).map(
        (s) => [s, (contexts.data ?? []).filter((c) => c.section === s)] as const,
      ),
    [sections.data, contexts.data],
  );

  if (project && projects.data && !projects.data.includes(project)) {
    return (
      <Callout type="alert" variant="outline">
        <Text weight="medium">Project not found</Text>
      </Callout>
    );
  }
  if (sections.isPending || contexts.isPending) return <Spinner size={3} />;

  return (
    <Flex direction="column" gap={6} className={classes.page}>
      <div className={classes.head}>
        <Flex direction="column" gap={2}>
          <Text size="small" variant="secondary">
            {project ? "Project" : "Global"}
          </Text>
          <Headline size="t3" render={renderH1}>
            {project ?? "Global"}
          </Headline>
          <Text variant="secondary">
            {count(groups.length, "section")} · {count(contexts.data?.length ?? 0, "context")}
          </Text>
        </Flex>
        <Flex gap={2}>
          <CreateSectionButton
            project={project}
            label="New section"
            icon={folderPlusIcon}
            className={`${classes.headAction} ${classes.success}`}
          />
          {project && <DeleteProjectButton project={project} />}
        </Flex>
      </div>
      {groups.length === 0 && <Text variant="secondary">No sections yet.</Text>}
      {groups.map(([s, cs]) => (
        <SectionCard key={s} section={s} project={project} contexts={cs} />
      ))}
    </Flex>
  );
}

// ProjectPage is the ScopePage of the project in the URL.
export function ProjectPage() {
  const { name } = useParams();
  return <ScopePage project={name} />;
}
