import { dateOnly, naira } from "@/lib/format";

import styles from "./Balance.module.css";

/** Plain-language reasons a shop's payouts are held. */
const holdText: Record<string, string> = {
  kyc_incomplete: "Your shop isn’t verified yet.",
  enforcement: "Your shop is suspended.",
  no_bank_account: "Add a payout account.",
  bank_account_changed:
    "A payout account was added in the last 24 hours. For your safety, payouts wait 24 hours.",
  credentials_changed:
    "Your sign-in details changed in the last 24 hours. For your safety, payouts wait 24 hours.",
  risk_case: "We’re reviewing your account. Support will contact you.",
  negative_balance:
    "Refunds are more than your earnings right now; they’ll be taken from your next sales.",
  manual: "Our finance team has paused payouts. Support will contact you.",
};

type BalanceData = {
  available: number;
  pending: number;
  onHold: boolean;
  holdReasons: string[] | null;
  minimum: number;
  nextPayoutDay: string;
};

/** Available, pending and the next payout day, with any hold explained. */
export function Balance({ b }: { b: BalanceData }) {
  const reasons = [
    ...new Set((b.holdReasons ?? []).map((r) => holdText[r] ?? r)),
  ];
  return (
    <section aria-labelledby="balance-title" className="stack">
      <h2 id="balance-title" className="text-xl">
        Your balance
      </h2>
      <ul role="list" className={styles.figures}>
        <li className={`card ${styles.figure}`}>
          <p className="text-sm text-muted">Available</p>
          <p className={`${styles.value} tabular-nums`}>{naira(b.available)}</p>
          <p className="text-xs text-muted">
            Paid on {dateOnly.format(new Date(b.nextPayoutDay))} if it’s at
            least {naira(b.minimum)}
          </p>
        </li>
        <li className={`card ${styles.figure}`}>
          <p className="text-sm text-muted">Pending</p>
          <p className={`${styles.value} tabular-nums`}>{naira(b.pending)}</p>
          <p className="text-xs text-muted">
            Available 7 days after delivery, once the return window closes
          </p>
        </li>
      </ul>
      {b.onHold ? (
        <div className="alert alert--warning stack">
          <p>
            <strong>Payouts are on hold.</strong>
          </p>
          <ul>
            {reasons.map((r) => (
              <li key={r}>{r}</li>
            ))}
          </ul>
        </div>
      ) : null}
      <p className="text-sm text-muted">
        We pay out on Tuesdays and Fridays (business days), after commission.
      </p>
    </section>
  );
}
