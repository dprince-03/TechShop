import type { Metadata } from "next";
import Link from "next/link";

import { FaqList, PageHeader, Prose } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Credit terms | TechShop Wholesale & Retail",
};

export default function CreditPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          eyebrow="Business"
          title="Credit terms"
          lead="Approved businesses can receive goods now and pay on invoice."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Credit terms" },
          ]}
          actions={
            <Link href="/business/register" className="btn btn--primary">
              Open a business account
            </Link>
          }
        />
        <Prose>
          <h2>Who can apply</h2>
          <p>
            Registered businesses, schools, government agencies and NGOs with a
            TechShop business account and verified CAC details.
          </p>
          <h2>How it works</h2>
          <ol>
            <li>Open a business account and complete verification.</li>
            <li>
              Apply for credit from your account. We may ask for recent bank
              statements or references.
            </li>
            <li>
              If approved, you’ll see your credit limit and payment period in
              your account.
            </li>
            <li>
              Order as usual and choose “Pay on invoice” at checkout. Pay by
              bank transfer before the due date.
            </li>
          </ol>
          <h2>Limits and periods</h2>
          <p>
            Credit limits and payment periods are set per business after review.
            They’re shown in your account and on every invoice.
          </p>
        </Prose>
        <FaqList
          items={[
            {
              q: "Is there a fee for credit terms?",
              a: "Fees, if any, are shown before you accept a credit offer. There are no hidden charges.",
            },
            {
              q: "What happens if we pay late?",
              a: "We’ll remind you before and after the due date. Late payment may pause credit on your account until it’s settled.",
            },
          ]}
        />
        <p className="text-xs text-muted">
          Sample content — credit policy to be confirmed by TechShop finance.
        </p>
      </div>
    </main>
  );
}
