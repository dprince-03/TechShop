"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { respond } from "../actions";

export function RespondForm({
  shopId,
  reviewId,
}: {
  shopId: string;
  reviewId: string;
}) {
  return (
    <ActionForm
      action={respond.bind(null, shopId, reviewId)}
      submitLabel="Post answer"
      pendingLabel="Posting…"
    >
      <Field
        name="text"
        label="Your answer"
        type="textarea"
        maxLength={1000}
        hint="Shown under the review. You can answer once."
      />
    </ActionForm>
  );
}
