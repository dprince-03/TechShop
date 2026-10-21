import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { Icon, Status, humanise } from "@techshop/ui/components";

import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";

export const metadata: Metadata = {
  title: "Overview | TechShop Seller Centre",
};

/** The next step, in words, and where to do it. */
const nextSteps = {
  submit_kyc: {
    text: "Verify your identity so we can approve your shop.",
    href: "verification",
    cta: "Verify your identity",
  },
  fix_kyc: {
    text: "We need you to fix your verification. See the reviewer’s note.",
    href: "verification",
    cta: "Fix verification",
  },
  add_bank_account: {
    text: "Add the bank account we should pay you into.",
    href: "payouts",
    cta: "Add a payout account",
  },
  await_review: {
    text: "We’re reviewing your verification. We’ll email you when it’s done.",
    href: "verification",
    cta: "See your verification",
  },
  done: {
    text: "Your shop is verified. You can list products.",
    href: "listings",
    cta: "Go to listings",
  },
} as const;

export default async function ShopOverview(
  props: PageProps<"/centre/[shopId]">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const o = await api.get("/api/v1/seller/shops/{shopId}/onboarding", {
    path: { shopId },
    init: { cache: "no-store" },
  });
  const next = nextSteps[o.nextStep];
  const steps = [
    { label: "Shop opened", done: true },
    {
      label: "Identity verification submitted",
      done:
        Boolean(o.kyc) &&
        o.kyc?.status !== "rejected" &&
        o.kyc?.status !== "needs_more_info",
    },
    { label: "Payout account added", done: (o.bankAccounts?.length ?? 0) > 0 },
    { label: "Verification approved", done: o.kyc?.status === "approved" },
  ];
  return (
    <ShopPage shop={shop} section="overview" title="Overview">
      <section aria-labelledby="next-title" className="card card--roomy stack">
        <h2 id="next-title" className="text-xl">
          {o.nextStep === "done" ? "You’re ready to sell" : "Next step"}
        </h2>
        <p>{next.text}</p>
        <div>
          <Link href={`${here}/${next.href}`} className="btn btn--primary">
            {next.cta}
          </Link>
        </div>
      </section>
      <section aria-labelledby="setup-title" className="stack">
        <h2 id="setup-title" className="text-xl">
          Shop setup
        </h2>
        <ol role="list" className="stack stack--sm">
          {steps.map((s) => (
            <li key={s.label} className="cluster">
              <Icon name={s.done ? "check" : "file"} size={20} />
              <span>{s.label}</span>
              <span className="visually-hidden">
                {s.done ? "(done)" : "(to do)"}
              </span>
            </li>
          ))}
        </ol>
        {o.kyc ? (
          <p className="text-sm text-muted">
            Verification:{" "}
            <Status
              label={humanise(o.kyc.status)}
              tone={
                o.kyc.status === "approved"
                  ? "success"
                  : o.kyc.status === "rejected"
                    ? "danger"
                    : "warning"
              }
            />
          </p>
        ) : null}
      </section>
    </ShopPage>
  );
}
