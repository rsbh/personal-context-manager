// API client for pcm-server. The UI is served from the same origin.

export class APIError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(path);
  if (!res.ok) {
    let message = res.statusText;
    try {
      message = (await res.json()).error ?? message;
    } catch {
      // body is not JSON
    }
    throw new APIError(res.status, message);
  }
  return res.json() as Promise<T>;
}

export const api = {
  health: () => getJSON<{ status: string }>("/health"),
  projects: () => getJSON<string[]>("/api/v1/projects"),
  globalSections: () => getJSON<string[]>("/api/v1/sections?scope=global"),
};
