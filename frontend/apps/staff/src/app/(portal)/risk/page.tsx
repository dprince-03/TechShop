import { connection } from "next/server";

import { Status, Table, humanise } from "@techshop/ui/components";

import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { RiskNav } from "@/features/risk/components/RiskNav";
import { cases, riskSummary } from "@/features/risk/server";
import { requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Risk & fraud | TechShop Staff" };

const STATUS_TABS = [
  { key: "active", label: "To work" },
  { key: "escalated", label: "Escalated" },
  { key: "cleared", label: "Cleared" },
  { key: "blocked", label: "Blocked" },
];

const tone = {
  open: "warning",
  in_review: "info",
  escalated: "danger",
  cleared: "success",
  blocked: "danger",
} as const;

export default async function RiskPage(props: PageProps<"/risk">) {
  await connection();
  await requireStaff("risk.view");
  const sp = await props.searchParams;
  const status =
    STATUS_TABS.find((t) => t.key === param(sp.status))?.key ?? "active";
  const [s, rows] = await Promise.all([
    riskSummary(),
    cases(status, param(sp.kind)),
  ]);
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Risk & fraud"
        description="Holds and blocks from checkout, payout, trade-in and return checks, ordered by score × money at risk."
      />
      <RiskNav current="cases" />
      <Figures
        items={[
          { label: "Cases to work", value: String(s.active) },
          { label: "Orders held", value: String(s.orders) },
          {
            label: "Past review target",
            value: String(s.overdue),
            note: "2 h for holds, a day for blocks",
          },
          { label: "Resolved (7 days)", value: String(s.resolved7d) },
          { label: "Blocked devices", value: String(s.blockedDevices) },
        ]}
      />
      <Tabs
        label="Case status"
        current={status}
        items={STATUS_TABS.map((t) => ({
          key: t.key,
          label: t.label,
          href: `/risk?status=${t.key}`,
        }))}
      />
      <Table
        caption="Risk cases"
        rows={rows ?? []}
        rowKey={(c) => c.reference}
        href={(c) => `/risk/cases/${c.reference}`}
        empty={{
          title: "Nothing to review",
          body: "Cases appear when a check holds or blocks something.",
        }}
        columns={[
          { label: "Case", cell: (c) => c.reference },
          { label: "Type", cell: (c) => humanise(c.kind) },
          { label: "Reason", cell: (c) => c.reason },
          { label: "Score", numeric: true, cell: (c) => c.score ?? "—" },
          {
            label: "At risk",
            numeric: true,
            cell: (c) => (c.moneyAtRisk ? formatKobo(c.moneyAtRisk) : "—"),
          },
          {
            label: "Raised",
            cell: (c) =>
              `${formatDateTime(c.createdAt)}${c.overdue ? " · overdue" : ""}`,
          },
          { label: "Analyst", cell: (c) => c.assignee ?? "—" },
          {
            label: "Status",
            cell: (c) => (
              <Status
                label={humanise(c.status)}
                tone={tone[c.status as keyof typeof tone] ?? "neutral"}
              />
            ),
          },
        ]}
      />
    </div>
  );
}
