"use client";

import {
  createContext,
  useActionState,
  useContext,
  useEffect,
  useId,
  useRef,
  type ReactNode,
} from "react";

/**
 * Result of a Server Action. Mirrors `ActionState` from `@techshop/bff` (structural, so the
 * design system doesn't depend on the server package).
 */
export type FormResult = {
  ok?: boolean;
  message?: string;
  errors?: Record<string, string>;
};

const Result = createContext<FormResult>({});

type Props = {
  /** A Server Action, usually bound to the record it changes. */
  action: (prev: FormResult, form: FormData) => Promise<FormResult>;
  submitLabel: string;
  pendingLabel?: string;
  variant?: "primary" | "secondary" | "inverse" | "ghost";
  /** Clear the fields after a success (add-style forms). */
  resetOnSuccess?: boolean;
  /** Also keeps multipart encoding for file fields. */
  children: ReactNode;
};

/**
 * A form that posts to a Server Action and shows its result: the message above the fields,
 * field errors next to each `Field`, focus moved to the first invalid field. Works without JS
 * (plain form post) as well.
 */
export function ActionForm({
  action,
  submitLabel,
  pendingLabel = "Saving…",
  variant = "secondary",
  resetOnSuccess = false,
  children,
}: Props) {
  const [state, formAction, pending] = useActionState(action, {} as FormResult);
  const form = useRef<HTMLFormElement>(null);
  useEffect(() => {
    if (state.ok && resetOnSuccess) form.current?.reset();
    if (state.errors)
      form.current
        ?.querySelector<HTMLElement>('[aria-invalid="true"]')
        ?.focus();
  }, [state, resetOnSuccess]);
  return (
    <form ref={form} action={formAction} className="stack" noValidate>
      {state.message ? (
        <p
          role={state.ok ? "status" : "alert"}
          className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
        >
          {state.message}
        </p>
      ) : null}
      <Result.Provider value={state}>{children}</Result.Provider>
      <div>
        <button className={`btn btn--${variant}`} disabled={pending}>
          {pending ? pendingLabel : submitLabel}
        </button>
      </div>
    </form>
  );
}

type FieldProps = {
  name: string;
  label: string;
  hint?: string;
  /** Marked "(optional)"; fields are required by default. */
  optional?: boolean;
  type?: "text" | "number" | "tel" | "textarea" | "select" | "file";
  options?: { value: string; label: string }[];
  defaultValue?: string | number;
  inputMode?: "numeric" | "decimal" | "tel" | "text";
  autoComplete?: string;
  accept?: string;
  maxLength?: number;
  min?: number;
  max?: number;
  step?: number | "any";
};

/** A labelled field (label above, hint and error below) wired to the enclosing `ActionForm`. */
export function Field({
  name,
  label,
  hint,
  optional = false,
  type = "text",
  options,
  ...rest
}: FieldProps) {
  const { errors } = useContext(Result);
  const error = errors?.[name];
  const id = useId();
  const describedBy =
    [hint ? `${id}-hint` : "", error ? `${id}-error` : ""]
      .filter(Boolean)
      .join(" ") || undefined;
  const common = {
    id,
    name,
    required: !optional,
    "aria-invalid": error ? true : undefined,
    "aria-describedby": describedBy,
  };
  let control: ReactNode;
  if (type === "select") {
    control = (
      <select className="input" {...common} defaultValue={rest.defaultValue}>
        {optional ? <option value="">—</option> : null}
        {(options ?? []).map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    );
  } else if (type === "textarea") {
    control = <textarea className="input" rows={4} {...common} {...rest} />;
  } else {
    control = <input className="input" type={type} {...common} {...rest} />;
  }
  return (
    <div className="field">
      <label htmlFor={id} className="field__label">
        {label}
        {optional ? <span className="text-muted"> (optional)</span> : null}
      </label>
      {control}
      {hint ? (
        <span id={`${id}-hint`} className="field__hint">
          {hint}
        </span>
      ) : null}
      {error ? (
        <span id={`${id}-error`} className="field__error">
          {error}
        </span>
      ) : null}
    </div>
  );
}
