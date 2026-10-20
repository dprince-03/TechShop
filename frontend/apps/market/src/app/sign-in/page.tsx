import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { safeReturnPath } from "@techshop/bff";
import { PageHeader } from "@techshop/ui/components";

import { SignInForm } from "@/features/auth/components/SignInForm";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Sign in | TechShop Market",
  robots: { index: false },
};

export default async function SignInPage(props: PageProps<"/sign-in">) {
  const sp = await props.searchParams;
  const next = safeReturnPath(typeof sp.next === "string" ? sp.next : "/");
  if (await bff.getSession()) redirect(next);
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Sign in or create an account"
          lead="Use your mobile number. No password needed."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Sign in" }]}
        />
        <div className="card card--roomy">
          <SignInForm next={next} />
        </div>
      </div>
    </main>
  );
}
