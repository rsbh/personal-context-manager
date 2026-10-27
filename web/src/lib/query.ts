import { QueryClient } from "@tanstack/react-query";
import { client } from "../client/client.gen";

// APIErrorBody is an API error with the HTTP status (added by withStatus).
export interface APIErrorBody {
  error?: string;
  status?: number;
}

// withStatus adds the HTTP status to an error body. The generated client throws only the body.
export function withStatus(error: unknown, response: Response | undefined): APIErrorBody {
  const body: APIErrorBody =
    typeof error === "object" && error !== null
      ? { ...(error as APIErrorBody) }
      : { error: String(error) };
  if (response) body.status = response.status;
  return body;
}

// shouldRetry retries server and network errors twice. A 4xx will not change, so no retry.
export function shouldRetry(failures: number, error: unknown): boolean {
  const status = (error as APIErrorBody | null)?.status;
  if (status !== undefined && status >= 400 && status < 500) return false;
  return failures < 2;
}

export function setupClient() {
  client.interceptors.error.use((error, response) => withStatus(error, response));
}

export function newQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: shouldRetry } } });
}

// contextQueries are the query ids that show context data, and the git status and log
// (every change to a context changes them).
const contextQueries = [
  "listContexts",
  "getContext",
  "searchContexts",
  "resolveContexts",
  "gitStatus",
  "gitLog",
];

// treeQueries are contextQueries plus the sidebar tree lists.
const treeQueries = [...contextQueries, "listSections", "listProjects"];

// invalidate refetches the queries with these ids.
// A key is [{ _id, ...options }]; a partial key matches all options.
function invalidate(queryClient: QueryClient, ids: string[]) {
  return Promise.all(ids.map((id) => queryClient.invalidateQueries({ queryKey: [{ _id: id }] })));
}

// invalidateContexts refetches all context data after a create, update or delete.
export function invalidateContexts(queryClient: QueryClient) {
  return invalidate(queryClient, contextQueries);
}

// invalidateTree refetches the tree and context data after a section or project change.
export function invalidateTree(queryClient: QueryClient) {
  return invalidate(queryClient, treeQueries);
}

// APIConflict is a sync conflict error (409): the files that conflict.
export interface APIConflict extends APIErrorBody {
  files?: string[];
}

// errorStatus is the HTTP status of an API error (added by withStatus).
export function errorStatus(error: unknown): number | undefined {
  return (error as APIErrorBody | null)?.status;
}

// errorMessage is the server message of an API error.
export function errorMessage(error: unknown): string {
  return (error as APIErrorBody | null)?.error ?? "Request failed";
}
