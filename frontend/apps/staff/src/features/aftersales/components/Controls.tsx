"use client";

import { useState } from "react";

import { ActionForm, Field } from "@techshop/ui/components";

import { ActionButton } from "@/components/ActionButton";

import {
  collectClaim,
  decideDispute,
  decideReturn,
  estimateTradeIn,
  inspectItem,
  inspectTradeIn,
  payTradeIn,
  receiveClaim,
  receiveReturn,
  receiveTradeIn,
  refundReturn,
  rejectClaim,
  resolveClaim,
  startRepair,
  updateRepair,
  walkInClaim,
} from "../actions";

type Option = { value: string; label: string };

/** A note box shared by a row of decisions; buttons that need the note stay off until it's written. */
function Noted({
  label,
  hint,
  buttons,
}: {
  label: string;
  hint?: string;
  buttons: {
    label: string;
    run: (note: string) => ReturnType<typeof decideReturn>;
    needsNote?: boolean;
    variant?: "secondary" | "inverse" | "ghost";
    confirm?: string;
  }[];
}) {
  const [note, setNote] = useState("");
  const noted = note.trim().length >= 3;
  const id = label.toLowerCase().replace(/\W+/g, "-");
  return (
    <div className="stack">
      <div className="field">
        <label htmlFor={id} className="field__label">
          {label}
        </label>
        <textarea
          id={id}
          className="input"
          rows={3}
          maxLength={1000}
          value={note}
          onChange={(e) => setNote(e.target.value)}
        />
        {hint ? <span className="field__hint">{hint}</span> : null}
      </div>
      <div className="cluster">
        {buttons.map((b) => (
          <ActionButton
            key={b.label}
            label={b.label}
            variant={b.variant ?? "secondary"}
            confirm={b.confirm}
            disabled={b.needsNote && !noted}
            action={() => b.run(note)}
          />
        ))}
      </div>
    </div>
  );
}

export function ReturnDecision({ rma }: { rma: string }) {
  return (
    <Noted
      label="Reason (required to reject)"
      hint="The customer sees it."
      buttons={[
        {
          label: "Approve",
          variant: "inverse",
          run: (n) => decideReturn(rma, "approve", n),
        },
        {
          label: "Reject",
          variant: "ghost",
          needsNote: true,
          confirm: "Reject this return?",
          run: (n) => decideReturn(rma, "reject", n),
        },
      ]}
    />
  );
}

export function ReceiveReturnButton({ rma }: { rma: string }) {
  return (
    <ActionButton
      label="Mark received"
      variant="inverse"
      action={() => receiveReturn(rma)}
    />
  );
}

export function RefundReturnButton({ rma }: { rma: string }) {
  return (
    <ActionButton
      label="Raise the refund"
      variant="inverse"
      action={() => refundReturn(rma)}
    />
  );
}

export function DisputeDecision({ rma }: { rma: string }) {
  return (
    <Noted
      label="Decision note"
      hint="The seller sees it. Upheld means the seller is credited and TechShop bears the cost."
      buttons={[
        {
          label: "Uphold (seller is right)",
          needsNote: true,
          confirm: "Credit the seller and write the cost off?",
          run: (n) => decideDispute(rma, "upheld", n),
        },
        {
          label: "Reject the dispute",
          variant: "ghost",
          needsNote: true,
          run: (n) => decideDispute(rma, "rejected", n),
        },
      ]}
    />
  );
}

const grades: Option[] = [
  { value: "A", label: "A: as new" },
  { value: "B", label: "B: light wear" },
  { value: "C", label: "C: heavy wear" },
  { value: "faulty", label: "Faulty" },
];

export function InspectItemForm({
  rma,
  itemId,
  serialised,
  sellerItem,
  locations,
}: {
  rma: string;
  itemId: string;
  serialised: boolean;
  sellerItem: boolean;
  locations: Option[];
}) {
  const dispositions: Option[] = sellerItem
    ? [{ value: "return_to_seller", label: "Back to the seller" }]
    : [
        { value: "restock_new", label: "Restock as new (sealed)" },
        { value: "restock_open_box", label: "Restock as open box" },
        { value: "repair", label: "Send for repair" },
        { value: "return_to_supplier", label: "Back to the supplier" },
        { value: "write_off", label: "Write off" },
      ];
  return (
    <ActionForm
      action={inspectItem.bind(null, rma, itemId)}
      submitLabel="Record inspection"
      pendingLabel="Saving…"
    >
      {serialised ? (
        <Field
          name="scannedSerial"
          label="IMEI or serial on the item"
          maxLength={40}
          autoComplete="off"
          hint="It must match the unit sold; a mismatch rejects the return and opens a risk case."
        />
      ) : null}
      <Field name="grade" label="Grade" type="select" options={grades} />
      <Field
        name="disposition"
        label="Where it goes"
        type="select"
        options={dispositions}
      />
      <Field
        name="warehouseId"
        label="Location"
        type="select"
        options={locations}
      />
      <Field
        name="notes"
        label="Notes"
        type="textarea"
        maxLength={500}
        optional
      />
    </ActionForm>
  );
}

// ---- Warranty

export function WalkInForm({ stores }: { stores: Option[] }) {
  return (
    <ActionForm
      action={walkInClaim}
      submitLabel="Take the claim"
      pendingLabel="Checking…"
      resetOnSuccess
    >
      <Field
        name="imeiOrSerial"
        label="IMEI or serial"
        maxLength={40}
        autoComplete="off"
        hint="We find the sale from it and check the stolen-device list."
      />
      <Field name="storeId" label="Store" type="select" options={stores} />
      <Field name="fault" label="Fault" type="textarea" maxLength={1000} />
    </ActionForm>
  );
}

export function ReceiveClaimForm({
  number,
  stores,
}: {
  number: string;
  stores: Option[];
}) {
  return (
    <ActionForm
      action={receiveClaim.bind(null, number)}
      submitLabel="Check the device in"
      pendingLabel="Checking…"
    >
      <Field name="storeId" label="Store" type="select" options={stores} />
      <Field
        name="scannedSerial"
        label="IMEI or serial of the device in hand"
        maxLength={40}
        autoComplete="off"
      />
    </ActionForm>
  );
}

export function ClaimDecisions({
  number,
  canRepair,
}: {
  number: string;
  canRepair: boolean;
}) {
  return (
    <Noted
      label="Diagnosis or reason"
      hint="Not covered: the customer sees the reason (physical or liquid damage, tampering)."
      buttons={[
        ...(canRepair
          ? [
              {
                label: "Send to the workshop",
                variant: "inverse" as const,
                run: (n: string) => startRepair(number, n),
              },
            ]
          : []),
        {
          label: "Not covered",
          variant: "ghost",
          needsNote: true,
          confirm: "Reject this claim?",
          run: (n) => rejectClaim(number, n),
        },
      ]}
    />
  );
}

const repairNext: Record<string, Option[]> = {
  queued: [
    { value: "diagnosing", label: "Diagnosing" },
    { value: "awaiting_parts", label: "Waiting for parts" },
    { value: "repairing", label: "Repairing" },
    { value: "done", label: "Done" },
    { value: "unrepairable", label: "Can't be repaired" },
  ],
  diagnosing: [
    { value: "awaiting_parts", label: "Waiting for parts" },
    { value: "repairing", label: "Repairing" },
    { value: "done", label: "Done" },
    { value: "unrepairable", label: "Can't be repaired" },
  ],
  awaiting_parts: [
    { value: "repairing", label: "Repairing" },
    { value: "unrepairable", label: "Can't be repaired" },
  ],
  repairing: [
    { value: "awaiting_parts", label: "Waiting for parts" },
    { value: "done", label: "Done" },
    { value: "unrepairable", label: "Can't be repaired" },
  ],
};

export function RepairUpdateForm({
  jobId,
  status,
}: {
  jobId: string;
  status: string;
}) {
  const options = repairNext[status];
  if (!options) return null;
  return (
    <ActionForm
      action={updateRepair.bind(null, jobId)}
      submitLabel="Update the job"
      pendingLabel="Saving…"
    >
      <Field name="status" label="Next step" type="select" options={options} />
      <Field
        name="diagnosis"
        label="Diagnosis"
        type="textarea"
        maxLength={1000}
        optional
        hint="Required before finishing, if not recorded yet."
      />
      <Field
        name="notes"
        label="Notes"
        type="textarea"
        maxLength={1000}
        optional
      />
    </ActionForm>
  );
}

export function ResolveClaimForm({
  number,
  warehouses,
  serialised,
}: {
  number: string;
  warehouses: Option[];
  serialised: boolean;
}) {
  return (
    <ActionForm
      action={resolveClaim.bind(null, number)}
      submitLabel="Settle the claim"
      pendingLabel="Saving…"
    >
      <Field
        name="resolution"
        label="Outcome"
        type="select"
        options={[
          { value: "replaced", label: "Replace from stock" },
          { value: "refunded", label: "Refund one unit (finance approves)" },
        ]}
      />
      {serialised ? (
        <Field
          name="replacementSerial"
          label="Replacement IMEI or serial"
          maxLength={40}
          optional
          hint="Replacing: scan the new unit (same model, in stock)."
        />
      ) : (
        <Field
          name="warehouseId"
          label="Replacement from"
          type="select"
          optional
          options={warehouses}
        />
      )}
      <Field name="note" label="Note" maxLength={500} optional />
    </ActionForm>
  );
}

export function CollectButton({ number }: { number: string }) {
  return (
    <ActionButton
      label="Customer collected it"
      variant="inverse"
      action={() => collectClaim(number)}
    />
  );
}

// ---- Trade-ins

export function EstimateForm({
  reference,
  models,
}: {
  reference: string;
  models: Option[];
}) {
  return (
    <ActionForm
      action={estimateTradeIn.bind(null, reference)}
      submitLabel="Send the estimate"
      pendingLabel="Sending…"
    >
      <Field name="estimate" label="Estimate (₦)" inputMode="decimal" />
      {models.length ? (
        <Field
          name="variantId"
          label="Catalogue model"
          type="select"
          optional
          options={models}
        />
      ) : null}
    </ActionForm>
  );
}

export function ReceiveTradeInForm({
  reference,
  stores,
  needsImei,
}: {
  reference: string;
  stores: Option[];
  needsImei: boolean;
}) {
  return (
    <ActionForm
      action={receiveTradeIn.bind(null, reference)}
      submitLabel="Take the device in"
      pendingLabel="Checking…"
    >
      <Field name="storeId" label="Store" type="select" options={stores} />
      {needsImei ? (
        <Field
          name="imei"
          label="IMEI (*#06#)"
          inputMode="numeric"
          maxLength={20}
          autoComplete="off"
          hint="Checked against the stolen-device list."
        />
      ) : (
        <Field
          name="serial"
          label="Serial number"
          maxLength={40}
          autoComplete="off"
        />
      )}
    </ActionForm>
  );
}

export function InspectTradeInForm({
  reference,
  models,
}: {
  reference: string;
  models: Option[];
}) {
  return (
    <ActionForm
      action={inspectTradeIn.bind(null, reference)}
      submitLabel="Send the offer"
      pendingLabel="Sending…"
    >
      <Field
        name="condition"
        label="Grade"
        type="select"
        options={[
          { value: "like_new", label: "Like new" },
          { value: "good", label: "Good" },
          { value: "fair", label: "Fair" },
          { value: "faulty", label: "Faulty" },
        ]}
      />
      {models.length ? (
        <Field
          name="variantId"
          label="Catalogue model"
          type="select"
          optional
          options={models}
          hint="Needed if the request had no model, so it can be stocked after refurbishment."
        />
      ) : null}
      <Field
        name="offer"
        label="Offer (₦)"
        inputMode="decimal"
        optional
        hint="Leave empty to price from today's catalogue and the grade."
      />
      <Field
        name="notes"
        label="What you checked"
        type="textarea"
        maxLength={1000}
        optional
        hint="Screen, battery health, iCloud or Google lock, accessories."
      />
    </ActionForm>
  );
}

export function PayTradeInControl({ reference }: { reference: string }) {
  const [ref, setRef] = useState("");
  return (
    <div className="cluster">
      <div className="field">
        <label htmlFor="ti-ref" className="field__label">
          Transfer reference
        </label>
        <input
          id="ti-ref"
          className="input"
          value={ref}
          maxLength={80}
          onChange={(e) => setRef(e.target.value)}
        />
      </div>
      <ActionButton
        label="Record as paid"
        variant="inverse"
        disabled={ref.trim().length < 3}
        action={() => payTradeIn(reference, ref)}
      />
    </div>
  );
}
