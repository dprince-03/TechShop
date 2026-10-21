"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { addBankAccount } from "../actions";

/** A new payout account (bank list from the payment provider). */
export function BankAccountForm({
  shopId,
  banks,
}: {
  shopId: string;
  banks: { value: string; label: string }[];
}) {
  return (
    <ActionForm
      action={addBankAccount.bind(null, shopId)}
      submitLabel="Add account"
      pendingLabel="Checking with your bank…"
      resetOnSuccess
    >
      <Field name="bankCode" label="Bank" type="select" options={banks} />
      <Field
        name="accountNumber"
        label="Account number"
        inputMode="numeric"
        maxLength={10}
        autoComplete="off"
        hint="10 digits (NUBAN)"
      />
      <Field
        name="accountName"
        label="Account name"
        maxLength={100}
        hint="It must match your verified name, or your registered business name."
      />
    </ActionForm>
  );
}
