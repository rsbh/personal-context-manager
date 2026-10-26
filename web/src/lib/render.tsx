import type { ComponentProps } from "react";

// renderH1 is a render prop for Apsara components (e.g. Headline) that renders an <h1>.
export function renderH1({ children, ...props }: ComponentProps<"h1">) {
  return <h1 {...props}>{children}</h1>;
}
