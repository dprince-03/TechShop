import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { PageHeader } from "@techshop/ui/components";

import { formatKobo } from "@techshop/api-client";

import { AddressBook } from "@/features/account/components/AddressBook";
import { AccountNav } from "@/features/aftersales/components/Records";
import { storeCredit } from "@/features/aftersales/server";
import { ProfileForm } from "@/features/account/components/ProfileForm";
import { Sessions } from "@/features/account/components/Sessions";
import { loadAccount } from "@/features/account/server";
import { signOut } from "@/features/auth/actions";
import { getStates } from "@/features/catalogue/api";
import { bff } from "@/lib/bff";

import styles from "./account.module.css";

export const metadata: Metadata = {
  title: "Your account | TechShop Market",
  robots: { index: false },
};

export default async function AccountPage() {
  await connection();
  await bff.requireSession("/account");
  const [{ me, addresses, sessions }, states, credit] = await Promise.all([
    loadAccount(),
    getStates(),
    storeCredit().catch(() => null),
  ]);
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title={`Hello, ${me.firstName}`}
          lead="Your details, delivery addresses and signed-in devices."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Account" }]}
          actions={
            <div className="cluster">
              <Link href="/orders" className="btn btn--secondary">
                Your orders
              </Link>
              <Link href="/saved" className="btn btn--secondary">
                Saved items
              </Link>
              <form action={signOut}>
                <button className="btn btn--ghost">Sign out</button>
              </form>
            </div>
          }
        />
        <AccountNav current="/account" />
        {credit && credit.balanceKobo > 0 ? (
          <p className="alert alert--success" role="status">
            You have{" "}
            <strong className="tabular-nums">
              {formatKobo(credit.balanceKobo)}
            </strong>{" "}
            store credit. Use it at checkout when it covers the whole order.
          </p>
        ) : null}
        <div className={styles.columns}>
          <section aria-labelledby="profile-title" className="card card--roomy">
            <h2 id="profile-title" className={styles.title}>
              Profile
            </h2>
            <ProfileForm me={me} />
          </section>
          <section
            aria-labelledby="addresses-title"
            className="card card--roomy"
          >
            <h2 id="addresses-title" className={styles.title}>
              Delivery addresses
            </h2>
            <AddressBook addresses={addresses} states={states} />
          </section>
        </div>
        <section aria-labelledby="devices-title" className="card card--roomy">
          <h2 id="devices-title" className={styles.title}>
            Signed-in devices
          </h2>
          <Sessions sessions={sessions} />
        </section>
      </div>
    </main>
  );
}
