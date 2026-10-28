import { Flex, IconButton, Search, Spinner, Text } from "@raystack/apsara";
import { useQueries } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router";
import {
  listContextsOptions,
  listProjectsOptions,
  listSectionsOptions,
} from "../client/@tanstack/react-query.gen";
import type { StoreContext } from "../client/types.gen";
import { useDebouncedValue } from "../lib/debounce";
import { contextPath } from "../lib/format";
import { SearchResults } from "./SearchResults";
import { Group, Item, Row } from "./SidebarNav";
import { CreateProjectButton, CreateSectionButton, RowMenu } from "./SidebarActions";
import classes from "./Sidebar.module.css";

function Loading() {
  return (
    <Row>
      <Spinner size={2} />
    </Row>
  );
}

interface EmptyProps {
  children: React.ReactNode;
}

function Empty({ children }: EmptyProps) {
  return (
    <Row>
      <Text size="small" variant="tertiary">
        {children}
      </Text>
    </Row>
  );
}

interface ContextItemProps {
  context: StoreContext;
}

function ContextItem({ context }: ContextItemProps) {
  const { pathname } = useLocation();
  const to = contextPath(context.id ?? "");
  return (
    <Item to={to} active={pathname === to}>
      {context.name}
    </Item>
  );
}

interface NewButtonProps {
  section: string;
  project?: string;
}

// NewButton opens the new context form with the scope and section filled in.
function NewButton({ section, project }: NewButtonProps) {
  const navigate = useNavigate();
  const onClick = useCallback(() => {
    const q = new URLSearchParams(
      project ? { scope: "project", project, section } : { scope: "global", section },
    );
    navigate(`/new?${q}`);
  }, [navigate, section, project]);
  return (
    <IconButton size={4} aria-label={`New context in ${section}`} onClick={onClick}>
      <Plus size={16} />
    </IconButton>
  );
}

interface SectionNodeProps {
  section: string;
  project?: string;
  contexts: StoreContext[];
}

function SectionNode({ section, project, contexts }: SectionNodeProps) {
  const actions = useMemo(
    () => (
      <Flex gap={1}>
        <NewButton section={section} project={project} />
        <RowMenu kind="section" name={section} project={project} />
      </Flex>
    ),
    [section, project],
  );
  return (
    <Group label={section} collapsible defaultOpen trailing={actions}>
      {contexts.map((c) => (
        <ContextItem key={c.id} context={c} />
      ))}
    </Group>
  );
}

// bySection groups contexts by section, for every section name given (also empty ones).
function bySection(sections: string[], contexts: StoreContext[]): [string, StoreContext[]][] {
  return sections.map((s) => [s, contexts.filter((c) => c.section === s)]);
}

interface ProjectNodeProps {
  project: string;
}

function ProjectNode({ project }: ProjectNodeProps) {
  const [opened, setOpened] = useState(false);
  const open = useCallback(() => setOpened(true), []);
  const actions = useMemo(
    () => (
      <Flex gap={1}>
        <CreateSectionButton project={project} onCreated={open} />
        <RowMenu kind="project" name={project} />
      </Flex>
    ),
    [project, open],
  );
  const scope = { scope: "project", project };
  const [sections, contexts] = useQueries({
    queries: [
      { ...listSectionsOptions({ query: scope }), enabled: opened },
      { ...listContextsOptions({ query: scope }), enabled: opened },
    ],
  });

  return (
    <Group label={project} collapsible open={opened} onOpenChange={setOpened} trailing={actions}>
      {opened && (sections.isPending || contexts.isPending) && <Loading />}
      {sections.data?.length === 0 && <Empty>No sections</Empty>}
      {bySection(sections.data ?? [], contexts.data ?? []).map(([s, cs]) => (
        <SectionNode key={s} section={s} project={project} contexts={cs} />
      ))}
    </Group>
  );
}

const globalActions = <CreateSectionButton />;
const projectsActions = <CreateProjectButton />;

export function Sidebar() {
  const [query, setQuery] = useState("");
  const debounced = useDebouncedValue(query.trim(), 250);
  const [sections, contexts, projects] = useQueries({
    queries: [
      listSectionsOptions({ query: { scope: "global" } }),
      listContextsOptions({ query: { scope: "global" } }),
      listProjectsOptions(),
    ],
  });
  const onChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => setQuery(e.currentTarget.value),
    [],
  );
  const onClear = useCallback(() => setQuery(""), []);

  return (
    <aside className={classes.sidebar}>
      <div className={classes.header}>
        <Search
          aria-label="Search contexts"
          placeholder="Search contexts"
          value={query}
          onChange={onChange}
          showClearButton
          onClear={onClear}
        />
      </div>
      <div className={classes.main}>
        <nav aria-label="Contexts">
          {debounced ? (
            <SearchResults query={debounced} />
          ) : (
            <ul className={classes.list}>
              <Group label="Global" trailing={globalActions}>
                {(sections.isPending || contexts.isPending) && <Loading />}
                {sections.data?.length === 0 && <Empty>No sections</Empty>}
                {bySection(sections.data ?? [], contexts.data ?? []).map(([s, cs]) => (
                  <SectionNode key={s} section={s} contexts={cs} />
                ))}
              </Group>
              <Group label="Projects" trailing={projectsActions}>
                {projects.isPending && <Loading />}
                {projects.data?.length === 0 && <Empty>No projects</Empty>}
                {projects.data?.map((p) => (
                  <ProjectNode key={p} project={p} />
                ))}
              </Group>
            </ul>
          )}
        </nav>
      </div>
    </aside>
  );
}
