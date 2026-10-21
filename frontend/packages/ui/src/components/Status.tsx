import styles from "./Status.module.css";

export type Tone = "success" | "warning" | "info" | "danger" | "neutral";

/** Status as a coloured dot plus words (never colour alone). */
export function Status({
  label,
  tone = "neutral",
}: {
  label: string;
  tone?: Tone;
}) {
  return <span className={`${styles.status} ${styles[tone]}`}>{label}</span>;
}

/** snake_case API status → "Sentence case" for display. */
export const humanise = (s: string) => {
  const t = s.replace(/_/g, " ");
  return t.charAt(0).toUpperCase() + t.slice(1);
};
