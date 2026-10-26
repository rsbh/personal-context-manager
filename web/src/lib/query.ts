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
