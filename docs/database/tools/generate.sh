#!/usr/bin/env bash
# Validates docs/database/schema.sql and regenerates the generated docs and diagrams.
#
#   1. starts a throwaway postgres:17-alpine container (never the project's own database)
#   2. loads schema.sql (must have zero errors)
#   3. runs tools/constraint-tests.sql (every rule must hold)
#   4. runs coverage checks (comments, FK indexes, updated_at triggers)
#   5. regenerates erd.md from the live schema
#   6. renders every Mermaid diagram to diagrams/*.svg (fails on any invalid diagram)
#   7. removes the container
#
# Requirements: Docker, Node 22+, Google Chrome/Chromium for rendering.
# Usage: docs/database/tools/generate.sh [--no-render]
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
DOCS="$(dirname "$HERE")"
CONTAINER="techshop-schema-check-$$"
RENDER=1
[[ "${1:-}" == "--no-render" ]] && RENDER=0

cleanup() { docker rm -f "$CONTAINER" >/dev/null 2>&1 || true; }
trap cleanup EXIT

psql_q() { docker exec -i "$CONTAINER" psql -U postgres -v ON_ERROR_STOP=1 -q "$@"; }

echo "→ Starting throwaway Postgres ($CONTAINER)"
docker run -d --rm --name "$CONTAINER" -e POSTGRES_PASSWORD=check postgres:17-alpine >/dev/null
for _ in $(seq 1 60); do
  docker exec "$CONTAINER" pg_isready -U postgres >/dev/null 2>&1 && break
  sleep 1
done
sleep 1

echo "→ Loading schema.sql"
psql_q < "$DOCS/schema.sql"

echo "→ Constraint tests"
psql_q < "$HERE/constraint-tests.sql" 2>&1 | grep -cE "PASS" | xargs -I{} echo "  {} tests passed"

echo "→ Coverage checks"
psql_q -At <<'SQL'
do $$
declare n int;
begin
  select count(*) into n from pg_class c join pg_namespace s on s.oid = c.relnamespace
   where s.nspname = 'public' and c.relkind = 'r' and coalesce(obj_description(c.oid), '') !~ '^\[[a-z-]+\] ';
  if n > 0 then raise exception '% tables lack a [domain] comment', n; end if;

  select count(*) into n from pg_constraint c where contype = 'f'
   and not exists (select 1 from pg_index i where i.indrelid = c.conrelid and i.indkey[0] = c.conkey[1]);
  if n > 0 then raise exception '% foreign keys are not indexed', n; end if;

  select count(*) into n from information_schema.columns col
   where table_schema = 'public' and column_name = 'updated_at'
     and not exists (select 1 from pg_trigger t where t.tgname = col.table_name || '_set_updated_at');
  if n > 0 then raise exception '% tables with updated_at lack the trigger', n; end if;
end $$;
select '  ' || count(*) || ' tables, ' || (select count(*) from pg_constraint where contype = 'f') || ' foreign keys — all commented, indexed and triggered'
from pg_tables where schemaname = 'public';
SQL

echo "→ Generating erd.md"
psql_q -At < "$HERE/introspect.sql" | node "$HERE/generate.mjs" "$DOCS/erd.md"

if [[ $RENDER == 1 ]]; then
  echo "→ Rendering diagrams"
  MMDC="${MMDC:-}"
  if [[ -z "$MMDC" ]]; then
    MMDC_DIR="$(mktemp -d)"
    (cd "$MMDC_DIR" && npm init -y >/dev/null && PUPPETEER_SKIP_DOWNLOAD=1 npm i --silent @mermaid-js/mermaid-cli >/dev/null)
    MMDC="$MMDC_DIR/node_modules/.bin/mmdc"
  fi
  CHROME="$(command -v google-chrome || command -v chromium || command -v chromium-browser || true)"
  PUPPETEER_CONFIG="$(mktemp --suffix=.json)"
  printf '{"executablePath":"%s","args":["--no-sandbox"]}' "$CHROME" > "$PUPPETEER_CONFIG"
  MMDC="$MMDC" PUPPETEER_CONFIG="$PUPPETEER_CONFIG" node "$HERE/render.mjs" "$DOCS/diagrams" \
    "$DOCS/../database.md" "$DOCS/erd.md" "$DOCS/states.md" "$DOCS/flows.md" "$DOCS/access.md"
fi

echo "✓ Done"
