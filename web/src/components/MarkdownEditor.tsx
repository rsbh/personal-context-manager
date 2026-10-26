import { markdown } from "@codemirror/lang-markdown";
import { EditorView } from "@codemirror/view";
import CodeMirror from "@uiw/react-codemirror";

// The theme uses Apsara tokens, so it follows light and dark mode.
const theme = EditorView.theme({
  "&": {
    height: "100%",
    color: "var(--rs-color-foreground-base-primary)",
    backgroundColor: "var(--rs-color-background-base-primary)",
    fontSize: "var(--rs-font-size-mono-regular, 14px)",
  },
  ".cm-content": { fontFamily: "var(--rs-font-mono)", caretColor: "currentColor" },
  ".cm-scroller": { fontFamily: "var(--rs-font-mono)" },
  ".cm-gutters": {
    color: "var(--rs-color-foreground-base-tertiary)",
    backgroundColor: "var(--rs-color-background-base-primary)",
    borderRight: "0.5px solid var(--rs-color-border-base-primary)",
  },
  ".cm-activeLine, .cm-activeLineGutter": { backgroundColor: "transparent" },
  "&.cm-focused": { outline: "none" },
});

const extensions = [
  markdown(),
  EditorView.lineWrapping,
  EditorView.contentAttributes.of({ "aria-label": "Markdown" }),
  theme,
];

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
}

// MarkdownEditor is a CodeMirror editor for markdown. Its textbox is named "Markdown".
export function MarkdownEditor({ value, onChange }: MarkdownEditorProps) {
  return (
    <CodeMirror
      value={value}
      onChange={onChange}
      extensions={extensions}
      theme="none"
      height="100%"
      minHeight="420px"
    />
  );
}
