"use client";

import { useActionState } from "react";

import type { Schema } from "@techshop/api-client";

import { updateProfile, type FormState } from "../actions";

import styles from "./Account.module.css";
import { FieldError, FormFeedback, fieldProps } from "./FormFeedback";

export function ProfileForm({ me }: { me: Schema<"Me"> }) {
  const [state, action, pending] = useActionState(
    updateProfile,
    {} as FormState,
  );
  return (
    <form action={action} className={styles.form} noValidate>
      <FormFeedback state={state} />
      <div className={styles.grid}>
        <label className="field">
          <span className="field__label">First name</span>
          <input
            {...fieldProps(state, "firstName")}
            className="input"
            autoComplete="given-name"
            defaultValue={me.firstName}
            required
          />
          <FieldError state={state} name="firstName" />
        </label>
        <label className="field">
          <span className="field__label">Last name</span>
          <input
            {...fieldProps(state, "lastName")}
            className="input"
            autoComplete="family-name"
            defaultValue={me.lastName}
            required
          />
          <FieldError state={state} name="lastName" />
        </label>
      </div>
      <dl className={styles.facts}>
        <div>
          <dt>Mobile</dt>
          <dd>{me.phone ? `0${me.phone.slice(4)}` : "Not set"}</dd>
        </div>
        <div>
          <dt>Email</dt>
          <dd>{me.email ?? "Not set"}</dd>
        </div>
      </dl>
      <label className="check">
        <input
          type="checkbox"
          name="marketingOptIn"
          defaultChecked={me.marketingOptIn}
        />
        Send me deals and news by SMS and email. You can stop any time.
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Saving…" : "Save profile"}
      </button>
    </form>
  );
}
