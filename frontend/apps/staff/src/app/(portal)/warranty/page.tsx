import Link from "next/link";
import { connection } from "next/server";

import { claimStatus } from "@techshop/api-client";
import { Status, Table, humanise } from "@techshop/ui/components";

import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import {
  RepairUpdateForm,
  WalkInForm,
} from "@/features/aftersales/components/Controls";
import {
  claims,
  locations,
  repairs,
  warrantySummary,
} from "@/features/aftersales/server";
import { requireStaff } from "@/features/session/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Warranty & repairs | TechShop Staff" };

const TABS = [
  { key: "active", label: "Open claims" },
  { key: "new", label: "Awaiting drop-off" },
  { key: "in_repair", label: "In repair" },
  { key: "ready", label: "Ready to collect" },
  { key: "workshop", label: "Workshop" },
  { key: "all", label: "All claims" },
];

export default async function WarrantyPage(props: PageProps<"/warranty">) {
  await connection();
  await requireStaff("aftersales.manage");
  const sp = await props.searchParams;
  const tab = TABS.find((t) => t.key === param(sp.tab))?.key ?? "active";
  const q = param(sp.q)?.trim().slice(0, 40);
  const [summary, places] = await Promise.all([
    warrantySummary(),
    locations().catch(() => []),
  ]);
  const stores = (places ?? []).map((l) => ({
    value: l.id,
    label: `${l.name}, ${l.city}`,
  }));
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Warranty & repairs"
        description="Claims checked in at stores, the workshop queue (warranty repairs and trade-in refurbishment), and device history by IMEI."
      />
      <Figures
        items={[
          { label: "Open claims", value: String(summary.open) },
          {
            label: "Awaiting drop-off",
            value: String(summary.awaitingDropoff),
          },
          { label: "In repair", value: String(summary.inRepair) },
          { label: "Ready to collect", value: String(summary.ready) },
          {
            label: "Avg turnaround",
            value: `${summary.avgTurnaroundDays.toFixed(1)} days`,
            note: "Received to ready, 30 days",
          },
        ]}
      />
      <details className="card stack">
        <summary className="link">Take a claim at the counter</summary>
        <WalkInForm stores={stores} />
      </details>
      <Tabs
        label="Warranty views"
        current={tab}
        items={TABS.map((t) => ({
          key: t.key,
          label: t.label,
          href: `/warranty?tab=${t.key}`,
        }))}
      />
      {tab === "workshop" ? (
        <Workshop />
      ) : (
        <Claims status={tab === "all" ? "" : tab} q={q} tab={tab} />
      )}
    </div>
  );
}

async function Claims({
  status,
  q,
  tab,
}: {
  status: string;
  q?: string;
  tab: string;
}) {
  const rows = (await claims(status, q)) ?? [];
  return (
    <>
      <form className="cluster" role="search">
        <input type="hidden" name="tab" value={tab} />
        <label htmlFor="claim-q" className="visually-hidden">
          Claim, order or IMEI
        </label>
        <input
          id="claim-q"
          name="q"
          defaultValue={q}
          className="input"
          placeholder="Claim, order or exact IMEI"
          maxLength={40}
        />
        <button className="btn btn--secondary">Search</button>
      </form>
      <Table
        caption="Warranty claims"
        rows={rows}
        rowKey={(c) => c.number}
        href={(c) => `/warranty/${c.number}`}
        empty={{
          title: "No claims here",
          body: "Claims appear when customers make them online or at a store.",
        }}
        columns={[
          { label: "Claim", cell: (c) => c.number },
          {
            label: "Device",
            cell: (c) => `${c.productName ?? "—"} · ${c.imeiOrSerial}`,
          },
          { label: "Fault", cell: (c) => c.fault },
          { label: "Customer", cell: (c) => c.customer },
          { label: "Opened", cell: (c) => formatDateTime(c.createdAt) },
          {
            label: "Status",
            cell: (c) => {
              const s = claimStatus[c.status] ?? {
                label: c.status,
                tone: "neutral" as const,
              };
              return <Status label={s.label} tone={s.tone} />;
            },
          },
        ]}
      />
    </>
  );
}

async function Workshop() {
  const jobs = (await repairs("active")) ?? [];
  if (!jobs.length) return <p className="text-muted">The workshop is clear.</p>;
  return (
    <ul role="list" className="stack stack--lg">
      {jobs.map((j) => (
        <li key={j.id} className="card stack">
          <p>
            <strong>{j.device || "Device"}</strong>{" "}
            <span className="text-sm text-muted">
              {j.imeiOrSerial ?? ""} ·{" "}
              {j.purpose === "refurbish"
                ? `Refurbishment (trade-in ${j.tradeIn ?? ""})`
                : "Warranty repair"}
              {j.claimNumber ? (
                <>
                  {" · "}
                  <Link href={`/warranty/${j.claimNumber}`} className="link">
                    {j.claimNumber}
                  </Link>
                </>
              ) : null}
            </span>
          </p>
          <p className="text-sm">
            {humanise(j.status)}
            {j.diagnosis ? ` · ${j.diagnosis}` : ""}
          </p>
          <RepairUpdateForm jobId={j.id} status={j.status} />
        </li>
      ))}
    </ul>
  );
}
