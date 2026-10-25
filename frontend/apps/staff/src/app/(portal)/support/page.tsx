import { connection } from "next/server";

import { ticketStatus, ticketTopic } from "@techshop/api-client";
import { Status, Table, humanise } from "@techshop/ui/components";

import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { requireStaff } from "@/features/session/server";
import { supportSummary, tickets } from "@/features/support/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Customer support | TechShop Staff" };

const TABS = [
  { key: "active", label: "Active", f: { status: "active" } },
  {
    key: "urgent",
    label: "Urgent and high",
    f: { status: "open", priority: "urgent_high" },
  },
  { key: "mine", label: "Mine", f: { status: "active", mine: true } },
  {
    key: "waiting",
    label: "Waiting on customer",
    f: { status: "pending_customer" },
  },
  { key: "resolved", label: "Resolved", f: { status: "resolved" } },
  { key: "all", label: "All", f: {} },
];

const priorityTone = {
  urgent: "danger",
  high: "warning",
  normal: "neutral",
  low: "neutral",
} as const;

export default async function SupportPage(props: PageProps<"/support">) {
  await connection();
  await requireStaff("support.view");
  const sp = await props.searchParams;
  const tab = TABS.find((t) => t.key === param(sp.tab)) ?? TABS[0];
  const q = param(sp.q)?.trim().slice(0, 80);
  const [s, rows] = await Promise.all([
    supportSummary(),
    tickets({ ...tab.f, q }),
  ]);
  const avg =
    s.avgFirstReplyMinutes >= 60
      ? `${(s.avgFirstReplyMinutes / 60).toFixed(1)} h`
      : `${s.avgFirstReplyMinutes} min`;
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Customer support"
        description="Tickets from the market, the app, Seller Centre and contact forms. Internal notes are never shown to customers."
      />
      <Figures
        items={[
          {
            label: "Open",
            value: String(s.open),
            note: `${s.urgent} urgent or high`,
          },
          {
            label: "Past reply target",
            value: String(s.overdue),
            note: "24 h, or 4 h for payments",
          },
          { label: "Waiting on customers", value: String(s.waiting) },
          { label: "Resolved today", value: String(s.resolvedToday) },
          { label: "Avg first reply", value: avg, note: "Last 7 days" },
        ]}
      />
      <Tabs
        label="Ticket filters"
        current={tab.key}
        items={TABS.map((t) => ({
          key: t.key,
          label: t.label,
          href: `/support?tab=${t.key}`,
        }))}
      />
      <form className="cluster" role="search">
        <input type="hidden" name="tab" value={tab.key} />
        <label htmlFor="t-q" className="visually-hidden">
          Search tickets
        </label>
        <input
          id="t-q"
          name="q"
          defaultValue={q}
          className="input"
          placeholder="Ticket, order, email or subject"
          maxLength={80}
        />
        <button className="btn btn--secondary">Search</button>
      </form>
      <Table
        caption="Tickets"
        rows={rows ?? []}
        rowKey={(t) => t.number}
        href={(t) => `/support/${t.number}`}
        empty={{
          title: "No tickets here",
          body: "Nothing waiting in this view.",
        }}
        columns={[
          { label: "Ticket", cell: (t) => t.number },
          { label: "Subject", cell: (t) => t.subject },
          {
            label: "From",
            cell: (t) => `${t.requester} (${humanise(t.requesterKind)})`,
          },
          { label: "Topic", cell: (t) => ticketTopic[t.topic] ?? t.topic },
          {
            label: "Priority",
            cell: (t) => (
              <Status
                label={humanise(t.priority)}
                tone={
                  priorityTone[t.priority as keyof typeof priorityTone] ??
                  "neutral"
                }
              />
            ),
          },
          { label: "Assignee", cell: (t) => t.assignee ?? "—" },
          {
            label: "Last message",
            cell: (t) =>
              `${formatDateTime(t.lastMessageAt)}${t.overdue ? " · overdue" : ""}`,
          },
          {
            label: "Status",
            cell: (t) => {
              const st = ticketStatus[t.status] ?? {
                label: t.status,
                tone: "neutral" as const,
              };
              return <Status label={st.label} tone={st.tone} />;
            },
          },
        ]}
      />
    </div>
  );
}
