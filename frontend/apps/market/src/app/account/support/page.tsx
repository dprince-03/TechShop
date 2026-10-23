import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { ticketStatus, ticketTopic } from "@techshop/api-client";
import { EmptyState, PageHeader } from "@techshop/ui/components";

import { AccountNav, Records } from "@/features/aftersales/components/Records";
import { myTickets } from "@/features/aftersales/server";
import { formatDateTime } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Support | TechShop Market",
  robots: { index: false },
};

export default async function SupportPage() {
  await connection();
  await bff.requireSession("/account/support");
  const tickets = (await myTickets()) ?? [];
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Support"
          lead="Your conversations with our team. We reply here and send you a notification."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Account", href: "/account" },
            { label: "Support" },
          ]}
          actions={
            <Link href="/account/support/new" className="btn btn--secondary">
              Ask a question
            </Link>
          }
        />
        <AccountNav current="/account/support" />
        <section className="card card--roomy">
          {tickets.length ? (
            <Records
              label="Your tickets"
              rows={tickets.map((t) => ({
                key: t.number,
                href: `/account/support/${t.number}`,
                title: `${t.unread ? "● " : ""}${t.subject}`,
                status: ticketStatus[t.status] ?? {
                  label: t.status,
                  tone: "neutral",
                },
                meta: `${t.number} · ${ticketTopic[t.topic] ?? t.topic}${t.orderNumber ? ` · ${t.orderNumber}` : ""} · ${t.unread ? "New reply · " : ""}${formatDateTime(t.lastMessageAt)}`,
              }))}
            />
          ) : (
            <EmptyState icon="support" title="No questions yet">
              Most answers are in the{" "}
              <Link href="/help" className="link">
                help centre
              </Link>
              . Still stuck?{" "}
              <Link href="/account/support/new" className="link">
                Ask us
              </Link>
              .
            </EmptyState>
          )}
        </section>
      </div>
    </main>
  );
}
