"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { ActionButton } from "@/components/ActionButton";

import {
  assignTicket,
  claimTicket,
  replyToTicket,
  setTicketStatus,
} from "../actions";

export function ReplyForm({ number }: { number: string }) {
  return (
    <ActionForm
      action={replyToTicket.bind(null, number)}
      submitLabel="Send"
      pendingLabel="Sending…"
      variant="inverse"
      resetOnSuccess
    >
      <Field name="body" label="Message" type="textarea" maxLength={5000} />
      <Field
        name="internal"
        type="checkbox"
        label="Internal note (staff only)"
        optional
      />
      <Field
        name="status"
        label="After a reply"
        type="select"
        options={[
          { value: "pending_customer", label: "Wait for the customer" },
          { value: "open", label: "Keep it open" },
          { value: "resolved", label: "Mark resolved" },
        ]}
      />
    </ActionForm>
  );
}

export function ClaimTicketButton({ number }: { number: string }) {
  return <ActionButton label="Take it" action={() => claimTicket(number)} />;
}

export function AssignForm({
  number,
  agents,
  current,
}: {
  number: string;
  agents: { value: string; label: string }[];
  current?: string;
}) {
  return (
    <ActionForm
      action={assignTicket.bind(null, number)}
      submitLabel="Assign"
      pendingLabel="Saving…"
    >
      <Field
        name="agentId"
        label="Assign to"
        type="select"
        optional
        options={agents}
        defaultValue={current}
      />
    </ActionForm>
  );
}

export function StatusForm({
  number,
  status,
  priority,
}: {
  number: string;
  status: string;
  priority: string;
}) {
  return (
    <ActionForm
      action={setTicketStatus.bind(null, number)}
      submitLabel="Update"
      pendingLabel="Saving…"
    >
      <Field
        name="status"
        label="Status"
        type="select"
        defaultValue={status}
        options={[
          { value: "open", label: "Open" },
          { value: "pending_customer", label: "Waiting on the customer" },
          { value: "resolved", label: "Resolved" },
          { value: "closed", label: "Closed" },
        ]}
      />
      <Field
        name="priority"
        label="Priority"
        type="select"
        defaultValue={priority}
        options={[
          { value: "low", label: "Low" },
          { value: "normal", label: "Normal" },
          { value: "high", label: "High" },
          { value: "urgent", label: "Urgent" },
        ]}
      />
    </ActionForm>
  );
}
