"use client";

import { useActionState } from "react";

import type { Schema } from "@techshop/api-client";

import { saveAddress, type FormState } from "../actions";

import styles from "./Account.module.css";
import { FieldError, FormFeedback, fieldProps } from "./FormFeedback";

type Address = Schema<"IdentityAddress">;
type State = Schema<"IdentityNigerianState">;

/** Add or edit a delivery address. State → LGA keeps delivery fees exact (Tesler's Law). */
export function AddressForm({
  states,
  address,
  onDone,
  submitLabel,
}: {
  states: State[];
  address?: Address;
  onDone?: () => void;
  submitLabel?: string;
}) {
  // Tell the parent once a save succeeds (e.g. close the editor, select the new address).
  const [state, action, pending] = useActionState(
    async (prev: FormState, form: FormData) => {
      const result = await saveAddress(prev, form);
      if (result.ok) onDone?.();
      return result;
    },
    {} as FormState,
  );

  return (
    <form action={action} className={styles.form} noValidate>
      {address ? <input type="hidden" name="id" value={address.id} /> : null}
      <FormFeedback state={state} />
      <div className={styles.grid}>
        <label className="field">
          <span className="field__label">Recipient’s name</span>
          <input
            {...fieldProps(state, "recipientName")}
            className="input"
            autoComplete="name"
            defaultValue={address?.recipientName}
            required
          />
          <FieldError state={state} name="recipientName" />
        </label>
        <label className="field">
          <span className="field__label">Phone</span>
          <input
            {...fieldProps(state, "phone")}
            className="input"
            type="tel"
            inputMode="tel"
            autoComplete="tel-national"
            placeholder="0803 123 4567"
            defaultValue={
              address?.phone ? `0${address.phone.slice(4)}` : undefined
            }
            required
          />
          <FieldError state={state} name="phone" />
        </label>
      </div>
      <label className="field">
        <span className="field__label">Street address</span>
        <input
          {...fieldProps(state, "line1")}
          className="input"
          autoComplete="address-line1"
          defaultValue={address?.line1}
          required
        />
        <FieldError state={state} name="line1" />
      </label>
      <label className="field">
        <span className="field__label">
          Apartment, floor or estate{" "}
          <span className="text-muted">(optional)</span>
        </span>
        <input
          {...fieldProps(state, "line2")}
          className="input"
          autoComplete="address-line2"
          defaultValue={address?.line2 ?? undefined}
        />
      </label>
      <div className={styles.grid}>
        <label className="field">
          <span className="field__label">State</span>
          <select
            {...fieldProps(state, "stateCode")}
            className="input"
            defaultValue={address?.stateCode ?? ""}
            required
          >
            <option value="" disabled>
              Choose a state
            </option>
            {states.map((s) => (
              <option key={s.code} value={s.code}>
                {s.name}
              </option>
            ))}
          </select>
          <FieldError state={state} name="stateCode" />
        </label>
        <label className="field">
          <span className="field__label">
            LGA <span className="text-muted">(optional)</span>
          </span>
          <input
            {...fieldProps(state, "lga")}
            className="input"
            defaultValue={address?.lga ?? undefined}
          />
        </label>
      </div>
      <div className={styles.grid}>
        <label className="field">
          <span className="field__label">City or town</span>
          <input
            {...fieldProps(state, "city")}
            className="input"
            autoComplete="address-level2"
            defaultValue={address?.city}
            required
          />
          <FieldError state={state} name="city" />
        </label>
        <label className="field">
          <span className="field__label">
            Landmark <span className="text-muted">(optional)</span>
          </span>
          <input
            {...fieldProps(state, "landmark")}
            className="input"
            placeholder="e.g. Opposite the filling station"
            defaultValue={address?.landmark ?? undefined}
          />
          <span className="field__hint">Helps the rider find you.</span>
        </label>
      </div>
      <label className="check">
        <input
          type="checkbox"
          name="isDefault"
          defaultChecked={address?.isDefault}
        />
        Use as my default address
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending
          ? "Saving…"
          : (submitLabel ?? (address ? "Save address" : "Add address"))}
      </button>
    </form>
  );
}
