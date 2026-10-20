import type { NextRequest } from "next/server";

import { createSessionProxy, loadConfig } from "@techshop/bff/proxy";

// Keeps the session fresh before each page renders (see @techshop/bff/proxy).
// Also passes the path to the portal layout, so sign-in can send people back where they were.
const refresh = createSessionProxy(loadConfig("staff", "staff"), {
  requestHeaders: (request) => ({
    "x-staff-path": request.nextUrl.pathname + request.nextUrl.search,
  }),
});

export function proxy(request: NextRequest) {
  return refresh(request);
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:png|jpg|jpeg|svg|webp|ico|css|js|woff2?)$).*)",
  ],
};
