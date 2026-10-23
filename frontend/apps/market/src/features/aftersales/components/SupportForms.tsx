"use client";

import {
  ActionForm,
  ContactForm as SharedContactForm,
  Field,
} from "@techshop/ui/components";

import { contactUs, openTicket, replyTicket } from "../actions";

const topics = [
  { value: "order", label: "An order" },
  { value: "delivery", label: "Delivery" },
  { value: "payment", label: "A payment" },
  { value: "refund", label: "A refund" },
  { value: "return", label: "A return" },
  { value: "warranty", label: "Warranty or repair" },
  { value: "trade_in", label: "A trade-in" },
  { value: "account", label: "My account" },
  { value: "other", label: "Something else" },
];

/** A new ticket for a signed-in customer. */
export function NewTicketForm({
  orders,
  orderNumber,
}: {
  orders: string[];
  orderNumber?: string;
}) {
  return (
    <ActionForm
      action={openTicket}
      submitLabel="Send to support"
      pendingLabel="Sending…"
      variant="primary"
    >
      <Field
        name="topic"
        label="What's it about?"
        type="select"
        options={topics}
      />
      {orders.length ? (
        <Field
          name="orderNumber"
          label="Which order?"
          type="select"
          optional
          defaultValue={orderNumber}
          options={orders.map((o) => ({ value: o, label: o }))}
        />
      ) : null}
      <Field name="subject" label="Subject" maxLength={140} />
      <Field
        name="message"
        label="How can we help?"
        type="textarea"
        maxLength={5000}
        hint="Payment and refund questions are answered within 4 hours; everything else within 24."
      />
    </ActionForm>
  );
}

export function TicketReplyForm({ number }: { number: string }) {
  return (
    <ActionForm
      action={replyTicket.bind(null, number)}
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

/** The public contact form. */
export function ContactForm() {
  return <SharedContactForm action={contactUs} topics={topics} />;
}
