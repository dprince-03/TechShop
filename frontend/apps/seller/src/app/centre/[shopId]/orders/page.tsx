import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { Status, Table } from "@techshop/ui/components";

import { orderStatus } from "@/features/orders/status";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime } from "@/lib/format";

export const metadata: Metadata = { title: "Orders | TechShop Seller Centre" };

const filters = [
  { value: "pending", label: "To accept" },
  { value: "accepted", label: "To pack" },
  { value: "packed", label: "Packed" },
  { value: "", label: "All" },
] as const;

export default async function OrdersPage(
  props: PageProps<"/centre/[shopId]/orders">,
) {
  await connection();
  const { shopId } = await props.params;
  const sp = await props.searchParams;
  const status = filters.find((f) => f.value === sp.status)?.value ?? "";
  const here = `/centre/${shopId}/orders`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const page = await api.get("/api/v1/seller/shops/{shopId}/orders", {
    path: { shopId },
    query: {
      status: status || undefined,
      cursor: typeof sp.cursor === "string" ? sp.cursor : undefined,
    },
    init: { cache: "no-store" },
  });
  return (
    <ShopPage
      shop={shop}
      section="orders"
      title="Orders"
      lead="Accept new orders within 24 hours and pack them within your handling time, or they’re cancelled automatically."
    >
      <nav aria-label="Filter orders" className="cluster">
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
        caption="Orders"
        rows={page.items ?? []}
        rowKey={(o) => o.id}
        href={(o) => `${here}/${o.id}`}
        empty={{
          title: "No orders here",
          body: "Orders for your listings appear here as soon as they’re paid.",
        }}
        columns={[
          { label: "Order", cell: (o) => o.orderNumber },
          {
            label: "Placed",
            cell: (o) => dateTime.format(new Date(o.placedAt)),
          },
          {
            label: "Deliver to",
            cell: (o) => `${o.recipientFirstName}, ${o.city}`,
          },
          {
            label: "Status",
            cell: (o) => (
              <Status
                label={
                  orderStatus[o.status as keyof typeof orderStatus]?.label ??
                  o.status
                }
                tone={
                  orderStatus[o.status as keyof typeof orderStatus]?.tone ??
                  "neutral"
                }
              />
            ),
          },
          {
            label: "Deadline",
            cell: (o) =>
              o.status === "pending" && o.acceptBy
                ? `Accept by ${dateTime.format(new Date(o.acceptBy))}`
                : o.status === "accepted" && o.packBy
                  ? `Pack by ${dateTime.format(new Date(o.packBy))}`
                  : "—",
          },
        ]}
      />
      {page.page.nextCursor ? (
        <Link
          href={`${here}?${new URLSearchParams({ ...(status ? { status } : {}), cursor: page.page.nextCursor })}`}
          className="link link--arrow"
        >
          Older orders
        </Link>
      ) : null}
    </ShopPage>
  );
}
