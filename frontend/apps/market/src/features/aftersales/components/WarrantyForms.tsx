"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { makeClaim } from "../actions";

export type CoveredOption = {
  value: string;
  label: string;
  serialised: boolean;
};

/** A warranty claim: the item, its IMEI or serial, and the fault. */
export function ClaimForm({
  items,
  selected,
}: {
  items: CoveredOption[];
  selected?: string;
}) {
  return (
    <ActionForm
      action={makeClaim}
      submitLabel="Make the claim"
      pendingLabel="Sending…"
      variant="primary"
    >
      <Field
        name="orderLineId"
        label="Which item?"
        type="select"
        options={items}
        defaultValue={selected}
      />
      <Field
        name="imeiOrSerial"
        label="IMEI or serial number"
        maxLength={40}
        autoComplete="off"
        hint="Phones: dial *#06# to see the IMEI. Other devices: the serial number on the label or box."
      />
      <Field
        name="fault"
        label="What's wrong?"
        type="textarea"
        maxLength={1000}
        hint="What happens, when it started, and anything you've tried. Physical or liquid damage isn't covered."
      />
    </ActionForm>
  );
}
