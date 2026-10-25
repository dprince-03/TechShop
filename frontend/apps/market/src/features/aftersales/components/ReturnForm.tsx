"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { requestReturn } from "../actions";

export type ReturnableLine = {
  id: string;
  name: string;
  variant: string;
  quantity: number;
};

const reasons = [
  { value: "faulty", label: "It's faulty" },
  { value: "damaged", label: "It arrived damaged" },
  { value: "not_as_described", label: "It's not as described" },
  { value: "wrong_item", label: "You sent the wrong item" },
  { value: "change_of_mind", label: "I changed my mind (unopened only)" },
];

/** Choose items and quantities, why, and how it comes back. */
export function ReturnForm({
  number,
  lines,
}: {
  number: string;
  lines: ReturnableLine[];
}) {
  return (
    <ActionForm
      action={requestReturn.bind(null, number)}
      submitLabel="Request the return"
      pendingLabel="Sending…"
      variant="primary"
    >
      <fieldset className="stack">
        <legend className="field__label">How many of each item?</legend>
        {lines.map((l) => (
          <Field
            key={l.id}
            name={`qty:${l.id}`}
            label={`${l.name} · ${l.variant}`}
            type="number"
            inputMode="numeric"
            min={0}
            max={l.quantity}
            defaultValue={lines.length === 1 ? l.quantity : 0}
            hint={`You bought ${l.quantity}.`}
          />
        ))}
      </fieldset>
      <Field
        name="reason"
        label="Why are you returning it?"
        type="select"
        options={reasons}
      />
      <Field
        name="details"
        label="What happened?"
        type="textarea"
        maxLength={1000}
        optional
        hint="A short description helps us check it faster."
      />
      <Field
        name="method"
        label="How will it come back?"
        type="select"
        options={[
          {
            value: "store_dropoff",
            label: "I'll bring it to a TechShop store",
          },
          { value: "pickup", label: "Collect it from my delivery address" },
        ]}
        hint="Change-of-mind returns are brought to a store. Collection is free in our delivery areas (Lagos and Abuja)."
      />
      <Field
        name="unopened"
        type="checkbox"
        label="It's unopened, with the seals intact"
        optional
        hint="Only needed for change-of-mind returns."
      />
    </ActionForm>
  );
}
