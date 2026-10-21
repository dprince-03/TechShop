import { notFound } from "next/navigation";
import { Table, Status, humanise } from "@techshop/ui/components";
import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { requireStaff } from "@/features/session/server";
import { claimPickList } from "@/features/warehouse/actions";
import { PackForm } from "@/features/warehouse/components/PackForm";
import { ScanForm } from "@/features/warehouse/components/ScanForm";
import { pickList } from "@/features/warehouse/server";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Pick list | TechShop Staff" };

export default async function PickListPage(
  props: PageProps<"/warehouse/pick/[id]">,
) {
  await connection();
  const me = await requireStaff("fulfilment.pick");
  const list = await pickList((await props.params).id);
  if (!list) notFound();
  const items = list.items ?? [];
  const mine = list.pickerId === me.id;
  // Parcels are packed per fulfilment once all of its lines are picked.
  const fulfilments = [...new Set(items.map((i) => i.fulfilmentId))].map(
    (f) => ({
      id: f,
      done: items
        .filter((i) => i.fulfilmentId === f)
        .every((i) => i.pickedQty >= i.quantity),
      count: items.filter((i) => i.fulfilmentId === f).length,
    }),
  );

  return (
    <div className="stack stack--lg">
      <PageHead
        title="Pick list"
        description={
          <>
            Released {formatDateTime(list.waveAt)} ·{" "}
            <Status label={humanise(list.status)} tone="info" />
          </>
        }
        actions={
          !list.pickerId ? (
            <ActionButton
              label="Claim this list"
              variant="inverse"
              action={claimPickList.bind(null, list.id)}
            />
          ) : (
            <span className="text-sm text-muted">
              {mine ? "You’re picking this list." : "Claimed by someone else."}
            </span>
          )
        }
      />
      <Table
        caption="Items to pick"
        rows={items}
        rowKey={(i) => i.id}
        columns={[
          { label: "Bin", cell: (i) => i.binCode || "—" },
          { label: "Item", cell: (i) => `${i.productName} · ${i.variantName}` },
          { label: "SKU", cell: (i) => i.sku },
          { label: "Condition", cell: (i) => humanise(i.condition) },
          {
            label: "Picked",
            cell: (i) => `${i.pickedQty} / ${i.quantity}`,
            numeric: true,
          },
          {
            label: "Pick",
            cell: (i) =>
              i.pickedQty >= i.quantity ? (
                <Status label="Done" tone="success" />
              ) : mine ? (
                <ScanForm
                  pickListId={list.id}
                  orderLineId={i.orderLineId}
                  remaining={i.quantity - i.pickedQty}
                  serialised={i.isSerialised}
                />
              ) : null,
          },
        ]}
      />
      <section className="card stack" aria-labelledby="pack-title">
        <h2 id="pack-title" className="text-xl">
          Pack
        </h2>
        <ul role="list" className="stack">
          {fulfilments.map((f) => (
            <li key={f.id} className="cluster cluster--between">
              <span className="text-sm">
                Parcel {f.id.slice(0, 8)} · {f.count} line
                {f.count === 1 ? "" : "s"}
              </span>
              {f.done ? (
                <PackForm fulfilmentId={f.id} />
              ) : (
                <span className="text-sm text-muted">
                  Pick every line first
                </span>
              )}
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
