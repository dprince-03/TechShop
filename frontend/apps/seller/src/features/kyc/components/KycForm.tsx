"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { submitKyc } from "../actions";

const idTypes = [
  { value: "nin", label: "National Identification Number (NIN)" },
  { value: "nin_slip", label: "NIN slip" },
  { value: "passport", label: "International passport" },
  { value: "drivers_licence", label: "Driver’s licence" },
  { value: "voters_card", label: "Voter’s card" },
];
const docs = "image/jpeg,image/png,application/pdf";

/** Identity (and, for businesses, CAC) verification with document uploads. */
export function KycForm({
  shopId,
  business,
}: {
  shopId: string;
  business: boolean;
}) {
  return (
    <ActionForm
      action={submitKyc.bind(null, shopId)}
      submitLabel="Submit for review"
      pendingLabel="Uploading…"
      variant="primary"
    >
      <input
        type="hidden"
        name="sellerType"
        value={business ? "business" : "individual"}
      />
      <Field
        name="idType"
        label="ID document"
        type="select"
        options={idTypes}
      />
      <Field
        name="idNumber"
        label="ID number"
        inputMode="text"
        autoComplete="off"
        maxLength={30}
        hint="Encrypted. Only the last 4 digits are ever shown."
      />
      {business ? (
        <>
          <Field
            name="cacRcNumber"
            label="CAC registration number"
            maxLength={20}
            hint="For example RC 1234567"
          />
          <Field
            name="tin"
            label="Tax identification number (TIN)"
            maxLength={20}
            optional
          />
        </>
      ) : null}
      <Field
        name="id_front"
        label="Photo of the front of your ID"
        type="file"
        accept={docs}
        hint="JPG, PNG or PDF, up to 10 MB"
      />
      <Field
        name="id_back"
        label="Photo of the back of your ID"
        type="file"
        accept={docs}
        optional
      />
      <Field
        name="selfie"
        label="A photo of you holding the ID"
        type="file"
        accept="image/jpeg,image/png"
        hint="Your face and the ID should both be clear"
      />
      {business ? (
        <Field
          name="cac_certificate"
          label="CAC certificate"
          type="file"
          accept={docs}
        />
      ) : null}
    </ActionForm>
  );
}
