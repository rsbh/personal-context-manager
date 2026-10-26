import { Theme, Toast } from "@raystack/apsara";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { useMemo, type ReactNode } from "react";
import { createMemoryRouter, RouterProvider, useLocation } from "react-router";
import { vi, type Mock } from "vitest";

const theme = { appearance: "light", accentColor: "indigo" } as const;

// Location shows the current path and query, so tests can check navigation.
function Location() {
  const { pathname, search } = useLocation();
  return <output aria-label="Location">{pathname + search}</output>;
}

interface ProvidersProps {
  ui: ReactNode;
  path: string;
  route: string;
}

function Providers({ ui, path, route }: ProvidersProps) {
  const client = useMemo(
    () => new QueryClient({ defaultOptions: { queries: { retry: false } } }),
    [],
  );
  const router = useMemo(
    () =>
      createMemoryRouter(
        [
          {
            path: route,
            element: (
              <>
                {ui}
                <Location />
              </>
            ),
          },
          { path: "*", element: <Location /> },
        ],
        { initialEntries: [path] },
      ),
    [ui, path, route],
  );
  return (
    <Theme defaultValue={theme}>
      <Toast.Provider>
        <QueryClientProvider client={client}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </Toast.Provider>
    </Theme>
  );
}

// renderUI renders ui at route, starting at path, with the app providers.
// Other paths render only the Location output. Queries do not retry in tests.
export function renderUI(ui: ReactNode, path = "/", route = "*") {
  return render(<Providers ui={ui} path={path} route={route} />);
}

// mockAPI stubs fetch: routes maps "<path with query>" (GET) or "<METHOD> <path>"
// to a JSON body or a Response. The generated client calls fetch with a Request.
export function mockAPI(routes: Record<string, unknown>) {
  const fetchMock = vi.fn<typeof fetch>(async (input) => {
    const req =
      input instanceof Request ? input : new Request(new URL(input.toString(), "http://localhost"));
    const parsed = new URL(req.url, "http://localhost");
    const url = parsed.pathname + parsed.search;
    const key = req.method === "GET" ? url : `${req.method} ${url}`;
    if (!(key in routes)) {
      return new Response(JSON.stringify({ error: "not found" }), { status: 404 });
    }
    const body = routes[key];
    if (body instanceof Response) return body;
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

// sentBody returns the JSON body of the first request with method.
export async function sentBody(fetchMock: Mock<typeof fetch>, method: string) {
  const call = fetchMock.mock.calls.find(([r]) => r instanceof Request && r.method === method);
  if (!call) throw new Error(`no ${method} request`);
  return (call[0] as Request).clone().json();
}

// json is a JSON Response with status, for mockAPI routes.
export function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
