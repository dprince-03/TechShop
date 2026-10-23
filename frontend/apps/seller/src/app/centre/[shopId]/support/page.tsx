import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { ticketStatus, ticketTopic } from "@techshop/api-client";
import { EmptyState, Status } from "@techshop/ui/components";

import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime } from "@/lib/format";

export const metadata: Metadata = { title: "Support | TechShop Seller Centre" };

export default async function SupportPage(
  props: PageProps<"/centre/[shopId]/support">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}/support`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const rows =
    (await api.get("/api/v1/seller/shops/{shopId}/tickets", {
      path: { shopId },
      init: { cache: "no-store" },
    })) ?? [];
  return (
    <ShopPage
      shop={shop}
      section="support"
      title="Seller support"
      lead="Questions to TechShop about selling, orders, drop-offs and payouts. Everyone in the shop sees these."
      actions={
        <Link href={`${here}/new`} className="btn btn--secondary">
          Ask a question
        </Link>
      }
    >
      {rows.length === 0 ? (
        <EmptyState icon="support" title="No questions yet">
          Ask us anything about selling on TechShop.
        </EmptyState>
      ) : (
        <ul role="list" className="stack">
          {rows.map((t) => {
            const s = ticketStatus[t.status] ?? {
              label: t.status,
              tone: "neutral" as const,
            };
            return (
              <li key={t.number} className="card cluster cluster--between">
                <span className="stack stack--sm">
                  <Link href={`${here}/${t.number}`} className="link">
                    <strong>
                      {t.unread ? "● " : ""}
                      {t.subject}
                    </strong>
                  </Link>
                  <span className="text-sm text-muted">
                    {t.number} · {ticketTopic[t.topic] ?? t.topic} ·{" "}
                    {dateTime.format(new Date(t.lastMessageAt))}
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
