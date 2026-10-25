"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { disputeReturn, openShopTicket, replyShopTicket } from "../actions";

export function DisputeForm({ shopId, rma }: { shopId: string; rma: string }) {
  return (
    <ActionForm
      action={disputeReturn.bind(null, shopId, rma)}
      submitLabel="Dispute the inspection"
      pendingLabel="Sending…"
    >
      <Field
        name="reason"
        label="What's wrong with our inspection?"
        type="textarea"
        maxLength={1000}
        hint="For example: it left you sealed, or the damage isn't in your packing photos. A support lead decides; if they agree, your money is given back."
      />
    </ActionForm>
  );
}

const topics = [
  { value: "selling", label: "Selling and listings" },
  { value: "order", label: "An order" },
  { value: "delivery", label: "Drop-off or delivery" },
  { value: "payment", label: "Payouts" },
  { value: "return", label: "A return" },
  { value: "account", label: "The shop account" },
  { value: "other", label: "Something else" },
];

export function NewShopTicketForm({ shopId }: { shopId: string }) {
  return (
    <ActionForm
      action={openShopTicket.bind(null, shopId)}
      submitLabel="Send to seller support"
      pendingLabel="Sending…"
      variant="primary"
    >
      <Field
        name="topic"
        label="What's it about?"
        type="select"
        options={topics}
      />
      <Field
        name="orderNumber"
        label="Order number"
        optional
        maxLength={20}
        hint="If it's about an order, e.g. TS-10482."
      />
      <Field name="subject" label="Subject" maxLength={140} />
      <Field
        name="message"
        label="How can we help?"
        type="textarea"
        maxLength={5000}
        hint="Payout questions are answered within 4 hours; everything else within 24."
      />
    </ActionForm>
  );
}

export function ShopTicketReply({
  shopId,
  number,
}: {
  shopId: string;
  number: string;
}) {
  return (
    <ActionForm
      action={replyShopTicket.bind(null, shopId, number)}
      submitLabel="Send reply"
      pendingLabel="Sending…"
      resetOnSuccess
    >
      <Field
        name="message"
        label="Your reply"
        type="textarea"
        maxLength={5000}
      />
    </ActionForm>
  );
}
