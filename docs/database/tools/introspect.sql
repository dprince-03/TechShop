-- Emits the loaded schema as one JSON document for generate.mjs:
-- tables (domain + purpose from comments), columns with keys, and foreign keys.
with tables as (
  select c.oid, c.relname as name,
         substring(obj_description(c.oid) from '^\[([a-z-]+)\]') as domain,
         regexp_replace(obj_description(c.oid), '^\[[a-z-]+\]\s*', '') as purpose
  from pg_class c
  join pg_namespace n on n.oid = c.relnamespace
  where n.nspname = 'public' and c.relkind = 'r'
),
pk as (
  select i.indrelid, unnest(i.indkey) as attnum from pg_index i where i.indisprimary
),
uk as (
  -- single-column unique constraints/indexes (not partial)
  select i.indrelid, i.indkey[0] as attnum from pg_index i
  where i.indisunique and not i.indisprimary and i.indnkeyatts = 1 and i.indpred is null and i.indkey[0] <> 0
),
fkcol as (
  select conrelid, conkey[1] as attnum from pg_constraint where contype = 'f' and array_length(conkey, 1) = 1
),
cols as (
  select t.name as table_name, a.attnum, a.attname as name,
         format_type(a.atttypid, a.atttypmod) as type,
         not a.attnotnull as nullable,
         exists (select 1 from pk where pk.indrelid = t.oid and pk.attnum = a.attnum) as is_pk,
         exists (select 1 from uk where uk.indrelid = t.oid and uk.attnum = a.attnum) as is_unique,
         exists (select 1 from fkcol f where f.conrelid = t.oid and f.attnum = a.attnum) as is_fk
  from tables t
  join pg_attribute a on a.attrelid = t.oid and a.attnum > 0 and not a.attisdropped
),
fks as (
  select src.relname as child, a.attname as column_name, dst.relname as parent,
         not a.attnotnull as nullable,
         exists (select 1 from uk where uk.indrelid = c.conrelid and uk.attnum = c.conkey[1])
           or exists (select 1 from pg_index i where i.indrelid = c.conrelid and i.indisprimary and i.indnkeyatts = 1 and i.indkey[0] = c.conkey[1]) as is_unique
  from pg_constraint c
  join pg_class src on src.oid = c.conrelid
  join pg_class dst on dst.oid = c.confrelid
  join pg_attribute a on a.attrelid = c.conrelid and a.attnum = c.conkey[1]
  where c.contype = 'f' and array_length(c.conkey, 1) = 1
)
select json_build_object(
  'tables', (select json_agg(json_build_object('name', name, 'domain', domain, 'purpose', purpose) order by name) from tables),
  'columns', (select json_agg(json_build_object('table', table_name, 'name', name, 'type', type, 'nullable', nullable,
                'pk', is_pk, 'unique', is_unique, 'fk', is_fk) order by table_name, attnum) from cols),
  'fks', (select json_agg(json_build_object('child', child, 'column', column_name, 'parent', parent,
            'nullable', nullable, 'unique', is_unique) order by child, column_name) from fks)
);
