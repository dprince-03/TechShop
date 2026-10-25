import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ApiError, ticketStatus } from "@techshop/api-client";
import { Status } from "@techshop/ui/components";

import { ShopTicketReply } from "@/features/aftersales/components/Forms";
import styles from "@/features/aftersales/components/Thread.module.css";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime } from "@/lib/format";

export const metadata: Metadata = { title: "Ticket | TechShop Seller Centre" };

export default async function TicketPage(
  props: PageProps<"/centre/[shopId]/support/[number]">,
) {
  await connection();
  const { shopId, number } = await props.params;
  const here = `/centre/${shopId}/support/${number}`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const t = await api
    .get("/api/v1/seller/shops/{shopId}/tickets/{number}", {
      path: { shopId, number },
      init: { cache: "no-store" },
    })
    .catch((e) => {
      if (e instanceof ApiError && (e.status === 404 || e.status === 422))
        notFound();
      throw e;
    });
  const s = ticketStatus[t.status] ?? {
    label: t.status,
    tone: "neutral" as const,
  };
  return (
    <ShopPage
      shop={shop}
      section="support"
      title={t.subject}
      lead={`${t.number}${t.orderNumber ? ` · ${t.orderNumber}` : ""}`}
    >
      <p>
        <Status label={s.label} tone={s.tone} />
      </p>
      <ol className={styles.thread} aria-label="Messages">
        {(t.messages ?? []).map((m) => (
          <li
            key={m.id}
            className={`${styles.message} ${m.authorKind === "seller" ? styles.mine : ""}`}
          >
            <p className={styles.meta}>
              {m.author} · {dateTime.format(new Date(m.createdAt))}
            </p>
            <p className={styles.body}>{m.body}</p>
          </li>
        ))}
      </ol>
      {t.status !== "closed" ? (
        <section className="card card--roomy">
          <ShopTicketReply shopId={shop.id} number={t.number} />
        </section>
      ) : (
        <p className="text-sm text-muted">
          This ticket is closed. Ask a new question and mention {t.number}.
        </p>
      )}
    </ShopPage>
  );
}
