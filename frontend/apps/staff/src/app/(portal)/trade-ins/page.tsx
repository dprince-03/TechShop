import { connection } from "next/server";

import {
  deviceCondition,
  formatKobo,
  tradeInStatus,
} from "@techshop/api-client";
import { Status, Table } from "@techshop/ui/components";

import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { tradeInSummary, tradeIns } from "@/features/aftersales/server";
import { requireStaff } from "@/features/session/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Trade-ins | TechShop Staff" };

const TABS = [
  { key: "active", label: "Open" },
  { key: "submitted", label: "Needs an estimate" },
  { key: "awaiting_inspection", label: "To inspect" },
  { key: "offer_sent", label: "Offers out" },
  { key: "payout_due", label: "Bank payouts due" },
  { key: "all", label: "All" },
];

export default async function TradeInsPage(props: PageProps<"/trade-ins">) {
  await connection();
  await requireStaff("aftersales.manage");
  const sp = await props.searchParams;
  const tab = TABS.find((t) => t.key === param(sp.tab))?.key ?? "active";
  const q = param(sp.q)?.trim().slice(0, 40);
  const [summary, rows] = await Promise.all([
    tradeInSummary(),
    tradeIns(tab === "all" ? "" : tab, q),
  ]);
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Trade-ins"
        description="Estimates, store appraisal with the IMEI and risk check, offers, and payouts. Accepted devices go to the workshop for refurbishment."
      />
      <Figures
        items={[
          { label: "Needs an estimate", value: String(summary.needsEstimate) },
          { label: "To inspect", value: String(summary.awaitingInspection) },
          { label: "Offers out", value: String(summary.offersOut) },
          { label: "Bank payouts due", value: String(summary.payoutsDue) },
          {
            label: "Paid this month",
            value: String(summary.paidThisMonth),
            note: summary.avgPaid
              ? `Average ${formatKobo(summary.avgPaid)}`
              : undefined,
          },
        ]}
      />
      <Tabs
        label="Trade-in filters"
        current={tab}
        items={TABS.map((t) => ({
          key: t.key,
          label: t.label,
          href: `/trade-ins?tab=${t.key}`,
        }))}
      />
      <form className="cluster" role="search">
        <input type="hidden" name="tab" value={tab} />
        <label htmlFor="ti-q" className="visually-hidden">
          Reference or IMEI
        </label>
        <input
          id="ti-q"
          name="q"
          defaultValue={q}
          className="input"
          placeholder="Reference or IMEI"
          maxLength={40}
        />
        <button className="btn btn--secondary">Search</button>
      </form>
      <Table
        caption="Trade-ins"
        rows={rows ?? []}
        rowKey={(t) => t.reference}
        href={(t) => `/trade-ins/${t.reference}`}
        empty={{
          title: "Nothing here",
          body: "Trade-ins appear when customers ask for an estimate.",
        }}
        columns={[
          { label: "Request", cell: (t) => t.reference },
          { label: "Device", cell: (t) => t.device },
          {
            label: "Condition (declared)",
            cell: (t) =>
              deviceCondition[t.declaredCondition] ?? t.declaredCondition,
          },
          {
            label: "Estimate",
            numeric: true,
            cell: (t) => (t.estimate != null ? formatKobo(t.estimate) : "—"),
          },
          {
            label: "Offer",
            numeric: true,
            cell: (t) => (t.offer != null ? formatKobo(t.offer) : "—"),
          },
          { label: "Customer", cell: (t) => t.customer },
          { label: "Opened", cell: (t) => formatDateTime(t.createdAt) },
          {
            label: "Status",
            cell: (t) => {
              const s = tradeInStatus[t.status] ?? {
                label: t.status,
                tone: "neutral" as const,
              };
              return <Status label={s.label} tone={s.tone} />;
            },
          },
        ]}
      />
    </div>
  );
}
