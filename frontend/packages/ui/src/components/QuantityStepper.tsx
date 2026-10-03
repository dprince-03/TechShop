"use client";

import { useId } from "react";

import { Icon } from "./Icon";
import styles from "./QuantityStepper.module.css";

type Props = {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  /** Visible label, e.g. "Quantity". Use `labelHidden` to keep it for screen readers only. */
  label?: string;
  labelHidden?: boolean;
};

/** − [n] + control. Buttons are 44px targets; the input accepts typed values and clamps them. */
export function QuantityStepper({
  value,
  onChange,
  min = 1,
  max = 99,
  label = "Quantity",
  labelHidden,
}: Props) {
  const id = useId();
  // "Quantity of Volt 20,000mAh" → "quantity of Volt 20,000mAh": only the first letter changes.
  const spoken = label.charAt(0).toLowerCase() + label.slice(1);
  const clamp = (n: number) =>
    Math.min(max, Math.max(min, Number.isFinite(n) ? Math.round(n) : min));

  return (
    <div className={styles.wrap}>
      <label
        htmlFor={id}
        className={labelHidden ? "visually-hidden" : styles.label}
      >
        {label}
      </label>
      <div className={styles.stepper}>
        <button
          type="button"
          className={styles.button}
          onClick={() => onChange(clamp(value - 1))}
          disabled={value <= min}
          aria-label={`Decrease ${spoken}`}
        >
          <Icon name="minus" size={16} />
        </button>
        <input
          id={id}
          className={styles.input}
          type="number"
          inputMode="numeric"
          min={min}
          max={max}
          value={value}
          onChange={(e) => onChange(clamp(e.target.valueAsNumber))}
        />
        <button
          type="button"
          className={styles.button}
          onClick={() => onChange(clamp(value + 1))}
          disabled={value >= max}
          aria-label={`Increase ${spoken}`}
        >
          <Icon name="plus" size={16} />
        </button>
      </div>
    </div>
  );
}
