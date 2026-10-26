import { Loader, NavLink, Stack, Text } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { getApiV1SearchOptions } from "../client/@tanstack/react-query.gen";
import { contextPath } from "../lib/format";

interface SearchResultsProps {
  query: string;
}

// stripTags removes the <mark> highlight from a search snippet: shown as plain text.
function stripTags(s: string | undefined): string {
  return (s ?? "").replace(/<[^>]*>/g, "");
}

export function SearchResults({ query }: SearchResultsProps) {
  const search = useQuery(getApiV1SearchOptions({ query: { q: query } }));

  return (
    <Stack gap={2} component="section" aria-label="Search results">
      <Text size="xs" fw={600} tt="uppercase" c="dimmed" px="sm" py={6}>
        Search results
      </Text>
      {search.isPending && <Loader size="xs" mx="sm" />}
      {search.isError && (
        <Text size="sm" c="red" px="sm">
          Search failed
        </Text>
      )}
      {search.data?.length === 0 && (
        <Text size="sm" c="dimmed" px="sm">
          No results
        </Text>
      )}
      {search.data?.map((hit) => (
        <NavLink
          key={hit.id}
          component={Link}
          to={contextPath(hit.id ?? "")}
          label={hit.name}
          description={`${hit.section} · ${hit.scope === "global" ? "global" : hit.project} — ${stripTags(hit.snippet)}`}
        />
      ))}
    </Stack>
  );
}
