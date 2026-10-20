import { connection } from "next/server";

import { sites } from "@techshop/ui/sites";

import { PageHead } from "@/components/PageHead";
import { Status, humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { orderTone } from "@/features/orders/status";
import { can, requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";
import { formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Search | TechShop Staff" };

/** One search box across orders, products and people, limited to what the person may see. */
export default async function StaffSearch(props: PageProps<"/search">) {
  await connection();
  const me = await requireStaff();
  const q =
    param((await props.searchParams).q)
      ?.trim()
      .slice(0, 80) ?? "";
  const api = await bff.api();
  const noStore = { cache: "no-store" } as const;
  const [orders, products, people] = q
    ? await Promise.all([
        can(me, "orders.view")
          ? api
              .get("/api/v1/staff/orders", {
                query: { q, limit: 10 },
                init: noStore,
              })
              .catch(() => null)
          : null,
        can(me, "catalog.view")
          ? bff
              .publicApi()
              .get("/api/v1/search", { query: { q, limit: 10 }, init: noStore })
              .catch(() => null)
          : null,
        can(me, "admin.users")
          ? api
              .get("/api/v1/staff/admin/users", { query: { q }, init: noStore })
              .catch(() => null)
          : null,
      ])
    : [null, null, null];

  return (
    <div className="stack stack--lg">
      <PageHead
        title={q ? `Results for “${q}”` : "Search"}
        description={
          q
            ? undefined
            : "Find orders (number or customer), products and people."
        }
      />
      {orders ? (
        <section className="stack" aria-labelledby="r-orders">
          <h2 id="r-orders" className="text-xl">
            Orders
          </h2>
          <Table
            caption="Matching orders"
            rows={orders.items ?? []}
            rowKey={(o) => o.id}
            href={(o) => `/orders/${o.orderNumber}`}
            empty={{
              title: "No orders",
              body: "No order number or customer matches.",
            }}
            columns={[
              { label: "Order", cell: (o) => o.orderNumber },
              { label: "Placed", cell: (o) => formatDateTime(o.placedAt) },
              {
                label: "Status",
                cell: (o) => (
                  <Status
                    label={humanise(o.status)}
                    tone={orderTone[o.status]}
                  />
                ),
              },
              {
                label: "Total",
                cell: (o) => formatKobo(o.totalKobo),
                numeric: true,
              },
            ]}
          />
        </section>
      ) : null}
      {products ? (
        <section className="stack" aria-labelledby="r-products">
          <h2 id="r-products" className="text-xl">
            Products
          </h2>
          <Table
            caption="Matching products"
            rows={products.items ?? []}
            rowKey={(p) => p.id}
            href={(p) => `${sites.market}/p/${p.slug}`}
            empty={{
              title: "No products",
              body: "Nothing in the catalogue matches.",
            }}
            columns={[
              { label: "Product", cell: (p) => p.name },
              { label: "Offers", cell: (p) => p.offerCount, numeric: true },
              {
                label: "From",
                cell: (p) => formatKobo(p.promoPrice ?? p.price ?? 0),
                numeric: true,
              },
            ]}
          />
        </section>
      ) : null}
      {people ? (
        <section className="stack" aria-labelledby="r-people">
          <h2 id="r-people" className="text-xl">
            People
          </h2>
          <Table
            caption="Matching people"
            rows={people}
            rowKey={(u) => u.id}
            empty={{
              title: "No people",
              body: "No customer or staff member matches.",
            }}
            columns={[
              { label: "Name", cell: (u) => u.name },
              { label: "Email", cell: (u) => u.email ?? "—" },
              { label: "Status", cell: (u) => humanise(u.status) },
            ]}
          />
        </section>
      ) : null}
    </div>
  );
}
