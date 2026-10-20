// Per-app BFF settings, read from the environment once. Secrets are refused when missing or weak
// in production (owner rule: no default or example secrets outside local development).
export type Audience = "market" | "wholesale" | "seller" | "staff";

export type BffConfig = {
  app: string;
  audience: Audience;
  apiUrl: string;
  secret: string;
  cookieName: string;
  /** Secure cookies (always in production; `__Host-` prefix needs it). */
  secure: boolean;
  /**
   * Request header holding the shopper's IP, set by the reverse proxy in front of this app
   * (nginx: `X-Real-IP`). Forwarded to the API so its per-IP limits see the real client.
   * Unset = don't forward (an app reached without a proxy must not trust client headers).
   */
  clientIpHeader?: string;
};

const MIN_SECRET_LENGTH = 32;

/** A fixed per-app key for local development only, so proxy and render agree without setup. */
const devSecret = (app: string) =>
  `techshop-local-development-only-session-key:${app}`;

export function loadConfig(
  app: string,
  audience: Audience,
  env: Record<string, string | undefined> = process.env,
): BffConfig {
  // `next build` / `next start` run with NODE_ENV=production, so a local production build
  // also needs a real SESSION_SECRET. That's deliberate: no escape hatch to forget in prod.
  const production = env.NODE_ENV === "production";
  const apiUrl = env.API_URL;
  if (!apiUrl) throw new Error(`${app}: API_URL is required`);

  let secret = env.SESSION_SECRET ?? "";
  if (!secret) {
    if (production)
      throw new Error(`${app}: SESSION_SECRET is required in production`);
    secret = devSecret(app);
  } else if (
    secret.length < MIN_SECRET_LENGTH ||
    /change|example|secret|placeholder/i.test(secret)
  ) {
    if (production)
      throw new Error(
        `${app}: SESSION_SECRET must be at least ${MIN_SECRET_LENGTH} random characters`,
      );
  }

  const secure = production || apiUrl.startsWith("https://");
  const ipHeader = env.CLIENT_IP_HEADER?.trim().toLowerCase();
  if (ipHeader && !/^[a-z0-9-]{1,64}$/.test(ipHeader))
    throw new Error(
      `${app}: CLIENT_IP_HEADER must be a header name, e.g. x-real-ip`,
    );
  return {
    app,
    audience,
    apiUrl,
    secret,
    // `__Host-` pins the cookie to this exact host over HTTPS (never shared across subdomains).
    cookieName: secure ? "__Host-ts_session" : "ts_session",
    secure,
    clientIpHeader: ipHeader || undefined,
  };
}
