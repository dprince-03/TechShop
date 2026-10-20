import type { BffConfig } from "./config";

/** The proxy-supplied client IP as `X-Forwarded-For` (one hop; never the browser's own header). */
export function forwardedFor(
  h: Headers,
  cfg: BffConfig,
): Record<string, string> {
  if (!cfg.clientIpHeader) return {};
  const ip = h.get(cfg.clientIpHeader)?.split(",")[0]?.trim();
  return ip && /^[0-9a-f.:]{3,45}$/i.test(ip) ? { "X-Forwarded-For": ip } : {};
}
