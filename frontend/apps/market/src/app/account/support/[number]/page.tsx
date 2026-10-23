import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ticketStatus, ticketTopic } from "@techshop/api-client";
import { PageHeader, Status } from "@techshop/ui/components";

import styles from "@/features/aftersales/components/Aftersales.module.css";
import { TicketReplyForm } from "@/features/aftersales/components/SupportForms";
import { myTicket } from "@/features/aftersales/server";
import { formatDateTime } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Support ticket | TechShop Market",
  robots: { index: false },
};

export default async function TicketPage(
  props: PageProps<"/account/support/[number]">,
) {
  await connection();
  const { number } = await props.params;
  await bff.requireSession(`/account/support/${number}`);
  const t = await myTicket(number);
  if (!t) notFound();
  const status = ticketStatus[t.status] ?? {
    label: t.status,
    tone: "neutral" as const,
  };
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title={t.subject}
          lead={
            <>
              {t.number} · {ticketTopic[t.topic] ?? t.topic}
              {t.orderNumber ? (
                <>
                  {" · "}
                  <Link href={`/orders/${t.orderNumber}`} className="link">
                    {t.orderNumber}
                  </Link>
                </>
              ) : null}{" "}
              · <Status label={status.label} tone={status.tone} />
            </>
          }
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Support", href: "/account/support" },
            { label: t.number },
          ]}
        />
        <ol className={styles.thread} aria-label="Messages">
          {(t.messages ?? []).map((m) => (
            <li
              key={m.id}
              className={`${styles.message} ${m.authorKind === "customer" ? styles.mine : ""}`}
            >
              <p className={styles.messageMeta}>
                {m.authorKind === "customer" ? "You" : m.author} ·{" "}
                {formatDateTime(m.createdAt)}
              </p>
              <p className={styles.body}>{m.body}</p>
            </li>
          ))}
        </ol>
        {t.status === "closed" ? (
          <p className="text-sm text-muted">
            This ticket is closed.{" "}
            <Link href="/account/support/new" className="link">
              Ask a new question
            </Link>{" "}
            and mention {t.number}.
          </p>
        ) : (
          <section className="card card--roomy" aria-label="Reply">
            <TicketReplyForm number={t.number} />
          </section>
        )}
      </div>
    </main>
  );
}
