"use client";

import { ActionForm, Field } from "@techshop/ui/components";

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

/** The public contact form. The hidden "website" field is a trap for bots. */
export function ContactForm() {
  return (
    <ActionForm
      action={contactUs}
      submitLabel="Send message"
      pendingLabel="Sending…"
      variant="primary"
      resetOnSuccess
    >
      <Field name="name" label="Full name" autoComplete="name" maxLength={80} />
      <Field
        name="email"
        label="Email address"
        type="email"
        autoComplete="email"
        maxLength={254}
        optional
        hint="Give us an email or a phone number."
      />
      <Field
        name="phone"
        label="Phone number"
        type="tel"
        autoComplete="tel"
        inputMode="tel"
        optional
      />
      <Field
        name="orderNumber"
        label="Order number"
        optional
        hint="For example TS-10482."
        maxLength={20}
      />
      <Field name="topic" label="Topic" type="select" options={topics} />
      <Field name="subject" label="Subject" maxLength={140} />
      <Field
        name="message"
        label="How can we help?"
        type="textarea"
        maxLength={5000}
      />
      <div className="visually-hidden" aria-hidden="true">
        <label>
          Leave this empty
          <input name="website" tabIndex={-1} autoComplete="off" />
        </label>
      </div>
    </ActionForm>
  );
}
