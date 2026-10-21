import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { Status, Table } from "@techshop/ui/components";

import { listingStatus } from "@/features/listings/options";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { naira } from "@/lib/format";

export const metadata: Metadata = {
  title: "Listings | TechShop Seller Centre",
};

const filters = [
  { value: "", label: "All" },
  { value: "active", label: "On sale" },
  { value: "in_review", label: "In review" },
  { value: "changes_requested", label: "Changes requested" },
  { value: "paused", label: "Paused" },
] as const;

export default async function ListingsPage(
  props: PageProps<"/centre/[shopId]/listings">,
) {
  await connection();
  const { shopId } = await props.params;
  const sp = await props.searchParams;
  const status = filters.find((f) => f.value === sp.status)?.value ?? "";
  const here = `/centre/${shopId}/listings`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const page = await api.get("/api/v1/seller/shops/{shopId}/listings", {
    path: { shopId },
    query: {
      status: status || undefined,
      cursor: typeof sp.cursor === "string" ? sp.cursor : undefined,
    },
    init: { cache: "no-store" },
  });
  const canSell = shop.status === "active" && shop.role !== "finance";
  return (
    <ShopPage
      shop={shop}
      section="listings"
      title="Listings"
      actions={
        canSell ? (
          <Link href={`${here}/new`} className="btn btn--primary">
            New listing
          </Link>
        ) : undefined
      }
    >
      {shop.status !== "active" ? (
        <p className="alert alert--warning">
          You can list products once your shop is verified.
        </p>
      ) : null}
      <nav aria-label="Filter listings" className="cluster">
        {filters.map((f) => (
          <Link
            key={f.value}
            href={f.value ? `${here}?status=${f.value}` : here}
            aria-current={f.value === status ? "page" : undefined}
            className={`btn btn--sm ${f.value === status ? "btn--inverse" : "btn--ghost"}`}
          >
            {f.label}
          </Link>
        ))}
      </nav>
      <Table
        caption="Your listings"
        rows={page.items ?? []}
        rowKey={(l) => l.id}
        href={(l) => `${here}/${l.id}`}
        empty={{
          title: "No listings here",
          body: canSell
            ? "Create a listing to start selling."
            : "Listings you create appear here.",
        }}
        columns={[
          {
            label: "Product",
            cell: (l) => `${l.productName} · ${l.variantName}`,
          },
          { label: "Condition", cell: (l) => l.condition.replace("_", " ") },
          { label: "Price", numeric: true, cell: (l) => naira(l.price) },
          { label: "Stock", numeric: true, cell: (l) => l.sellerStock ?? "—" },
          {
            label: "Status",
            cell: (l) => (
              <Status
                label={listingStatus[l.status].label}
                tone={listingStatus[l.status].tone}
              />
            ),
          },
        ]}
      />
      {page.page.nextCursor ? (
        <Link
          href={`${here}?${new URLSearchParams({ ...(status ? { status } : {}), cursor: page.page.nextCursor })}`}
          className="link link--arrow"
        >
          Older listings
        </Link>
      ) : null}
    </ShopPage>
  );
}
