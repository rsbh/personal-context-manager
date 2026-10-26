import { Badge, Flex, Headline, Label, Select, Table, Text } from "@raystack/apsara";
import { useQueries } from "@tanstack/react-query";
import { useCallback } from "react";
import { Link, useSearchParams } from "react-router";
import {
  listProjectsOptions,
  resolveContextsOptions,
  listSectionsOptions,
} from "../client/@tanstack/react-query.gen";
import { age, contextPath } from "../lib/format";
import { renderH1 } from "../lib/render";
import classes from "./CurrentContexts.module.css";

// ALL is the select value for "no filter". Names cannot start with "_".
const ALL = "_all";

// NONE is a stable empty list, for props while data loads.
const NONE: string[] = [];

interface FilterSelectProps {
  id: string;
  label: string;
  allLabel: string;
  items: string[];
  value: string;
  onChange: (value: string) => void;
}

function FilterSelect({ id, label, allLabel, items, value, onChange }: FilterSelectProps) {
  const onValueChange = useCallback((v: string) => onChange(v === ALL ? "" : v), [onChange]);
  return (
    <Flex direction="column" gap={2}>
      <Label htmlFor={id}>{label}</Label>
      <Select value={value || ALL} onValueChange={onValueChange}>
        <Select.Trigger id={id} aria-label={label} className={classes.select}>
          <Select.Value />
        </Select.Trigger>
        <Select.Content>
          <Select.Item value={ALL}>{allLabel}</Select.Item>
          {items.map((it) => (
            <Select.Item key={it} value={it}>
              {it}
            </Select.Item>
          ))}
        </Select.Content>
      </Select>
    </Flex>
  );
}

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
      listProjectsOptions(),
      listSectionsOptions({ query: { scope: "global" } }),
      { ...listSectionsOptions({ query: { scope: "project", project } }), enabled: project !== "" },
      resolveContextsOptions(
        Object.keys(resolveQuery).length ? { query: resolveQuery } : undefined,
      ),
    ],
  });

  const sections = [
    ...new Set([...(globalSections.data ?? []), ...(projectSections.data ?? [])]),
  ].toSorted();
  const set = useCallback(
    (key: string, value: string) => {
      const next = new URLSearchParams(params);
      if (value) next.set(key, value);
      else next.delete(key);
      setParams(next);
    },
    [params, setParams],
  );
  const setProject = useCallback((v: string) => set("project", v), [set]);
  const setSection = useCallback((v: string) => set("section", v), [set]);

  return (
    <Flex direction="column" gap={7}>
      <Flex justify="between" align="end" wrap="wrap" gap={5}>
        <Flex direction="column" gap={2}>
          <Headline size="t2" render={renderH1}>
            Current contexts
          </Headline>
          <Text variant="secondary">
            Global first, then the project. A project context replaces a global one with the same
            section and name.
          </Text>
        </Flex>
        <Flex gap={4} wrap="wrap">
          <FilterSelect
            id="filter-project"
            label="Project"
            allLabel="Global only"
            items={projects.data ?? NONE}
            value={project}
            onChange={setProject}
          />
          <FilterSelect
            id="filter-section"
            label="Section"
            allLabel="All sections"
            items={sections}
            value={section}
            onChange={setSection}
          />
        </Flex>
      </Flex>

      {resolved.data?.length === 0 && <Text variant="secondary">No contexts found.</Text>}
      {!!resolved.data?.length && (
        <div className={classes.tableBox}>
          <Table className={classes.table}>
            <Table.Header>
              <Table.Row>
                <Table.Head>Scope</Table.Head>
                <Table.Head>Name</Table.Head>
                <Table.Head>Section</Table.Head>
                <Table.Head>Title</Table.Head>
                <Table.Head>Description</Table.Head>
                <Table.Head>Updated</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {resolved.data.map((c) => (
                <Table.Row key={c.id}>
                  <Table.Cell>
                    <Badge variant={c.scope === "global" ? "accent" : "warning"} size="small">
                      {c.scope === "global" ? "global" : "project"}
                    </Badge>
                  </Table.Cell>
                  <Table.Cell>
                    <Link to={contextPath(c.id ?? "")}>{c.name}</Link>
                  </Table.Cell>
                  <Table.Cell>{c.section}</Table.Cell>
                  <Table.Cell>{c.frontmatter?.title}</Table.Cell>
                  <Table.Cell>{c.frontmatter?.description}</Table.Cell>
                  <Table.Cell>{age(c.updated_at)}</Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </div>
      )}
    </Flex>
  );
}
