import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach, vi } from "vitest";
import { client } from "../client/client.gen";

// Node's Request needs an absolute URL; the browser resolves relative ones.
client.setConfig({ baseUrl: "http://localhost" });

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

// Mantine needs these browser APIs. jsdom does not have them.
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  }),
});
window.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
};
