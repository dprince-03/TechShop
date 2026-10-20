import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { safeReturnPath } from "@techshop/bff";
import { Logo } from "@techshop/ui/components";

import { StaffSignInForm } from "@/features/auth/components/StaffSignInForm";
import { getStaff } from "@/features/session/server";

import styles from "./sign-in.module.css";

export const metadata: Metadata = { title: "Sign in | TechShop Staff" };

export default async function SignInPage(props: PageProps<"/sign-in">) {
  const sp = await props.searchParams;
  const next = safeReturnPath(typeof sp.next === "string" ? sp.next : "/");
  if (await getStaff()) redirect(next);
  return (
    <main id="main" className={styles.page}>
      <div className={`card card--roomy stack stack--lg ${styles.panel}`}>
        <Logo product="Staff" />
        <h1 className={styles.title}>Sign in to the staff portal</h1>
        <StaffSignInForm next={next} />
      </div>
    </main>
  );
}
