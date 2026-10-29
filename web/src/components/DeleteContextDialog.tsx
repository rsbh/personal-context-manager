import { AlertDialog, Button, Flex, useToastManager } from "@raystack/apsara";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import { deleteContextMutation } from "../client/@tanstack/react-query.gen";
import { errorMessage, invalidateContexts } from "../lib/query";

interface DeleteContextDialogProps {
  id: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // onDeleted runs after a delete (for example, to leave the deleted context's page).
  onDeleted?: () => void;
}

// DeleteContextDialog asks for a confirm, then deletes the context.
export function DeleteContextDialog({
  id,
  open,
  onOpenChange,
  onDeleted,
}: DeleteContextDialogProps) {
  const queryClient = useQueryClient();
  const toast = useToastManager();
  const remove = useMutation({
    ...deleteContextMutation(),
    onSuccess: () => {
      onOpenChange(false);
      toast.add({ title: "Deleted", type: "success" });
      onDeleted?.();
      void invalidateContexts(queryClient);
    },
    onError: (error) => {
      toast.add({ title: errorMessage(error), type: "error" });
    },
  });
  const onDelete = useCallback(() => remove.mutate({ path: { id } }), [remove, id]);
  const onCancel = useCallback(() => onOpenChange(false), [onOpenChange]);

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
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
  );
}
