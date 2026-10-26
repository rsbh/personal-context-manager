import { Theme } from "@raystack/apsara";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { useMemo, type ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { vi } from "vitest";

const theme = { appearance: "light", accentColor: "indigo" } as const;

interface ProvidersProps {
  path: string;
  children: ReactNode;
}

function Providers({ path, children }: ProvidersProps) {
  const client = useMemo(
    () => new QueryClient({ defaultOptions: { queries: { retry: false } } }),
    [],
  );
  const entries = useMemo(() => [path], [path]);
  return (
    <Theme defaultValue={theme}>
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={entries}>{children}</MemoryRouter>
      </QueryClientProvider>
    </Theme>
  );
}

// renderUI renders with the app providers. Queries do not retry in tests.
export function renderUI(ui: ReactNode, path = "/") {
  return render(<Providers path={path}>{ui}</Providers>);
}

// mockAPI stubs fetch: routes maps a path (with query) to a JSON body or a Response.
// The generated client calls fetch with a Request.
export function mockAPI(routes: Record<string, unknown>) {
  const fetchMock = vi.fn<typeof fetch>(async (input) => {
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
