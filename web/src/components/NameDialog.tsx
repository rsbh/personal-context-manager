import { Button, Dialog, Flex, Input, Label } from "@raystack/apsara";
import { useCallback, useId, useState } from "react";

const cancelButton = <Button type="button" variant="outline" color="neutral" />;

interface NameDialogProps {
  title: string;
  description?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  pending: boolean;
  onSubmit: (name: string) => void;
}

// NameDialog asks for one name (a new section or project). The parent closes it on success.
export function NameDialog({
  title,
  description,
  open,
  onOpenChange,
  pending,
  onSubmit,
}: NameDialogProps) {
  const id = useId();
  const [name, setName] = useState("");
  const trimmed = name.trim();
  const onInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => setName(e.currentTarget.value),
    [],
  );
  const submit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      if (trimmed) onSubmit(trimmed);
    },
    [trimmed, onSubmit],
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content>
        <form onSubmit={submit}>
          <Dialog.Header>
            <Dialog.Title>{title}</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={4}>
              {description && <Dialog.Description>{description}</Dialog.Description>}
              <Flex direction="column" gap={2}>
                <Label htmlFor={id}>Name</Label>
                <Input id={id} value={name} onChange={onInput} />
              </Flex>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Flex gap={3} justify="end">
              <Dialog.Close render={cancelButton}>Cancel</Dialog.Close>
              <Button type="submit" disabled={!trimmed || pending} loading={pending}>
                Create
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
