import Link from "next/link";
import { Table, Status, humanise } from "@techshop/ui/components";
import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { closePeriod } from "@/features/finance/actions";
import { ExceptionActions } from "@/features/finance/components/ExceptionActions";
import { RefundActions } from "@/features/finance/components/RefundActions";
import {
  financeOverview,
  listExceptions,
  listJournals,
  listPeriods,
  listRefunds,
  trialBalance,
} from "@/features/finance/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Finance | TechShop Staff" };

const TABS = [
  { key: "overview", label: "Overview" },
  { key: "refunds", label: "Refunds" },
  { key: "reconciliation", label: "Reconciliation" },
  { key: "ledger", label: "Ledger" },
  { key: "periods", label: "Periods" },
];

export default async function FinancePage(props: PageProps<"/finance">) {
  await connection();
  const sp = await props.searchParams;
  const tab = TABS.some((t) => t.key === param(sp.tab))
    ? param(sp.tab)!
    : "overview";
  const me = await requireStaff("finance.view");
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Finance"
        description="Ledger figures, refunds, payment reconciliation and month-end."
      />
      <Tabs
        label="Finance sections"
        current={tab}
        items={TABS.map((t) => ({ ...t, href: `/finance?tab=${t.key}` }))}
      />
      {tab === "overview" ? <Overview /> : null}
      {tab === "refunds" ? (
        <Refunds
          status={param(sp.status) ?? "requested"}
          approver={
            can(me, "refunds.approve") || can(me, "refunds.approve_small")
          }
        />
      ) : null}
      {tab === "reconciliation" ? <Reconciliation /> : null}
      {tab === "ledger" ? <Ledger cursor={param(sp.cursor)} /> : null}
      {tab === "periods" ? (
        <Periods canClose={can(me, "finance.period_close")} />
      ) : null}
    </div>
  );
}

async function Overview() {
  const { dashboard: d, providers } = await financeOverview();
  return (
    <>
      <Figures
        items={[
          { label: "Revenue", value: formatKobo(d.revenueKobo) },
          { label: "In the bank", value: formatKobo(d.bankKobo) },
          {
            label: "Clearing with providers",
            value: formatKobo(d.clearingKobo),
          },
          { label: "Owed to sellers", value: formatKobo(d.owedToSellersKobo) },
          { label: "Refunds owed", value: formatKobo(d.refundsOwedKobo) },
          { label: "VAT collected", value: formatKobo(d.vatKobo) },
          { label: "Expenses", value: formatKobo(d.expensesKobo) },
        ]}
      />
      <section className="card stack" aria-labelledby="providers-title">
        <h2 id="providers-title" className="text-xl">
          Payment providers{" "}
          {providers.fake ? (
            <span className="badge badge--outline">Test mode</span>
          ) : null}
        </h2>
        <Table
          caption="Provider health"
          rows={providers.health ?? []}
          rowKey={(h) => h.provider}
          empty={{
            title: "No provider activity yet",
            body: "Health appears after the first payments.",
          }}
          columns={[
            { label: "Provider", cell: (h) => humanise(h.provider) },
            {
              label: "Circuit",
              cell: (h) => (
                <Status
                  label={humanise(h.breakerState ?? "closed")}
                  tone={
                    (h.breakerState ?? "closed") === "closed"
                      ? "success"
                      : "danger"
                  }
                />
              ),
            },
            {
              label: "Recent failures",
              cell: (h) => h.failures ?? 0,
              numeric: true,
            },
            {
              label: "Last failure",
              cell: (h) => formatDateTime(h.lastFailureAt),
            },
          ]}
        />
      </section>
    </>
  );
}

async function Refunds({
  status,
  approver,
}: {
  status: string;
  approver: boolean;
}) {
  const refunds = await listRefunds(status);
  const filters = ["requested", "approved", "succeeded", "failed", "rejected"];
  return (
    <section className="stack">
      <div className="cluster">
        {filters.map((f) => (
          <Link
            key={f}
            href={`/finance?tab=refunds&status=${f}`}
            className={`btn btn--sm ${f === status ? "btn--inverse" : "btn--secondary"}`}
            aria-current={f === status ? "page" : undefined}
          >
            {humanise(f)}
          </Link>
        ))}
      </div>
      <Table
        caption="Refunds"
        rows={refunds}
        rowKey={(r) => r.id}
        href={(r) => `/orders/${r.orderNumber}`}
        empty={{ title: "No refunds here", body: "Nothing in this state." }}
        columns={[
          { label: "Order", cell: (r) => r.orderNumber },
          { label: "Reason", cell: (r) => r.reason },
          { label: "Method", cell: (r) => humanise(r.method) },
          { label: "Requested", cell: (r) => formatDateTime(r.createdAt) },
          {
            label: "Status",
            cell: (r) => (
              <Status
                label={humanise(r.status)}
                tone={
                  r.status === "requested"
                    ? "warning"
                    : r.status === "succeeded"
                      ? "success"
                      : r.status === "failed" || r.status === "rejected"
                        ? "danger"
                        : "info"
                }
              />
            ),
          },
          {
            label: "Amount",
            cell: (r) => formatKobo(r.amountKobo),
            numeric: true,
          },
          {
            label: "Decision",
            cell: (r) =>
              approver && r.status === "requested" ? (
                <RefundActions id={r.id} />
              ) : null,
          },
        ]}
      />
      <p className="text-xs text-muted">
        Approvers can’t approve refunds they requested. Large refunds need a
        finance manager.
      </p>
    </section>
  );
}

async function Reconciliation() {
  const items = await listExceptions("open");
  return (
    <Table
      caption="Open reconciliation exceptions"
      rows={items}
      rowKey={(x) => x.id}
      empty={{
        title: "Nothing to reconcile",
        body: "Ledger and provider records agree.",
      }}
      columns={[
        { label: "Kind", cell: (x) => humanise(x.kind) },
        { label: "Provider", cell: (x) => humanise(x.provider ?? "—") },
        { label: "Reference", cell: (x) => x.reference },
        { label: "Detail", cell: (x) => x.detail },
        { label: "Raised", cell: (x) => formatDateTime(x.createdAt) },
        {
          label: "Amount",
          cell: (x) => (x.amountKobo === null ? "—" : formatKobo(x.amountKobo)),
          numeric: true,
        },
        { label: "Resolve", cell: (x) => <ExceptionActions id={x.id} /> },
      ]}
    />
  );
}

async function Ledger({ cursor }: { cursor?: string }) {
  const [tb, journals] = await Promise.all([
    trialBalance(),
    listJournals(cursor),
  ]);
  return (
    <>
      <section className="card stack" aria-labelledby="tb-title">
        <h2 id="tb-title" className="text-xl">
          Trial balance{" "}
          <Status
            label={tb.balanced ? "Balanced" : "Out of balance"}
            tone={tb.balanced ? "success" : "danger"}
          />
        </h2>
        <Table
          caption="Trial balance"
          rows={tb.accounts ?? []}
          rowKey={(a) => a.id}
          columns={[
            { label: "Code", cell: (a) => a.code },
            { label: "Account", cell: (a) => a.name },
            { label: "Type", cell: (a) => humanise(a.kind) },
            {
              label: "Balance",
              cell: (a) => formatKobo(a.balanceKobo),
              numeric: true,
            },
          ]}
        />
      </section>
      <section className="card stack" aria-labelledby="journals-title">
        <h2 id="journals-title" className="text-xl">
          Journals
        </h2>
        <Table
          caption="Journals"
          rows={journals.items ?? []}
          rowKey={(j) => j.id}
          columns={[
            { label: "Posted", cell: (j) => formatDateTime(j.postedAt) },
            { label: "Kind", cell: (j) => humanise(j.kind) },
            { label: "Memo", cell: (j) => j.memo },
            {
              label: "Reference",
              cell: (j) =>
                `${humanise(j.referenceType)} ${j.referenceId.slice(0, 8)}`,
            },
            { label: "Period", cell: (j) => j.period },
          ]}
        />
        {journals.page.hasMore && journals.page.nextCursor ? (
          <Link
            href={`/finance?tab=ledger&cursor=${encodeURIComponent(journals.page.nextCursor)}`}
            className="btn btn--secondary"
          >
            Older journals
          </Link>
        ) : null}
      </section>
    </>
  );
}

async function Periods({ canClose }: { canClose: boolean }) {
  const periods = await listPeriods();
  return (
    <Table
      caption="Accounting periods"
      rows={periods}
      rowKey={(p) => p.month}
      empty={{
        title: "No periods yet",
        body: "Periods appear as journals are posted.",
      }}
      columns={[
        { label: "Month", cell: (p) => p.month },
        {
          label: "Status",
          cell: (p) => (
            <Status
              label={humanise(p.status)}
              tone={p.status === "closed" ? "neutral" : "success"}
            />
          ),
        },
        { label: "Closed", cell: (p) => formatDateTime(p.closedAt) },
        {
          label: "Close",
          cell: (p) =>
            canClose && p.status !== "closed" ? (
              <ActionButton
                label="Close period"
                variant="ghost"
                confirm={`Close ${p.month}? No more postings will be allowed in it.`}
                action={closePeriod.bind(null, p.month)}
              />
            ) : null,
        },
      ]}
    />
  );
}
