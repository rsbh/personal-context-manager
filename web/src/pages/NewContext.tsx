import { Flex, Label, Select, useToastManager } from "@raystack/apsara";
import { useMutation, useQueries, useQueryClient } from "@tanstack/react-query";
import { useCallback, useId, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import {
  createContextMutation,
  listProjectsOptions,
  listSectionsOptions,
} from "../client/@tanstack/react-query.gen";
import type { ServerCreateContextRequest } from "../client/types.gen";
import { ContextForm, TextField } from "../components/ContextForm";
import { emptyDraft, sameDraft, toFrontmatter } from "../lib/draft";
import { contextPath } from "../lib/format";
import { SAVED } from "../lib/navigation";
import { errorMessage, invalidateContexts } from "../lib/query";

// GLOBAL is the scope select value for global. Project names cannot start with "_".
const GLOBAL = "_global";

const NONE: string[] = [];

// pair is a select item whose label is its value.
function pair(v: string): [string, string] {
  return [v, v];
}

interface PickerProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  items: [value: string, label: string][];
}

function Picker({ label, value, onChange, items }: PickerProps) {
  const id = useId();
  return (
    <Flex direction="column" gap={2}>
      <Label htmlFor={id}>{label}</Label>
      <Select value={value} onValueChange={onChange}>
        <Select.Trigger id={id} aria-label={label}>
          <Select.Value placeholder={`Select ${label.toLowerCase()}`} />
        </Select.Trigger>
        <Select.Content>
          {items.map(([v, l]) => (
            <Select.Item key={v} value={v}>
              {l}
            </Select.Item>
          ))}
        </Select.Content>
      </Select>
    </Flex>
  );
}

// NewContext creates a context at /new?scope=&project=&section=.
export function NewContext() {
  const [params] = useSearchParams();
  const [scope, setScope] = useState(() =>
    params.get("scope") === "project" ? (params.get("project") ?? "") : GLOBAL,
  );
  const [section, setSection] = useState(() => params.get("section") ?? "");
  const [name, setName] = useState("");
  const [draft, setDraft] = useState(emptyDraft);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastManager();

  const isGlobal = scope === GLOBAL;
  const [projects, sections] = useQueries({
    queries: [
      listProjectsOptions(),
      {
        ...listSectionsOptions({
          query: isGlobal ? { scope: "global" } : { scope: "project", project: scope },
        }),
        enabled: scope !== "",
      },
    ],
  });

  const create = useMutation({
    ...createContextMutation(),
    onSuccess: (c) => {
      toast.add({ title: "Saved", type: "success" });
      navigate(contextPath(c.id ?? ""), { state: SAVED });
      void invalidateContexts(queryClient);
    },
    onError: (error) => {
      toast.add({ title: errorMessage(error), type: "error" });
    },
  });

  const onScope = useCallback((v: string) => {
    setScope(v);
    setSection("");
  }, []);
  const onSave = useCallback(() => {
    const body: ServerCreateContextRequest = isGlobal
      ? {
          scope: "global",
          section,
          name,
          frontmatter: toFrontmatter(draft),
          content: draft.content,
        }
      : {
          scope: "project",
          project: scope,
          section,
          name,
          frontmatter: toFrontmatter(draft),
          content: draft.content,
        };
    create.mutate({ body });
  }, [isGlobal, scope, section, name, draft, create]);
  const onDiscard = useCallback(() => {
    setName("");
    setDraft(emptyDraft);
  }, []);

  const scopeItems = useMemo(
    (): [string, string][] => [[GLOBAL, "Global"], ...(projects.data ?? NONE).map(pair)],
    [projects.data],
  );
  const sectionItems = useMemo(() => (sections.data ?? NONE).map(pair), [sections.data]);
  const dirty = name !== "" || !sameDraft(draft, emptyDraft);

  return (
    <ContextForm
      heading="New context"
      draft={draft}
      onChange={setDraft}
      dirty={dirty}
      canSave={scope !== "" && section !== "" && name.trim() !== "" && !create.isPending}
      saving={create.isPending}
      onSave={onSave}
      onDiscard={onDiscard}
      fields={
        <>
          <Picker label="Scope" value={scope} onChange={onScope} items={scopeItems} />
          <Picker label="Section" value={section} onChange={setSection} items={sectionItems} />
          <TextField label="Name" value={name} onChange={setName} />
        </>
      }
    />
  );
}
