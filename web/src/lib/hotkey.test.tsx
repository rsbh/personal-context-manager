import { renderHook } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { useHotkey } from "./hotkey";

describe("useHotkey", () => {
  it("calls the handler on Ctrl+K", async () => {
    const onKey = vi.fn<() => void>();
    renderHook(() => useHotkey("k", onKey));
    await userEvent.keyboard("{Control>}k{/Control}");
    expect(onKey).toHaveBeenCalledTimes(1);
  });

  it("calls the handler on Cmd+K", async () => {
    const onKey = vi.fn<() => void>();
    renderHook(() => useHotkey("k", onKey));
    await userEvent.keyboard("{Meta>}k{/Meta}");
    expect(onKey).toHaveBeenCalledTimes(1);
  });

  it("ignores k without a modifier and other keys", async () => {
    const onKey = vi.fn<() => void>();
    renderHook(() => useHotkey("k", onKey));
    await userEvent.keyboard("k");
    await userEvent.keyboard("{Control>}j{/Control}");
    expect(onKey).not.toHaveBeenCalled();
  });

  it("stops after unmount", async () => {
    const onKey = vi.fn<() => void>();
    const { unmount } = renderHook(() => useHotkey("k", onKey));
    unmount();
    await userEvent.keyboard("{Control>}k{/Control}");
    expect(onKey).not.toHaveBeenCalled();
  });
});
