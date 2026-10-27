import {
  Button,
  Callout,
  Flex,
  Headline,
  Spinner,
  Table,
  Text,
  useToastManager,
} from "@raystack/apsara";
import { useMutation, useQueries, useQueryClient } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import { useCallback, useState } from "react";
import {
  gitLogOptions,
  gitStatusOptions,
  gitSyncMutation,
} from "../client/@tanstack/react-query.gen";
import { errorMessage, errorStatus, invalidateTree, type APIConflict } from "../lib/query";
import { renderH1 } from "../lib/render";
import { parseChange, syncSummary } from "../lib/sync";
import classes from "./SyncPage.module.css";

const syncIcon = <RefreshCw size={14} />;
// A floated legend is a normal flex item inside the card, not drawn on the fieldset border.
const legend = <legend className={classes.legend} />;

interface CardProps {
  label: string;
  value: number;
}

// Card is one status number: a fieldset (group) named by its legend.
function Card({ label, value }: CardProps) {
  return (
    <fieldset className={classes.card}>
      <Text size="small" variant="tertiary" render={legend}>
        {label}
      </Text>
      <span className={classes.cardValue}>{value}</span>
    </fieldset>
  );
}

interface ConflictProps {
  files: string[];
}

function Conflict({ files }: ConflictProps) {
  return (
    <section aria-label="Conflict">
      <Callout type="alert">
        <Flex direction="column" gap={2}>
          <Text weight="medium">Sync conflict, rebase aborted</Text>
          <Text>
            Fix{" "}
            {files.map((f, i) => (
              <span key={f}>
                {i > 0 && ", "}
                <code className={classes.mono}>{f}</code>
              </span>
            ))}{" "}
            in <code className={classes.mono}>~/.pcm/repo</code>, then sync again.
          </Text>
        </Flex>
      </Callout>
    </section>
  );
}

function NoRemote() {
  return (
    <Callout type="normal">
      <Flex direction="column" gap={2}>
        <Text weight="medium">Remote not set</Text>
        <Text>
          Contexts are local only. To sync, run{" "}
          <code className={classes.mono}>pcm config set remote &lt;git-url&gt;</code>
        </Text>
      </Flex>
    </Callout>
  );
}

// SyncPage shows the git status of the context repo and syncs it with the remote.
export function SyncPage() {
  const [status, log] = useQueries({
    queries: [gitStatusOptions(), gitLogOptions({ query: { n: 10 } })],
  });
  const [conflict, setConflict] = useState<string[] | null>(null);
  const queryClient = useQueryClient();
  const toast = useToastManager();

  const sync = useMutation({
    ...gitSyncMutation(),
    onSuccess: (r) => {
      setConflict(null);
      toast.add({ title: syncSummary(r), type: "success" });
    },
    onError: (error) => {
      if (errorStatus(error) === 409) setConflict((error as APIConflict).files ?? []);
      else toast.add({ title: errorMessage(error), type: "error" });
    },
    // A pull can change contexts, so refresh the tree too.
    onSettled: () => invalidateTree(queryClient),
  });
  const onSync = useCallback(() => sync.mutate({}), [sync]);

  if (status.isPending) return <Spinner size={3} />;
  if (status.isError) {
    return (
      <Callout type="alert">
        <Text>Git status not available: {errorMessage(status.error)}</Text>
      </Callout>
    );
  }

  const s = status.data;
  const changes = (s.changes ?? []).map(parseChange);
  const hasRemote = !!s.remote;

  return (
    <Flex direction="column" gap={7} className={classes.page}>
      <Flex align="center" gap={5} wrap="wrap">
        <Flex direction="column" gap={2}>
          <Headline size="t3" render={renderH1}>
            Sync
          </Headline>
          <Text size="small" variant="tertiary" className={classes.remote}>
            {hasRemote ? `${s.remote} · ${s.branch}` : s.branch}
          </Text>
        </Flex>
        <Button
          className={classes.syncButton}
          leadingIcon={syncIcon}
          disabled={!hasRemote}
          loading={sync.isPending}
          onClick={onSync}
        >
          Sync now
        </Button>
      </Flex>

      {!hasRemote && <NoRemote />}
      {conflict && <Conflict files={conflict} />}

      <div className={classes.cards}>
        <Card label="Changes" value={changes.length} />
        <Card label="Ahead" value={s.ahead ?? 0} />
        <Card label="Behind" value={s.behind ?? 0} />
      </div>

      <div className={classes.panels}>
        <section aria-label="Changed files" className={`${classes.panel} ${classes.files}`}>
          <Text size="small" weight="medium" variant="tertiary" className={classes.panelTitle}>
            Changed files
          </Text>
          {changes.length === 0 ? (
            <Text size="small" variant="secondary" className={classes.empty}>
              No changes
            </Text>
          ) : (
            <ul className={classes.fileList}>
              {changes.map((c) => (
                <li key={c.path} className={classes.file}>
                  <span className={classes.code}>{c.code}</span>
                  <span>{c.path}</span>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section aria-label="Recent commits" className={`${classes.panel} ${classes.commits}`}>
          <Text size="small" weight="medium" variant="tertiary" className={classes.panelTitle}>
            Recent commits
          </Text>
          {log.data?.length ? (
            <Table>
              <Table.Header>
                <Table.Row>
                  <Table.Head>Hash</Table.Head>
                  <Table.Head>Date</Table.Head>
                  <Table.Head>Message</Table.Head>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {log.data.map((c) => (
                  <Table.Row key={c.hash}>
                    <Table.Cell className={classes.mono}>{c.hash}</Table.Cell>
                    <Table.Cell className={classes.mono}>{c.date}</Table.Cell>
                    <Table.Cell>{c.message}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          ) : (
            <Text size="small" variant="secondary" className={classes.empty}>
              No commits
            </Text>
          )}
        </section>
      </div>
    </Flex>
  );
}
