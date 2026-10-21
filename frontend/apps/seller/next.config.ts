import path from "node:path";

import type { NextConfig } from "next";

// Repo root, so Turbopack can compile the shared/ packages that live outside frontend/.
const repoRoot = path.join(__dirname, "../../..");

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  transpilePackages: [
    "@techshop/api-client",
    "@techshop/bff",
    "@techshop/design-tokens",
    "@techshop/fixtures",
    "@techshop/ui",
  ],
  turbopack: { root: repoRoot },
  // KYC documents (up to four, 10 MB each) are posted to a Server Action, which uploads them to
  // storage; the browser never talks to the API directly.
  experimental: { serverActions: { bodySizeLimit: "41mb" } },
  outputFileTracingRoot: repoRoot,
};

export default nextConfig;
