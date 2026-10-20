// Type plumbing that turns the generated OpenAPI `paths` into typed request options and
// response bodies. Nothing here exists at runtime.
import type { paths } from "../generated/schema";

export type HttpMethod = "get" | "post" | "put" | "patch" | "delete";

/** Paths that support a given method. */
export type PathsFor<M extends HttpMethod> = {
  [P in keyof paths]: NonNullable<paths[P][M]> extends never ? never : P;
}[keyof paths];

export type Operation<P extends keyof paths, M extends HttpMethod> = NonNullable<paths[P][M]>;

type Params<O> = O extends { parameters: infer X } ? X : never;
type Field<T, K extends string> = K extends keyof T ? NonNullable<T[K]> : never;

type PathOf<O> = Field<Params<O>, "path">;
type QueryOf<O> = Field<Params<O>, "query">;
type HeaderOf<O> = Field<Params<O>, "header">;

// Required `requestBody` → `body` is required; optional → optional; none → not allowed.
type BodyOption<O> = O extends { requestBody: { content: { "application/json": infer B } } }
  ? { body: B }
  : O extends { requestBody?: { content: { "application/json": infer B } } }
    ? { body?: B }
    : { body?: never };

type PathOption<O> = [PathOf<O>] extends [never] ? { path?: never } : { path: PathOf<O> };
// A query with any required parameter is a required `query` option.
type QueryOption<O> = [QueryOf<O>] extends [never]
  ? { query?: never }
  : Params<O> extends { query: object }
    ? { query: QueryOf<O> }
    : { query?: QueryOf<O> };
type HeaderOption<O> = [HeaderOf<O>] extends [never]
  ? { headers?: Record<string, string> }
  : { headers?: Partial<HeaderOf<O>> & Record<string, string | undefined> };

/** Options shared by every call, independent of the operation. */
export type CallOptions = {
  /** Bearer token for this call (overrides the client's `getToken`). */
  token?: string | null;
  /** Sent as `Idempotency-Key` (POST /orders and other money-moving calls require one). */
  idempotencyKey?: string;
  signal?: AbortSignal;
  /** Extra fetch options (for example Next.js `next: { revalidate }` or `cache`). */
  init?: RequestInit;
};

export type RequestOptions<O> = PathOption<O> & QueryOption<O> & HeaderOption<O> & BodyOption<O> & CallOptions;

/** Options become a required argument only when the operation needs a path or body. */
export type OptionsArg<O> = object extends RequestOptions<O> ? [options?: RequestOptions<O>] : [options: RequestOptions<O>];

type JsonOf<R> = R extends { content: { "application/json": infer J } } ? J : undefined;

/** The JSON body of the operation's success response (200, 201 or 204). */
export type SuccessBody<O> = O extends { responses: infer R }
  ? 200 extends keyof R
    ? JsonOf<R[200]>
    : 201 extends keyof R
      ? JsonOf<R[201]>
      : 202 extends keyof R
        ? JsonOf<R[202]>
        : undefined
  : never;

export type ApiResponse<T> = {
  data: T;
  status: number;
  headers: Headers;
};
