"use client";

import { useEffect, useState } from "react";

import styles from "./Countdown.module.css";

const WAT_OFFSET_MS = 60 * 60 * 1000; // West Africa Time = UTC+1

/** Next midnight in West Africa Time. */
function nextMidnightWAT(now: number) {
  const wat = new Date(now + WAT_OFFSET_MS);
  return (
    Date.UTC(wat.getUTCFullYear(), wat.getUTCMonth(), wat.getUTCDate() + 1) -
    WAT_OFFSET_MS
  );
}

const pad = (n: number) => String(n).padStart(2, "0");

/**
 * Countdown to a real end time (default: tonight's midnight WAT, when daily deals end).
 * Only use with genuine deadlines — fake urgency is banned.
 */
export function Countdown({
  endsAt,
  label = "Ends in",
}: {
  endsAt?: number;
  label?: string;
}) {
  const [remaining, setRemaining] = useState<number | null>(null);

  useEffect(() => {
    const target = endsAt ?? nextMidnightWAT(Date.now());
    const tick = () => setRemaining(Math.max(0, target - Date.now()));
    tick();
    const id = window.setInterval(tick, 1000);
    return () => window.clearInterval(id);
  }, [endsAt]);

  const s = Math.floor((remaining ?? 0) / 1000);
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;

  return (
    <span className={styles.countdown}>
      <span className="text-sm text-muted">{label}</span>
      <span
        className={styles.time}
        role="timer"
        aria-label={remaining === null ? undefined : `${h} hours ${m} minutes`}
      >
        {remaining === null ? "--:--:--" : `${pad(h)}:${pad(m)}:${pad(sec)}`}
      </span>
    </span>
  );
}
