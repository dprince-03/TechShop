import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { formatMoney } from "@techshop/api-client";
import { cars, getCar } from "@techshop/fixtures";
import { Breadcrumbs, CategoryArt, Icon } from "@techshop/ui/components";

import { carConditionLabel, formatMileage } from "@/components/CarCard";
import { ViewingForm } from "@/components/ViewingForm";

import styles from "./car.module.css";

export const dynamicParams = false;

export function generateStaticParams() {
  return cars.map((c) => ({ slug: c.slug }));
}

export async function generateMetadata(
  props: PageProps<"/cars/[slug]">,
): Promise<Metadata> {
  const car = getCar((await props.params).slug);
  return car
    ? {
        title: `${car.year} ${car.title} | TechShop Cars`,
        description: `${carConditionLabel[car.condition]}, ${formatMileage(car.mileageKm)}, ${car.location}.`,
      }
    : {};
}

/** SAMPLE inspection checklist — the real report comes from the inspection partner. */
const inspection = [
  "Engine and gearbox",
  "Brakes and suspension",
  "Body and paintwork",
  "Electricals and AC",
  "Documents and VIN check",
];

export default async function CarPage(props: PageProps<"/cars/[slug]">) {
  const car = getCar((await props.params).slug);
  if (!car) notFound();
  const title = `${car.year} ${car.title}`;

  const facts = [
    { label: "Year", value: String(car.year) },
    { label: "Mileage", value: formatMileage(car.mileageKm) },
    { label: "Condition", value: carConditionLabel[car.condition] },
    {
      label: "Transmission",
      value: car.transmission === "automatic" ? "Automatic" : "Manual",
    },
    { label: "Location", value: car.location },
    { label: "Seller", value: car.seller.name },
  ];

  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <Breadcrumbs
          items={[
            { label: "Home", href: "/" },
            { label: "Cars", href: "/c/cars" },
            { label: title },
          ]}
        />

        <div className={styles.layout}>
          <div className="stack stack--lg">
            <div className="frame frame--wide">
              <CategoryArt category="cars" size="lg" />
            </div>

            <div className="stack stack--sm">
              <div className="cluster">
                <span className="badge">
                  {carConditionLabel[car.condition]}
                </span>
                {car.inspected ? (
                  <span className="badge badge--accent">
                    <Icon name="check" size={12} />
                    Inspected
                  </span>
                ) : (
                  <span className="badge badge--outline">
                    Inspection pending
                  </span>
                )}
              </div>
              <h1 className={styles.title}>{title}</h1>
              <p className={styles.price}>{formatMoney(car.price)}</p>
              <p className="text-sm text-muted">
                Price set by {car.seller.name}. Financing may be available.
              </p>
            </div>

            <section aria-labelledby="facts-title">
              <h2 id="facts-title" className={styles.sectionTitle}>
                Key facts
              </h2>
              <dl className={styles.facts}>
                {facts.map((f) => (
                  <div key={f.label} className={styles.fact}>
                    <dt className="text-xs text-muted">{f.label}</dt>
                    <dd className={styles.factValue}>{f.value}</dd>
                  </div>
                ))}
              </dl>
            </section>

            <section aria-labelledby="inspection-title">
              <h2 id="inspection-title" className={styles.sectionTitle}>
                Inspection
              </h2>
              {car.inspected ? (
                <ul role="list" className={styles.checks}>
                  {inspection.map((item) => (
                    <li key={item} className={styles.check}>
                      <Icon
                        name="check"
                        size={18}
                        className={styles.checkIcon}
                      />
                      {item}
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-muted">
                  This car hasn’t been inspected yet. You can request an
                  inspection before your viewing.
                </p>
              )}
              <p className="text-xs text-muted">
                Inspection details shown are sample information.
              </p>
            </section>
          </div>

          <aside className={styles.aside}>
            <ViewingForm carTitle={title} />
          </aside>
        </div>
      </div>
    </main>
  );
}
