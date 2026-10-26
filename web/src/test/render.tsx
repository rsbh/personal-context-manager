import { Theme } from "@raystack/apsara";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { vi } from "vitest";

// renderUI renders with the app providers. Queries do not retry in tests.
export function renderUI(ui: ReactNode, path = "/") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <Theme defaultValue={{ appearance: "light", accentColor: "indigo" }}>
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={[path]}>{ui}</MemoryRouter>
      </QueryClientProvider>
    </Theme>,
  );
}

// mockAPI stubs fetch: routes maps a path (with query) to a JSON body or a Response.
// The generated client calls fetch with a Request.
export function mockAPI(routes: Record<string, unknown>) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const raw = input instanceof Request ? input.url : input.toString();
    const parsed = new URL(raw, "http://localhost");
    const url = parsed.pathname + parsed.search;
    if (!(url in routes)) {
      return new Response(JSON.stringify({ error: "not found" }), { status: 404 });
    }
    const body = routes[url];
    if (body instanceof Response) return body;
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}
