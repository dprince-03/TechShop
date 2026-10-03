import type { CSSProperties } from "react";
import Link from "next/link";

import { formatMoney, type CarListing } from "@techshop/api-client";
import { CategoryArt, Icon } from "@techshop/ui/components";

import styles from "./CarCard.module.css";

export const carConditionLabel = {
  "brand-new": "Brand new",
  "foreign-used": "Foreign used",
  "nigerian-used": "Nigerian used",
} as const;

export const formatMileage = (km: number) => `${km.toLocaleString("en-NG")} km`;

/** Car listing card. Cars lead to a viewing request, never add-to-cart. */
export function CarCard({ car }: { car: CarListing }) {
  return (
    <article className={`card card--interactive ${styles.car}`}>
      <div className="frame frame--wide">
        <CategoryArt category="cars" />
      </div>
      <div
        className="cluster"
        style={{ "--cluster-gap": "var(--space-1)" } as CSSProperties}
      >
        <span className="badge">{carConditionLabel[car.condition]}</span>
        {car.inspected ? (
          <span className="badge badge--accent">
            <Icon name="check" size={12} />
            Inspected
          </span>
        ) : null}
      </div>
      <h3 className={styles.title}>
        <Link href={`/cars/${car.slug}`} className={styles.stretched}>
          {car.year} {car.title}
        </Link>
      </h3>
      <p className="text-xs text-muted tabular-nums">
        {formatMileage(car.mileageKm)} ·{" "}
        {car.transmission === "automatic" ? "Automatic" : "Manual"}
      </p>
      <p className={`text-xs text-muted ${styles.location}`}>
        <Icon name="pin" size={14} />
        {car.location}
      </p>
      <p className={styles.price}>{formatMoney(car.price)}</p>
      <span className="link link--arrow text-sm">Book a viewing</span>
    </article>
  );
}
