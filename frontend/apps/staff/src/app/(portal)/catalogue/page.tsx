import Link from "next/link";
import { connection } from "next/server";

import { sites } from "@techshop/ui/sites";

import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { Status, humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { Tabs } from "@/components/Tabs";
import { decideSuggestion, reindex } from "@/features/catalogue/actions";
import { ModerationActions } from "@/features/catalogue/components/ModerationActions";
import {
  moderationQueue,
  searchProducts,
  suggestions,
} from "@/features/catalogue/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Catalogue | TechShop Staff" };

const TABS = [
  { key: "products", label: "Products" },
  { key: "moderation", label: "Moderation" },
  { key: "suggestions", label: "Seller suggestions" },
];

export default async function CataloguePage(props: PageProps<"/catalogue">) {
  await connection();
  const me = await requireStaff("catalog.view");
  const sp = await props.searchParams;
  const tab = TABS.some((t) => t.key === param(sp.tab))
    ? param(sp.tab)!
    : "products";
  const created = param(sp.created);
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Catalogue"
        description="Products, listing moderation and seller change requests."
        actions={
          <>
            {can(me, "catalog.edit") ? (
              <Link href="/catalogue/new" className="btn btn--inverse btn--sm">
                Add a product
              </Link>
            ) : null}
            {can(me, "catalog.edit") ? (
              <ActionButton label="Rebuild search index" action={reindex} />
            ) : null}
          </>
        }
      />
      {created ? (
        <p className="alert alert--success" role="status">
          Product created.{" "}
          <a className="link" href={`${sites.market}/p/${created}`}>
            View it on the market
          </a>{" "}
          (drafts aren’t visible there).
        </p>
      ) : null}
      <Tabs
        label="Catalogue sections"
        current={tab}
        items={TABS.map((t) => ({ ...t, href: `/catalogue?tab=${t.key}` }))}
      />
      {tab === "products" ? (
        <Products q={param(sp.q)} cursor={param(sp.cursor)} />
      ) : null}
      {tab === "moderation" ? (
        <Moderation canModerate={can(me, "catalog.moderate")} />
      ) : null}
      {tab === "suggestions" ? (
        <Suggestions canDecide={can(me, "catalog.moderate")} />
      ) : null}
    </div>
  );
}

async function Products({ q, cursor }: { q?: string; cursor?: string }) {
  const r = await searchProducts(q?.slice(0, 80), cursor);
  return (
    <section className="stack">
      <form method="get" className="cluster" role="search">
        <input type="hidden" name="tab" value="products" />
        <label className="field" style={{ flex: "1 1 18rem" }}>
          <span className="field__label">Search products</span>
          <input name="q" defaultValue={q} className="input" />
        </label>
        <button
          className="btn btn--secondary"
          style={{ alignSelf: "flex-end" }}
        >
          Search
        </button>
      </form>
      <p className="text-sm text-muted" role="status">
        {r.total} products
      </p>
      <Table
        caption="Products"
        rows={r.items ?? []}
        rowKey={(p) => p.id}
        href={(p) => `${sites.market}/p/${p.slug}`}
        columns={[
          { label: "Product", cell: (p) => p.name },
          { label: "Brand", cell: (p) => p.brandName ?? "—" },
          { label: "Offers", cell: (p) => p.offerCount, numeric: true },
          {
            label: "Stock",
            cell: (p) => (
              <Status
                label={
                  p.inStock
                    ? p.stockBand === "low"
                      ? "Low"
                      : "In stock"
                    : "Out of stock"
                }
                tone={
                  p.inStock
                    ? p.stockBand === "low"
                      ? "warning"
                      : "success"
                    : "danger"
                }
              />
            ),
          },
          {
            label: "From",
            cell: (p) => formatKobo(p.promoPrice ?? p.price ?? 0),
            numeric: true,
          },
        ]}
      />
      {r.page.hasMore && r.page.nextCursor ? (
        <Link
          className="btn btn--secondary"
          href={`/catalogue?tab=products${q ? `&q=${encodeURIComponent(q)}` : ""}&cursor=${encodeURIComponent(r.page.nextCursor)}`}
        >
          More products
        </Link>
      ) : null}
    </section>
  );
}

async function Moderation({ canModerate }: { canModerate: boolean }) {
  const rows = await moderationQueue();
  return (
    <Table
      caption="Listings waiting for review"
      rows={rows}
      rowKey={(l) => l.id}
      empty={{
        title: "Nothing to review",
        body: "New and changed seller listings appear here.",
      }}
      columns={[
        {
          label: "Listing",
          cell: (l) => `${l.productName} · ${l.variantName}`,
        },
        { label: "Seller", cell: (l) => l.sellerName },
        { label: "Condition", cell: (l) => humanise(l.condition) },
        {
          label: "Why flagged",
          cell: (l) => (l.reasons ?? []).map(humanise).join(", ") || "—",
        },
        { label: "Risk", cell: (l) => l.riskScore, numeric: true },
        { label: "Price", cell: (l) => formatKobo(l.priceKobo), numeric: true },
        {
          label: "Decision",
          cell: (l) => (canModerate ? <ModerationActions id={l.id} /> : null),
        },
      ]}
    />
  );
}

async function Suggestions({ canDecide }: { canDecide: boolean }) {
  const rows = await suggestions();
  return (
    <Table
      caption="Seller suggestions"
      rows={rows}
      rowKey={(s) => s.id}
      empty={{
        title: "No suggestions",
        body: "Sellers’ proposed product changes appear here.",
      }}
      columns={[
        { label: "Kind", cell: (s) => humanise(s.kind) },
        {
          label: "Details",
          cell: (s) => (
            <code className="text-xs">
              {JSON.stringify(s.payload).slice(0, 120)}
            </code>
          ),
        },
        { label: "Sent", cell: (s) => formatDateTime(s.createdAt) },
        { label: "Status", cell: (s) => humanise(s.status) },
        {
          label: "Decision",
          cell: (s) =>
            canDecide && s.status === "pending" ? (
              <div className="cluster">
                <ActionButton
                  label="Accept"
                  variant="inverse"
                  action={decideSuggestion.bind(null, s.id, "accepted")}
                />
                <ActionButton
                  label="Reject"
                  variant="ghost"
                  action={decideSuggestion.bind(null, s.id, "rejected")}
                />
              </div>
            ) : null,
        },
      ]}
    />
  );
}
