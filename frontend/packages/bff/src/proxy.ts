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

export function createSessionProxy(cfg: BffConfig) {
  return async function sessionProxy(
    request: NextRequest,
  ): Promise<NextResponse> {
    const raw = request.cookies.get(cfg.cookieName)?.value;
    if (!raw) return NextResponse.next();

    const session = await decodeSession(raw, cfg);
    if (!session) {
      const res = NextResponse.next({
        request: { headers: forward(request, cfg, null) },
      });
      res.cookies.delete(cfg.cookieName);
      return res;
    }
    if (!needsRefresh(session)) return NextResponse.next();

    const outcome = await refreshSession(cfg, session);
    if (outcome.kind === "unavailable") return NextResponse.next();
    if (outcome.kind === "revoked") {
      const res = NextResponse.next({
        request: { headers: forward(request, cfg, null) },
      });
      res.cookies.delete(cfg.cookieName);
      return res;
    }
    const sealed = await encodeSession(outcome.session, cfg);
    const res = NextResponse.next({
      request: { headers: forward(request, cfg, sealed) },
    });
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
