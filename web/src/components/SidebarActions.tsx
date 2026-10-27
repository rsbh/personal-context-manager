import { IconButton, Menu, useToastManager } from "@raystack/apsara";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Ellipsis, Plus } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router";
import {
  createProjectMutation,
  createSectionMutation,
  deleteProjectMutation,
  deleteSectionMutation,
} from "../client/@tanstack/react-query.gen";
import { errorMessage, errorStatus, invalidateTree } from "../lib/query";
import { DeleteDialog, type DeleteKind } from "./DeleteDialog";
import { NameDialog } from "./NameDialog";

// useCreated closes the dialog, shows a toast and refreshes the tree after a create.
function useCreated(title: string, close: () => void, after?: () => void) {
  const queryClient = useQueryClient();
  const toast = useToastManager();
  return {
    onSuccess: () => {
      close();
      toast.add({ title, type: "success" as const });
      after?.();
      void invalidateTree(queryClient);
    },
    onError: (error: unknown) => {
      toast.add({ title: errorMessage(error), type: "error" as const });
    },
  };
}

interface CreateSectionButtonProps {
  // project is empty for a global section.
  project?: string;
  // onCreated runs after a create (for example, to open the project row).
  onCreated?: () => void;
}

// CreateSectionButton is a "+" that creates a section, global or in a project.
export function CreateSectionButton({ project, onCreated }: CreateSectionButtonProps) {
  const [open, setOpen] = useState(false);
  const close = useCallback(() => setOpen(false), []);
  const show = useCallback(() => setOpen(true), []);
  const create = useMutation({
    ...createSectionMutation(),
    ...useCreated("Section created", close, onCreated),
  });
  const onSubmit = useCallback(
    (name: string) =>
      create.mutate({
        body: project ? { scope: "project", project, name } : { scope: "global", name },
      }),
    [create, project],
  );
  return (
    <>
      <IconButton
        size={1}
        aria-label={project ? `New section in ${project}` : "New global section"}
        onClick={show}
      >
        <Plus size={12} />
      </IconButton>
      {open && (
        <NameDialog
          title="New section"
          description={project ? `In project ${project}` : "Global: applies to all projects"}
          open
          onOpenChange={setOpen}
          pending={create.isPending}
          onSubmit={onSubmit}
        />
      )}
    </>
  );
}

// CreateProjectButton is a "+" that creates a project.
export function CreateProjectButton() {
  const [open, setOpen] = useState(false);
  const close = useCallback(() => setOpen(false), []);
  const show = useCallback(() => setOpen(true), []);
  const create = useMutation({
    ...createProjectMutation(),
    ...useCreated("Project created", close),
  });
  const onSubmit = useCallback((name: string) => create.mutate({ body: { name } }), [create]);
  return (
    <>
      <IconButton size={1} aria-label="New project" onClick={show}>
        <Plus size={12} />
      </IconButton>
      {open && (
        <NameDialog
          title="New project"
          open
          onOpenChange={setOpen}
          pending={create.isPending}
          onSubmit={onSubmit}
        />
      )}
    </>
  );
}

// idPrefix is the start of the ids of all contexts in a section or project.
function idPrefix(kind: DeleteKind, name: string, project?: string): string {
  if (kind === "project") return `projects/${name}/`;
  return project ? `projects/${project}/${name}/` : `global/${name}/`;
}

// insideDeleted is true when the current page shows a context in the deleted section or project.
function insideDeleted(pathname: string, prefix: string): boolean {
  return pathname.startsWith(`/contexts/${prefix}`) || pathname.startsWith(`/edit/${prefix}`);
}

interface RowMenuProps {
  kind: DeleteKind;
  name: string;
  // project of a project section.
  project?: string;
}

// RowMenu is the "⋯" menu of a section or project row, with Delete.
export function RowMenu({ kind, name, project }: RowMenuProps) {
  const [open, setOpen] = useState(false);
  const [notEmpty, setNotEmpty] = useState(false);
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastManager();

  const onSuccess = () => {
    setOpen(false);
    toast.add({ title: "Deleted", type: "success" });
    if (insideDeleted(pathname, idPrefix(kind, name, project))) navigate("/");
    void invalidateTree(queryClient);
  };
  const onError = (error: unknown) => {
    if (errorStatus(error) === 409) setNotEmpty(true);
    else toast.add({ title: errorMessage(error), type: "error" });
  };
  const removeSection = useMutation({ ...deleteSectionMutation(), onSuccess, onError });
  const removeProject = useMutation({ ...deleteProjectMutation(), onSuccess, onError });
  const pending = removeSection.isPending || removeProject.isPending;

  const onDelete = useCallback(
    (force: boolean) => {
      const f = force ? { force: true } : {};
      if (kind === "project") {
        removeProject.mutate({ path: { project: name }, query: f });
        return;
      }
      const scope = project ? { scope: "project", project } : { scope: "global" };
      removeSection.mutate({ path: { section: name }, query: { ...scope, ...f } });
    },
    [kind, name, project, removeProject, removeSection],
  );
  const show = useCallback(() => {
    setNotEmpty(false);
    setOpen(true);
  }, []);

  const label =
    kind === "project"
      ? `Project ${name} actions`
      : `Section ${project ? `${project}/${name}` : name} actions`;
  const trigger = useMemo(() => <IconButton size={1} aria-label={label} />, [label]);
  return (
    <>
      <Menu>
        <Menu.Trigger render={trigger}>
          <Ellipsis size={12} />
        </Menu.Trigger>
        <Menu.Content>
          <Menu.Item onClick={show}>Delete</Menu.Item>
        </Menu.Content>
      </Menu>
      <DeleteDialog
        kind={kind}
        name={project ? `${project}/${name}` : name}
        open={open}
        onOpenChange={setOpen}
        notEmpty={notEmpty}
        pending={pending}
        onDelete={onDelete}
      />
    </>
  );
}
