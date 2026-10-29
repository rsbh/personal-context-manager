import { IconButton, useToastManager } from "@raystack/apsara";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { type ReactNode, useCallback, useState } from "react";
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
  // label replaces the default accessible name.
  label?: string;
  // icon replaces the default "+" icon.
  icon?: ReactNode;
  className?: string;
}

const plusIcon = <Plus size={16} />;

// CreateSectionButton is a "+" that creates a section, global or in a project.
export function CreateSectionButton({
  project,
  onCreated,
  label,
  icon = plusIcon,
  className,
}: CreateSectionButtonProps) {
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
        size={4}
        aria-label={label ?? (project ? `New section in ${project}` : "New global section")}
        title={label}
        className={className}
        onClick={show}
      >
        {icon}
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

// useCreateProject creates a project from a name dialog. show opens the dialog; onCreated
// gets the new project name.
export function useCreateProject(onCreated?: (name: string) => void) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const close = useCallback(() => setOpen(false), []);
  const show = useCallback(() => setOpen(true), []);
  const after = useCallback(() => onCreated?.(name), [onCreated, name]);
  const create = useMutation({
    ...createProjectMutation(),
    ...useCreated("Project created", close, after),
  });
  const onSubmit = useCallback(
    (n: string) => {
      setName(n);
      create.mutate({ body: { name: n } });
    },
    [create],
  );
  const dialog = open && (
    <NameDialog
      title="New project"
      open
      onOpenChange={setOpen}
      pending={create.isPending}
      onSubmit={onSubmit}
    />
  );
  return { show, dialog };
}

// idPrefix is the start of the ids of all contexts in a section or project.
function idPrefix(kind: DeleteKind, name: string, project?: string): string {
  if (kind === "project") return `projects/${name}/`;
  return project ? `projects/${project}/${name}/` : `global/${name}/`;
}

// insideDeleted is true when the current page shows the deleted project or a context in the
// deleted section or project.
function insideDeleted(pathname: string, prefix: string): boolean {
  return (
    `${pathname}/` === `/${prefix}` ||
    pathname.startsWith(`/contexts/${prefix}`) ||
    pathname.startsWith(`/edit/${prefix}`)
  );
}

interface ScopeDeleteProps {
  kind: DeleteKind;
  name: string;
  // project of a project section.
  project?: string;
}

// useScopeDelete deletes a section or project after a confirm. On 409 (not empty) the dialog
// asks again and "Delete all" sends force=true. show opens the dialog.
export function useScopeDelete({ kind, name, project }: ScopeDeleteProps) {
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

  const dialog = (
    <DeleteDialog
      kind={kind}
      name={project ? `${project}/${name}` : name}
      open={open}
      onOpenChange={setOpen}
      notEmpty={notEmpty}
      pending={pending}
      onDelete={onDelete}
    />
  );
  return { show, dialog };
}
