import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { formatKobo } from "@techshop/api-client";
import { PageHeader } from "@techshop/ui/components";

import { getDeliveryEstimate, getStates } from "@/features/catalogue/api";

export const metadata: Metadata = {
  title: "Delivery & fees | TechShop Market",
  description: "Check delivery fees and times for your state.",
};

const days = (min?: number, max?: number) =>
  min === undefined || max === undefined
    ? undefined
    : max === 0
      ? "today"
      : min === max
        ? `${max} working day${max === 1 ? "" : "s"}`
        : `${min}–${max} working days`;

/** Live delivery checker: a GET form, so it works without JavaScript and the result is linkable. */
export default async function DeliveryPage(props: PageProps<"/delivery">) {
  await connection();
  const states = await getStates();
  const raw = (await props.searchParams).state;
  const state =
    typeof raw === "string" && states.some((s) => s.code === raw)
      ? raw
      : undefined;
  const [own, seller] = state
    ? await Promise.all([
        getDeliveryEstimate(state, 1000, "techshop"),
        getDeliveryEstimate(state, 1000, "seller"),
      ])
    : [null, null];
  const stateName = states.find((s) => s.code === state)?.name;

  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Delivery & fees"
          lead="Choose your state to see what delivery costs and how long it takes. Your exact fee is confirmed at checkout."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Delivery" }]}
        />
        <form
          method="get"
          className="card card--roomy cluster"
          style={{ alignItems: "flex-end" }}
        >
          <label className="field" style={{ flex: "1 1 14rem" }}>
            <span className="field__label">State</span>
            <select
              name="state"
              className="input"
              defaultValue={state ?? ""}
              required
            >
              <option value="" disabled>
                Choose a state
              </option>
              {states.map((s) => (
                <option key={s.code} value={s.code}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
          <button className="btn btn--inverse">Check delivery</button>
        </form>

        {state ? (
          <section
            className="card card--roomy stack"
            aria-live="polite"
            aria-labelledby="result-title"
          >
            <h2 id="result-title" className="text-xl">
              Delivery to {stateName}
            </h2>
            {own === null && seller === null ? (
              <p className="alert alert--error" role="alert">
                We couldn’t check delivery right now. Please try again.
              </p>
            ) : (
              <dl className="stack">
                {[
                  { label: "Items sold by TechShop", e: own },
                  { label: "Items shipped by marketplace sellers", e: seller },
                ].map(({ label, e }) => (
                  <div key={label} className="stack stack--sm">
                    <dt className="text-sm text-muted">{label}</dt>
                    <dd>
                      {e?.covered ? (
                        <>
                          <strong className="tabular-nums">
                            {formatKobo(e.feeKobo ?? 0)}
                          </strong>{" "}
                          for a parcel up to 1 kg
                          {days(e.etaMinDays, e.etaMaxDays)
                            ? `, usually ${days(e.etaMinDays, e.etaMaxDays)}`
                            : ""}
                          {e.zone ? ` (${e.zone} zone)` : ""}.
                        </>
                      ) : (
                        <>
                          Delivered by our courier partners. The fee is shown at
                          checkout.
                        </>
                      )}
                    </dd>
                  </div>
                ))}
              </dl>
            )}
          </section>
        ) : null}
        <p className="text-sm text-muted">
          <Link href="/help/delivery" className="link">
            How delivery works
          </Link>
        </p>
      </div>
    </main>
  );
}
