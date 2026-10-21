import Link from "next/link";
import { connection } from "next/server";

import { Status, Table, humanise } from "@techshop/ui/components";

import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { requireStaff } from "@/features/session/server";
import { vendors, type VendorFilter } from "@/features/vendors/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Vendors | TechShop Staff" };

const TABS: { key: VendorFilter; label: string }[] = [
  { key: "to_review", label: "To review" },
  { key: "all", label: "All" },
  { key: "active", label: "Active" },
  { key: "suspended", label: "Suspended" },
];

const kycTone = {
  submitted: "info",
  in_review: "info",
  approved: "success",
  rejected: "danger",
  needs_more_info: "warning",
} as const;

export default async function VendorsPage(props: PageProps<"/vendors">) {
  await connection();
  await requireStaff("vendors.view");
  const sp = await props.searchParams;
  const filter = TABS.find((t) => t.key === param(sp.tab))?.key ?? "to_review";
  const q = param(sp.q)?.trim().slice(0, 60);
  const page = await vendors(filter, q, param(sp.cursor));
  const href = (extra: Record<string, string>) =>
    `/vendors?${new URLSearchParams({ tab: filter, ...(q ? { q } : {}), ...extra })}`;
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Vendors"
        description="Marketplace sellers: verification, payout accounts and duplicate signals."
      />
      <Tabs
        label="Vendor filters"
        current={filter}
        items={TABS.map((t) => ({ ...t, href: `/vendors?tab=${t.key}` }))}
      />
      <form className="cluster" role="search">
        <input type="hidden" name="tab" value={filter} />
        <label htmlFor="vendor-q" className="visually-hidden">
          Shop name
        </label>
        <input
          id="vendor-q"
          name="q"
          defaultValue={q}
          className="input"
          placeholder="Shop name"
          maxLength={60}
        />
        <button className="btn btn--secondary">Search</button>
      </form>
      <Table
        caption="Vendors"
        rows={page.items ?? []}
        rowKey={(v) => v.id}
        href={(v) => `/vendors/${v.id}`}
        empty={{
          title: filter === "to_review" ? "Nothing to review" : "No vendors",
          body: "Sellers appear here when they open a shop.",
        }}
        columns={[
          { label: "Shop", cell: (v) => v.displayName },
          { label: "Type", cell: (v) => humanise(v.type) },
          {
            label: "Location",
            cell: (v) =>
              [v.city, v.stateCode].filter(Boolean).join(", ") || "—",
          },
          { label: "Shop status", cell: (v) => humanise(v.status) },
          {
            label: "Verification",
            cell: (v) =>
              v.kycStatus ? (
                <Status
                  label={humanise(v.kycStatus)}
                  tone={
                    kycTone[v.kycStatus as keyof typeof kycTone] ?? "neutral"
                  }
                />
              ) : (
                "Not submitted"
              ),
          },
          {
            label: "Submitted",
            cell: (v) =>
              v.kycSubmittedAt ? formatDateTime(v.kycSubmittedAt) : "—",
          },
        ]}
      />
      {page.page.nextCursor ? (
        <Link
          href={href({ cursor: page.page.nextCursor })}
          className="link link--arrow"
        >
          Older vendors
        </Link>
      ) : null}
    </div>
  );
}
