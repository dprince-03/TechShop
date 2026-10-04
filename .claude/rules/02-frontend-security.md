# 02 — Frontend Security

For browser-based applications: React, Next.js, Vue, Angular, plain HTML/JS, and admin dashboards.

> Golden rule: **the frontend is never a security boundary.** Everything here reduces client-side attack surface; the backend must still enforce all rules.

---

## 1. Cross-Site Scripting (XSS)

- [ ] `[CRITICAL]` No rendering of untrusted HTML without sanitization (DOMPurify)
- [ ] `[CRITICAL]` React: `dangerouslySetInnerHTML` only with sanitized input; audit every usage
- [ ] `[CRITICAL]` No `eval()`, `new Function()`, `setTimeout(string)` with user input
- [ ] `[HIGH]` `href`/`src` attributes validated: block `javascript:`, `data:` (except safe images), `vbscript:` URLs
- [ ] `[HIGH]` DOM sinks audited: `innerHTML`, `outerHTML`, `document.write`, `insertAdjacentHTML`
- [ ] `[HIGH]` Markdown rendering sanitized (react-markdown without `rehype-raw` or with sanitize plugin)
- [ ] `[HIGH]` Stored, reflected, and DOM-based XSS tested
- [ ] `[MEDIUM]` **Trusted Types** enforced via CSP where supported
- [ ] `[MEDIUM]` Vue `v-html` and Angular `bypassSecurityTrust*` usages audited

**Test payloads:** `<img src=x onerror=alert(1)>`, `"><svg onload=alert(1)>`, `javascript:alert(1)` in every input, URL param, and hash.

---

## 2. Content Security Policy (CSP)

- [ ] `[HIGH]` CSP header deployed (not just meta tag)
- [ ] `[HIGH]` No `unsafe-inline` or `unsafe-eval` in `script-src`; use nonces or hashes
- [ ] `[HIGH]` `object-src 'none'`, `base-uri 'self'` (or `'none'`), `frame-ancestors` set
- [ ] `[MEDIUM]` `strict-dynamic` with nonces for modern apps
- [ ] `[MEDIUM]` CSP violation reporting (`report-to` / `report-uri`)
- [ ] `[MEDIUM]` Rolled out via `Content-Security-Policy-Report-Only` first

**Example starter policy:**
```
Content-Security-Policy: default-src 'self'; script-src 'self' 'nonce-{RANDOM}' 'strict-dynamic'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https://api.example.com; frame-ancestors 'none'; base-uri 'self'; object-src 'none'; form-action 'self'; upgrade-insecure-requests
```

**Tools:** Google CSP Evaluator, Mozilla Observatory

---

## 3. Security Headers

| Header | Recommended value | Priority |
|---|---|---|
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains; preload` | HIGH |
| `X-Content-Type-Options` | `nosniff` | HIGH |
| `X-Frame-Options` | `DENY` (or CSP `frame-ancestors`) | HIGH |
| `Referrer-Policy` | `strict-origin-when-cross-origin` or `no-referrer` | MEDIUM |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` (adjust) | MEDIUM |
| `Cross-Origin-Opener-Policy` | `same-origin` | MEDIUM |
| `Cross-Origin-Resource-Policy` | `same-origin` / `same-site` | MEDIUM |
| `Cross-Origin-Embedder-Policy` | `require-corp` (if needed) | LOW |

- [ ] `[HIGH]` All headers above reviewed and set
- [ ] `[MEDIUM]` Domain submitted to HSTS preload list (after confirming all subdomains support HTTPS)
- [ ] `[LOW]` Deprecated `X-XSS-Protection` removed or set to `0`

**Tools:** securityheaders.com, Mozilla Observatory

---

## 4. CSRF

- [ ] `[HIGH]` Cookie-authenticated apps use CSRF tokens (synchronizer or double-submit) or rely on `SameSite` + origin checks
- [ ] `[HIGH]` State-changing actions never use GET
- [ ] `[HIGH]` `Origin` / `Referer` validated on sensitive requests
- [ ] `[MEDIUM]` `SameSite=Lax` minimum on session cookies; `Strict` where UX allows

---

## 5. Token & Session Storage

- [ ] `[HIGH]` Auth tokens in `HttpOnly; Secure; SameSite` cookies preferred over `localStorage`/`sessionStorage`
- [ ] `[HIGH]` If tokens must live in JS memory, keep access tokens short-lived and never persist refresh tokens in `localStorage`
- [ ] `[HIGH]` Logout clears tokens client-side and invalidates server-side
- [ ] `[MEDIUM]` No sensitive data cached in `localStorage`, IndexedDB, or service worker caches
- [ ] `[MEDIUM]` Sensitive pages send `Cache-Control: no-store`

---

## 6. Clickjacking

- [ ] `[HIGH]` `frame-ancestors 'none'` (or specific origins) / `X-Frame-Options: DENY`
- [ ] `[MEDIUM]` Embeddable widgets isolated on separate origin

---

## 7. CORS

- [ ] `[CRITICAL]` Never `Access-Control-Allow-Origin: *` together with credentials
- [ ] `[HIGH]` Origin allow-list exact-match; don't reflect arbitrary `Origin` header
- [ ] `[HIGH]` No `null` origin allowed
- [ ] `[MEDIUM]` Regex origin matching anchored (avoid `evil-example.com` matching `example.com`)
- [ ] `[MEDIUM]` Allowed methods and headers minimized

---

## 8. Third-Party Scripts & Supply Chain

- [ ] `[HIGH]` **Subresource Integrity (SRI)** on all CDN scripts and styles: `integrity="sha384-..." crossorigin="anonymous"`
- [ ] `[HIGH]` Inventory of all third-party scripts (analytics, chat, ads, tag managers)
- [ ] `[HIGH]` Tag manager access restricted (it can inject arbitrary JS)
- [ ] `[MEDIUM]` Third-party scripts loaded only on pages that need them; never on payment pages unless required
- [ ] `[MEDIUM]` Self-host critical libraries where possible
- [ ] `[MEDIUM]` npm dependencies audited (see file 00)
- [ ] `[MEDIUM]` **PCI DSS 4.0 req. 6.4.3 & 11.6.1** if payment pages exist: script inventory, authorization, and tamper detection

---

## 9. Sensitive Data Exposure in the Client

- [ ] `[CRITICAL]` No API secrets, private keys, or DB credentials in bundles
- [ ] `[CRITICAL]` Next.js: only `NEXT_PUBLIC_*` vars reach the browser; confirm no secrets use that prefix
- [ ] `[HIGH]` Vite/CRA: `VITE_*` / `REACT_APP_*` values treated as public
- [ ] `[HIGH]` Source maps not publicly served in production (or uploaded privately to Sentry)
- [ ] `[HIGH]` API responses don't over-share data that the UI merely hides
- [ ] `[MEDIUM]` Comments, TODOs, internal URLs stripped from production builds
- [ ] `[MEDIUM]` Autocomplete disabled on sensitive fields where appropriate (`autocomplete="off"` / `one-time-code`)

**Test:** Download the production JS bundle and grep for `key`, `secret`, `token`, `password`, `sk_`, `AKIA`, internal hostnames.

---

## 10. Client-Side Validation

- [ ] `[HIGH]` Client validation exists for UX only; every rule duplicated server-side
- [ ] `[MEDIUM]` Hidden/disabled UI elements not relied upon for access control
- [ ] `[MEDIUM]` Role-based UI rendering matches backend permissions (avoid confusion, not a security control)

---

## 11. Open Redirects & URL Handling

- [ ] `[HIGH]` `?redirect=` / `?next=` / `returnUrl` parameters validated against allow-list or relative paths only
- [ ] `[HIGH]` Reject `//evil.com`, `/\evil.com`, `https:evil.com` variants
- [ ] `[MEDIUM]` `target="_blank"` links use `rel="noopener noreferrer"`
- [ ] `[MEDIUM]` `postMessage` handlers validate `event.origin`; senders specify exact target origin

---

## 12. Framework-Specific Risks

### React
- [ ] `dangerouslySetInnerHTML` audited
- [ ] User-controlled URLs in `href` validated (React warns but doesn't block `javascript:` in all versions)
- [ ] No server state leaked via `window.__INITIAL_STATE__` without escaping

### Next.js
- [ ] **Server Actions** treated as public POST endpoints: authenticate and authorize inside each action
- [ ] **Middleware** not the only auth layer (enforce auth again in route handlers / data layer)
- [ ] Next.js kept updated (past middleware bypass CVEs, e.g. `x-middleware-subrequest`)
- [ ] Server Components don't pass secrets or full DB objects to Client Components
- [ ] `server-only` package used for modules that must never reach the client
- [ ] Image optimizer `remotePatterns` restricted (SSRF/abuse)
- [ ] API routes have rate limiting
- [ ] Security headers configured in `next.config.js` or middleware

### Vue / Angular
- [ ] `v-html` / `[innerHTML]` audited
- [ ] Angular sanitizer bypasses justified and reviewed

---

## 13. Forms, Inputs & Bot Protection

- [ ] `[HIGH]` CAPTCHA / bot challenge on signup, login, password reset, contact forms (Cloudflare Turnstile, hCaptcha, reCAPTCHA)
- [ ] `[MEDIUM]` Honeypot fields for low-friction spam protection
- [ ] `[MEDIUM]` Paste disabled only where strictly necessary (never on password fields; it hurts password managers)

---

## 14. Privacy & Accessibility UX

- [ ] `[HIGH]` Cookie consent banner where required (GDPR/ePrivacy, NDPA): non-essential cookies blocked until consent
- [ ] `[HIGH]` Privacy policy and terms linked and accessible
- [ ] `[MEDIUM]` Consent choices recorded and changeable
- [ ] `[MEDIUM]` **WCAG 2.2 AA**: keyboard navigation, contrast, alt text, focus states, labels, error messages
- [ ] `[MEDIUM]` Accessible CAPTCHA alternatives

**Tools:** axe DevTools, Lighthouse, WAVE, Pa11y

---

## 15. Build & Deployment

- [ ] `[HIGH]` Production builds minified, debug flags off, React DevTools production mode
- [ ] `[MEDIUM]` Static hosting buckets not listable
- [ ] `[MEDIUM]` Old deployments/preview URLs not publicly accessible with prod data (Vercel/Netlify previews protected)
- [ ] `[MEDIUM]` Service worker scope and caching reviewed

---

## Tools Summary

| Purpose | Tools |
|---|---|
| Headers & CSP | securityheaders.com, Mozilla Observatory, CSP Evaluator |
| XSS / DAST | OWASP ZAP, Burp Suite, DOM Invader |
| Dependency scan | npm audit, Snyk, Socket.dev |
| Lint | eslint-plugin-security, eslint-plugin-react, eslint-plugin-no-unsanitized |
| Accessibility | axe, Lighthouse, WAVE |
| Secret scan in bundles | TruffleHog, custom grep |

## References
- OWASP Cheat Sheets: XSS Prevention, DOM XSS, CSP, CSRF, HTML5 Security, Clickjacking
- MDN Web Security docs
- WCAG 2.2
