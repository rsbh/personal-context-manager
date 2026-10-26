import { Badge, Button, Flex, Headline, Input, Label, Spinner, Text } from "@raystack/apsara";
import { lazy, Suspense, useCallback, useId, type ReactNode } from "react";
import type { Draft } from "../lib/draft";
import { renderH1 } from "../lib/render";
import classes from "./ContextForm.module.css";
import { MarkdownBody } from "./MarkdownBody";
import { TagsInput } from "./TagsInput";
import { UnsavedGuard } from "./UnsavedGuard";

// CodeMirror is large, so the editor loads only on the edit pages.
const MarkdownEditor = lazy(() =>
  import("./MarkdownEditor").then((m) => ({ default: m.MarkdownEditor })),
);

const editorFallback = <Spinner size={3} />;

interface TextFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
}

// TextField is a labelled Apsara Input.
export function TextField({ label, value, onChange }: TextFieldProps) {
  const id = useId();
  const onInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => onChange(e.currentTarget.value),
    [onChange],
  );
  return (
    <Flex direction="column" gap={2}>
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} value={value} onChange={onInput} />
    </Flex>
  );
}

interface ContextFormProps {
  heading: string;
  path?: string;
  draft: Draft;
  onChange: (draft: Draft) => void;
  dirty: boolean;
  canSave: boolean;
  saving: boolean;
  onSave: () => void;
  onDiscard: () => void;
  // fields are shown before the frontmatter fields (scope, section, name for a new context).
  fields?: ReactNode;
}

// ContextForm edits a context: frontmatter fields, markdown editor and live preview.
export function ContextForm({
  heading,
  path,
  draft,
  onChange,
  dirty,
  canSave,
  saving,
  onSave,
  onDiscard,
  fields,
}: ContextFormProps) {
  const set = useCallback(
    <K extends keyof Draft>(key: K) =>
      (value: Draft[K]) =>
        onChange({ ...draft, [key]: value }),
    [draft, onChange],
  );
  const onSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      onSave();
    },
    [onSave],
  );

  return (
    <form onSubmit={onSubmit} className={classes.form}>
      <UnsavedGuard when={dirty} />
      <Flex direction="column" gap={7}>
        <Flex align="end" gap={5} wrap="wrap" className={classes.toolbar}>
          <Flex direction="column" gap={2}>
            {path && (
              <Text size="small" variant="tertiary" className={classes.path}>
                {path}
              </Text>
            )}
            <Headline size="t3" render={renderH1}>
              {heading}
            </Headline>
          </Flex>
          <Flex align="center" gap={3} className={classes.actions}>
            {dirty && (
              <Badge variant="warning" size="small">
                Unsaved changes
              </Badge>
            )}
            <Button
              type="button"
              variant="outline"
              color="neutral"
              disabled={!dirty || saving}
              onClick={onDiscard}
            >
              Discard
            </Button>
            <Button type="submit" disabled={!canSave} loading={saving}>
              Save
            </Button>
          </Flex>
        </Flex>

        <section aria-label="Frontmatter" className={classes.fields}>
          {fields}
          <TextField label="Title" value={draft.title} onChange={set("title")} />
          <TextField label="Description" value={draft.description} onChange={set("description")} />
          <TagsInput label="Tags" value={draft.tags} onChange={set("tags")} />
        </section>

        <div className={classes.panes}>
          <section aria-label="Markdown editor" className={classes.pane}>
            <Text size="small" weight="medium" variant="tertiary" className={classes.paneTitle}>
              Markdown
            </Text>
            <div className={classes.paneBody}>
              <Suspense fallback={editorFallback}>
                <MarkdownEditor value={draft.content} onChange={set("content")} />
              </Suspense>
            </div>
          </section>
          <section aria-label="Preview" className={classes.pane}>
            <Text size="small" weight="medium" variant="tertiary" className={classes.paneTitle}>
              Preview
            </Text>
            <div className={classes.preview}>
              <MarkdownBody>{draft.content}</MarkdownBody>
            </div>
          </section>
        </div>
      </Flex>
    </form>
  );
}
