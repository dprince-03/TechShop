import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { deviceType, formatKobo } from "@techshop/api-client";
import { PageHeader } from "@techshop/ui/components";

import { TradeInForm } from "@/features/aftersales/components/TradeInForms";
import { tradeInDevices } from "@/features/aftersales/server";
import { bff } from "@/lib/bff";

import styles from "./trade-in.module.css";

export const metadata: Metadata = {
  title: "Trade in your device | TechShop Market",
  description:
    "Get an estimate for your old phone, laptop or console and put it towards something new.",
};

const steps = [
  {
    title: "Tell us about your device",
    text: "Model, storage and condition. It takes a minute.",
  },
  {
    title: "Get an estimate",
    text: "Listed models are priced at once from today's prices; others within a working day.",
  },
  {
    title: "Bring it to a store",
    text: "We check the condition, the IMEI or serial, and that it isn't reported stolen. Bring an ID.",
  },
  {
    title: "Accept the offer",
    text: "Take it as store credit straight away, or by bank transfer within 2 working days. You have 7 days to decide.",
  },
];

const TYPES = ["phone", "laptop", "tablet", "console", "smartwatch"] as const;
type DeviceType = (typeof TYPES)[number];

export default async function TradeInPage(props: PageProps<"/trade-in">) {
  await connection();
  const raw = (await props.searchParams).type;
  const type = TYPES.find((t) => t === raw) as DeviceType | undefined;
  const session = await bff.getSession();
  const models = type
    ? ((await tradeInDevices(type).catch(() => [])) ?? [])
    : [];

  return (
    <main id="main" className="section section--page">
      <div className="container">
        <PageHeader
          eyebrow="Trade-in"
          title="Turn your old device into credit"
          lead="Phones, laptops, tablets, consoles and smartwatches. The final value is confirmed at inspection."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Trade-in" }]}
        />
        <div className={styles.layout}>
          <div className="card card--roomy stack">
            <form method="get" className="cluster" aria-label="Device type">
              <div className="field">
                <label htmlFor="type" className="field__label">
                  What are you trading in?
                </label>
                <select
                  id="type"
                  name="type"
                  className="input"
                  defaultValue={type ?? ""}
                >
                  <option value="" disabled>
                    Choose a device type
                  </option>
                  {TYPES.map((t) => (
                    <option key={t} value={t}>
                      {deviceType[t]}
                    </option>
                  ))}
                </select>
              </div>
              <button className="btn btn--secondary">Continue</button>
            </form>
            {type ? (
              session ? (
                <TradeInForm
                  key={type}
                  deviceType={type}
                  models={models.map((m) => ({
                    value: m.variantId,
                    label: `${m.brand} ${m.productName} ${m.variantName} · up to ${formatKobo(m.estimates.like_new ?? 0)}`,
                  }))}
                />
              ) : (
                <p className="text-sm">
                  <Link
                    href={`/sign-in?next=${encodeURIComponent(`/trade-in?type=${type}`)}`}
                    className="btn btn--primary"
                  >
                    Sign in to get your estimate
                  </Link>
                </p>
              )
            ) : null}
            {type && models.length ? (
              <p className="text-xs text-muted">
                “Up to” is for a device in like-new condition. Good, fair and
                faulty devices get less.
              </p>
            ) : null}
          </div>
          <aside aria-labelledby="how-title" className="stack">
            <h2 id="how-title" className="text-xl">
              How trade-in works
            </h2>
            <ol className={styles.steps}>
              {steps.map((s, i) => (
                <li key={s.title} className={styles.step}>
                  <span className={styles.num} aria-hidden="true">
                    {i + 1}
                  </span>
                  <div>
                    <p className={styles.stepTitle}>{s.title}</p>
                    <p className="text-sm text-muted">{s.text}</p>
                  </div>
                </li>
              ))}
            </ol>
            {session ? (
              <Link href="/account/trade-ins" className="link link--arrow">
                Your trade-ins
              </Link>
            ) : null}
          </aside>
        </div>
      </div>
    </main>
  );
}
