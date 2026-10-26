import { Flex, Input, Label } from "@raystack/apsara";
import { useCallback, useId, useMemo, useState } from "react";

interface TagsInputProps {
  label: string;
  value: string[];
  onChange: (tags: string[]) => void;
}

// TagsInput edits a list of tags: type and press Enter to add, × to remove.
export function TagsInput({ label, value, onChange }: TagsInputProps) {
  const id = useId();
  const [text, setText] = useState("");
  const chips = useMemo(
    () => value.map((t) => ({ label: t, onRemove: () => onChange(value.filter((v) => v !== t)) })),
    [value, onChange],
  );
  const onInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => setText(e.currentTarget.value),
    [],
  );
  const onKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key !== "Enter") return;
      e.preventDefault();
      const tag = text.trim();
      if (tag && !value.includes(tag)) onChange([...value, tag]);
      setText("");
    },
    [text, value, onChange],
  );
  return (
    <Flex direction="column" gap={2}>
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        chips={chips}
        placeholder="Add tag"
        value={text}
        onChange={onInput}
        onKeyDown={onKeyDown}
      />
    </Flex>
  );
}
