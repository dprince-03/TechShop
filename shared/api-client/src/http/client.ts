// Typed client for the TechShop Go API. Paths, parameters and bodies are checked against the
// generated OpenAPI contract, so a backend change that breaks a screen fails the typecheck.
import { ApiError, networkError } from "./errors";
import type {
  ApiResponse,
  HttpMethod,
  Operation,
  OptionsArg,
  PathsFor,
  SuccessBody,
} from "./types";

export type ApiClientOptions = {
  /** API origin, e.g. http://localhost:8080 (paths already start with /api/v1). */
  baseUrl: string;
  /** Default bearer token for calls that don't pass `token`. */
  getToken?: () =>
    string | null | undefined | Promise<string | null | undefined>;
  /** Override for tests, token refresh wrappers or custom runtimes. Defaults to global fetch. */
  fetch?: typeof fetch;
  /** Headers added to every call (for example a device ID on mobile). */
  headers?: Record<string, string>;
};

type RawOptions = {
  path?: Record<string, string | number>;
  query?: Record<string, QueryValue>;
  headers?: Record<string, string | undefined>;
  body?: unknown;
  token?: string | null;
  idempotencyKey?: string;
  signal?: AbortSignal;
  init?: RequestInit;
};

type QueryValue =
  | string
  | number
  | boolean
  | null
  | undefined
  | readonly (string | number | boolean)[];

/** Replaces `{name}` segments with encoded path values. */
export function buildPath(
  template: string,
  path?: Record<string, string | number>,
): string {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => {
    const value = path?.[key];
    if (value === undefined || value === "")
      throw new Error(`Missing path parameter "${key}" for ${template}`);
    return encodeURIComponent(String(value));
  });
}

/** Serialises a query object; arrays repeat the key, empty values are dropped. */
export function buildQuery(query?: Record<string, QueryValue>): string {
  if (!query) return "";
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === "") continue;
    if (Array.isArray(value))
      value.forEach((v) => params.append(key, String(v)));
    else params.append(key, String(value));
  }
  const s = params.toString();
  return s ? `?${s}` : "";
}

async function readBody(res: Response): Promise<unknown> {
  if (res.status === 204) return undefined;
  const text = await res.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

export function createApiClient(options: ApiClientOptions) {
  const fetchImpl =
    options.fetch ?? ((...args: Parameters<typeof fetch>) => fetch(...args));
  const baseUrl = options.baseUrl.replace(/\/$/, "");

  // Untyped core; the typed surface is the verb methods below (typing this generically over
  // every path makes TypeScript give up on the union).
  async function call(
    method: HttpMethod,
    path: string,
    opts: RawOptions | undefined,
  ): Promise<ApiResponse<unknown>> {
    const o = opts ?? {};
    const url = baseUrl + buildPath(path, o.path) + buildQuery(o.query);

    const headers = new Headers({
      Accept: "application/json, application/problem+json",
    });
    for (const [k, v] of Object.entries(options.headers ?? {}))
      headers.set(k, v);
    for (const [k, v] of Object.entries(o.headers ?? {}))
      if (v !== undefined && v !== "") headers.set(k, v);
    const token = o.token !== undefined ? o.token : await options.getToken?.();
    if (token) headers.set("Authorization", `Bearer ${token}`);
    if (o.idempotencyKey) headers.set("Idempotency-Key", o.idempotencyKey);
    if (o.body !== undefined) headers.set("Content-Type", "application/json");

    let res: Response;
    try {
      res = await fetchImpl(url, {
        ...o.init,
        method: method.toUpperCase(),
        headers,
        body: o.body === undefined ? undefined : JSON.stringify(o.body),
        signal: o.signal ?? o.init?.signal,
      });
    } catch (cause) {
      throw networkError(cause);
    }

    const data = await readBody(res);
    if (!res.ok) {
      const problem =
        data && typeof data === "object"
          ? data
          : { detail: typeof data === "string" ? data : undefined };
      throw new ApiError(
        res.status,
        problem,
        res.headers.get("x-request-id") ?? undefined,
      );
    }
    return { data, status: res.status, headers: res.headers };
  }

  // One typed method per HTTP verb; `request` variants also return status and headers.
  const verbWithResponse =
    <M extends HttpMethod>(method: M) =>
    <P extends PathsFor<M>>(path: P, ...args: OptionsArg<Operation<P, M>>) =>
      call(method, path, args[0] as RawOptions | undefined) as Promise<
        ApiResponse<SuccessBody<Operation<P, M>>>
      >;
  const verb =
    <M extends HttpMethod>(method: M) =>
    async <P extends PathsFor<M>>(
      path: P,
      ...args: OptionsArg<Operation<P, M>>
    ) =>
      (await verbWithResponse(method)(path, ...args)).data;

  return {
    get: verb("get"),
    post: verb("post"),
    put: verb("put"),
    patch: verb("patch"),
    delete: verb("delete"),
    /** Like the verb methods, but also returns status and response headers. */
    request: {
      get: verbWithResponse("get"),
      post: verbWithResponse("post"),
      put: verbWithResponse("put"),
      patch: verbWithResponse("patch"),
      delete: verbWithResponse("delete"),
    },
    baseUrl,
  };
}

export type ApiClient = ReturnType<typeof createApiClient>;
