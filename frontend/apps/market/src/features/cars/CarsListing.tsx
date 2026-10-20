import { cars } from "@techshop/fixtures";
import { Breadcrumbs, SampleDataNotice } from "@techshop/ui/components";

import type { MarketCategory } from "@/features/catalogue/mappers";

import { CarCard } from "./components/CarCard";
import styles from "./CarsListing.module.css";

/**
 * Cars are listings with a viewing flow, not add-to-cart products. The cars API arrives with
 * backend M7, so the listings here are SAMPLE data.
 */
export function CarsListing({ category }: { category: MarketCategory }) {
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <SampleDataNotice>
          Preview: car listings are sample data until the cars service is live.
        </SampleDataNotice>
        <div className="stack stack--sm">
          <Breadcrumbs
            items={[{ label: "Home", href: "/" }, { label: category.name }]}
          />
          <h1 className={styles.title}>{category.name}</h1>
          <p className="text-muted measure">
            {category.description} Book a viewing before you pay.
          </p>
        </div>
        <p className="text-sm text-muted" role="status">
          {cars.length} cars
        </p>
        <ul role="list" className={`grid ${styles.carGrid}`}>
          {cars.map((car) => (
            <li key={car.id}>
              <CarCard car={car} />
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
