import type { Metadata } from "next";
import { connection } from "next/server";

import { Status, Table } from "@techshop/ui/components";

import { RowAction } from "@/components/RowAction";
import { acceptOrder, packOrder } from "@/features/orders/actions";
import { CancelOrderForm } from "@/features/orders/components/CancelOrderForm";
import { DropoffPanel } from "@/features/orders/components/DropoffPanel";
import { orderStatus } from "@/features/orders/status";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime, naira } from "@/lib/format";

export const metadata: Metadata = { title: "Order | TechShop Seller Centre" };

export default async function OrderPage(
  props: PageProps<"/centre/[shopId]/orders/[orderId]">,
) {
  await connection();
  const { shopId, orderId } = await props.params;
  const here = `/centre/${shopId}/orders/${orderId}`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const o = await api.get("/api/v1/seller/shops/{shopId}/orders/{orderId}", {
    path: { shopId, orderId },
    init: { cache: "no-store" },
  });
  const st = orderStatus[o.status as keyof typeof orderStatus] ?? {
    label: o.status,
    tone: "neutral" as const,
  };
  const manage = shop.role !== "finance";
  const points =
    o.status === "packed"
      ? await api.get("/api/v1/seller/dropoff-points", {
          init: { cache: "no-store" },
        })
      : [];
  return (
    <ShopPage
      shop={shop}
      section="orders"
      title={`Order ${o.orderNumber}`}
      actions={<Status label={st.label} tone={st.tone} />}
    >
      <section aria-labelledby="what-title" className="stack">
        <h2 id="what-title" className="text-xl">
          Items
        </h2>
        <Table
          caption="Items to ship"
          rows={o.lines ?? []}
          rowKey={(l) => l.sku + l.productName}
          columns={[
            {
              label: "Item",
              cell: (l) =>
                `${l.productName} · ${l.variantName}${l.cancelledAt ? " (cancelled)" : ""}`,
            },
            { label: "SKU", cell: (l) => l.sku },
            { label: "Qty", numeric: true, cell: (l) => l.quantity },
            { label: "Total", numeric: true, cell: (l) => naira(l.lineTotal) },
          ]}
        />
        <p className="text-sm text-muted">
          For {o.recipientFirstName} in {o.city}. Placed{" "}
          {dateTime.format(new Date(o.placedAt))}.
          {o.status === "pending" && o.acceptBy
            ? ` Accept by ${dateTime.format(new Date(o.acceptBy))}.`
            : ""}
          {o.status === "accepted" && o.packBy
            ? ` Pack by ${dateTime.format(new Date(o.packBy))}.`
            : ""}
        </p>
        {o.cancelledReason ? (
          <p className="alert alert--warning">{o.cancelledReason}</p>
        ) : null}
      </section>
      {manage && (o.status === "pending" || o.status === "accepted") ? (
        <section aria-labelledby="do-title" className="card card--roomy stack">
          <h2 id="do-title" className="text-xl">
            {o.status === "pending" ? "Accept the order" : "Pack the order"}
          </h2>
          {o.status === "pending" ? (
            <RowAction
              action={acceptOrder.bind(null, shop.id, o.id)}
              label="Accept order"
              pendingLabel="Accepting…"
            />
          ) : (
            <RowAction
              action={packOrder.bind(null, shop.id, o.id)}
              label="Mark as packed"
              pendingLabel="Saving…"
            />
          )}
          <details className="stack">
            <summary className="link">Can’t fulfil it?</summary>
            <div className="stack">
              <p className="text-sm text-muted">
                Cancelling refunds the customer and counts against your shop’s
                performance.
              </p>
              <CancelOrderForm shopId={shop.id} orderId={o.id} />
            </div>
          </details>
        </section>
      ) : null}
      {o.status === "packed" && o.label ? (
        <DropoffPanel
          label={o.label}
          handoverBy={o.handoverBy}
          points={points ?? []}
        />
      ) : null}
    </ShopPage>
  );
}
