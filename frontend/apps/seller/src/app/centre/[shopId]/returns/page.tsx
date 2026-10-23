import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { returnReason, returnStatus } from "@techshop/api-client";
import { EmptyState, Status } from "@techshop/ui/components";

import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime } from "@/lib/format";

export const metadata: Metadata = { title: "Returns | TechShop Seller Centre" };

export default async function ReturnsPage(
  props: PageProps<"/centre/[shopId]/returns">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}/returns`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const rows =
    (await api.get("/api/v1/seller/shops/{shopId}/returns", {
      path: { shopId },
      init: { cache: "no-store" },
    })) ?? [];
  return (
    <ShopPage
      shop={shop}
      section="returns"
      title="Returns"
      lead="Customers return your items through TechShop. We inspect them and send them back to you at the hub; the refund comes from your payable. You can dispute an inspection within 3 days."
    >
      {rows.length === 0 ? (
        <EmptyState icon="returns" title="No returns">
          Returns of your items show here.
        </EmptyState>
      ) : (
        <ul role="list" className="stack">
          {rows.map((r) => {
            const s = returnStatus[r.status] ?? {
              label: r.status,
              tone: "neutral" as const,
            };
            return (
              <li key={r.rma} className="card cluster cluster--between">
                <span className="stack stack--sm">
                  <Link href={`${here}/${r.rma}`} className="link">
                    <strong>{r.rma}</strong>
                  </Link>
                  <span className="text-sm text-muted">
                    Order {r.orderNumber} · {returnReason[r.reason] ?? r.reason}{" "}
                    · {dateTime.format(new Date(r.createdAt))}
                  </span>
                </span>
                <Status label={s.label} tone={s.tone} />
              </li>
            );
          })}
        </ul>
      )}
    </ShopPage>
  );
}
