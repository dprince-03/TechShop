#!/usr/bin/env bash
# Validates the database migrations and regenerates the generated docs and diagrams.
#
#   1. starts a throwaway postgres:17-alpine container (never the project's own database)
#   2. loads the Up section of every goose migration in backend/db/migrations, in order (zero errors)
#   3. runs tools/constraint-tests.sql (every rule must hold)
#   4. runs coverage checks (table comments, FK indexes, updated_at triggers) across all schemas
#   5. regenerates schema.sql (pg_dump --schema-only) and erd.md from the live schema
#   6. renders every Mermaid diagram to diagrams/*.svg (fails on any invalid diagram)
#   7. removes the container
#
# The full goose cycle (up → down → up) is tested separately with the goose binary; this script
# only needs Docker, Node 22+ and (for rendering) Chrome/Chromium.
# Usage: docs/database/tools/generate.sh [--no-render]
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
DOCS="$(dirname "$HERE")"
MIGRATIONS="$(cd "$DOCS/../../backend/db/migrations" && pwd)"
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

echo "→ Loading migrations (Up sections)"
for f in "$MIGRATIONS"/*.sql; do
  # Everything between '-- +goose Up' and '-- +goose Down'; the StatementBegin/End markers are comments.
  awk '/^-- \+goose Up/{up=1; next} /^-- \+goose Down/{up=0} up' "$f" | psql_q
done

echo "→ Constraint tests"
psql_q < "$HERE/constraint-tests.sql" 2>&1 | grep -cE "PASS" | xargs -I{} echo "  {} tests passed"

echo "→ Coverage checks"
psql_q -At <<'SQL'
create temp view app_tables as
  select c.oid, n.nspname, c.relname from pg_class c join pg_namespace n on n.oid = c.relnamespace
  where n.nspname not in ('public', 'information_schema') and n.nspname not like 'pg\_%'
    and c.relkind in ('r', 'p') and not c.relispartition;
do $$
declare n int;
begin
  select count(*) into n from app_tables t where coalesce(obj_description(t.oid), '') !~ '^\[[a-z0-9_-]+\] ';
  if n > 0 then raise exception '% tables lack a [domain] comment', n; end if;

  select count(*) into n from pg_constraint c where contype = 'f' and conparentid = 0
   and not exists (select 1 from pg_index i where i.indrelid = c.conrelid and i.indkey[0] = c.conkey[1]);
  if n > 0 then raise exception '% foreign keys are not indexed', n; end if;

  select count(*) into n from app_tables t
   where exists (select 1 from pg_attribute a where a.attrelid = t.oid and a.attname = 'updated_at' and not a.attisdropped)
     and not exists (select 1 from pg_trigger g where g.tgrelid = t.oid and g.tgname = t.relname || '_set_updated_at');
  if n > 0 then raise exception '% tables with updated_at lack the trigger', n; end if;
end $$;
select '  ' || count(*) || ' tables in ' || count(distinct nspname) || ' schemas, '
       || (select count(*) from pg_constraint where contype = 'f' and conparentid = 0)
       || ' foreign keys — all commented, indexed and triggered'
from app_tables;
SQL

echo "→ Regenerating schema.sql"
{
  printf '%s\n' '-- =============================================================================' \
    '-- TechShop database schema (PostgreSQL 17). GENERATED: do not edit by hand.' '--' \
    '-- Source of truth: backend/db/migrations (goose). This file is a pg_dump --schema-only' \
    '-- of a scratch database after `goose up`, kept so the whole schema can be read in one' \
    '-- place. Regenerate it with docs/database/tools/generate.sh after changing migrations.' \
    '-- =============================================================================' ''
  docker exec "$CONTAINER" pg_dump -U postgres --schema-only --no-owner --no-privileges --exclude-schema=public postgres \
    | grep -v '^SET \|^SELECT pg_catalog.set_config\|^-- Dumped \|^\\restrict\|^\\unrestrict'
} > "$DOCS/schema.sql"

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
  PUPPETEER_CONFIG="${PUPPETEER_CONFIG:-$(mktemp --suffix=.json)}"
  [[ -s "$PUPPETEER_CONFIG" ]] || printf '{"executablePath":"%s","args":["--no-sandbox"]}' "$CHROME" > "$PUPPETEER_CONFIG"
  MMDC="$MMDC" PUPPETEER_CONFIG="$PUPPETEER_CONFIG" node "$HERE/render.mjs" "$DOCS/diagrams" \
    "$DOCS/../database.md" "$DOCS/erd.md" "$DOCS/states.md" "$DOCS/flows.md" "$DOCS/access.md"
fi

echo "✓ Done"
