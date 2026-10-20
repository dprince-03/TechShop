import Link from "next/link";
import { connection } from "next/server";

import { PageHead } from "@/components/PageHead";
import { Status, humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { Tabs } from "@/components/Tabs";
import {
  AdjustForm,
  ReceiveUnitsForm,
} from "@/features/inventory/components/StockForms";
import { can, requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Inventory | TechShop Staff" };

const TABS = [
  { key: "stock", label: "Stock" },
  { key: "movements", label: "Movements" },
  { key: "warehouses", label: "Warehouses" },
];

export default async function InventoryPage(props: PageProps<"/inventory">) {
  await connection();
  const me = await requireStaff("inventory.view");
  const sp = await props.searchParams;
  const wanted = param(sp.tab);
  const tab = TABS.some((t) => t.key === wanted) ? wanted! : "stock";
  const api = await bff.api();
  const noStore = { cache: "no-store" } as const;
  const warehouses = (
    (await api.get("/api/v1/staff/inventory/warehouses", { init: noStore })) ??
    []
  ).filter((w) => w.isActive);
  const warehouseId =
    warehouses.find((w) => w.id === param(sp.w))?.id ?? warehouses[0]?.id;
  const variantId = param(sp.variant);
  const choices = warehouses.map((w) => ({ id: w.id, name: w.name }));

  return (
    <div className="stack stack--lg">
      <PageHead
        title="Inventory"
        description="Stock by warehouse, every movement, and receiving."
      />
      <Tabs
        label="Inventory sections"
        current={tab}
        items={TABS.map((t) => ({ ...t, href: `/inventory?tab=${t.key}` }))}
      />
      {tab === "stock" && warehouseId ? (
        <>
          <nav aria-label="Warehouse" className="cluster">
            {warehouses.map((w) => (
              <Link
                key={w.id}
                href={`/inventory?w=${w.id}`}
                className={`btn btn--sm ${w.id === warehouseId ? "btn--inverse" : "btn--secondary"}`}
              >
                {w.name}
              </Link>
            ))}
          </nav>
          <Stock warehouseId={warehouseId} />
          {can(me, "inventory.adjust") ? (
            <div className="grid" style={{ ["--grid-min" as string]: "26rem" }}>
              <section
                className="card card--roomy stack"
                aria-labelledby="adjust-title"
              >
                <h2 id="adjust-title" className="text-xl">
                  Adjust stock
                </h2>
                <AdjustForm warehouses={choices} />
              </section>
              <section
                className="card card--roomy stack"
                aria-labelledby="units-title"
              >
                <h2 id="units-title" className="text-xl">
                  Receive serialised devices
                </h2>
                <ReceiveUnitsForm warehouses={choices} />
              </section>
            </div>
          ) : null}
        </>
      ) : null}
      {tab === "movements" ? <Movements variantId={variantId} /> : null}
      {tab === "warehouses" ? (
        <Table
          caption="Warehouses"
          rows={warehouses}
          rowKey={(w) => w.id}
          columns={[
            { label: "Warehouse", cell: (w) => w.name },
            { label: "Code", cell: (w) => w.code },
            { label: "Kind", cell: (w) => humanise(w.kind) },
            { label: "Location", cell: (w) => `${w.city}, ${w.stateCode}` },
          ]}
        />
      ) : null}
    </div>
  );
}

async function Stock({ warehouseId }: { warehouseId: string }) {
  const rows =
    (await (
      await bff.api()
    ).get("/api/v1/staff/inventory/levels", {
      query: { warehouseId },
      init: { cache: "no-store" },
    })) ?? [];
  return (
    <Table
      caption="Stock levels"
      rows={rows}
      rowKey={(r) =>
        `${r.variantId}:${r.condition}:${r.binId}:${r.ownerSellerId}`
      }
      href={(r) => `/inventory?tab=movements&variant=${r.variantId}`}
      empty={{ title: "No stock here", body: "Receive stock to see it." }}
      columns={[
        { label: "Item", cell: (r) => `${r.productName} · ${r.variantName}` },
        { label: "SKU", cell: (r) => r.sku },
        { label: "Condition", cell: (r) => humanise(r.condition) },
        { label: "On hand", cell: (r) => r.onHand, numeric: true },
        { label: "Reserved", cell: (r) => r.reserved, numeric: true },
        { label: "Available", cell: (r) => r.available, numeric: true },
        {
          label: "",
          cell: (r) =>
            r.available <= (r.reorderPoint ?? 0) ? (
              <Status label="Reorder" tone="warning" />
            ) : null,
        },
        {
          label: "Variant ID",
          cell: (r) => <code className="text-xs">{r.variantId}</code>,
        },
      ]}
    />
  );
}

async function Movements({ variantId }: { variantId?: string }) {
  const rows =
    (await (
      await bff.api()
    ).get("/api/v1/staff/inventory/movements", {
      query: { variantId },
      init: { cache: "no-store" },
    })) ?? [];
  return (
    <Table
      caption="Stock movements"
      rows={rows}
      rowKey={(m) => m.id}
      empty={{
        title: "No movements",
        body: variantId
          ? "Nothing has moved for this item yet."
          : "Movements appear as stock is received, sold or adjusted.",
      }}
      columns={[
        { label: "When", cell: (m) => formatDateTime(m.createdAt) },
        { label: "Reason", cell: (m) => humanise(m.reason) },
        {
          label: "Change",
          cell: (m) =>
            m.quantityDelta > 0
              ? `+${m.quantityDelta}`
              : String(m.quantityDelta),
          numeric: true,
        },
        { label: "Condition", cell: (m) => humanise(m.condition) },
        {
          label: "Reference",
          cell: (m) => (m.referenceType ? humanise(m.referenceType) : "—"),
        },
      ]}
    />
  );
}
