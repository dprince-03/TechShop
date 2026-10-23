import { connection } from "next/server";

import { returnReason, returnStatus } from "@techshop/api-client";
import { Status, Table } from "@techshop/ui/components";

import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { returnsList } from "@/features/aftersales/server";
import { requireStaff } from "@/features/session/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Returns | TechShop Staff" };

const TABS = [
  { key: "requested", label: "To decide", status: "requested" },
  { key: "approved", label: "Coming to a store", status: "approved" },
  { key: "awaiting_pickup", label: "Rider pickup", status: "awaiting_pickup" },
  { key: "received", label: "To inspect", status: "received" },
  { key: "inspected", label: "To refund", status: "inspected" },
  { key: "disputed", label: "Seller disputes", status: "" },
  { key: "all", label: "All", status: "" },
];

export default async function ReturnsPage(props: PageProps<"/returns">) {
  await connection();
  await requireStaff("returns.manage");
  const sp = await props.searchParams;
  const tab = TABS.find((t) => t.key === param(sp.tab)) ?? TABS[0];
  const page = await returnsList(
    tab.status,
    tab.key === "disputed",
    param(sp.cursor),
  );
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Returns"
        description="Approve, receive, inspect (the IMEI must match the unit sold) and refund. Seller items go back to the seller."
      />
      <Tabs
        label="Return filters"
        current={tab.key}
        items={TABS.map((t) => ({
          key: t.key,
          label: t.label,
          href: `/returns?tab=${t.key}`,
        }))}
      />
      <Table
        caption="Returns"
        rows={page.items ?? []}
        rowKey={(r) => r.rma}
        href={(r) => `/returns/${r.rma}`}
        empty={{
          title: "Nothing here",
          body: "Returns appear when customers request them from an order.",
        }}
        columns={[
          { label: "Return", cell: (r) => r.rma },
          { label: "Order", cell: (r) => r.orderNumber },
          { label: "Reason", cell: (r) => returnReason[r.reason] ?? r.reason },
          { label: "Requested", cell: (r) => formatDateTime(r.createdAt) },
          {
            label: "Status",
            cell: (r) => {
              const s = returnStatus[r.status] ?? {
                label: r.status,
                tone: "neutral" as const,
              };
              return <Status label={s.label} tone={s.tone} />;
            },
          },
        ]}
      />
      {page.page.nextCursor ? (
        <a
          className="link link--arrow"
          href={`/returns?tab=${tab.key}&cursor=${encodeURIComponent(page.page.nextCursor)}`}
        >
          Older returns
        </a>
      ) : null}
    </div>
  );
}
