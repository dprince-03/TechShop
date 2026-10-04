#!/usr/bin/env bash
# Quick defensive grep sweep for common security smells. Hits are leads, not confirmed findings.
# Usage: quick_scan.sh [path]   (default: current dir)
set -u
ROOT="${1:-.}"
EXCL="--exclude-dir=node_modules --exclude-dir=.git --exclude-dir=vendor --exclude-dir=dist --exclude-dir=build --exclude-dir=.next --exclude-dir=coverage"

section() { printf "\n==== %s ====\n" "$1"; }
scan() { grep -rnIE $EXCL "$2" "$ROOT" 2>/dev/null | head -n "${3:-25}" || true; }

section "Possible hardcoded secrets"
scan secrets '(api[_-]?key|secret|passw(or)?d|private[_-]?key|access[_-]?token)\s*[:=]\s*["'"'"'][^"'"'"'$\{]{8,}'
scan aws 'AKIA[0-9A-Z]{16}'
scan pk '-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----'
scan stripe 'sk_(live|test)_[0-9a-zA-Z]{16,}'

section "Committed env files"
find "$ROOT" -name ".env*" -not -name ".env.example" -not -path "*/node_modules/*" -not -path "*/.git/*" 2>/dev/null | head -20

section "Dangerous code execution"
scan exec '\beval\(|new Function\(|child_process|exec\.Command\(|Runtime\.getRuntime\(\)\.exec'

section "Possible SQL string building"
scan sql '(SELECT|INSERT|UPDATE|DELETE)[^;]*(\+\s*[a-zA-Z_]|\$\{|%s|fmt\.Sprintf)'

section "XSS sinks"
scan xss 'dangerouslySetInnerHTML|\.innerHTML\s*=|v-html|document\.write\('

section "Weak crypto / randomness"
scan crypto '\b(md5|sha1)\b|createHash\(["'"'"'](md5|sha1)|Math\.random\(\)|math/rand"'

section "JWT pitfalls"
scan jwt 'jwt\.decode\(|algorithms?\s*[:=]\s*\[?["'"'"']none|ignoreExpiration\s*:\s*true'

section "Permissive CORS"
scan cors 'Access-Control-Allow-Origin["'"'"']?\s*[:,]\s*["'"'"']\*|origin\s*:\s*(true|["'"'"']\*["'"'"'])|AllowAllOrigins\s*[:=]\s*true'

section "Debug / insecure flags"
scan debug 'gin\.DebugMode|debug\s*[:=]\s*true|NODE_TLS_REJECT_UNAUTHORIZED|InsecureSkipVerify:\s*true|rejectUnauthorized:\s*false|android:debuggable="true"|usesCleartextTraffic="true"|NSAllowsArbitraryLoads'

section "Secrets exposed to browser (Next.js/Vite)"
scan public 'NEXT_PUBLIC_[A-Z_]*(SECRET|PRIVATE|PASSWORD|KEY)|VITE_[A-Z_]*(SECRET|PRIVATE|PASSWORD)'

section "Dockerfile running as root (no USER)"
for f in $(find "$ROOT" -name "Dockerfile*" -not -path "*/node_modules/*" 2>/dev/null); do
  grep -q '^USER ' "$f" || echo "$f: no USER directive"
done

printf "\nDone. Verify each hit manually; follow up with semgrep, gitleaks, trivy, npm audit / govulncheck.\n"
