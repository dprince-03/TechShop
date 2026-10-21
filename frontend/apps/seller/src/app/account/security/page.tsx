import type { Metadata } from "next";
import { connection } from "next/server";

import { PageHeader } from "@techshop/ui/components";

import { AuthenticatorSetup } from "@/features/account/components/AuthenticatorSetup";
import { sellerApi } from "@/features/session/server";

export const metadata: Metadata = {
  title: "Security | TechShop Seller Centre",
};

/** Only same-site paths are followed after setup (no open redirects). */
const safeNext = (v: unknown) =>
  typeof v === "string" &&
  v.startsWith("/") &&
  !v.startsWith("//") &&
  !v.startsWith("/\\")
    ? v
    : undefined;

export default async function SecurityPage(
  props: PageProps<"/account/security">,
) {
  await connection();
  const next = safeNext((await props.searchParams).next);
  const api = await sellerApi(
    `/account/security${next ? `?next=${encodeURIComponent(next)}` : ""}`,
  );
  const me = await api.get("/api/v1/me", { init: { cache: "no-store" } });
  return (
    <main className="section section--page">
      <div className="container container--narrow stack">
        <PageHeader
          title="Security"
          lead="An authenticator app protects your shop’s money. Finance members must use one."
        />
        <section
          className="card card--roomy stack"
          aria-labelledby="totp-title"
        >
          <h2 id="totp-title" className="text-xl">
            Authenticator app
          </h2>
          {me.mfaEnabled ? (
            <p role="status">
              It’s on. Signing in asks for a code from your app.
            </p>
          ) : (
            <AuthenticatorSetup next={next} />
          )}
        </section>
      </div>
    </main>
  );
}
