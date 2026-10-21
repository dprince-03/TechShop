import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { safeReturnPath } from "@techshop/bff";
import { PageHeader, SignInForm } from "@techshop/ui/components";

import { signInStep } from "@/features/auth/actions";
import { bff } from "@/lib/bff";

export const metadata: Metadata = { title: "Sign in | TechShop Seller Centre" };

export default async function SignInPage(props: PageProps<"/sign-in">) {
  const sp = await props.searchParams;
  const next = safeReturnPath(
    typeof sp.next === "string" ? sp.next : "/centre",
  );
  if (await bff.getSession()) redirect(next);
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Sign in to Seller Centre"
          lead="Use your mobile number. On a new phone or computer we’ll text you a second code."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Sign in" },
          ]}
        />
        <div className="card card--roomy">
          <SignInForm
            next={next}
            action={signInStep}
            welcome="Welcome to TechShop Seller Centre. What’s your name, as on your ID?"
          />
        </div>
        <p className="text-sm text-muted">
          New seller?{" "}
          <Link href="/register" className="link">
            Open a shop
          </Link>
        </p>
      </div>
    </main>
  );
}
