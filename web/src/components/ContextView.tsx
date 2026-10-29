import { Badge, Flex, IconButton, Spinner, Text } from "@raystack/apsara";
import { useQuery } from "@tanstack/react-query";
import { Pencil, Trash2 } from "lucide-react";
import { useCallback, useState } from "react";
import { Link, useNavigate } from "react-router";
import { getContextOptions } from "../client/@tanstack/react-query.gen";
import { age, rawFile } from "../lib/format";
import classes from "./ContextView.module.css";
import { DeleteContextDialog } from "./DeleteContextDialog";
import { MarkdownBody } from "./MarkdownBody";
import { NotFound } from "./NotFound";

interface ActionsProps {
  id: string;
}

// Actions are the Edit and Delete buttons. Delete asks for a confirm first.
function Actions({ id }: ActionsProps) {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const onEdit = useCallback(() => navigate(`/edit/${id}`), [navigate, id]);
  const onDeleted = useCallback(() => navigate("/"), [navigate]);
  const show = useCallback(() => setOpen(true), []);

  return (
    <Flex gap={2} className={classes.actions}>
      <IconButton
        size={4}
        aria-label="Edit"
        title="Edit"
        className={classes.action}
        onClick={onEdit}
      >
        <Pencil size={16} />
      </IconButton>
      <IconButton
        size={4}
        aria-label="Delete"
        title="Delete"
        className={`${classes.action} ${classes.danger}`}
        onClick={show}
      >
        <Trash2 size={16} />
      </IconButton>
      <DeleteContextDialog id={id} open={open} onOpenChange={setOpen} onDeleted={onDeleted} />
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
  const overview = global ? "/global" : `/projects/${encodeURIComponent(c.project ?? "")}`;
  return (
    <Flex direction="column" gap={7}>
      <Flex direction="column" gap={4}>
        <div className={classes.head}>
          <div className={classes.titles}>
            <nav aria-label="Breadcrumb" className={classes.crumbs}>
              <Link to={overview} className={classes.crumbLink}>
                {global ? "Global" : c.project}
              </Link>
              <span aria-hidden="true">/</span>
              <span>{c.section}</span>
            </nav>
            <h1 className={classes.title}>{c.frontmatter?.title || c.name}</h1>
            {c.frontmatter?.description && (
              <Text variant="secondary" className={classes.desc}>
                {c.frontmatter.description}
              </Text>
            )}
          </div>
          <Actions id={id} />
        </div>
        <Flex align="center" gap={3} wrap="wrap">
          <Badge variant={global ? "accent" : "warning"} size="small">
            {global ? "global" : "project"}
          </Badge>
          {c.frontmatter?.tags?.map((t) => (
            <Badge key={t} variant="neutral" size="small" className={classes.tag}>
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
