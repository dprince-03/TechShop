"use server";

import { type ActionState } from "@techshop/bff";

import { act, text } from "@/lib/actions";

/** Adds a payout account; the API checks the name with the bank first. */
export async function addBankAccount(
  shopId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const accountNumber = text(form, "accountNumber", 20).replace(/\s/g, "");
  const errors: Record<string, string> = {};
  if (!/^\d{10}$/.test(accountNumber))
    errors.accountNumber = "Enter the 10-digit account number.";
  if (!text(form, "bankCode")) errors.bankCode = "Choose your bank.";
  if (text(form, "accountName").length < 3)
    errors.accountName = "Enter the name on the account.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/bank-accounts", {
        path: { shopId },
        body: {
          bankCode: text(form, "bankCode", 6),
          accountNumber,
          accountName: text(form, "accountName", 100),
        },
      }),
    "Account added. Payouts can use it after 24 hours.",
  );
}

export async function makeDefault(
  shopId: string,
  accountId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.post(
        "/api/v1/seller/shops/{shopId}/bank-accounts/{accountId}/default",
        { path: { shopId, accountId } },
      ),
    "Payouts will go to this account.",
  );
}

export async function removeAccount(
  shopId: string,
  accountId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.delete("/api/v1/seller/shops/{shopId}/bank-accounts/{accountId}", {
        path: { shopId, accountId },
      }),
    "Account removed.",
  );
}
