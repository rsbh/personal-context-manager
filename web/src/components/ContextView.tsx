import { Badge, Callout, Flex, Headline, Spinner, Text } from "@raystack/apsara";
import { useQuery } from "@tanstack/react-query";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { getContextOptions } from "../client/@tanstack/react-query.gen";
import { age, rawFile } from "../lib/format";
import classes from "./ContextView.module.css";

interface ContextViewProps {
  id: string;
}

export function ContextView({ id }: ContextViewProps) {
  const ctx = useQuery(getContextOptions({ path: { id } }));

  if (ctx.isPending) return <Spinner size={3} />;
  if (ctx.isError || !ctx.data) {
    return (
      <Callout type="alert" variant="outline">
        <Flex direction="column" gap={2}>
          <Text weight="medium">Context not found</Text>
          <Text size="small" className={classes.path}>
            {id}
          </Text>
        </Flex>
      </Callout>
    );
  }

  const c = ctx.data;
  const global = c.scope === "global";
  return (
    <Flex direction="column" gap={7}>
      <Flex direction="column" gap={3}>
        <Text size="small" variant="tertiary" className={classes.path}>
          {id.split("/").join(" / ")}
        </Text>
        <Headline size="t3" render={<h1 />}>
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
        <div className={classes.markdown}>
          <Markdown remarkPlugins={[remarkGfm]}>{c.content ?? ""}</Markdown>
        </div>
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
