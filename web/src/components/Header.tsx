import { Flex, Text } from "@raystack/apsara";
import { useQuery } from "@tanstack/react-query";
import { NavLink } from "react-router";
import { getHealthOptions } from "../client/@tanstack/react-query.gen";
import classes from "./Header.module.css";

function navClass({ isActive }: { isActive: boolean }) {
  return isActive ? `${classes.navLink} ${classes.active}` : classes.navLink;
}

export function Header() {
  const health = useQuery({ ...getHealthOptions(), refetchInterval: 10_000 });
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
      </nav>
      {!health.isPending && (
        <Flex align="center" gap={2} className={classes.status}>
          <span className={`${classes.dot} ${ok ? classes.ok : classes.down}`} aria-hidden="true" />
          <Text size="small" variant="secondary">
            {ok ? "server ok" : "server down"}
          </Text>
        </Flex>
      )}
    </header>
  );
}
