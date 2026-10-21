import Link from "next/link";
import { Table, Status, humanise } from "@techshop/ui/components";
import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { riders } from "@/features/dispatch/server";
import { can, requireStaff } from "@/features/session/server";
import { releaseWave } from "@/features/warehouse/actions";
import { DropoffForm } from "@/features/warehouse/components/DropoffForm";
import { ManifestForm } from "@/features/warehouse/components/ManifestForm";
import {
  carriers,
  manifests,
  pickLists,
  warehouses,
} from "@/features/warehouse/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Warehouse | TechShop Staff" };

export default async function WarehousePage(props: PageProps<"/warehouse">) {
  await connection();
  const me = await requireStaff("fulfilment.pick");
  const wanted = param((await props.searchParams).w);
  const all = (await warehouses()).filter((w) => w.isActive);
  const chosen = all.find((w) => w.id === wanted) ?? all[0];
  if (!chosen)
    return (
      <p className="alert alert--warning">
        No active warehouses yet. Add one in Inventory.
      </p>
    );
  const [lists, mans, rs, cs] = await Promise.all([
    pickLists(chosen.id),
    manifests(chosen.id),
    can(me, "dispatch.view") ? riders().catch(() => []) : Promise.resolve([]),
    carriers(),
  ]);
  const manage = can(me, "fulfilment.manage");

  return (
    <div className="stack stack--lg">
      <PageHead
        title="Warehouse"
        description="Pick, pack and hand over orders, and take in sellers’ parcels."
        actions={
          manage ? (
            <ActionButton
              label="Release a wave"
              variant="inverse"
              action={releaseWave.bind(null, chosen.id)}
            />
          ) : null
        }
      />
      {all.length > 1 ? (
        <nav aria-label="Warehouse" className="cluster">
          {all.map((w) => (
            <Link
              key={w.id}
              href={`/warehouse?w=${w.id}`}
              className={`btn btn--sm ${w.id === chosen.id ? "btn--inverse" : "btn--secondary"}`}
            >
              {w.name}
            </Link>
          ))}
        </nav>
      ) : (
        <p className="text-sm text-muted">
          {chosen.name} · {chosen.city}
        </p>
      )}

      <section className="stack" aria-labelledby="picks-title">
        <h2 id="picks-title" className="text-xl">
          Pick lists
        </h2>
        <Table
          caption="Pick lists"
          rows={lists}
          rowKey={(l) => l.id}
          href={(l) => `/warehouse/pick/${l.id}`}
          empty={{
            title: "No pick lists",
            body: "Release a wave to group paid orders into a pick list.",
          }}
          columns={[
            { label: "Released", cell: (l) => formatDateTime(l.waveAt) },
            { label: "Zone", cell: (l) => l.zone || "All" },
            { label: "Progress", cell: (l) => `${l.itemsDone} of ${l.items}` },
            {
              label: "Picker",
              cell: (l) => (l.pickerId ? "Claimed" : "Unclaimed"),
            },
            {
              label: "Status",
              cell: (l) => (
                <Status
                  label={humanise(l.status)}
                  tone={
                    l.status === "done" || l.status === "completed"
                      ? "success"
                      : "info"
                  }
                />
              ),
            },
          ]}
        />
      </section>

      {chosen.kind === "warehouse" || chosen.kind === "hub" ? (
        <section className="stack" aria-labelledby="dropoff-title">
          <h2 id="dropoff-title" className="text-xl">
            Seller drop-offs
          </h2>
          <p className="text-sm text-muted">
            Sellers bring packed parcels here. Scanning one in sends it on by
            rider or carrier, like our own parcels.
          </p>
          <div className="card card--roomy">
            <DropoffForm warehouseId={chosen.id} />
          </div>
        </section>
      ) : null}

      <section className="stack" aria-labelledby="manifests-title">
        <h2 id="manifests-title" className="text-xl">
          Manifests
        </h2>
        <Table
          caption="Manifests"
          rows={mans}
          rowKey={(m) => m.id}
          href={(m) => `/warehouse/manifests/${m.id}`}
          empty={{
            title: "No manifests",
            body: "A manifest lists the parcels leaving with one rider or carrier.",
          }}
          columns={[
            { label: "Created", cell: (m) => formatDateTime(m.createdAt) },
            {
              label: "Collected by",
              cell: (m) => (m.riderId ? "Rider" : "Carrier"),
            },
            {
              label: "Signed",
              cell: (m) =>
                m.signedAt
                  ? `${m.signedByName} · ${formatDateTime(m.signedAt)}`
                  : "Not yet",
            },
          ]}
        />
        {manage ? (
          <div className="card card--roomy">
            <ManifestForm
              warehouseId={chosen.id}
              riders={rs
                .filter((r) => r.status === "active")
                .map((r) => ({
                  id: r.userId,
                  name: `${r.firstName} ${r.lastName}`,
                }))}
              carriers={cs
                .filter((c) => c.isActive)
                .map((c) => ({ id: c.id, name: c.name }))}
            />
          </div>
        ) : null}
      </section>
    </div>
  );
}
