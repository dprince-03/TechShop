import type { Metadata } from "next";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

import styles from "./account.module.css";

export const metadata: Metadata = {
  title: "Sign in | TechShop Market",
  robots: { index: false },
};

export default function AccountPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container">
        <PageHeader
          title="Sign in or create an account"
          lead="Track orders, save items and check out faster. Guest checkout is also available."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Account" }]}
        />
        <div className={styles.columns}>
          <section aria-labelledby="signin-title" className="card card--roomy">
            <h2 id="signin-title" className={styles.title}>
              Sign in
            </h2>
            <PreviewForm
              id="signin"
              groups={[
                {
                  fields: [
                    {
                      name: "email",
                      label: "Email address",
                      type: "email",
                      required: true,
                      autoComplete: "email",
                    },
                    {
                      name: "password",
                      label: "Password",
                      type: "password",
                      required: true,
                      autoComplete: "current-password",
                    },
                  ],
                },
              ]}
              submitLabel="Sign in"
              successTitle="Details accepted"
              successMessage="In the live store you’d now be signed in."
              note="Forgot your password? Password reset arrives with real accounts."
            />
          </section>
          <section
            aria-labelledby="register-title"
            className="card card--roomy"
          >
            <h2 id="register-title" className={styles.title}>
              Create an account
            </h2>
            <PreviewForm
              id="register"
              groups={[
                {
                  fields: [
                    {
                      name: "firstName",
                      label: "First name",
                      required: true,
                      autoComplete: "given-name",
                      half: true,
                    },
                    {
                      name: "lastName",
                      label: "Last name",
                      required: true,
                      autoComplete: "family-name",
                      half: true,
                    },
                    {
                      name: "email",
                      label: "Email address",
                      type: "email",
                      required: true,
                      autoComplete: "email",
                    },
                    {
                      name: "phone",
                      label: "Mobile number",
                      type: "tel",
                      required: true,
                      autoComplete: "tel",
                      placeholder: "0803 123 4567",
                      hint: "For delivery updates. We never share it with sellers’ marketing.",
                    },
                    {
                      name: "password",
                      label: "Password",
                      type: "password",
                      required: true,
                      autoComplete: "new-password",
                      minLength: 8,
                      hint: "At least 8 characters.",
                    },
                    {
                      name: "terms",
                      label: "I agree to the Terms of use and Privacy policy",
                      type: "checkbox",
                      required: true,
                    },
                    {
                      name: "news",
                      label:
                        "Send me deals and new arrivals by email (you can unsubscribe any time)",
                      type: "checkbox",
                    },
                  ],
                },
              ]}
              submitLabel="Create account"
              successTitle="Account details accepted"
              successMessage="In the live store we’d now send a verification code to your phone."
            />
          </section>
        </div>
      </div>
    </main>
  );
}
