"use client";

import { useRef, useState, type FormEvent } from "react";

import { normaliseNigerianPhone } from "@techshop/api-client";

import styles from "./PreviewForm.module.css";

export type FormField = {
  name: string;
  label: string;
  type?:
    | "text"
    | "email"
    | "tel"
    | "password"
    | "number"
    | "date"
    | "select"
    | "textarea"
    | "checkbox"
    | "file";
  required?: boolean;
  autoComplete?: string;
  inputMode?: "text" | "numeric" | "tel" | "email" | "decimal";
  placeholder?: string;
  hint?: string;
  options?: string[];
  minLength?: number;
  /** Name of another field this must equal (e.g. confirm password). */
  matches?: string;
  /** Half width beside the next half-width field on wide screens. */
  half?: boolean;
  accept?: string;
};

export type FormGroup = { legend?: string; fields: FormField[] };

type Props = {
  id: string;
  groups: FormGroup[];
  submitLabel: string;
  successTitle: string;
  /** Shown after a valid submit. Always states that nothing was sent (interface preview). */
  successMessage: string;
  /** Small print under the submit button. */
  note?: string;
};

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

function validateField(
  f: FormField,
  value: string,
  data: FormData,
): string | undefined {
  const label = f.label.toLowerCase();
  if (f.type === "checkbox")
    return f.required && value !== "on"
      ? `Please confirm: ${f.label.charAt(0).toLowerCase()}${f.label.slice(1)}`
      : undefined;
  if (f.type === "file")
    return f.required && !value ? `Upload ${label}.` : undefined;
  const v = value.trim();
  if (!v)
    return f.required
      ? f.type === "select"
        ? `Choose ${label}.`
        : `Enter ${label}.`
      : undefined;
  if (f.type === "email" && !EMAIL.test(v))
    return "Enter an email address like name@example.com.";
  if (f.type === "tel" && !normaliseNigerianPhone(v))
    return "Enter a Nigerian mobile number, for example 0803 123 4567.";
  if (f.minLength && v.length < f.minLength)
    return `${f.label} must be at least ${f.minLength} characters.`;
  if (f.matches && v !== String(data.get(f.matches) ?? ""))
    return `${f.label} doesn’t match.`;
  return undefined;
}

/**
 * Accessible form for interface previews: labels above fields, optional fields
 * marked, validation on blur and submit, focus on the first error, input kept
 * on error. Nothing is submitted anywhere until the matching API exists.
 */
export function PreviewForm({
  id,
  groups,
  submitLabel,
  successTitle,
  successMessage,
  note,
}: Props) {
  const formRef = useRef<HTMLFormElement>(null);
  const [errors, setErrors] = useState<Record<string, string | undefined>>({});
  const [done, setDone] = useState(false);
  const fields = groups.flatMap((g) => g.fields);

  const valueOf = (data: FormData, f: FormField) => {
    const v = data.get(f.name);
    return v instanceof File ? v.name : String(v ?? "");
  };

  const validateAll = (form: HTMLFormElement) => {
    const data = new FormData(form);
    return Object.fromEntries(
      fields.map((f) => [f.name, validateField(f, valueOf(data, f), data)]),
    );
  };

  const onBlur = (name: string) => () => {
    const all = validateAll(formRef.current!);
    setErrors((prev) => ({ ...prev, [name]: all[name] }));
  };

  const onSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const all = validateAll(e.currentTarget);
    setErrors(all);
    const first = fields.find((f) => all[f.name]);
    if (first) {
      e.currentTarget
        .querySelector<HTMLElement>(`[name="${first.name}"]`)
        ?.focus();
      return;
    }
    setDone(true);
  };

  if (done) {
    return (
      <div className={styles.done} role="status">
        <p className={styles.doneTitle}>{successTitle}</p>
        <p className="text-sm">{successMessage}</p>
        <p className="text-xs text-muted">Preview only — nothing was sent.</p>
        <button
          type="button"
          className="btn btn--secondary btn--sm"
          onClick={() => setDone(false)}
        >
          Edit details
        </button>
      </div>
    );
  }

  return (
    <form
      ref={formRef}
      id={id}
      className={styles.form}
      onSubmit={onSubmit}
      noValidate
    >
      {groups.map((g, gi) => (
        <fieldset key={gi} className={styles.group}>
          {g.legend ? (
            <legend className={styles.legend}>{g.legend}</legend>
          ) : null}
          <div className={styles.grid}>
            {g.fields.map((f) => {
              const fid = `${id}-${f.name}`;
              const err = errors[f.name];
              const describedBy = err
                ? `${fid}-err`
                : f.hint
                  ? `${fid}-hint`
                  : undefined;
              const common = {
                id: fid,
                name: f.name,
                onBlur: onBlur(f.name),
                "aria-invalid": err ? true : undefined,
                "aria-describedby": describedBy,
                "aria-required": f.required || undefined,
              };
              if (f.type === "checkbox") {
                return (
                  <div
                    key={f.name}
                    className={`${styles.field} ${styles.full}`}
                  >
                    <label className={styles.check}>
                      <input type="checkbox" {...common} />
                      <span>{f.label}</span>
                    </label>
                    {err ? (
                      <p id={`${fid}-err`} className={styles.error}>
                        {err}
                      </p>
                    ) : null}
                  </div>
                );
              }
              return (
                <div
                  key={f.name}
                  className={`${styles.field} ${f.half ? "" : styles.full}`}
                >
                  <label htmlFor={fid} className={styles.label}>
                    {f.label}
                    {f.required ? null : (
                      <span className={styles.optional}> (optional)</span>
                    )}
                  </label>
                  {f.type === "select" ? (
                    <select {...common} defaultValue="">
                      <option value="" disabled>
                        Select…
                      </option>
                      {f.options?.map((o) => (
                        <option key={o}>{o}</option>
                      ))}
                    </select>
                  ) : f.type === "textarea" ? (
                    <textarea
                      {...common}
                      rows={4}
                      placeholder={f.placeholder}
                    />
                  ) : (
                    <input
                      {...common}
                      type={f.type ?? "text"}
                      autoComplete={f.autoComplete}
                      inputMode={
                        f.inputMode ?? (f.type === "tel" ? "tel" : undefined)
                      }
                      placeholder={f.placeholder}
                      accept={f.accept}
                    />
                  )}
                  {err ? (
                    <p id={`${fid}-err`} className={styles.error}>
                      {err}
                    </p>
                  ) : f.hint ? (
                    <p id={`${fid}-hint`} className={styles.hint}>
                      {f.hint}
                    </p>
                  ) : null}
                </div>
              );
            })}
          </div>
        </fieldset>
      ))}
      <div className={styles.footer}>
        <button type="submit" className="btn btn--primary btn--lg">
          {submitLabel}
        </button>
        {note ? <p className={styles.hint}>{note}</p> : null}
      </div>
    </form>
  );
}
