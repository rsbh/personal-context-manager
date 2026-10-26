import { Loader, NavLink, Stack, Text } from "@mantine/core";
import { useQueries } from "@tanstack/react-query";
import { getApiV1ProjectsOptions, getApiV1SectionsOptions } from "../client/@tanstack/react-query.gen";

interface SidebarGroupProps {
  label: string;
  items?: string[];
  empty: string;
  loading: boolean;
}

function SidebarGroup({ label, items, empty, loading }: SidebarGroupProps) {
  return (
    <Stack gap={2} role="group" aria-label={label}>
      <Text size="xs" fw={600} tt="uppercase" c="dimmed" px="sm" py={6}>
        {label}
      </Text>
      {loading && <Loader size="xs" mx="sm" />}
      {items?.length === 0 && (
        <Text size="sm" c="dimmed" px="sm">
          {empty}
        </Text>
      )}
      {items?.map((name) => <NavLink key={name} label={name} />)}
    </Stack>
  );
}

export function Sidebar() {
  const [sections, projects] = useQueries({
    queries: [
      getApiV1SectionsOptions({ query: { scope: "global" } }),
      getApiV1ProjectsOptions(),
    ],
  });

  return (
    <Stack component="nav" aria-label="Contexts" gap="md" p="sm">
      <SidebarGroup label="Global" items={sections.data} empty="No sections" loading={sections.isPending} />
      <SidebarGroup label="Projects" items={projects.data} empty="No projects" loading={projects.isPending} />
    </Stack>
  );
}
