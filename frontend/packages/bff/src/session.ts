// The session the BFF keeps in its encrypted cookie, and how it is refreshed. The browser never
// sees either token; server code sends the access token to the API as a bearer token.
import { ApiError, createApiClient, type Schema } from "@techshop/api-client";

import type { BffConfig } from "./config";
import { seal, unseal } from "./crypto";

export type Session = {
  sessionId: string;
  userId: string;
  accessToken: string;
  /** Epoch milliseconds. */
  accessExpiresAt: number;
  refreshToken: string;
  refreshExpiresAt: number;
};

/** Refresh this long before the access token expires, so a render never starts with a stale one. */
export const REFRESH_SKEW_MS = 60_000;

export function sessionFromTokens(t: Schema<"TokenPair">): Session {
  return {
    sessionId: t.sessionId,
    userId: t.userId,
    accessToken: t.accessToken,
    accessExpiresAt: Date.parse(t.accessExpiresAt),
    refreshToken: t.refreshToken,
    refreshExpiresAt: Date.parse(t.refreshExpiresAt),
  };
}

export const needsRefresh = (s: Session, now = Date.now()) =>
  s.accessExpiresAt - REFRESH_SKEW_MS <= now;
export const isExpired = (s: Session, now = Date.now()) =>
  s.refreshExpiresAt <= now;

export const encodeSession = (s: Session, cfg: BffConfig) =>
  seal(s, cfg.secret);

export async function decodeSession(
  raw: string | undefined,
  cfg: BffConfig,
): Promise<Session | null> {
  const s = await unseal<Session>(raw, cfg.secret);
  if (!s || typeof s.refreshToken !== "string" || isExpired(s)) return null;
  return s;
}

export function cookieOptions(cfg: BffConfig, s: Session) {
  return {
    httpOnly: true,
    secure: cfg.secure,
    sameSite: "lax" as const,
    path: "/",
    maxAge: Math.max(0, Math.floor((s.refreshExpiresAt - Date.now()) / 1000)),
  };
}

export type RefreshOutcome =
  | { kind: "refreshed"; session: Session }
  | { kind: "revoked" }
  | { kind: "unavailable" };

/**
 * Rotates the refresh token. "revoked" means the API refused it (expired, signed out, or reuse
 * detected) and the cookie must go; "unavailable" keeps the cookie so a blip doesn't sign anyone out.
 */
export async function refreshSession(
  cfg: BffConfig,
  s: Session,
  fetchImpl?: typeof fetch,
): Promise<RefreshOutcome> {
  const api = createApiClient({ baseUrl: cfg.apiUrl, fetch: fetchImpl });
  try {
    const t = await api.post("/api/v1/auth/refresh", {
      body: { refreshToken: s.refreshToken },
      init: { cache: "no-store" },
    });
    return { kind: "refreshed", session: sessionFromTokens(t) };
  } catch (e) {
    if (
      e instanceof ApiError &&
      e.status >= 400 &&
      e.status < 500 &&
      e.status !== 429
    )
      return { kind: "revoked" };
    return { kind: "unavailable" };
  }
}
