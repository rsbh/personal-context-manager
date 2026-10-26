import { Alert, Badge, Code, Group, Loader, Paper, Stack, Text, Title, Typography } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { getContextOptions } from "../client/@tanstack/react-query.gen";
import { age, rawFile } from "../lib/format";

interface ContextViewProps {
  id: string;
}

export function ContextView({ id }: ContextViewProps) {
  const ctx = useQuery(getContextOptions({ path: { id } }));

  if (ctx.isPending) return <Loader />;
  if (ctx.isError || !ctx.data) {
    return (
      <Alert color="red" title="Context not found">
        <Code>{id}</Code>
      </Alert>
    );
  }

  const c = ctx.data;
  const scope = c.scope === "global" ? "global" : "project";
  return (
    <Stack gap="lg">
      <Stack gap={6}>
        <Text size="sm" c="dimmed" ff="monospace">
          {id.split("/").join(" / ")}
        </Text>
        <Title order={1}>{c.frontmatter?.title || c.name}</Title>
        {c.frontmatter?.description && <Text c="dimmed">{c.frontmatter.description}</Text>}
        <Group gap="xs">
          <Badge variant="light" color={scope === "global" ? "blue" : "orange"}>
            {scope}
          </Badge>
          {c.frontmatter?.tags?.map((t) => (
            <Badge key={t} variant="outline" color="gray">
              {t}
            </Badge>
          ))}
          <Text size="xs" c="dimmed">
            updated {age(c.updated_at)} ago
          </Text>
        </Group>
      </Stack>

      <Paper withBorder p="lg" radius="md">
        <Typography>
          <Markdown remarkPlugins={[remarkGfm]}>{c.content ?? ""}</Markdown>
        </Typography>
      </Paper>

      <Stack gap={6} component="section" aria-label="Raw file">
        <Text size="xs" fw={600} tt="uppercase" c="dimmed">
          File · {id}.md
        </Text>
        <Code block>{rawFile(c.frontmatter, c.content)}</Code>
      </Stack>
    </Stack>
  );
}
