/**
 * Accepts a provider checkout URL only if it is HTTPS, or the local API's own development
 * checkout page (fake providers). Anything else is dropped rather than redirected to.
 */
export function safePaymentUrl(
  url: string | undefined,
  apiUrl: string,
): string | undefined {
  if (!url) return undefined;
  try {
    const u = new URL(url);
    if (u.protocol === "https:") return u.toString();
    const api = new URL(apiUrl);
    const sameApiHost =
      u.port === api.port &&
      ["localhost", "127.0.0.1", api.hostname].includes(u.hostname);
    if (
      u.protocol === "http:" &&
      sameApiHost &&
      u.pathname.startsWith("/dev/pay/")
    )
      return u.toString();
  } catch {
    return undefined;
  }
  return undefined;
}
