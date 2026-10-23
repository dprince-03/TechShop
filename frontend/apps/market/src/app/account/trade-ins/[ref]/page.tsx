import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import {
  deviceCondition,
  formatKobo,
  tradeInStatus,
} from "@techshop/api-client";
import { PageHeader, Status } from "@techshop/ui/components";

import styles from "@/features/aftersales/components/Aftersales.module.css";
import { Facts } from "@/features/aftersales/components/Records";
import {
  AcceptOfferForm,
  CancelTradeInForm,
  DeclineOfferForm,
} from "@/features/aftersales/components/TradeInForms";
import { banks, myTradeIn } from "@/features/aftersales/server";
import { formatDateTime } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Trade-in | TechShop Market",
  robots: { index: false },
};

export default async function TradeInPage(
  props: PageProps<"/account/trade-ins/[ref]">,
) {
  await connection();
  const { ref } = await props.params;
  await bff.requireSession(`/account/trade-ins/${ref}`);
  const t = await myTradeIn(ref);
  if (!t) notFound();
  const isNew = (await props.searchParams).new === "1";
  const offerOpen = t.status === "offer_sent";
  const bankList = offerOpen ? ((await banks()) ?? []) : [];
  const status = tradeInStatus[t.status] ?? {
    label: t.status,
    tone: "neutral" as const,
  };

  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title={`Trade-in ${t.reference}`}
          lead={<Status label={status.label} tone={status.tone} />}
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Trade-ins", href: "/account/trade-ins" },
            { label: t.reference },
          ]}
        />
        {isNew ? (
          <p className="alert alert--success" role="status">
            {t.estimate != null
              ? `Your estimate is ${formatKobo(t.estimate)}. Bring the device and an ID to a store below; the final offer follows the inspection.`
              : "Thanks. We'll price it within a working day and let you know."}
          </p>
        ) : null}
        {offerOpen && t.offer != null ? (
          <section
            className="card card--roomy stack"
            aria-labelledby="offer-title"
          >
            <h2 id="offer-title" className={styles.h2}>
              Our offer:{" "}
              <span className="tabular-nums">{formatKobo(t.offer)}</span>
            </h2>
            <p className="text-sm text-muted">
              Graded{" "}
              {deviceCondition[t.inspectedCondition ?? ""] ??
                t.inspectedCondition}
              .
              {t.offerExpiresAt
                ? ` Answer by ${formatDateTime(t.offerExpiresAt)}.`
                : ""}
            </p>
            <AcceptOfferForm
              reference={t.reference}
              banks={bankList.map((b) => ({ value: b.code, label: b.name }))}
            />
            <DeclineOfferForm reference={t.reference} />
          </section>
        ) : null}
        <section
          className="card card--roomy stack"
          aria-labelledby="ti-details"
        >
          <h2 id="ti-details" className={styles.h2}>
            Details
          </h2>
          <Facts
            items={[
              ["Device", t.device],
              [
                "Your description",
                deviceCondition[t.declaredCondition] ?? t.declaredCondition,
              ],
              [
                "Estimate",
                t.estimate != null ? formatKobo(t.estimate) : "Being priced",
              ],
              ["Inspection", t.inspectionNotes],
              ["Store", t.store],
              [
                "Paid",
                t.paidAt
                  ? `${formatDateTime(t.paidAt)}${t.payoutMethod === "store_credit" ? " as store credit" : t.accountLast4 ? ` to account ending ${t.accountLast4}` : ""}`
                  : null,
              ],
            ]}
          />
        </section>
        {(t.dropOff ?? []).length ? (
          <section
            className="card card--roomy stack"
            aria-labelledby="ti-stores"
          >
            <h2 id="ti-stores" className={styles.h2}>
              Where to bring it
            </h2>
            <ul role="list" className={styles.list}>
              {t.dropOff!.map((s) => (
                <li key={s.id} className={styles.row}>
                  <span className={styles.title}>{s.name}</span>
                  <span className="text-sm text-muted">{s.city}</span>
                  <span className={styles.meta}>{s.address}</span>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
        {t.status === "submitted" || t.status === "estimated" ? (
          <CancelTradeInForm reference={t.reference} />
        ) : null}
      </div>
    </main>
  );
}
