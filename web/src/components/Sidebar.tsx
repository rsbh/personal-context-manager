import { Search, Sidebar as ApsaraSidebar, Spinner, Text } from "@raystack/apsara";
import { useQueries } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useLocation } from "react-router";
import {
  listContextsOptions,
  listProjectsOptions,
  listSectionsOptions,
} from "../client/@tanstack/react-query.gen";
import type { StoreContext } from "../client/types.gen";
import { useDebouncedValue } from "../lib/debounce";
import { contextPath } from "../lib/format";
import { SearchResults } from "./SearchResults";
import classes from "./Sidebar.module.css";

interface EmptyProps {
  children: React.ReactNode;
}

function Empty({ children }: EmptyProps) {
  return (
    <Text size="small" variant="tertiary" className={classes.empty}>
      {children}
    </Text>
  );
}

interface ContextItemProps {
  context: StoreContext;
}

function ContextItem({ context }: ContextItemProps) {
  const { pathname } = useLocation();
  const to = contextPath(context.id ?? "");
  return (
    <ApsaraSidebar.Item active={pathname === to} render={<Link to={to} />}>
      {context.name}
    </ApsaraSidebar.Item>
  );
}

interface SectionNodeProps {
  section: string;
  contexts: StoreContext[];
}

function SectionNode({ section, contexts }: SectionNodeProps) {
  return (
    <ApsaraSidebar.Group label={section} collapsible defaultOpen>
      {contexts.map((c) => (
        <ContextItem key={c.id} context={c} />
      ))}
    </ApsaraSidebar.Group>
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
  const scope = { scope: "project", project };
  const [sections, contexts] = useQueries({
    queries: [
      { ...listSectionsOptions({ query: scope }), enabled: opened },
      { ...listContextsOptions({ query: scope }), enabled: opened },
    ],
  });

  return (
    <ApsaraSidebar.Group label={project} collapsible open={opened} onOpenChange={setOpened}>
      {opened && (sections.isPending || contexts.isPending) && <Spinner size={2} />}
      {sections.data?.length === 0 && <Empty>No sections</Empty>}
      {bySection(sections.data ?? [], contexts.data ?? []).map(([s, cs]) => (
        <SectionNode key={s} section={s} contexts={cs} />
      ))}
    </ApsaraSidebar.Group>
  );
}

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

  return (
    <ApsaraSidebar collapsible="none" className={classes.sidebar}>
      <ApsaraSidebar.Header>
        <Search
          aria-label="Search contexts"
          placeholder="Search contexts"
          value={query}
          onChange={(e) => setQuery(e.currentTarget.value)}
          showClearButton
          onClear={() => setQuery("")}
        />
      </ApsaraSidebar.Header>
      <ApsaraSidebar.Main>
        <nav aria-label="Contexts">
          {debounced ? (
            <SearchResults query={debounced} />
          ) : (
            <>
              <ApsaraSidebar.Group label="Global" role="group" aria-label="Global">
                {(sections.isPending || contexts.isPending) && <Spinner size={2} />}
                {sections.data?.length === 0 && <Empty>No sections</Empty>}
                {bySection(sections.data ?? [], contexts.data ?? []).map(([s, cs]) => (
                  <SectionNode key={s} section={s} contexts={cs} />
                ))}
              </ApsaraSidebar.Group>
              <ApsaraSidebar.Group label="Projects" role="group" aria-label="Projects">
                {projects.isPending && <Spinner size={2} />}
                {projects.data?.length === 0 && <Empty>No projects</Empty>}
                {projects.data?.map((p) => (
                  <ProjectNode key={p} project={p} />
                ))}
              </ApsaraSidebar.Group>
            </>
          )}
        </nav>
      </ApsaraSidebar.Main>
    </ApsaraSidebar>
  );
}
