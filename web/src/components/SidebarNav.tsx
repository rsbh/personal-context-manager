import { ChevronRight } from "lucide-react";
import { useCallback, useState } from "react";
import { Link } from "react-router";
import classes from "./SidebarNav.module.css";

interface GroupProps {
  label: string;
  // collapsible adds a toggle button on the heading.
  collapsible?: boolean;
  defaultOpen?: boolean;
  // open and onOpenChange control the group from the parent.
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  // trailing is shown at the end of the heading, outside the toggle button.
  trailing?: React.ReactNode;
  children?: React.ReactNode;
}

// Group is a heading with children. Rows share one height, gap and indent.
export function Group({
  label,
  collapsible,
  defaultOpen = true,
  open,
  onOpenChange,
  trailing,
  children,
}: GroupProps) {
  const [inner, setInner] = useState(defaultOpen);
  const shown = !collapsible || (open ?? inner);
  const toggle = useCallback(() => {
    setInner(!shown);
    onOpenChange?.(!shown);
  }, [shown, onOpenChange]);

  return (
    <li aria-label={label} className={classes.group}>
      <div className={classes.heading}>
        {collapsible ? (
          <button type="button" aria-expanded={shown} className={classes.toggle} onClick={toggle}>
            <ChevronRight size={16} className={shown ? classes.chevronOpen : classes.chevron} />
            <span className={classes.label}>{label}</span>
          </button>
        ) : (
          <span className={classes.title}>{label}</span>
        )}
        {trailing && <span className={classes.actions}>{trailing}</span>}
      </div>
      {shown && <ul className={classes.children}>{children}</ul>}
    </li>
  );
}

interface ItemProps {
  to: string;
  active?: boolean;
  children: React.ReactNode;
}

// Row is a plain list row for text that is not a link (empty and loading states).
export function Row({ children }: { children: React.ReactNode }) {
  return <li className={classes.row}>{children}</li>;
}

// Item is a link row. The current page has aria-current="page".
export function Item({ to, active, children }: ItemProps) {
  return (
    <li>
      <Link to={to} aria-current={active ? "page" : undefined} className={classes.item}>
        {children}
      </Link>
    </li>
  );
}
