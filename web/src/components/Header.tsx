import { Badge, Group, Text } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { api } from "../api";

export function Header() {
  const health = useQuery({ queryKey: ["health"], queryFn: api.health, refetchInterval: 10_000 });
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
      </Group>
      {!health.isPending && (
        <Badge variant="dot" color={ok ? "green" : "red"}>
          {ok ? "server ok" : "server down"}
        </Badge>
      )}
    </Group>
  );
}
