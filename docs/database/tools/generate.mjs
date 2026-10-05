// Generates docs/database/erd.md from the introspected schema JSON (stdin).
// Usage: psql … -f introspect.sql | node generate.mjs ../erd.md
import { readFileSync, writeFileSync } from "node:fs";

const out = process.argv[2];
const { tables, columns, fks } = JSON.parse(readFileSync(0, "utf8"));

// One entry per Postgres schema (backend/db/migrations/00002_foundation.sql), in reading order.
const DOMAINS = [
  ["identity", "Identity & access"],
  ["sellers", "Sellers"],
  ["b2b", "Business buyers (B2B)"],
  ["catalog", "Catalogue & reviews"],
  ["search", "Search"],
  ["inventory", "Inventory"],
  ["purchasing", "Purchasing"],
  ["pos", "Point of sale"],
  ["sales", "Sales"],
  ["fulfilment", "Fulfilment & shipping"],
  ["logistics", "Logistics"],
  ["payments", "Payments"],
  ["finance", "Finance & ledger"],
  ["aftersales", "After-sales"],
  ["messaging", "Messaging"],
  ["marketing", "Marketing"],
  ["content", "Content"],
  ["support", "Support"],
  ["risk", "Trust & safety"],
  ["cars", "Cars"],
  ["personalisation", "Personalisation & recommendations"],
  ["platform", "Platform"],
];

const dupes = tables.map((t) => t.name).filter((n, i, a) => a.indexOf(n) !== i);
if (dupes.length) throw new Error(`Table names must be unique across schemas for the diagrams: ${dupes.join(", ")}`);

const unknown = tables.filter((t) => !DOMAINS.some(([k]) => k === t.domain));
if (unknown.length) throw new Error(`Tables with unknown domain: ${unknown.map((t) => t.name).join(", ")}`);

/** Mermaid attribute types must be a single token. */
const shortType = (t) =>
  t
    .replace("timestamp with time zone", "timestamptz")
    .replace("character", "char")
    .replace(/\(.*\)/, "")
    .replace(/\[\]$/, "_array")
    .replace(/\s+/g, "_");

const colsOf = (table) => columns.filter((c) => c.table === table);

function entity(table) {
  const lines = colsOf(table).map((c) => {
    const keys = [c.pk && "PK", c.fk && "FK", c.unique && !c.pk && "UK"].filter(Boolean).join(", ");
    const note = c.nullable && !c.pk ? ' "nullable"' : "";
    return `    ${shortType(c.type)} ${c.name}${keys ? ` ${keys}` : ""}${note}`;
  });
  return `  ${table} {\n${lines.join("\n")}\n  }`;
}

/** parent ||--o{ child : "column"  (left: parent optional?  right: one or many children?) */
function relation(fk) {
  const left = fk.nullable ? "|o" : "||";
  const right = fk.unique ? "o|" : "o{";
  return `  ${fk.parent} ${left}--${right} ${fk.child} : "${fk.column}"`;
}

const sections = [];

// Table index
const domainName = Object.fromEntries(DOMAINS);
const index = [
  "## Table index",
  "",
  `${tables.length} tables, ${fks.length} foreign keys. Generated from the live schema.`,
  "",
  "| Table | Domain | Columns | Purpose |",
  "|---|---|---|---|",
  ...DOMAINS.flatMap(([key]) =>
    tables
      .filter((t) => t.domain === key)
      .map((t) => `| \`${key}.${t.name}\` | ${domainName[key]} | ${colsOf(t.name).length} | ${t.purpose.replace(/\|/g, "\\|")} |`),
  ),
];

// Overview: all tables, keys only. users/files are referenced by most tables; omitted for readability.
const HUBS = new Set(["users", "files"]);
const overviewRels = fks.filter((f) => !HUBS.has(f.parent));
const overview = [
  "## Overview: all tables",
  "",
  "Every table and relationship, without columns. `users` and `files` are referenced by most tables (who created/approved something, attachments); those links are left out here and shown in the domain diagrams.",
  "",
  "```mermaid",
  "erDiagram",
  ...tables.map((t) => {
    const pk = colsOf(t.name).filter((c) => c.pk);
    return `  ${t.name} {\n${pk.map((c) => `    ${shortType(c.type)} ${c.name} PK`).join("\n")}\n  }`;
  }),
  ...overviewRels.map(relation),
  "```",
];

// One diagram per domain: its tables with columns, plus links to parents in other domains.
const domainSections = DOMAINS.map(([key, label]) => {
  const own = tables.filter((t) => t.domain === key).map((t) => t.name);
  const rels = fks.filter((f) => own.includes(f.child));
  const external = [...new Set(rels.map((f) => f.parent).filter((p) => !own.includes(p)))].sort();
  return [
    `## ${label}`,
    "",
    own.map((t) => `- \`${key}.${t}\`: ${tables.find((x) => x.name === t).purpose}`).join("\n"),
    "",
    external.length ? `Links to other domains: ${external.map((e) => `\`${e}\``).join(", ")} (shown without columns).` : "",
    "",
    "```mermaid",
    "erDiagram",
    ...own.map(entity),
    ...rels.map(relation),
    "```",
  ].join("\n");
});

const legend = [
  "## How to read these diagrams",
  "",
  "| Notation | Meaning |",
  "|---|---|",
  "| `PK` / `FK` / `UK` | Primary key / foreign key / unique |",
  '| `"nullable"` | Column may be empty |',
  "| `\\|\\|--o{` | Each child row has exactly one parent; a parent has zero or many children |",
  "| `\\|o--o{` | The parent link is optional (nullable foreign key) |",
  "| `\\|\\|--o\\|` | One-to-one (the foreign key is unique) |",
  "| Label on a line | The foreign-key column on the child table |",
];

writeFileSync(
  out,
  [
    "# TechShop entity-relationship diagrams",
    "",
    "> **Generated** by `docs/database/tools/generate.sh` from the goose migrations (`backend/db/migrations`) loaded into Postgres 17. Do not edit by hand. Change a migration and regenerate. Each section is one Postgres schema; tables are shown without the schema prefix.",
    "",
    "Overview and design notes: [`../database.md`](../database.md).",
    "",
    legend.join("\n"),
    "",
    domainSections.join("\n\n"),
    "",
    overview.join("\n"),
    "",
    index.join("\n"),
    "",
  ].join("\n"),
);

console.log(`Wrote ${out}: ${tables.length} tables, ${fks.length} relationships, ${DOMAINS.length + 1} diagrams.`);
