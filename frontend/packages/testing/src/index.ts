// Drives a running Next.js app over HTTP like a browser: keeps cookies, loads pages, and calls
// Server Actions by name (IDs come from the dev server's action manifest). Used by each app's
// test/*.test.ts against `next dev` + a seeded local API.
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";

// React's own argument encoding (what the browser sends), from the Next install.
const require_ = createRequire(import.meta.url);
const { encodeReply } = require_(
  "next/dist/compiled/react-server-dom-turbopack/client.node",
) as {
  encodeReply: (value: unknown) => Promise<string | FormData>;
};

/** A random address from 198.18.0.0/15 (reserved for benchmarking, never a real client). */
const randomIp = () =>
  `198.${18 + Math.floor(Math.random() * 2)}.${Math.floor(Math.random() * 256)}.${1 + Math.floor(Math.random() * 254)}`;

type ManifestEntry = {
  exportedName?: string;
  filename?: string;
  workers?: Record<string, unknown>;
};

export type PageResult = {
  status: number;
  html: string;
  headers: Headers;
  url: string;
};

export class WebClient {
  readonly baseUrl: string;
  private readonly appDir: string;
  private readonly cookies = new Map<string, string>();

  /** Simulated client IP, sent as `X-Real-IP` the way nginx does (each browser gets its own). */
  readonly ip: string;

  /** `appDir` is the app's folder (where `.next/` lives). */
  constructor(baseUrl: string, appDir: string, ip = randomIp()) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.appDir = appDir;
    this.ip = ip;
  }

  cookie(name: string): string | undefined {
    return this.cookies.get(name);
  }

  clearCookies() {
    this.cookies.clear();
  }

  private cookieHeader() {
    return [...this.cookies].map(([k, v]) => `${k}=${v}`).join("; ");
  }

  private store(headers: Headers) {
    for (const line of headers.getSetCookie()) {
      const [pair, ...attrs] = line.split(";");
      const i = pair.indexOf("=");
      const name = pair.slice(0, i).trim();
      const value = pair.slice(i + 1).trim();
      const expired = attrs.some(
        (a) =>
          /max-age=0\b/i.test(a.trim()) ||
          /expires=thu, 01 jan 1970/i.test(a.trim()),
      );
      if (!value || expired) this.cookies.delete(name);
      else this.cookies.set(name, value);
    }
  }

  /** GET a page, following redirects manually so cookies set on the way are kept. */
  async get(pathname: string, maxRedirects = 5): Promise<PageResult> {
    let url = this.baseUrl + pathname;
    for (let i = 0; i <= maxRedirects; i++) {
      const res = await fetch(url, {
        headers: { cookie: this.cookieHeader(), "x-real-ip": this.ip },
        redirect: "manual",
      });
      this.store(res.headers);
      const loc = res.headers.get("location");
      if (res.status >= 300 && res.status < 400 && loc) {
        url = new URL(loc, url).toString();
        continue;
      }
      return {
        status: res.status,
        html: await res.text(),
        headers: res.headers,
        url,
      };
    }
    throw new Error(`too many redirects from ${pathname}`);
  }

  /** The action ID for an exported Server Action, as the given page sees it. */
  actionId(page: string, exportedName: string, file?: string): string {
    const manifest = JSON.parse(
      readFileSync(
        path.join(
          this.appDir,
          ".next/dev/server/server-reference-manifest.json",
        ),
        "utf8",
      ),
    ) as { node: Record<string, ManifestEntry> };
    const worker = `app${page === "/" ? "" : page}/page`;
    // Route groups ("(portal)") are in the worker names but not in URLs.
    const ungrouped = (w: string) => w.replace(/\/\([^/]+\)/g, "");
    const match = Object.entries(manifest.node).find(
      ([, v]) =>
        v.exportedName === exportedName &&
        (!file || v.filename?.endsWith(file)) &&
        Object.keys(v.workers ?? {}).some((w) => ungrouped(w) === worker),
    );
    if (!match)
      throw new Error(
        `no action ${exportedName} on ${page} (load the page once first)`,
      );
    return match[0];
  }

  /**
   * Calls a Server Action as the page's client code would. `page` is the route pattern (for the
   * manifest), `url` the concrete path posted to. Returns the action's value and any redirect.
   */
  async action<T = unknown>(
    page: string,
    exportedName: string,
    args: unknown[],
    opts: { url?: string; file?: string; origin?: string } = {},
  ): Promise<{ value: T; redirect?: string; status: number; raw: string }> {
    const id = this.actionId(page, exportedName, opts.file);
    const body = await encodeReply(args);
    const headers: Record<string, string> = {
      "Next-Action": id,
      Accept: "text/x-component",
      Origin: opts.origin ?? this.baseUrl,
      cookie: this.cookieHeader(),
      "x-real-ip": this.ip,
    };
    // Plain arguments go as text; FormData arguments as multipart (fetch sets the boundary).
    if (typeof body === "string")
      headers["Content-Type"] = "text/plain;charset=UTF-8";
    const res = await fetch(this.baseUrl + (opts.url ?? page), {
      method: "POST",
      headers,
      body,
      redirect: "manual",
    });
    this.store(res.headers);
    const text = await res.text();
    const redirect =
      res.headers.get("x-action-redirect")?.split(";")[0] ?? undefined;
    return {
      value: actionValue(text) as T,
      redirect,
      status: res.status,
      raw: text,
    };
  }
}

/**
 * The action's return value from a React Flight response. Row 0 is the root
 * (`{"a":"$@<id>", "f": <tree>}`); `a` points at the row holding the value.
 */
function actionValue(flight: string): unknown {
  const rows = new Map<string, string>();
  for (const line of flight.split("\n")) {
    // Model rows hold JSON; tagged rows (I, D, T, E… e.g. dev debug info) never start lowercase-JSON.
    const m = line.match(/^([0-9a-f]+):(.*)$/);
    if (m && !/^[A-Z]/.test(m[2]) && !rows.has(m[1])) rows.set(m[1], m[2]);
  }
  const root = rows.get("0");
  if (!root) return undefined;
  let ref: string | undefined;
  try {
    ref = (JSON.parse(root) as { a?: string }).a;
  } catch {
    return undefined;
  }
  if (typeof ref !== "string" || !ref.startsWith("$@")) return ref;
  const row = rows.get(ref.slice(2));
  if (row === undefined) return undefined;
  try {
    return JSON.parse(row);
  } catch {
    return row;
  }
}

/** Builds FormData from a plain object (for `useActionState` form actions). */
export function form(fields: Record<string, string>): FormData {
  const f = new FormData();
  for (const [k, v] of Object.entries(fields)) f.set(k, v);
  return f;
}

/** Visible text of an HTML page (tags and scripts stripped), for simple content assertions. */
export function textOf(html: string): string {
  return html
    .replace(/<script[\s\S]*?<\/script>/g, " ")
    .replace(/<style[\s\S]*?<\/style>/g, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&quot;/g, '"')
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/\s+/g, " ");
}

// API-side helpers (seed accounts, OTP codes from the dev message log, TOTP) shared with the
// api-client contract tests.
export {
  account,
  api,
  API_URL,
  customerToken,
  deliveredOrder,
  latestCode,
  peekCode,
  placePaidOrder,
  seed,
  staffToken,
  tokenForPhone,
  totp,
} from "../../../../shared/api-client/test/kit";
