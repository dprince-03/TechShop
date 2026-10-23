import { notFound } from "next/navigation";
import { connection } from "next/server";

import {
  ApiError,
  deviceCondition,
  deviceType,
  formatKobo,
  tradeInStatus,
} from "@techshop/api-client";
import { Status, humanise } from "@techshop/ui/components";

import { Details } from "@/components/Details";
import { PageHead } from "@/components/PageHead";
import {
  EstimateForm,
  InspectTradeInForm,
  PayTradeInControl,
  ReceiveTradeInForm,
} from "@/features/aftersales/components/Controls";
import {
  locations,
  tradeIn,
  tradeInModels,
} from "@/features/aftersales/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Trade-in | TechShop Staff" };

export default async function TradeInPage(
  props: PageProps<"/trade-ins/[ref]">,
) {
  await connection();
  const me = await requireStaff("aftersales.manage");
  const { ref } = await props.params;
  const t = await tradeIn(ref).catch((e) => {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      notFound();
    throw e;
  });
  const needsModels =
    t.status === "submitted" || t.status === "awaiting_inspection";
  const [places, models] = await Promise.all([
    locations().catch(() => []),
    needsModels
      ? tradeInModels(t.deviceType).catch(() => [])
      : Promise.resolve([]),
  ]);
  const stores = (places ?? []).map((l) => ({
    value: l.id,
    label: `${l.name}, ${l.city}`,
  }));
  const modelOptions = (models ?? []).map((m) => ({
    value: m.variantId,
    label: `${m.brand} ${m.productName} ${m.variantName}`,
  }));
  const status = tradeInStatus[t.status] ?? {
    label: t.status,
    tone: "neutral" as const,
  };
  return (
    <div className="stack stack--lg">
      <PageHead
        title={`Trade-in ${t.reference}`}
        description={<Status label={status.label} tone={status.tone} />}
      />
      <section className="card stack" aria-labelledby="ti-sum">
        <h2 id="ti-sum" className="text-xl">
          Device
        </h2>
        <Details
          items={[
            ["Customer", t.customer],
            ["Type", deviceType[t.deviceType] ?? t.deviceType],
            ["Device", t.device],
            [
              "Declared",
              deviceCondition[t.declaredCondition] ?? t.declaredCondition,
            ],
            [
              "Inspected",
              t.inspectedCondition
                ? (deviceCondition[t.inspectedCondition] ??
                  t.inspectedCondition)
                : null,
            ],
            ["IMEI", t.imei],
            ["Serial", t.serial],
            ["Estimate", t.estimate != null ? formatKobo(t.estimate) : null],
            ["Offer", t.offer != null ? formatKobo(t.offer) : null],
            [
              "Offer expires",
              t.offerExpiresAt ? formatDateTime(t.offerExpiresAt) : null,
            ],
            ["Store", t.store],
            ["Inspection notes", t.inspectionNotes],
            ["Payout", t.payoutMethod ? humanise(t.payoutMethod) : null],
            [
              "Bank account",
              t.accountLast4
                ? `${t.accountName ?? ""} · ending ${t.accountLast4}`
                : null,
            ],
            [
              "Paid",
              t.paidAt
                ? `${formatDateTime(t.paidAt)} · ${t.payoutReference ?? ""}`
                : null,
            ],
          ]}
        />
      </section>
      {t.status === "submitted" ? (
        <section className="card stack" aria-labelledby="ti-est">
          <h2 id="ti-est" className="text-xl">
            Estimate
          </h2>
          <EstimateForm reference={t.reference} models={modelOptions} />
        </section>
      ) : null}
      {t.status === "estimated" ? (
        <section className="card stack" aria-labelledby="ti-recv">
          <h2 id="ti-recv" className="text-xl">
            Take the device in
          </h2>
          <ReceiveTradeInForm
            reference={t.reference}
            stores={stores}
            needsImei={t.deviceType === "phone" || t.deviceType === "tablet"}
          />
        </section>
      ) : null}
      {t.status === "awaiting_inspection" ? (
        <section className="card stack" aria-labelledby="ti-insp">
          <h2 id="ti-insp" className="text-xl">
            Inspect and offer
          </h2>
          <p className="text-sm text-muted">
            Check the risk queue first if this device was flagged when it came
            in.
          </p>
          <InspectTradeInForm reference={t.reference} models={modelOptions} />
        </section>
      ) : null}
      {t.status === "accepted" && t.payoutMethod === "bank" ? (
        <section className="card stack" aria-labelledby="ti-pay">
          <h2 id="ti-pay" className="text-xl">
            Bank payout
          </h2>
          {can(me, "refunds.approve") ? (
            <PayTradeInControl reference={t.reference} />
          ) : (
            <p className="text-sm text-muted">
              A finance manager records the transfer once it’s sent.
            </p>
          )}
        </section>
      ) : null}
    </div>
  );
}
