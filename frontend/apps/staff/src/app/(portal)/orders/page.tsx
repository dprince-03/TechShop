import Link from "next/link";
import { connection } from "next/server";

import { PageHead } from "@/components/PageHead";
import { Status, humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { Tabs } from "@/components/Tabs";
import { listOrders } from "@/features/orders/server";
import { orderTone } from "@/features/orders/status";
import { requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Orders | TechShop Staff" };

const TABS = [
  { key: "", label: "All" },
  { key: "pending_payment", label: "Awaiting payment" },
  { key: "paid", label: "Paid" },
  { key: "processing", label: "Processing" },
  { key: "shipped", label: "Shipped" },
  { key: "delivered", label: "Delivered" },
  { key: "cancelled", label: "Cancelled" },
  { key: "refunded", label: "Refunded" },
];

export default async function OrdersPage(props: PageProps<"/orders">) {
  await connection();
  await requireStaff("orders.view");
  const sp = await props.searchParams;
  const status = param(sp.status) ?? "";
  const q = param(sp.q)?.trim().slice(0, 60);
  const cursor = param(sp.cursor);
  const page = await listOrders({ status: status || undefined, q, cursor });
  const href = (s: string) =>
    `/orders${s || q ? "?" : ""}${new URLSearchParams({ ...(s ? { status: s } : {}), ...(q ? { q } : {}) })}`;
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Orders"
        description="Every order from Market, Wholesale and the app. Contact details are masked."
      />
      <form method="get" className="cluster" role="search">
        {status ? <input type="hidden" name="status" value={status} /> : null}
        <label className="field" style={{ flex: "1 1 18rem" }}>
          <span className="field__label">Search</span>
          <input
            name="q"
            defaultValue={q}
            className="input"
            placeholder="Order number or customer name"
          />
        </label>
        <button
          className="btn btn--secondary"
          style={{ alignSelf: "flex-end" }}
        >
          Search
        </button>
      </form>
      <Tabs
        label="Order status"
        current={status}
        items={TABS.map((t) => ({ ...t, href: href(t.key) }))}
      />
      <Table
        caption="Orders"
        rows={page.items ?? []}
        rowKey={(o) => o.id}
        href={(o) => `/orders/${o.orderNumber}`}
        empty={{ title: "No orders", body: "Nothing matches this filter." }}
        columns={[
          { label: "Order", cell: (o) => o.orderNumber },
          {
            label: "Customer",
            cell: (o) =>
              `${o.firstName ?? ""} ${(o.lastName ?? "").charAt(0)}.`.trim(),
          },
          { label: "Channel", cell: (o) => humanise(o.channel) },
          { label: "Placed", cell: (o) => formatDateTime(o.placedAt) },
          {
            label: "Status",
            cell: (o) => (
              <Status label={humanise(o.status)} tone={orderTone[o.status]} />
            ),
          },
          {
            label: "Total",
            cell: (o) => formatKobo(o.totalKobo),
            numeric: true,
          },
        ]}
      />
      {page.page.hasMore && page.page.nextCursor ? (
        <Link
          className="btn btn--secondary"
          href={`${href(status)}${status || q ? "&" : "?"}cursor=${encodeURIComponent(page.page.nextCursor)}`}
        >
          More orders
        </Link>
      ) : null}
    </div>
  );
}
