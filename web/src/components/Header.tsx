import { Anchor, Badge, Group, Text } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { getHealthOptions } from "../client/@tanstack/react-query.gen";

export function Header() {
  const health = useQuery({ ...getHealthOptions(), refetchInterval: 10_000 });
  const ok = health.data?.status === "ok";

  return (
    <Group h="100%" px="md" justify="space-between">
      <Group gap="xs">
        <Badge size="lg" radius="sm" variant="filled" aria-hidden="true">
          p
        </Badge>
        <Text fw={700} size="lg">
          pcm
        </Text>
        <Text size="sm" c="dimmed" visibleFrom="sm">
          personal context manager
        </Text>
        <Anchor component={Link} to="/current" size="sm" ml="lg">
          Current contexts
        </Anchor>
      </Group>
      {!health.isPending && (
        <Badge variant="dot" color={ok ? "green" : "red"}>
          {ok ? "server ok" : "server down"}
        </Badge>
      )}
    </Group>
  );
}
