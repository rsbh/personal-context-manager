import { AlertDialog, Button, Flex } from "@raystack/apsara";
import { useCallback } from "react";

export type DeleteKind = "section" | "project";

const labels: Record<DeleteKind, { title: string; notEmpty: string; contents: string }> = {
  section: { title: "Section", notEmpty: "Section is not empty", contents: "contexts" },
  project: {
    title: "Project",
    notEmpty: "Project is not empty",
    contents: "sections and contexts",
  },
};

interface DeleteDialogProps {
  kind: DeleteKind;
  name: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // notEmpty is true after the server said the section or project is not empty (409).
  notEmpty: boolean;
  pending: boolean;
  onDelete: (force: boolean) => void;
}

// DeleteDialog confirms a section or project delete. When it is not empty, it asks again
// and "Delete all" deletes with force.
export function DeleteDialog({
  kind,
  name,
  open,
  onOpenChange,
  notEmpty,
  pending,
  onDelete,
}: DeleteDialogProps) {
  const l = labels[kind];
  const cancel = useCallback(() => onOpenChange(false), [onOpenChange]);
  const remove = useCallback(() => onDelete(notEmpty), [onDelete, notEmpty]);

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialog.Content>
        <AlertDialog.Header>
          <AlertDialog.Title>{notEmpty ? l.notEmpty : `Delete ${kind}?`}</AlertDialog.Title>
        </AlertDialog.Header>
        <AlertDialog.Body>
          <AlertDialog.Description>
            {notEmpty ? (
              <>
                <code>{name}</code> has {l.contents}. Delete it with all its {l.contents}?
              </>
            ) : (
              <>
                {l.title} <code>{name}</code> will be deleted.
              </>
            )}
          </AlertDialog.Description>
        </AlertDialog.Body>
        <AlertDialog.Footer>
          <Flex gap={3} justify="end">
            <Button variant="outline" color="neutral" onClick={cancel}>
              Cancel
            </Button>
            <Button color="danger" loading={pending} onClick={remove}>
              {notEmpty ? "Delete all" : "Delete"}
            </Button>
          </Flex>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
}
