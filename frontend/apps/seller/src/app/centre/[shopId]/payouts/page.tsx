import type { Metadata } from "next";
import { connection } from "next/server";

import { Status, Table } from "@techshop/ui/components";

import { makeDefault, removeAccount } from "@/features/payouts/actions";
import { BankAccountForm } from "@/features/payouts/components/BankAccountForm";
import { RowAction } from "@/components/RowAction";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";

export const metadata: Metadata = {
  title: "Payout accounts | TechShop Seller Centre",
};

const match = {
  match: { label: "Name matches", tone: "success" },
  partial: { label: "Name partly matches", tone: "warning" },
  mismatch: { label: "Name doesn’t match", tone: "danger" },
  unknown: { label: "Not checked", tone: "neutral" },
} as const;

const date = new Intl.DateTimeFormat("en-NG", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "Africa/Lagos",
});

export default async function PayoutsPage(
  props: PageProps<"/centre/[shopId]/payouts">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}/payouts`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const [o, banks] = await Promise.all([
    api.get("/api/v1/seller/shops/{shopId}/onboarding", {
      path: { shopId },
      init: { cache: "no-store" },
    }),
    api.get("/api/v1/seller/banks").catch(() => null),
  ]);
  const accounts = o.bankAccounts ?? [];
  const bankName = new Map((banks ?? []).map((b) => [b.code, b.name]));
  const canManage = shop.role === "owner" || shop.role === "finance";
  return (
    <ShopPage
      shop={shop}
      section="payouts"
      title="Payout accounts"
      lead="Where we pay your earnings. A new account can receive payouts 24 hours after it’s added."
    >
      <Table
        caption="Payout accounts"
        rows={accounts}
        rowKey={(a) => a.id}
        empty={{
          title: "No payout account yet",
          body: "Add the account we should pay you into.",
        }}
        columns={[
          {
            label: "Account",
            cell: (a) =>
              `${bankName.get(a.bankCode) ?? a.bankCode} ••••${a.accountNumberLast4}`,
          },
          {
            label: "Name at the bank",
            cell: (a) => a.resolvedAccountName ?? "—",
          },
          {
            label: "Check",
            cell: (a) => (
              <Status
                label={match[a.nameMatch].label}
                tone={match[a.nameMatch].tone}
              />
            ),
          },
          {
            label: "Payouts from",
            cell: (a) => date.format(new Date(a.payoutsFrom)),
          },
          {
            label: "Default",
            cell: (a) =>
              a.isDefault ? (
                <Status label="Default" tone="info" />
              ) : canManage ? (
                <div className="cluster">
                  <RowAction
                    action={makeDefault.bind(null, shop.id, a.id)}
                    label="Make default"
                    pendingLabel="Saving…"
                  />
                  <RowAction
                    action={removeAccount.bind(null, shop.id, a.id)}
                    label="Remove"
                    pendingLabel="Removing…"
                  />
                </div>
              ) : null,
          },
        ]}
      />
      {canManage ? (
        <section aria-labelledby="add-title" className="card card--roomy stack">
          <h2 id="add-title" className="text-xl">
            Add a payout account
          </h2>
          {banks ? (
            <BankAccountForm
              shopId={shop.id}
              banks={banks.map((b) => ({ value: b.code, label: b.name }))}
            />
          ) : (
            <p className="alert alert--warning">
              The bank list isn’t available right now. Try again shortly.
            </p>
          )}
        </section>
      ) : (
        <p className="text-sm text-muted">
          Only the shop owner or a finance member can change payout accounts.
        </p>
      )}
    </ShopPage>
  );
}
