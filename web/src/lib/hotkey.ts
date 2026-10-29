import { useEffect } from "react";

// useHotkey calls onKey on Cmd+key or Ctrl+key, and stops the browser default.
export function useHotkey(key: string, onKey: () => void) {
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === key) {
        e.preventDefault();
        onKey();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [key, onKey]);
}
