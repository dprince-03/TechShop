import type { Metadata } from "next";
import { connection } from "next/server";

import { Status, Table } from "@techshop/ui/components";

import { payoutStatus } from "@/features/payouts/status";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateOnly, dateTime, naira } from "@/lib/format";

export const metadata: Metadata = {
  title: "Payout statement | TechShop Seller Centre",
};

export default async function PayoutStatementPage(
  props: PageProps<"/centre/[shopId]/payouts/[payoutId]">,
) {
  await connection();
  const { shopId, payoutId } = await props.params;
  const here = `/centre/${shopId}/payouts/${payoutId}`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const p = await api.get("/api/v1/seller/shops/{shopId}/payouts/{payoutId}", {
    path: { shopId, payoutId },
    init: { cache: "no-store" },
  });
  const st = payoutStatus[p.status];
  return (
    <ShopPage
      shop={shop}
      section="payouts"
      title={`Payout of ${naira(p.amount)}`}
      lead={
        p.paidAt
          ? `Sent ${dateTime.format(new Date(p.paidAt))} to the account ending ${p.accountLast4}.`
          : `Scheduled for ${dateOnly.format(new Date(p.scheduledFor))}, to the account ending ${p.accountLast4}.`
      }
      actions={<Status label={st.label} tone={st.tone} />}
    >
      {p.failureReason ? (
        <p className="alert alert--error">
          {p.failureReason} The money stays in your balance and goes out in the
          next payout run once your account is fixed.
        </p>
      ) : null}
      <Table
        caption="Orders in this payout"
        rows={p.items ?? []}
        rowKey={(i) => i.orderNumber}
        columns={[
          { label: "Order", cell: (i) => i.orderNumber },
          {
            label: "Delivered",
            cell: (i) =>
              i.deliveredAt ? dateOnly.format(new Date(i.deliveredAt)) : "—",
          },
          {
            label: "Earned",
            numeric: true,
            cell: (i) => naira(i.amount),
          },
        ]}
      />
      <p className="text-sm text-muted">
        Amounts are after TechShop’s commission. Refunds on returned items are
        taken from your balance before we pay out.
      </p>
    </ShopPage>
  );
}
