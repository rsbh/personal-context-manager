import { Sidebar, Spinner, Text } from "@raystack/apsara";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { searchContextsOptions } from "../client/@tanstack/react-query.gen";
import { contextPath } from "../lib/format";
import classes from "./SearchResults.module.css";

interface SearchResultsProps {
  query: string;
}

// stripTags removes the <mark> highlight from a search snippet: shown as plain text.
function stripTags(s: string | undefined): string {
  return (s ?? "").replace(/<[^>]*>/g, "");
}

export function SearchResults({ query }: SearchResultsProps) {
  const search = useQuery(searchContextsOptions({ query: { q: query } }));

  return (
    <section aria-label="Search results" className={classes.results}>
      <Text size="mini" weight="medium" variant="tertiary" className={classes.label}>
        Search results
      </Text>
      {search.isPending && <Spinner size={2} />}
      {search.isError && (
        <Text size="small" variant="danger" className={classes.label}>
          Search failed
        </Text>
      )}
      {search.data?.length === 0 && (
        <Text size="small" variant="tertiary" className={classes.label}>
          No results
        </Text>
      )}
      {search.data?.map((hit) => (
        <Sidebar.Item key={hit.id} render={<Link to={contextPath(hit.id ?? "")} />}>
          <span className={classes.hit}>
            <span>{hit.name}</span>
            <Text size="mini" variant="tertiary" lineClamp={2}>
              {hit.section} · {hit.scope === "global" ? "global" : hit.project} — {stripTags(hit.snippet)}
            </Text>
          </span>
        </Sidebar.Item>
      ))}
    </section>
  );
}
