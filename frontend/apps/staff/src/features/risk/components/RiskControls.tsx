"use client";

import { useState } from "react";

import { ActionForm, Field } from "@techshop/ui/components";

import { ActionButton } from "@/components/ActionButton";

import {
  blockDevice,
  claimCase,
  discardRules,
  draftRules,
  escalateCase,
  noteCase,
  publishRules,
  resolveCase,
  setThresholds,
  testRules,
  unblockDevice,
  updateRule,
} from "../actions";

/** Everything an analyst does on a case; the note goes into the case timeline and audit log. */
export function CaseActions({
  reference,
  status,
  canResolve,
}: {
  reference: string;
  status: string;
  canResolve: boolean;
}) {
  const [note, setNote] = useState("");
  const noted = note.trim().length >= 3;
  const open =
    status === "open" || status === "in_review" || status === "escalated";
  if (!open || !canResolve) return null;
  return (
    <div className="stack">
      <div className="field">
        <label htmlFor="case-note" className="field__label">
          Note
        </label>
        <textarea
          id="case-note"
          className="input"
          rows={3}
          maxLength={1000}
          value={note}
          onChange={(e) => setNote(e.target.value)}
        />
        <span className="field__hint">
          What you checked and why. Required to note, escalate, clear or block.
        </span>
      </div>
      <div className="cluster">
        {status !== "in_review" ? (
          <ActionButton
            label="Take this case"
            action={() => claimCase(reference)}
          />
        ) : null}
        <ActionButton
          label="Add note"
          disabled={!noted}
          action={() => noteCase(reference, note)}
        />
        {status !== "escalated" ? (
          <ActionButton
            label="Escalate"
            disabled={!noted}
            action={() => escalateCase(reference, note)}
          />
        ) : null}
        <ActionButton
          label="Clear (legitimate)"
          variant="inverse"
          disabled={!noted}
          action={() => resolveCase(reference, "clear", "legit", note)}
        />
        <ActionButton
          label="Clear (unclear)"
          disabled={!noted}
          action={() => resolveCase(reference, "clear", "unclear", note)}
        />
        <ActionButton
          label="Block (fraud)"
          variant="ghost"
          disabled={!noted}
          confirm="Block? Held orders are cancelled and refunded, and the label 'fraud' feeds rule tuning."
          action={() => resolveCase(reference, "block", "", note)}
        />
      </div>
    </div>
  );
}

export function BlockDeviceForm() {
  return (
    <ActionForm
      action={blockDevice}
      submitLabel="Block this IMEI"
      pendingLabel="Blocking…"
      resetOnSuccess
    >
      <Field
        name="imei"
        label="IMEI"
        inputMode="numeric"
        maxLength={20}
        autoComplete="off"
      />
      <Field
        name="reason"
        label="Reason"
        type="select"
        options={[
          { value: "reported_stolen", label: "Reported stolen" },
          { value: "fraud", label: "Fraud" },
          { value: "counterfeit", label: "Counterfeit" },
        ]}
      />
      <Field
        name="source"
        label="Reported by"
        maxLength={80}
        hint="For example: customer with police extract, case RK-301."
      />
      <Field
        name="reference"
        label="Reference"
        maxLength={80}
        optional
        hint="Police extract number, case or ticket."
      />
    </ActionForm>
  );
}

export function UnblockControl({ imei }: { imei: string }) {
  const [note, setNote] = useState("");
  return (
    <details>
      <summary className="link text-sm">Remove</summary>
      <div className="cluster">
        <label className="visually-hidden" htmlFor={`un-${imei}`}>
          Why remove {imei}
        </label>
        <input
          id={`un-${imei}`}
          className="input"
          placeholder="Why (e.g. phone recovered)"
          value={note}
          maxLength={200}
          onChange={(e) => setNote(e.target.value)}
        />
        <ActionButton
          label="Remove from blocklist"
          variant="ghost"
          disabled={note.trim().length < 3}
          confirm={`Unblock ${imei}?`}
          action={() => unblockDevice(imei, note)}
        />
      </div>
    </details>
  );
}

export function StartDraft({ checkpoint }: { checkpoint: string }) {
  const [notes, setNotes] = useState("");
  return (
    <div className="cluster">
      <label className="visually-hidden" htmlFor={`draft-${checkpoint}`}>
        Why change the {checkpoint} rules
      </label>
      <input
        id={`draft-${checkpoint}`}
        className="input"
        placeholder="Why the change"
        value={notes}
        maxLength={500}
        onChange={(e) => setNotes(e.target.value)}
      />
      <ActionButton
        label="Start a draft"
        action={() => draftRules(checkpoint, notes)}
      />
    </div>
  );
}

export function RuleForm({
  setId,
  ruleId,
  weight,
  enabled,
  label,
}: {
  setId: string;
  ruleId: string;
  weight: number;
  enabled: boolean;
  label: string;
}) {
  return (
    <ActionForm
      action={updateRule.bind(null, setId, ruleId)}
      submitLabel="Save"
      pendingLabel="Saving…"
    >
      <Field
        name="weight"
        label={`Weight: ${label}`}
        type="number"
        min={-100}
        max={100}
        defaultValue={weight}
      />
      <Field
        name="enabled"
        type="checkbox"
        label="On"
        optional
        defaultValue={enabled ? "on" : ""}
      />
    </ActionForm>
  );
}

export function ThresholdsForm({
  setId,
  step,
  hold,
  block,
}: {
  setId: string;
  step: number;
  hold: number;
  block: number;
}) {
  return (
    <ActionForm
      action={setThresholds.bind(null, setId)}
      submitLabel="Save bands"
      pendingLabel="Saving…"
    >
      <Field
        name="step"
        label="Step up from"
        type="number"
        min={1}
        max={99}
        defaultValue={step}
      />
      <Field
        name="hold"
        label="Hold from"
        type="number"
        min={2}
        max={99}
        defaultValue={hold}
      />
      <Field
        name="block"
        label="Block from"
        type="number"
        min={3}
        max={100}
        defaultValue={block}
      />
    </ActionForm>
  );
}

export function DraftActions({ setId }: { setId: string }) {
  return (
    <div className="cluster">
      <ActionButton
        label="Test against the last 30 days"
        action={() => testRules(setId)}
      />
      <ActionButton
        label="Publish"
        variant="inverse"
        confirm="Publish these rules? New decisions use them at once."
        action={() => publishRules(setId)}
      />
      <ActionButton
        label="Discard draft"
        variant="ghost"
        confirm="Discard this draft?"
        action={() => discardRules(setId)}
      />
    </div>
  );
}
