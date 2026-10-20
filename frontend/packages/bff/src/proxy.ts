// Session refresh for each app's `proxy.ts`. Server Components can read cookies but not set them,
// and the API rotates refresh tokens on every use, so the refresh must happen here: the new
// cookie goes to the browser and, in the same request, to the page being rendered.
import { NextResponse, type NextRequest } from "next/server";

import type { BffConfig } from "./config";

export { loadConfig } from "./config";
import {
  cookieOptions,
  decodeSession,
  encodeSession,
  needsRefresh,
  refreshSession,
} from "./session";

/** Paths that never need a session refresh (static assets, images, metadata files). */
export const PROXY_MATCHER =
  "/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:png|jpg|jpeg|svg|webp|ico|css|js|woff2?)$).*)";

/** Rewrites one cookie in a raw Cookie header (or removes it when value is null). */
function withCookie(
  header: string | null,
  name: string,
  value: string | null,
): string {
  const parts = (header ?? "")
    .split(";")
    .map((p) => p.trim())
    .filter((p) => p && !p.startsWith(`${name}=`));
  if (value !== null) parts.push(`${name}=${value}`);
  return parts.join("; ");
}

type ProxyOptions = {
  /** Extra request headers for the render, e.g. the current path for sign-in redirects. */
  requestHeaders?: (request: NextRequest) => Record<string, string>;
};

export function createSessionProxy(cfg: BffConfig, opts: ProxyOptions = {}) {
  return async function sessionProxy(
    request: NextRequest,
  ): Promise<NextResponse> {
    const extra = opts.requestHeaders?.(request) ?? {};
    // `undefined` keeps the cookie as is; null drops it; a string replaces it.
    const next = (sealed?: string | null) => {
      const headers =
        sealed === undefined
          ? new Headers(request.headers)
          : forward(request, cfg, sealed);
      for (const [k, v] of Object.entries(extra)) headers.set(k, v);
      return NextResponse.next({ request: { headers } });
    };

    const raw = request.cookies.get(cfg.cookieName)?.value;
    if (!raw) return next();

    const session = await decodeSession(raw, cfg);
    if (!session) {
      const res = next(null);
      res.cookies.delete(cfg.cookieName);
      return res;
    }
    if (!needsRefresh(session)) return next();

    const outcome = await refreshSession(cfg, session);
    if (outcome.kind === "unavailable") return next();
    if (outcome.kind === "revoked") {
      const res = next(null);
      res.cookies.delete(cfg.cookieName);
      return res;
    }
    const sealed = await encodeSession(outcome.session, cfg);
    const res = next(sealed);
    res.cookies.set(
      cfg.cookieName,
      sealed,
      cookieOptions(cfg, outcome.session),
    );
    return res;
  };
}

function forward(
  request: NextRequest,
  cfg: BffConfig,
  sealed: string | null,
): Headers {
  const headers = new Headers(request.headers);
  headers.set(
    "cookie",
    withCookie(request.headers.get("cookie"), cfg.cookieName, sealed),
  );
  return headers;
}
