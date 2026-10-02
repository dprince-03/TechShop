import path from "node:path";

import type { NextConfig } from "next";

// Repo root, so Turbopack can compile the shared/ packages that live outside frontend/.
const repoRoot = path.join(__dirname, "../../..");

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  transpilePackages: [
    "@techshop/api-client",
    "@techshop/design-tokens",
    "@techshop/ui",
  ],
  turbopack: { root: repoRoot },
  outputFileTracingRoot: repoRoot,
};

export default nextConfig;
