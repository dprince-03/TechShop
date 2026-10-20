import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { Status, humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { verifyPayment } from "@/features/orders/actions";
import { CancelOrder } from "@/features/orders/components/CancelOrder";
import { getOrder } from "@/features/orders/server";
import { orderTone } from "@/features/orders/status";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo } from "@/lib/format";

export const metadata = { title: "Order | TechShop Staff" };

const CANCELLABLE = new Set(["pending_payment", "paid", "processing"]);

export default async function StaffOrderPage(
  props: PageProps<"/orders/[number]">,
) {
  await connection();
  const { number } = await props.params;
  const me = await requireStaff("orders.view");
  const order = await getOrder(number);
  if (!order) notFound();
  const lines = (order.fulfilments ?? []).flatMap((f) =>
    (f.lines ?? []).map((l) => ({
      ...l,
      seller: f.sellerName,
      parcel: f.status,
    })),
  );

  return (
    <div className="stack stack--lg">
      <PageHead
        title={`Order ${order.orderNumber}`}
        description={
          <>
            {humanise(order.channel)} · placed {formatDateTime(order.placedAt)}{" "}
            ·{" "}
            <Status
              label={humanise(order.status)}
              tone={orderTone[order.status]}
            />
          </>
        }
      />
      <section className="card stack" aria-labelledby="lines-title">
        <h2 id="lines-title" className="text-xl">
          Items
        </h2>
        <Table
          caption="Order lines"
          rows={lines}
          rowKey={(l) => l.id}
          columns={[
            {
              label: "Product",
              cell: (l) => `${l.productName} · ${l.variantName}`,
            },
            { label: "SKU", cell: (l) => l.sku },
            { label: "Seller", cell: (l) => l.seller },
            { label: "Parcel", cell: (l) => humanise(l.parcel) },
            { label: "Qty", cell: (l) => l.quantity, numeric: true },
            {
              label: "Total",
              cell: (l) =>
                l.cancelledAt ? "Cancelled" : formatKobo(l.lineTotalKobo),
              numeric: true,
            },
          ]}
        />
        <dl className="cluster text-sm">
          <div>
            <dt className="text-muted">Delivery</dt>
            <dd>{formatKobo(order.deliveryFeeKobo)}</dd>
          </div>
          <div>
            <dt className="text-muted">Discount</dt>
            <dd>{formatKobo(order.discountKobo)}</dd>
          </div>
          <div>
            <dt className="text-muted">VAT</dt>
            <dd>{formatKobo(order.vatKobo)}</dd>
          </div>
          <div>
            <dt className="text-muted">Total</dt>
            <dd>
              <strong>{formatKobo(order.totalKobo)}</strong>
            </dd>
          </div>
          <div>
            <dt className="text-muted">Contact</dt>
            <dd>{order.contactPhone ?? "—"}</dd>
          </div>
        </dl>
      </section>

      <section className="card stack" aria-labelledby="payments-title">
        <h2 id="payments-title" className="text-xl">
          Payments
        </h2>
        <Table
          caption="Payments"
          rows={order.payments ?? []}
          rowKey={(p) => p.id}
          empty={{
            title: "No payments yet",
            body: "The customer hasn’t started paying.",
          }}
          columns={[
            { label: "Provider", cell: (p) => humanise(p.provider) },
            {
              label: "Status",
              cell: (p) => (
                <Status
                  label={humanise(p.status)}
                  tone={
                    p.status === "succeeded"
                      ? "success"
                      : p.status === "failed"
                        ? "danger"
                        : "warning"
                  }
                />
              ),
            },
            { label: "Started", cell: (p) => formatDateTime(p.createdAt) },
            { label: "Paid", cell: (p) => formatDateTime(p.paidAt) },
            {
              label: "Amount",
              cell: (p) => formatKobo(p.amountKobo),
              numeric: true,
            },
            {
              label: "Check",
              cell: (p) =>
                can(me, "finance.view") && p.status !== "succeeded" ? (
                  <ActionButton
                    label="Re-verify"
                    action={verifyPayment.bind(null, p.id)}
                  />
                ) : null,
            },
          ]}
        />
      </section>

      {(order.history ?? []).length ? (
        <section className="card stack" aria-labelledby="history-title">
          <h2 id="history-title" className="text-xl">
            History
          </h2>
          <ol className="stack stack--sm text-sm">
            {order.history!.map((h) => (
              <li key={h.id}>
                <strong>{humanise(h.toStatus)}</strong> ·{" "}
                {formatDateTime(h.createdAt)}
                {h.note ? (
                  <span className="text-muted"> · {h.note}</span>
                ) : null}
              </li>
            ))}
          </ol>
        </section>
      ) : null}

      {can(me, "orders.cancel") && CANCELLABLE.has(order.status) ? (
        <section className="card stack" aria-labelledby="cancel-title">
          <h2 id="cancel-title" className="text-xl">
            Cancel
          </h2>
          <CancelOrder number={order.orderNumber} />
        </section>
      ) : null}
    </div>
  );
}
