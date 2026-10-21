"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { rateSeller, reportReview, reviewItem } from "../actions";

const starOptions = [5, 4, 3, 2, 1].map((n) => ({
  value: String(n),
  label: `${n} star${n === 1 ? "" : "s"}`,
}));

export function ReviewItemForm({ lineId }: { lineId: string }) {
  return (
    <ActionForm
      action={reviewItem.bind(null, lineId)}
      submitLabel="Post review"
      pendingLabel="Posting…"
      variant="primary"
    >
      <Field
        name="rating"
        label="Your rating"
        type="select"
        options={starOptions}
      />
      <Field name="title" label="Headline" maxLength={120} optional />
      <Field
        name="body"
        label="Your review"
        type="textarea"
        maxLength={3000}
        optional
        hint="What did you like or not? Please don’t include phone numbers, emails or links."
      />
    </ActionForm>
  );
}

export function RateSellerForm({
  fulfilmentId,
  seller,
}: {
  fulfilmentId: string;
  seller: string;
}) {
  return (
    <ActionForm
      action={rateSeller.bind(null, fulfilmentId)}
      submitLabel={`Rate ${seller}`}
      pendingLabel="Saving…"
    >
      <Field
        name="rating"
        label={`How was ${seller}?`}
        type="select"
        options={starOptions}
      />
      <Field name="comment" label="Comment" maxLength={500} optional />
    </ActionForm>
  );
}

export function ReportReviewForm({ reviewId }: { reviewId: string }) {
  return (
    <details>
      <summary className="link link--subtle text-xs">
        Report this review
      </summary>
      <ActionForm
        action={reportReview.bind(null, reviewId)}
        submitLabel="Send report"
        pendingLabel="Sending…"
        variant="ghost"
      >
        <Field
          name="reason"
          label="What’s wrong with it?"
          type="select"
          options={[
            { value: "fake", label: "It looks fake" },
            { value: "abuse", label: "Abusive or offensive" },
            { value: "personal_data", label: "Contains personal details" },
            { value: "off_topic", label: "Not about the product" },
            { value: "other", label: "Something else" },
          ]}
        />
      </ActionForm>
    </details>
  );
}
