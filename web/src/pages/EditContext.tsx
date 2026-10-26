import { Spinner, useToastManager } from "@raystack/apsara";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { getContextOptions, updateContextMutation } from "../client/@tanstack/react-query.gen";
import type { StoreContext } from "../client/types.gen";
import { ContextForm } from "../components/ContextForm";
import { NotFound } from "../components/NotFound";
import { fromContext, sameDraft, toFrontmatter } from "../lib/draft";
import { contextPath } from "../lib/format";
import { SAVED } from "../lib/navigation";
import { errorMessage, invalidateContexts } from "../lib/query";

interface EditFormProps {
  context: StoreContext;
}

function EditForm({ context }: EditFormProps) {
  const id = context.id ?? "";
  const [initial] = useState(() => fromContext(context));
  const [draft, setDraft] = useState(initial);
  const dirty = !sameDraft(draft, initial);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastManager();

  const save = useMutation({
    ...updateContextMutation(),
    onSuccess: () => {
      toast.add({ title: "Saved", type: "success" });
      navigate(contextPath(id), { state: SAVED });
      void invalidateContexts(queryClient);
    },
    onError: (error) => {
      toast.add({ title: errorMessage(error), type: "error" });
    },
  });
  const onSave = useCallback(
    () =>
      save.mutate({
        path: { id },
        body: { frontmatter: toFrontmatter(draft), content: draft.content },
      }),
    [save, id, draft],
  );
  const onDiscard = useCallback(() => setDraft(initial), [initial]);

  return (
    <ContextForm
      heading="Edit context"
      path={id.split("/").join(" / ")}
      draft={draft}
      onChange={setDraft}
      dirty={dirty}
      canSave={dirty && !save.isPending}
      saving={save.isPending}
      onSave={onSave}
      onDiscard={onDiscard}
    />
  );
}

// EditContext edits the context at /edit/<id>.
export function EditContext() {
  const id = useParams()["*"] ?? "";
  const ctx = useQuery(getContextOptions({ path: { id } }));
  if (ctx.isPending) return <Spinner size={3} />;
  if (ctx.isError || !ctx.data) return <NotFound id={id} />;
  return <EditForm key={id} context={ctx.data} />;
}
