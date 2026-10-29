import { Badge, Button, Flex, Text } from "@raystack/apsara";
import { useQueries } from "@tanstack/react-query";
import { NavLink } from "react-router";
import { getHealthOptions, gitStatusOptions } from "../client/@tanstack/react-query.gen";
import { pendingCount } from "../lib/sync";
import classes from "./Header.module.css";

function navClass({ isActive }: { isActive: boolean }) {
  return isActive ? `${classes.navLink} ${classes.active}` : classes.navLink;
}

interface HeaderProps {
  // onSearch opens the search palette.
  onSearch: () => void;
}

export function Header({ onSearch }: HeaderProps) {
  const [health, status] = useQueries({
    queries: [{ ...getHealthOptions(), refetchInterval: 10_000 }, gitStatusOptions()],
  });
  const pending = pendingCount(status.data);
  const ok = health.data?.status === "ok";

  return (
    <header className={classes.header}>
      <Flex align="center" gap={3}>
        <span className={classes.logo} aria-hidden="true">
          p
        </span>
        <Text weight="medium">pcm</Text>
        <Text size="small" variant="tertiary">
          personal context manager
        </Text>
      </Flex>
      <nav aria-label="Main" className={classes.nav}>
        <NavLink to="/" end className={navClass}>
          Browse
        </NavLink>
        <NavLink to="/current" className={navClass}>
          Current contexts
        </NavLink>
        <NavLink to="/sync" className={navClass}>
          Sync
          {pending > 0 && (
            <Badge variant="accent" size="small" className={classes.badge}>
              {pending}
            </Badge>
          )}
        </NavLink>
      </nav>
      <Button
        variant="outline"
        color="neutral"
        size="small"
        onClick={onSearch}
        className={classes.search}
      >
        Search{" "}
        <kbd aria-hidden="true" className={classes.kbd}>
          ⌘K
        </kbd>
      </Button>
      {!health.isPending && (
        <Flex align="center" gap={2}>
          <span className={`${classes.dot} ${ok ? classes.ok : classes.down}`} aria-hidden="true" />
          <Text size="small" variant="secondary">
            {ok ? "server ok" : "server down"}
          </Text>
        </Flex>
      )}
    </header>
  );
}
