import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { claimResolution, claimStatus } from "@techshop/api-client";
import { PageHeader, Status } from "@techshop/ui/components";

import styles from "@/features/aftersales/components/Aftersales.module.css";
import { Facts } from "@/features/aftersales/components/Records";
import { myClaim } from "@/features/aftersales/server";
import { formatDate, formatDateTime } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Warranty claim | TechShop Market",
  robots: { index: false },
};

const repairWords: Record<string, string> = {
  queued: "Waiting for a technician",
  diagnosing: "Being diagnosed",
  awaiting_parts: "Waiting for parts",
  repairing: "Being repaired",
  done: "Repaired",
  unrepairable: "Can't be repaired",
};

export default async function ClaimPage(
  props: PageProps<"/account/warranty/[number]">,
) {
  await connection();
  const { number } = await props.params;
  await bff.requireSession(`/account/warranty/${number}`);
  const c = await myClaim(number);
  if (!c) notFound();
  const isNew = (await props.searchParams).new === "1";
  const status = claimStatus[c.status] ?? {
    label: c.status,
    tone: "neutral" as const,
  };
  const latest = (c.repairs ?? []).at(-1);

  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title={`Claim ${c.number}`}
          lead={<Status label={status.label} tone={status.tone} />}
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Warranty", href: "/account/warranty" },
            { label: c.number },
          ]}
        />
        {isNew ? (
          <p className="alert alert--success" role="status">
            Claim received. Bring the device, its charger and box to a store
            below and quote {c.number}.
          </p>
        ) : null}
        {c.status === "rejected" && c.rejectionReason ? (
          <p className="alert alert--error" role="status">
            Not covered: {c.rejectionReason}
          </p>
        ) : null}
        {c.status === "ready" ? (
          <p className="alert alert--success" role="status">
            {c.resolution === "replaced" ? "Your replacement" : "Your device"}{" "}
            is ready to collect
            {c.store ? ` at ${c.store}` : ""}. Bring an ID and quote {c.number}.
          </p>
        ) : null}
        <section
          className="card card--roomy stack"
          aria-labelledby="claim-details"
        >
          <h2 id="claim-details" className={styles.h2}>
            Details
          </h2>
          <Facts
            items={[
              [
                "Device",
                `${c.productName ?? ""} ${c.variantName ?? ""}`.trim(),
              ],
              ["IMEI or serial", c.imeiOrSerial],
              ["Fault", c.fault],
              [
                "Warranty until",
                c.warrantyEndsOn ? formatDate(c.warrantyEndsOn) : null,
              ],
              ["Store", c.store],
              ["Received", c.receivedAt ? formatDateTime(c.receivedAt) : null],
              [
                "Workshop",
                latest ? (repairWords[latest.status] ?? latest.status) : null,
              ],
              [
                "Outcome",
                c.resolution
                  ? (claimResolution[c.resolution] ?? c.resolution)
                  : null,
              ],
              [
                "Order",
                c.orderNumber ? (
                  <Link
                    key="o"
                    href={`/orders/${c.orderNumber}`}
                    className="link"
                  >
                    {c.orderNumber}
                  </Link>
                ) : null,
              ],
            ]}
          />
        </section>
        {(c.dropOff ?? []).length ? (
          <section
            className="card card--roomy stack"
            aria-labelledby="dropoff-title"
          >
            <h2 id="dropoff-title" className={styles.h2}>
              Where to bring it
            </h2>
            <ul role="list" className={styles.list}>
              {c.dropOff!.map((s) => (
                <li key={s.id} className={styles.row}>
                  <span className={styles.title}>{s.name}</span>
                  <span className="text-sm text-muted">{s.city}</span>
                  <span className={styles.meta}>{s.address}</span>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
      </div>
    </main>
  );
}
