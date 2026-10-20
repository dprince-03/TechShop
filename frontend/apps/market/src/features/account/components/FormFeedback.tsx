import type { FormState } from "../actions";

/** Status line for a form action: success announced politely, errors as alerts. */
export function FormFeedback({ state }: { state: FormState }) {
  if (!state.message) return <p role="status" className="visually-hidden" />;
  return state.ok ? (
    <p role="status" className="alert alert--success">
      {state.message}
    </p>
  ) : (
    <p role="alert" className="alert alert--error">
      {state.message}
    </p>
  );
}

/** Props for an input tied to a field error (aria-invalid + aria-describedby). */
export function fieldProps(state: FormState, name: string) {
  const error = state.errors?.[name];
  return {
    name,
    id: `f-${name}`,
    "aria-invalid": error ? true : undefined,
    "aria-describedby": error ? `f-${name}-error` : undefined,
  };
}

export function FieldError({
  state,
  name,
}: {
  state: FormState;
  name: string;
}) {
  const error = state.errors?.[name];
  return error ? (
    <span id={`f-${name}-error`} className="field__error">
      {error}
    </span>
  ) : null;
}
