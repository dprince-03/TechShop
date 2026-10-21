"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { createListing, resubmitListing, setPrice, setStock } from "../actions";
import { conditions, grades, warranties } from "../options";

const naira = (kobo?: number | null) => (kobo ? String(kobo / 100) : undefined);

/** A new offer on one of a product's variants. */
export function NewListingForm({
  shopId,
  variants,
}: {
  shopId: string;
  variants: { value: string; label: string }[];
}) {
  return (
    <ActionForm
      action={createListing.bind(null, shopId)}
      submitLabel="Create listing"
      pendingLabel="Checking…"
      variant="primary"
    >
      <Field name="variantId" label="Model" type="select" options={variants} />
      <Field
        name="condition"
        label="Condition"
        type="select"
        options={conditions}
      />
      <Field
        name="grade"
        label="Grade"
        type="select"
        options={grades}
        optional
        hint="For used, refurbished and open-box items"
      />
      <Field
        name="conditionNotes"
        label="Condition notes"
        type="textarea"
        maxLength={500}
        optional
        hint="Required unless new: scratches, battery, what’s in the box"
      />
      <Field
        name="batteryHealthPct"
        label="Battery health (%)"
        type="number"
        inputMode="numeric"
        min={1}
        max={100}
        optional
      />
      <Field
        name="price"
        label="Price (₦)"
        inputMode="decimal"
        hint="Customers pay this. No “was” price on a new listing: it needs 30 days of price history."
      />
      <Field
        name="sellerStock"
        label="In stock"
        type="number"
        inputMode="numeric"
        min={0}
      />
      <Field
        name="handlingDays"
        label="Days to pack after you accept"
        type="number"
        inputMode="numeric"
        min={0}
        max={10}
        defaultValue={1}
      />
      <Field
        name="warrantyProvider"
        label="Warranty from"
        type="select"
        options={warranties}
      />
      <Field
        name="warrantyMonths"
        label="Warranty (months)"
        type="number"
        inputMode="numeric"
        min={0}
        max={60}
        defaultValue={0}
      />
    </ActionForm>
  );
}

export function PriceForm({
  shopId,
  listingId,
  price,
  compareAt,
}: {
  shopId: string;
  listingId: string;
  price: number;
  compareAt?: number | null;
}) {
  return (
    <ActionForm
      action={setPrice.bind(null, shopId, listingId)}
      submitLabel="Save price"
    >
      <Field
        name="price"
        label="Price (₦)"
        inputMode="decimal"
        defaultValue={naira(price)}
      />
      <Field
        name="compareAt"
        label="Was (₦)"
        inputMode="decimal"
        defaultValue={naira(compareAt)}
        optional
        hint="At most your highest price in the last 30 days"
      />
    </ActionForm>
  );
}

export function StockForm({
  shopId,
  listingId,
  stock,
}: {
  shopId: string;
  listingId: string;
  stock?: number | null;
}) {
  return (
    <ActionForm
      action={setStock.bind(null, shopId, listingId)}
      submitLabel="Save stock"
    >
      <Field
        name="stock"
        label="In stock"
        type="number"
        inputMode="numeric"
        min={0}
        defaultValue={stock ?? 0}
      />
    </ActionForm>
  );
}

type Editable = {
  grade?: string | null;
  conditionNotes?: string | null;
  batteryHealthPct?: number | null;
  warrantyProvider: string;
  warrantyMonths: number;
  handlingDays: number;
};

export function ResubmitForm({
  shopId,
  listingId,
  listing,
}: {
  shopId: string;
  listingId: string;
  listing: Editable;
}) {
  return (
    <ActionForm
      action={resubmitListing.bind(null, shopId, listingId)}
      submitLabel="Resubmit for review"
      variant="primary"
    >
      <Field
        name="grade"
        label="Grade"
        type="select"
        options={grades}
        optional
        defaultValue={listing.grade ?? ""}
      />
      <Field
        name="conditionNotes"
        label="Condition notes"
        type="textarea"
        maxLength={500}
        optional
        defaultValue={listing.conditionNotes ?? ""}
      />
      <Field
        name="batteryHealthPct"
        label="Battery health (%)"
        type="number"
        min={1}
        max={100}
        optional
        defaultValue={listing.batteryHealthPct ?? undefined}
      />
      <Field
        name="warrantyProvider"
        label="Warranty from"
        type="select"
        options={warranties}
        defaultValue={listing.warrantyProvider}
      />
      <Field
        name="warrantyMonths"
        label="Warranty (months)"
        type="number"
        min={0}
        max={60}
        defaultValue={listing.warrantyMonths}
      />
      <Field
        name="handlingDays"
        label="Days to pack after you accept"
        type="number"
        min={0}
        max={10}
        defaultValue={listing.handlingDays}
      />
    </ActionForm>
  );
}
