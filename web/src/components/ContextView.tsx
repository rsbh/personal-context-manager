import {
  AlertDialog,
  Badge,
  Button,
  Flex,
  Headline,
  Spinner,
  Text,
  useToastManager,
} from "@raystack/apsara";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Trash2 } from "lucide-react";
import { useCallback, useState } from "react";
import { useNavigate } from "react-router";
import { deleteContextMutation, getContextOptions } from "../client/@tanstack/react-query.gen";
import { age, rawFile } from "../lib/format";
import { errorMessage, invalidateContexts } from "../lib/query";
import { renderH1 } from "../lib/render";
import classes from "./ContextView.module.css";
import { MarkdownBody } from "./MarkdownBody";
import { NotFound } from "./NotFound";

interface ActionsProps {
  id: string;
}

const editIcon = <Pencil size={14} />;
const deleteIcon = <Trash2 size={14} />;
const deleteTrigger = <Button variant="outline" color="danger" leadingIcon={deleteIcon} />;

// Actions are the Edit and Delete buttons. Delete asks for a confirm first.
function Actions({ id }: ActionsProps) {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastManager();
  const remove = useMutation({
    ...deleteContextMutation(),
    onSuccess: () => {
      setOpen(false);
      toast.add({ title: "Deleted", type: "success" });
      navigate("/");
      void invalidateContexts(queryClient);
    },
    onError: (error) => {
      toast.add({ title: errorMessage(error), type: "error" });
    },
  });
  const onEdit = useCallback(() => navigate(`/edit/${id}`), [navigate, id]);
  const onDelete = useCallback(() => remove.mutate({ path: { id } }), [remove, id]);
  const onCancel = useCallback(() => setOpen(false), []);

  return (
    <Flex gap={3} className={classes.actions}>
      <Button variant="outline" color="neutral" leadingIcon={editIcon} onClick={onEdit}>
        Edit
      </Button>
      <AlertDialog open={open} onOpenChange={setOpen}>
        <AlertDialog.Trigger render={deleteTrigger}>Delete</AlertDialog.Trigger>
        <AlertDialog.Content>
          <AlertDialog.Header>
            <AlertDialog.Title>Delete context?</AlertDialog.Title>
          </AlertDialog.Header>
          <AlertDialog.Body>
            <AlertDialog.Description>
              This deletes <code>{id}</code>. Sync to remove it from the remote too.
            </AlertDialog.Description>
          </AlertDialog.Body>
          <AlertDialog.Footer>
            <Flex gap={3} justify="end">
              <Button variant="outline" color="neutral" onClick={onCancel}>
                Cancel
              </Button>
              <Button color="danger" loading={remove.isPending} onClick={onDelete}>
                Delete
              </Button>
            </Flex>
          </AlertDialog.Footer>
        </AlertDialog.Content>
      </AlertDialog>
    </Flex>
  );
}

interface ContextViewProps {
  id: string;
}

export function ContextView({ id }: ContextViewProps) {
  const ctx = useQuery(getContextOptions({ path: { id } }));

  if (ctx.isPending) return <Spinner size={3} />;
  if (ctx.isError || !ctx.data) return <NotFound id={id} />;

  const c = ctx.data;
  const global = c.scope === "global";
  return (
    <Flex direction="column" gap={7}>
      <Flex direction="column" gap={3}>
        <Flex align="center" gap={3} wrap="wrap">
          <Text size="small" variant="tertiary" className={classes.path}>
            {id.split("/").join(" / ")}
          </Text>
          <Actions id={id} />
        </Flex>
        <Headline size="t3" render={renderH1}>
          {c.frontmatter?.title || c.name}
        </Headline>
        {c.frontmatter?.description && <Text variant="secondary">{c.frontmatter.description}</Text>}
        <Flex align="center" gap={3} wrap="wrap">
          <Badge variant={global ? "accent" : "warning"} size="small">
            {global ? "global" : "project"}
          </Badge>
          {c.frontmatter?.tags?.map((t) => (
            <Badge key={t} variant="neutral" size="small">
              {t}
            </Badge>
          ))}
          <Text size="small" variant="tertiary">
            updated {age(c.updated_at)} ago
          </Text>
        </Flex>
      </Flex>

      <div className={classes.panel}>
        <MarkdownBody>{c.content ?? ""}</MarkdownBody>
      </div>

      <section aria-label="Raw file">
        <Flex direction="column" gap={3}>
          <Text size="small" weight="medium" variant="tertiary">
            File · {id}.md
          </Text>
          <pre className={classes.raw}>{rawFile(c.frontmatter, c.content)}</pre>
        </Flex>
      </section>
    </Flex>
  );
}
