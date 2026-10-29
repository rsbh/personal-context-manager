import { Command, Spinner, Text } from "@raystack/apsara";
import { useQuery } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import { useNavigate } from "react-router";
import { searchContextsOptions } from "../client/@tanstack/react-query.gen";
import type { SearchHit } from "../client/types.gen";
import { useDebouncedValue } from "../lib/debounce";
import { contextPath } from "../lib/format";
import classes from "./SearchPalette.module.css";

const noHits: SearchHit[] = [];

// stripTags removes the <mark> highlight from a search snippet: shown as plain text.
function stripTags(s: string | undefined): string {
  return (s ?? "").replace(/<[^>]*>/g, "");
}

interface PaletteBodyProps {
  onClose: () => void;
}

// PaletteBody holds the query, so it starts empty each time the palette opens.
function PaletteBody({ onClose }: PaletteBodyProps) {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const debounced = useDebouncedValue(query.trim(), 250);
  const search = useQuery({
    ...searchContextsOptions({ query: { q: debounced } }),
    enabled: debounced !== "",
  });
  const hits = search.data ?? noHits;

  const onValueChange = useCallback((value: string) => setQuery(value), []);
  const open = useCallback(
    (hit: SearchHit) => {
      navigate(contextPath(hit.id ?? ""));
      onClose();
    },
    [navigate, onClose],
  );

  return (
    <Command items={hits} filter={null} value={query} onValueChange={onValueChange}>
      <Command.Input aria-label="Search contexts" placeholder="Search contexts" />
      <Command.Content>
        {search.isFetching && (
          <div className={classes.note}>
            <Spinner size={2} />
          </div>
        )}
        {search.isError && (
          <Text size="small" variant="danger" className={classes.note}>
            Search failed
          </Text>
        )}
        {search.isSuccess && hits.length === 0 && (
          <Text size="small" variant="tertiary" className={classes.note}>
            No results
          </Text>
        )}
        {hits.map((hit) => (
          <PaletteHit key={hit.id} hit={hit} onOpen={open} />
        ))}
      </Command.Content>
    </Command>
  );
}

interface PaletteHitProps {
  hit: SearchHit;
  onOpen: (hit: SearchHit) => void;
}

function PaletteHit({ hit, onOpen }: PaletteHitProps) {
  const onClick = useCallback(() => onOpen(hit), [hit, onOpen]);
  return (
    <Command.Item value={hit.id} onClick={onClick}>
      <span className={classes.hit}>
        <span>{hit.name}</span>
        <Text size="mini" variant="tertiary" lineClamp={2}>
          {hit.section} · {hit.scope === "global" ? "global" : hit.project} —{" "}
          {stripTags(hit.snippet)}
        </Text>
      </span>
    </Command.Item>
  );
}

interface SearchPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// SearchPalette is the Cmd+K dialog: full-text search of contexts.
export function SearchPalette({ open, onOpenChange }: SearchPaletteProps) {
  const onClose = useCallback(() => onOpenChange(false), [onOpenChange]);
  return (
    <Command.Dialog open={open} onOpenChange={onOpenChange}>
      <Command.DialogContent aria-label="Search contexts">
        <PaletteBody onClose={onClose} />
      </Command.DialogContent>
    </Command.Dialog>
  );
}
