import type { HealthResponse } from "./types";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export type ApiClientOptions = {
  baseUrl: string;
  /** Returns a bearer token for authenticated requests, if any. */
  getToken?: () => string | null | undefined | Promise<string | null | undefined>;
  /** Override for tests or custom runtimes. Defaults to global fetch. */
  fetch?: typeof fetch;
};

export function createApiClient({ baseUrl, getToken, fetch: fetchImpl = fetch }: ApiClientOptions) {
  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const token = await getToken?.();
    const res = await fetchImpl(`${baseUrl}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init.headers,
      },
    });

    if (!res.ok) {
      throw new ApiError(res.status, `${res.status} ${res.statusText}`);
    }
    return (await res.json()) as T;
  }

  return {
    request,
    health: () => request<HealthResponse>("/health"),
  };
}

export type ApiClient = ReturnType<typeof createApiClient>;
