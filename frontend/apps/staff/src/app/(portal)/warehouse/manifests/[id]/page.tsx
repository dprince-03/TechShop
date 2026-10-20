import { notFound } from "next/navigation";
import { connection } from "next/server";

import { PageHead } from "@/components/PageHead";
import { Table } from "@/components/Table";
import { can, requireStaff } from "@/features/session/server";
import { SignManifest } from "@/features/warehouse/components/SignManifest";
import { manifest } from "@/features/warehouse/server";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Manifest | TechShop Staff" };

export default async function ManifestPage(
  props: PageProps<"/warehouse/manifests/[id]">,
) {
  await connection();
  const me = await requireStaff("fulfilment.pick");
  const m = await manifest((await props.params).id);
  if (!m) notFound();
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Manifest"
        description={`Created ${formatDateTime(m.createdAt)} · ${m.riderId ? "rider" : "carrier"} collection`}
      />
      <Table
        caption="Parcels on this manifest"
        rows={m.parcels ?? []}
        rowKey={(p) => p.id}
        columns={[
          { label: "Label", cell: (p) => p.labelCode },
          { label: "Box", cell: (p) => p.boxCode || "—" },
          {
            label: "Weight",
            cell: (p) =>
              p.weightGrams === null
                ? "—"
                : `${p.weightGrams.toLocaleString("en-NG")} g`,
            numeric: true,
          },
          { label: "Packed", cell: (p) => formatDateTime(p.createdAt) },
        ]}
      />
      {m.signedAt ? (
        <p className="alert alert--success">
          Handed over to {m.signedByName} on {formatDateTime(m.signedAt)}.
        </p>
      ) : can(me, "fulfilment.manage") ? (
        <section className="card stack" aria-labelledby="sign-title">
          <h2 id="sign-title" className="text-xl">
            Hand over
          </h2>
          <SignManifest id={m.id} />
        </section>
      ) : null}
    </div>
  );
}
