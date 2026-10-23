"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import {
  acceptTradeIn,
  cancelTradeIn,
  declineTradeIn,
  startTradeIn,
} from "../actions";

const conditions = [
  { value: "like_new", label: "Like new: no marks" },
  { value: "good", label: "Good: light scratches" },
  { value: "fair", label: "Fair: visible wear" },
  { value: "faulty", label: "Faulty: screen, battery or other issue" },
];

/** Step 2 of a trade-in: the model (or brand and model typed in), condition and IMEI. */
export function TradeInForm({
  deviceType,
  models,
}: {
  deviceType: string;
  models: { value: string; label: string }[];
}) {
  const typed = models.length === 0;
  return (
    <ActionForm
      action={startTradeIn}
      submitLabel="Get my estimate"
      pendingLabel="Pricing…"
      variant="primary"
    >
      <input type="hidden" name="deviceType" value={deviceType} />
      {!typed ? (
        <Field
          name="variantId"
          label="Model"
          type="select"
          optional
          options={models}
          hint="Not listed? Leave this empty and type the brand and model below; we'll price it by hand."
        />
      ) : null}
      <Field name="brand" label="Brand" maxLength={60} optional={!typed} />
      <Field name="model" label="Model" maxLength={80} optional={!typed} />
      <Field
        name="storage"
        label="Storage"
        maxLength={20}
        optional
        hint="For example 128GB."
      />
      <Field
        name="condition"
        label="Condition"
        type="select"
        options={conditions}
      />
      {deviceType === "phone" || deviceType === "tablet" ? (
        <Field
          name="imei"
          label="IMEI"
          inputMode="numeric"
          maxLength={20}
          optional
          hint="Dial *#06# to see it. We check it isn't reported stolen."
        />
      ) : null}
      <Field
        name="owner"
        type="checkbox"
        label="I own this device and can show proof of purchase or ownership"
      />
    </ActionForm>
  );
}

/** Take the offer: store credit at once, or a bank transfer. */
export function AcceptOfferForm({
  reference,
  banks,
}: {
  reference: string;
  banks: { value: string; label: string }[];
}) {
  return (
    <ActionForm
      action={acceptTradeIn.bind(null, reference)}
      submitLabel="Accept the offer"
      pendingLabel="Accepting…"
      variant="primary"
    >
      <Field
        name="payoutMethod"
        label="How do you want to be paid?"
        type="select"
        options={[
          { value: "store_credit", label: "Store credit, straight away" },
          { value: "bank", label: "Bank transfer, within 2 working days" },
        ]}
      />
      <fieldset className="stack">
        <legend className="field__label">For a bank transfer</legend>
        <Field
          name="bankCode"
          label="Bank"
          type="select"
          optional
          options={banks}
        />
        <Field
          name="accountNumber"
          label="Account number"
          inputMode="numeric"
          maxLength={10}
          optional
          autoComplete="off"
        />
        <Field
          name="accountName"
          label="Account name"
          maxLength={120}
          optional
          hint="As your bank has it. We check it with the bank."
        />
      </fieldset>
    </ActionForm>
  );
}

export function DeclineOfferForm({ reference }: { reference: string }) {
  return (
    <ActionForm
      action={declineTradeIn.bind(null, reference)}
      submitLabel="Decline and collect my device"
      pendingLabel="Declining…"
      variant="ghost"
    >
      <></>
    </ActionForm>
  );
}

export function CancelTradeInForm({ reference }: { reference: string }) {
  return (
    <ActionForm
      action={cancelTradeIn.bind(null, reference)}
      submitLabel="Cancel this trade-in"
      pendingLabel="Cancelling…"
      variant="ghost"
    >
      <></>
    </ActionForm>
  );
}
