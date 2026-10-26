import { Badge, Group, Select, Stack, Table, Text, Title } from "@mantine/core";
import { useQueries } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router";
import {
  getApiV1ProjectsOptions,
  getApiV1ResolveOptions,
  getApiV1SectionsOptions,
} from "../client/@tanstack/react-query.gen";
import { age, contextPath } from "../lib/format";

// CurrentContexts is the table of contexts that apply: global first, then the project.
export function CurrentContexts() {
  const [params, setParams] = useSearchParams();
  const project = params.get("project") ?? "";
  const section = params.get("section") ?? "";

  const resolveQuery: { project?: string; section?: string } = {};
  if (project) resolveQuery.project = project;
  if (section) resolveQuery.section = section;

  const [projects, globalSections, projectSections, resolved] = useQueries({
    queries: [
      getApiV1ProjectsOptions(),
      getApiV1SectionsOptions({ query: { scope: "global" } }),
      { ...getApiV1SectionsOptions({ query: { scope: "project", project } }), enabled: project !== "" },
      getApiV1ResolveOptions(Object.keys(resolveQuery).length ? { query: resolveQuery } : undefined),
    ],
  });

  const sections = [...new Set([...(globalSections.data ?? []), ...(projectSections.data ?? [])])].sort();
  const set = (key: string, value: string | null) => {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next);
  };

  return (
    <Stack gap="lg">
      <Group justify="space-between" align="flex-end" wrap="wrap">
        <Stack gap={4}>
          <Title order={1}>Current contexts</Title>
          <Text size="sm" c="dimmed">
            Global first, then the project. A project context replaces a global one with the same section and name.
          </Text>
        </Stack>
        <Group gap="sm">
          <Select
            label="Project"
            placeholder="Global only"
            data={projects.data ?? []}
            value={project || null}
            onChange={(v) => set("project", v)}
            clearable
            w={240}
          />
          <Select
            label="Section"
            placeholder="All sections"
            data={sections}
            value={section || null}
            onChange={(v) => set("section", v)}
            clearable
            w={180}
          />
        </Group>
      </Group>

      {resolved.data?.length === 0 && <Text c="dimmed">No contexts found.</Text>}
      {!!resolved.data?.length && (
        <Table.ScrollContainer minWidth={760}>
          <Table striped highlightOnHover withTableBorder>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Scope</Table.Th>
                <Table.Th>Name</Table.Th>
                <Table.Th>Section</Table.Th>
                <Table.Th>Title</Table.Th>
                <Table.Th>Description</Table.Th>
                <Table.Th>Updated</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {resolved.data.map((c) => (
                <Table.Tr key={c.id}>
                  <Table.Td>
                    <Badge variant="light" color={c.scope === "global" ? "blue" : "orange"}>
                      {c.scope === "global" ? "global" : "project"}
                    </Badge>
                  </Table.Td>
                  <Table.Td>
                    <Link to={contextPath(c.id ?? "")}>{c.name}</Link>
                  </Table.Td>
                  <Table.Td>{c.section}</Table.Td>
                  <Table.Td>{c.frontmatter?.title}</Table.Td>
                  <Table.Td>{c.frontmatter?.description}</Table.Td>
                  <Table.Td>{age(c.updated_at)}</Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Table.ScrollContainer>
      )}
    </Stack>
  );
}
