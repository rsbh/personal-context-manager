import { AlertDialog, Button, Flex } from "@raystack/apsara";
import { useCallback, useEffect } from "react";
import { useBlocker, type BlockerFunction } from "react-router";
import type { SAVED } from "../lib/navigation";

interface UnsavedGuardProps {
  when: boolean;
}

// UnsavedGuard asks before the user leaves a page with unsaved changes:
// a dialog for in-app navigation, the browser prompt for tab close and reload.
export function UnsavedGuard({ when }: UnsavedGuardProps) {
  const shouldBlock = useCallback<BlockerFunction>(
    ({ nextLocation }) => when && (nextLocation.state as typeof SAVED | null)?.saved !== true,
    [when],
  );
  const blocker = useBlocker(shouldBlock);

  useEffect(() => {
    if (!when) return;
    const onBeforeUnload = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [when]);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) blocker.reset?.();
    },
    [blocker],
  );
  const stay = useCallback(() => blocker.reset?.(), [blocker]);
  const leave = useCallback(() => blocker.proceed?.(), [blocker]);

  return (
    <AlertDialog open={blocker.state === "blocked"} onOpenChange={onOpenChange}>
      <AlertDialog.Content>
        <AlertDialog.Header>
          <AlertDialog.Title>Unsaved changes</AlertDialog.Title>
        </AlertDialog.Header>
        <AlertDialog.Body>
          <AlertDialog.Description>
            You have changes that are not saved. Leave and lose them?
          </AlertDialog.Description>
        </AlertDialog.Body>
        <AlertDialog.Footer>
          <Flex gap={3} justify="end">
            <Button variant="outline" color="neutral" onClick={stay}>
              Stay
            </Button>
            <Button color="danger" onClick={leave}>
              Leave
            </Button>
          </Flex>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
}
