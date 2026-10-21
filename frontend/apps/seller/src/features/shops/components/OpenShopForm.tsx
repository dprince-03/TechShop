"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { openShop } from "../actions";

/** Name, seller type and location of a new shop. */
export function OpenShopForm({
  states,
}: {
  states: { value: string; label: string }[];
}) {
  return (
    <ActionForm
      action={openShop}
      submitLabel="Open shop"
      pendingLabel="Opening…"
      variant="primary"
    >
      <Field
        name="type"
        label="You’re selling as"
        type="select"
        options={[
          { value: "individual", label: "An individual" },
          { value: "business", label: "A registered business (CAC)" },
        ]}
      />
      <Field
        name="displayName"
        label="Shop name"
        maxLength={60}
        hint="Shown to customers. Businesses: use your registered name."
      />
      <Field name="stateCode" label="State" type="select" options={states} />
      <Field
        name="city"
        label="City or town"
        maxLength={60}
        autoComplete="address-level2"
      />
    </ActionForm>
  );
}
