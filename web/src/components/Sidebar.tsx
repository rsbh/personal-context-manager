import { Loader, NavLink, Stack, Text, TextInput } from "@mantine/core";
import { useDebouncedValue } from "@mantine/hooks";
import { useQueries } from "@tanstack/react-query";
import { useState } from "react";
import { Link } from "react-router";
import {
  getApiV1ContextsOptions,
  getApiV1ProjectsOptions,
  getApiV1SectionsOptions,
} from "../client/@tanstack/react-query.gen";
import type { StoreContext } from "../client/types.gen";
import { contextPath } from "../lib/format";
import { SearchResults } from "./SearchResults";

interface SidebarGroupProps {
  label: string;
  loading: boolean;
  empty: string;
  isEmpty: boolean;
  children: React.ReactNode;
}

function SidebarGroup({ label, loading, empty, isEmpty, children }: SidebarGroupProps) {
  return (
    <Stack gap={2} role="group" aria-label={label}>
      <Text size="xs" fw={600} tt="uppercase" c="dimmed" px="sm" py={6}>
        {label}
      </Text>
      {loading && <Loader size="xs" mx="sm" />}
      {!loading && isEmpty && (
        <Text size="sm" c="dimmed" px="sm">
          {empty}
        </Text>
      )}
      {children}
    </Stack>
  );
}

interface SectionNodeProps {
  section: string;
  contexts: StoreContext[];
}

function SectionNode({ section, contexts }: SectionNodeProps) {
  return (
    <NavLink label={section} defaultOpened childrenOffset="md">
      {contexts.map((c) => (
        <NavLink key={c.id} component={Link} to={contextPath(c.id ?? "")} label={c.name} />
      ))}
    </NavLink>
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
      { ...getApiV1SectionsOptions({ query: scope }), enabled: opened },
      { ...getApiV1ContextsOptions({ query: scope }), enabled: opened },
    ],
  });

  return (
    <NavLink label={project} opened={opened} onChange={setOpened} childrenOffset="md">
      {opened && (sections.isPending || contexts.isPending) && <Loader size="xs" mx="sm" />}
      {sections.data?.length === 0 && (
        <Text size="sm" c="dimmed" px="sm">
          No sections
        </Text>
      )}
      {bySection(sections.data ?? [], contexts.data ?? []).map(([s, cs]) => (
        <SectionNode key={s} section={s} contexts={cs} />
      ))}
    </NavLink>
  );
}

export function Sidebar() {
  const [query, setQuery] = useState("");
  const [debounced] = useDebouncedValue(query.trim(), 250);
  const [sections, contexts, projects] = useQueries({
    queries: [
      getApiV1SectionsOptions({ query: { scope: "global" } }),
      getApiV1ContextsOptions({ query: { scope: "global" } }),
      getApiV1ProjectsOptions(),
    ],
  });

  return (
    <Stack component="nav" aria-label="Contexts" gap="md" p="sm">
      <TextInput
        aria-label="Search contexts"
        placeholder="Search contexts"
        value={query}
        onChange={(e) => setQuery(e.currentTarget.value)}
      />
      {debounced ? (
        <SearchResults query={debounced} />
      ) : (
        <>
          <SidebarGroup
            label="Global"
            loading={sections.isPending || contexts.isPending}
            empty="No sections"
            isEmpty={sections.data?.length === 0}
          >
            {bySection(sections.data ?? [], contexts.data ?? []).map(([s, cs]) => (
              <SectionNode key={s} section={s} contexts={cs} />
            ))}
          </SidebarGroup>
          <SidebarGroup label="Projects" loading={projects.isPending} empty="No projects" isEmpty={projects.data?.length === 0}>
            {projects.data?.map((p) => <ProjectNode key={p} project={p} />)}
          </SidebarGroup>
        </>
      )}
    </Stack>
  );
}
