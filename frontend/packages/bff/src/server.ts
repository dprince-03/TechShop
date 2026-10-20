import "server-only";

import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";

import {
  createApiClient,
  type ApiClient,
  type Schema,
} from "@techshop/api-client";

import { loadConfig, type Audience, type BffConfig } from "./config";
import { signInPath } from "./paths";
import {
  cookieOptions,
  decodeSession,
  encodeSession,
  needsRefresh,
  refreshSession,
  sessionFromTokens,
  type Session,
} from "./session";

export type Bff = ReturnType<typeof createBff>;

/**
 * One instance per app (`src/lib/bff.ts`). Everything here runs on the server: pages read the
 * session, Server Functions sign in and out, and every API call carries the bearer token.
 */
export function createBff(app: string, audience: Audience) {
  let cfg: BffConfig | undefined;
  const config = () => (cfg ??= loadConfig(app, audience));

  /** Unauthenticated client for public reads (catalogue, content). */
  const publicApi = (): ApiClient =>
    createApiClient({ baseUrl: config().apiUrl });

  async function writeSession(s: Session) {
    (await cookies()).set(
      config().cookieName,
      await encodeSession(s, config()),
      cookieOptions(config(), s),
    );
  }

  async function clearSession() {
    (await cookies()).delete(config().cookieName);
  }

  /**
   * The current session, or null. Pages rely on `proxy.ts` having refreshed it. Server Functions
   * and Route Handlers pass `{ refresh: true }`, because they can write the rotated cookie.
   */
  async function getSession(
    opts: { refresh?: boolean } = {},
  ): Promise<Session | null> {
    const raw = (await cookies()).get(config().cookieName)?.value;
    const s = await decodeSession(raw, config());
    if (!s) return null;
    if (opts.refresh && needsRefresh(s)) {
      const outcome = await refreshSession(config(), s);
      if (outcome.kind === "refreshed") {
        await writeSession(outcome.session);
        return outcome.session;
      }
      if (outcome.kind === "revoked") {
        await clearSession();
        return null;
      }
    }
    return s;
  }

  /** Session or redirect to the app's sign-in page (with a safe return path). */
  async function requireSession(returnTo?: string): Promise<Session> {
    const s = await getSession();
    if (s) return s;
    redirect(signInPath(returnTo));
  }

  /** API client acting as the signed-in user (anonymous when signed out). */
  async function api(opts: { refresh?: boolean } = {}): Promise<ApiClient> {
    const s = await getSession(opts);
    return createApiClient({
      baseUrl: config().apiUrl,
      getToken: () => s?.accessToken,
    });
  }

  /** Stores the tokens from a successful sign-in (Server Functions only). */
  async function signIn(tokens: Schema<"TokenPair">) {
    await writeSession(sessionFromTokens(tokens));
  }

  /** Revokes the session at the API, then drops the cookie, even if the API call fails. */
  async function signOut() {
    const s = await getSession();
    if (s) {
      try {
        await createApiClient({ baseUrl: config().apiUrl }).post(
          "/api/v1/auth/logout",
          {
            token: s.accessToken,
            body: { refreshToken: s.refreshToken },
          },
        );
      } catch {
        // The cookie still goes; the session expires on its own if the API was unreachable.
      }
    }
    await clearSession();
  }

  return {
    config,
    publicApi,
    api,
    getSession,
    requireSession,
    signIn,
    signOut,
  };
}

/**
 * CSRF check for custom Route Handlers that change state (Server Actions already compare Origin).
 * Accepts same-origin requests only.
 */
export async function assertSameOrigin(): Promise<void> {
  const h = await headers();
  const site = h.get("sec-fetch-site");
  if (site && site !== "same-origin")
    throw new Response("Forbidden", { status: 403 });
  const origin = h.get("origin");
  const host = h.get("x-forwarded-host") ?? h.get("host");
  if (origin && host && new URL(origin).host !== host)
    throw new Response("Forbidden", { status: 403 });
}
