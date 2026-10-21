"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { cancelOrder } from "../actions";

export function CancelOrderForm({
  shopId,
  orderId,
}: {
  shopId: string;
  orderId: string;
}) {
  return (
    <ActionForm
      action={cancelOrder.bind(null, shopId, orderId)}
      submitLabel="Cancel this order"
      pendingLabel="Cancelling…"
      variant="ghost"
    >
      <Field
        name="reason"
        label="Reason"
        type="select"
        options={[
          { value: "out_of_stock", label: "Out of stock" },
          { value: "damaged", label: "Item damaged" },
          { value: "price_error", label: "Price was wrong" },
          { value: "other", label: "Other" },
        ]}
      />
      <Field name="note" label="Details" maxLength={300} optional />
    </ActionForm>
  );
}
