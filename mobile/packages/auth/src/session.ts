// Mobile session for both apps (docs/identity-access.md §5.2): the refresh token lives in secure
// storage, the access token only in memory. Pure TypeScript; storage and fetch are injected.
import { ApiError, createApiClient, type ApiClient, type Schema } from "@techshop/api-client";

export interface TokenStore {
  get(): Promise<string | null>;
  set(refreshToken: string): Promise<void>;
  clear(): Promise<void>;
}

type Tokens = Schema<"TokenPair">;

export type SessionEvents = {
  /** Called when the session ends (signed out, or the API refused the refresh token). */
  onSignedOut?: () => void;
};

export class MobileSession {
  private access: { token: string; expiresAt: number } | null = null;
  private refreshing: Promise<boolean> | null = null;
  readonly api: ApiClient;
  private readonly baseUrl: string;
  private readonly store: TokenStore;
  private readonly fetchImpl: typeof fetch;
  private readonly events: SessionEvents;

  // Plain fields, not parameter properties: Node's type stripping (used by the unit tests)
  // doesn't support them.
  constructor(
    baseUrl: string,
    store: TokenStore,
    fetchImpl: typeof fetch = (...a) => fetch(...a),
    events: SessionEvents = {},
    headers: Record<string, string> = {},
  ) {
    this.baseUrl = baseUrl;
    this.store = store;
    this.fetchImpl = fetchImpl;
    this.events = events;
    // Every call carries the access token; a 401 refreshes once and retries.
    this.api = createApiClient({ baseUrl, headers, fetch: (input, init) => this.authedFetch(input, init) });
  }

  /** Stores a fresh token pair from sign-in or refresh. */
  async accept(t: Tokens) {
    this.access = { token: t.accessToken, expiresAt: Date.parse(t.accessExpiresAt) };
    await this.store.set(t.refreshToken);
  }

  async hasSession(): Promise<boolean> {
    return Boolean(this.access) || Boolean(await this.store.get());
  }

  /** Rotates the refresh token. Concurrent callers share one request (rotation must not race). */
  refresh(): Promise<boolean> {
    this.refreshing ??= this.doRefresh().finally(() => {
      this.refreshing = null;
    });
    return this.refreshing;
  }

  private async doRefresh(): Promise<boolean> {
    const refreshToken = await this.store.get();
    if (!refreshToken) return false;
    try {
      const plain = createApiClient({ baseUrl: this.baseUrl, fetch: this.fetchImpl });
      await this.accept(await plain.post("/api/v1/auth/refresh", { body: { refreshToken } }));
      return true;
    } catch (e) {
      // Refused (expired, revoked, reused): sign out. A network error keeps the token for later.
      if (e instanceof ApiError && e.status >= 400 && e.status < 500 && e.status !== 429) await this.end();
      return false;
    }
  }

  private async token(): Promise<string | null> {
    if (this.access && this.access.expiresAt - 30_000 > Date.now()) return this.access.token;
    return (await this.refresh()) ? this.access!.token : null;
  }

  private async authedFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
    const send = async (token: string | null) => {
      const headers = new Headers(init?.headers);
      if (token && !headers.has("Authorization")) headers.set("Authorization", `Bearer ${token}`);
      return this.fetchImpl(input, { ...init, headers });
    };
    const res = await send(await this.token());
    if (res.status !== 401) return res;
    // The access token may have been revoked or rotated elsewhere: refresh once and retry.
    this.access = null;
    return (await this.refresh()) ? send(this.access!.token) : res;
  }

  /** Signs out at the API (best effort) and forgets both tokens. */
  async signOut() {
    const refreshToken = await this.store.get();
    try {
      if (refreshToken) await this.api.post("/api/v1/auth/logout", { body: { refreshToken } });
    } catch {
      // Still sign out locally; the session expires on its own.
    }
    await this.end();
  }

  private async end() {
    this.access = null;
    await this.store.clear();
    this.events.onSignedOut?.();
  }
}
