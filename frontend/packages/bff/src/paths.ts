// Redirect targets built from user input (`?next=`). Kept free of Next imports so it is unit-tested.
/** Builds `/sign-in?next=…` accepting only same-site relative paths (no open redirects). */
export function signInPath(returnTo?: string): string {
  const next = safeReturnPath(returnTo);
  return next && next !== "/"
    ? `/sign-in?next=${encodeURIComponent(next)}`
    : "/sign-in";
}

/** Returns `path` when it is a plain relative path, else "/" (rejects `//x`, `/\x`, `https:x`). */
export function safeReturnPath(path: string | undefined | null): string {
  if (
    !path ||
    !path.startsWith("/") ||
    path.startsWith("//") ||
    path.startsWith("/\\")
  )
    return "/";
  if (/[\u0000-\u001f]/.test(path)) return "/";
  return path;
}
