import { dateTime } from "@/lib/format";

import styles from "./DropoffPanel.module.css";

type Point = { id: string; name: string; address: string; city: string };

/** What a seller does after packing: label the parcel and bring it to a TechShop hub. */
export function DropoffPanel({
  label,
  handoverBy,
  points,
}: {
  label: string;
  handoverBy: string | null;
  points: Point[];
}) {
  return (
    <section aria-labelledby="drop-title" className="card card--roomy stack">
      <h2 id="drop-title" className="text-xl">
        Drop it at a TechShop hub
      </h2>
      <p>Write this code clearly on the parcel:</p>
      <p className={styles.label}>{label}</p>
      {handoverBy ? (
        <p className="alert alert--warning">
          Drop it off by {dateTime.format(new Date(handoverBy))}, or the order
          is cancelled, the customer is refunded and it counts against your
          shop.
        </p>
      ) : null}
      <h3 className="text-lg">Drop-off points</h3>
      <ul className={styles.points}>
        {points.map((p) => (
          <li key={p.id}>
            <strong>{p.name}</strong>
            <br />
            <span className="text-sm text-muted">
              {p.address}, {p.city}
            </span>
          </li>
        ))}
      </ul>
      <p className="text-sm text-muted">
        Hub staff scan the code. From there TechShop delivers the order, and
        you’re paid after the customer receives it.
      </p>
    </section>
  );
}
