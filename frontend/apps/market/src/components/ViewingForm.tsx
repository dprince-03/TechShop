"use client";

import { useState, type FormEvent } from "react";

import { normaliseNigerianPhone } from "@techshop/api-client";

import styles from "./ViewingForm.module.css";

type Errors = Partial<Record<"name" | "phone" | "date", string>>;

const today = () => new Date().toISOString().slice(0, 10);

/**
 * Viewing request form. INTERFACE ONLY: validates and confirms, but nothing is
 * sent until the bookings API exists — the confirmation says so.
 */
export function ViewingForm({ carTitle }: { carTitle: string }) {
  const [errors, setErrors] = useState<Errors>({});
  const [done, setDone] = useState<string | null>(null);

  const validate = (form: HTMLFormElement): Errors => {
    const data = new FormData(form);
    const e: Errors = {};
    if (!String(data.get("name") ?? "").trim())
      e.name = "Enter your name so the dealer knows who to expect.";
    if (!normaliseNigerianPhone(String(data.get("phone") ?? "")))
      e.phone = "Enter a Nigerian mobile number, for example 0803 123 4567.";
    const date = String(data.get("date") ?? "");
    if (!date) e.date = "Choose a day for the viewing.";
    else if (date < today()) e.date = "Choose today or a later date.";
    return e;
  };

  const onBlur =
    (field: keyof Errors) => (ev: React.FocusEvent<HTMLInputElement>) => {
      const e = validate(ev.currentTarget.form!);
      setErrors((prev) => ({ ...prev, [field]: e[field] }));
    };

  const onSubmit = (ev: FormEvent<HTMLFormElement>) => {
    ev.preventDefault();
    const e = validate(ev.currentTarget);
    setErrors(e);
    const firstError = Object.keys(e)[0];
    if (firstError) {
      ev.currentTarget
        .querySelector<HTMLInputElement>(`[name="${firstError}"]`)
        ?.focus();
      return;
    }
    const data = new FormData(ev.currentTarget);
    setDone(
      `Viewing request for ${carTitle} on ${new Date(String(data.get("date"))).toLocaleDateString("en-NG", { weekday: "long", day: "numeric", month: "long" })} (${data.get("slot")}) is ready. Preview only — it hasn’t been sent to the dealer.`,
    );
  };

  if (done) {
    return (
      <div className={styles.done} role="status">
        <p className={styles.doneTitle}>Request ready</p>
        <p className="text-sm">{done}</p>
        <button
          type="button"
          className="btn btn--secondary btn--sm"
          onClick={() => setDone(null)}
        >
          Change details
        </button>
      </div>
    );
  }

  return (
    <form
      className={styles.form}
      onSubmit={onSubmit}
      noValidate
      aria-labelledby="viewing-title"
    >
      <h2 id="viewing-title" className={styles.title}>
        Book a viewing
      </h2>

      <div className={styles.field}>
        <label htmlFor="v-name">Full name</label>
        <input
          id="v-name"
          name="name"
          autoComplete="name"
          onBlur={onBlur("name")}
          aria-invalid={!!errors.name}
          aria-describedby={errors.name ? "v-name-err" : undefined}
        />
        {errors.name ? (
          <p id="v-name-err" className={styles.error}>
            {errors.name}
          </p>
        ) : null}
      </div>

      <div className={styles.field}>
        <label htmlFor="v-phone">Phone number</label>
        <input
          id="v-phone"
          name="phone"
          type="tel"
          inputMode="tel"
          autoComplete="tel"
          placeholder="0803 123 4567"
          onBlur={onBlur("phone")}
          aria-invalid={!!errors.phone}
          aria-describedby={errors.phone ? "v-phone-err" : "v-phone-hint"}
        />
        {errors.phone ? (
          <p id="v-phone-err" className={styles.error}>
            {errors.phone}
          </p>
        ) : (
          <p id="v-phone-hint" className={styles.hint}>
            The dealer will call to confirm the time.
          </p>
        )}
      </div>

      <div className={styles.row}>
        <div className={styles.field}>
          <label htmlFor="v-date">Day</label>
          <input
            id="v-date"
            name="date"
            type="date"
            min={today()}
            onBlur={onBlur("date")}
            aria-invalid={!!errors.date}
            aria-describedby={errors.date ? "v-date-err" : undefined}
          />
          {errors.date ? (
            <p id="v-date-err" className={styles.error}>
              {errors.date}
            </p>
          ) : null}
        </div>
        <div className={styles.field}>
          <label htmlFor="v-slot">Time</label>
          <select id="v-slot" name="slot" defaultValue="Morning (9am–12pm)">
            <option>Morning (9am–12pm)</option>
            <option>Afternoon (12pm–3pm)</option>
            <option>Late afternoon (3pm–6pm)</option>
          </select>
        </div>
      </div>

      <div className={styles.field}>
        <label htmlFor="v-note">
          Questions for the dealer{" "}
          <span className={styles.optional}>(optional)</span>
        </label>
        <textarea id="v-note" name="note" rows={3} />
      </div>

      <button type="submit" className="btn btn--primary btn--lg btn--block">
        Request viewing
      </button>
      <p className={styles.hint}>
        Free to book. You only pay after inspecting the car.
      </p>
    </form>
  );
}
