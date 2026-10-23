"use client";

import { ActionForm, Field, type FormResult } from "./ActionForm";

export type ContactTopic = { value: string; label: string };

/**
 * The public contact form (every site): name, a way to reply, topic, subject and message. The
 * hidden "website" field is a trap only bots fill in. `action` is the app's Server Action.
 */
export function ContactForm({
  action,
  topics,
  askOrder = true,
}: {
  action: (prev: FormResult, form: FormData) => Promise<FormResult>;
  topics: ContactTopic[];
  askOrder?: boolean;
}) {
  return (
    <ActionForm
      action={action}
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
      {askOrder ? (
        <Field
          name="orderNumber"
          label="Order number"
          optional
          hint="For example TS-10482."
          maxLength={20}
        />
      ) : null}
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
