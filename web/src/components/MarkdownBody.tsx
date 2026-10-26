import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import classes from "./Markdown.module.css";

const remarkPlugins = [remarkGfm];

interface MarkdownBodyProps {
  children: string;
}

// MarkdownBody renders context markdown (GitHub flavored).
export function MarkdownBody({ children }: MarkdownBodyProps) {
  return (
    <div className={classes.markdown}>
      <Markdown remarkPlugins={remarkPlugins}>{children}</Markdown>
    </div>
  );
}
