import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ApiError, claimResolution, claimStatus } from "@techshop/api-client";
import { Status, humanise } from "@techshop/ui/components";

import { Details } from "@/components/Details";
import { PageHead } from "@/components/PageHead";
import {
  ClaimDecisions,
  CollectButton,
  ReceiveClaimForm,
  RepairUpdateForm,
  ResolveClaimForm,
} from "@/features/aftersales/components/Controls";
import { claim, locations } from "@/features/aftersales/server";
import { requireStaff } from "@/features/session/server";
import { formatDate, formatDateTime } from "@/lib/format";

export const metadata = { title: "Warranty claim | TechShop Staff" };

export default async function ClaimPage(
  props: PageProps<"/warranty/[number]">,
) {
  await connection();
  await requireStaff("aftersales.manage");
  const { number } = await props.params;
  const c = await claim(number).catch((e) => {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      notFound();
    throw e;
  });
  const places = ((await locations().catch(() => [])) ?? []).map((l) => ({
    value: l.id,
    label: `${l.name}, ${l.city}`,
  }));
  const status = claimStatus[c.status] ?? {
    label: c.status,
    tone: "neutral" as const,
  };
  const jobs = c.repairs ?? [];
  const openJob = jobs.find(
    (j) => j.status !== "done" && j.status !== "unrepairable",
  );
  const tracked = /^\d{15}$/.test(c.imeiOrSerial);
  return (
    <div className="stack stack--lg">
      <PageHead
        title={`Claim ${c.number}`}
        description={<Status label={status.label} tone={status.tone} />}
      />
      <section className="card stack" aria-labelledby="claim-sum">
        <h2 id="claim-sum" className="text-xl">
          Claim
        </h2>
        <Details
          items={[
            ["Customer", c.customer],
            ["Device", `${c.productName ?? ""} ${c.variantName ?? ""}`.trim()],
            ["IMEI or serial", c.imeiOrSerial],
            ["Order", c.orderNumber],
            ["Fault", c.fault],
            [
              "Warranty",
              c.provider
                ? `${humanise(c.provider)} until ${formatDate(c.warrantyEndsOn)}`
                : null,
            ],
            ["Store", c.store],
            ["Received", c.receivedAt ? formatDateTime(c.receivedAt) : null],
            [
              "Outcome",
              c.resolution
                ? (claimResolution[c.resolution] ?? c.resolution)
                : null,
            ],
            ["Rejected because", c.rejectionReason],
            ["Note", c.resolutionNote],
          ]}
        />
        {c.status === "new" ? (
          <ReceiveClaimForm number={c.number} stores={places} />
        ) : null}
        {c.status === "received" || (c.status === "in_repair" && !openJob) ? (
          <ClaimDecisions number={c.number} canRepair={!openJob} />
        ) : null}
        {(c.status === "received" || c.status === "in_repair") && !openJob ? (
          <details>
            <summary className="link">Replace or refund instead</summary>
            <ResolveClaimForm
              number={c.number}
              warehouses={places}
              serialised={tracked}
            />
          </details>
        ) : null}
        {c.status === "ready" ? <CollectButton number={c.number} /> : null}
      </section>
      {jobs.length ? (
        <section className="card stack" aria-labelledby="claim-jobs">
          <h2 id="claim-jobs" className="text-xl">
            Workshop
          </h2>
          <ul role="list" className="stack">
            {jobs.map((j) => (
              <li key={j.id} className="stack stack--sm">
                <p className="text-sm">
                  <strong>{humanise(j.status)}</strong>
                  {j.technician ? ` · ${j.technician}` : ""} · opened{" "}
                  {formatDateTime(j.createdAt)}
                </p>
                {j.diagnosis ? (
                  <p className="text-sm">Diagnosis: {j.diagnosis}</p>
                ) : null}
                {j.notes ? (
                  <p className="text-sm text-muted">{j.notes}</p>
                ) : null}
                {j.id === openJob?.id ? (
                  <RepairUpdateForm jobId={j.id} status={j.status} />
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
      {(c.history ?? []).length ? (
        <section className="card stack" aria-labelledby="claim-history">
          <h2 id="claim-history" className="text-xl">
            Earlier claims on this device
          </h2>
          <ul role="list" className="stack stack--sm text-sm">
            {c.history!.map((h) => (
              <li key={h.number}>
                {h.number} · {h.fault} · {humanise(h.status)} ·{" "}
                {formatDate(h.createdAt)}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
