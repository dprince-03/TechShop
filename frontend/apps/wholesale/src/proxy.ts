import type { NextRequest } from "next/server";

import { createSessionProxy, loadConfig } from "@techshop/bff/proxy";

// Keeps the session fresh before each page renders (see @techshop/bff/proxy).
const refresh = createSessionProxy(loadConfig("wholesale", "wholesale"));

export function proxy(request: NextRequest) {
  return refresh(request);
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:png|jpg|jpeg|svg|webp|ico|css|js|woff2?)$).*)",
  ],
};
