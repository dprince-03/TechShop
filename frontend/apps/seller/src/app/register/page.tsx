import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { signInPath } from "@techshop/bff";
import { PageHeader } from "@techshop/ui/components";

import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Open a shop | TechShop Seller Centre",
};

const steps = [
  {
    title: "Sign in with your phone",
    body: "We text you a code. New here? The same code creates your account.",
  },
  {
    title: "Open your shop",
    body: "Your shop name, whether you sell as an individual or a registered business, and where you’re based.",
  },
  {
    title: "Verify your identity",
    body: "Your ID number and a photo of the ID and of yourself. Registered businesses also add their CAC number and certificate.",
  },
  {
    title: "Add a payout account",
    body: "We check the account name with your bank before paying into it.",
  },
  {
    title: "Start listing",
    body: "Once we’ve approved your shop, list products from the TechShop catalogue.",
  },
];

export default async function RegisterPage() {
  const start = "/centre/open-shop";
  if (await bff.getSession()) redirect(start);
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Open a shop on TechShop"
          lead="Five steps. We review verifications within a few working days."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Open a shop" },
          ]}
        />
        <ol className="stack" role="list">
          {steps.map((s, i) => (
            <li
              key={s.title}
              className="card cluster"
              style={{ alignItems: "flex-start" }}
            >
              <span
                className="badge badge--outline tabular-nums"
                aria-hidden="true"
              >
                {i + 1}
              </span>
              <div className="stack stack--xs">
                <h2 className="text-lg">{s.title}</h2>
                <p className="text-sm text-muted">{s.body}</p>
              </div>
            </li>
          ))}
        </ol>
        <div>
          <Link
            href={signInPath(start)}
            className="btn btn--primary btn--lg btn--block-mobile"
          >
            Continue with your phone
          </Link>
        </div>
      </div>
    </main>
  );
}
