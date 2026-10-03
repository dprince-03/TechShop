import type { Metadata } from "next";
import Link from "next/link";

import { nigerianStates } from "@techshop/fixtures";
import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Register as a seller | TechShop Seller Centre",
};

export default function RegisterPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--md stack stack--lg">
        <PageHeader
          title="Create your seller account"
          lead="It takes about 10 minutes. We review applications within a few working days."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Register" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="seller"
            groups={[
              {
                legend: "1. About you",
                fields: [
                  {
                    name: "sellerType",
                    label: "I’m selling as",
                    type: "select",
                    required: true,
                    options: ["An individual", "A registered business"],
                  },
                  {
                    name: "name",
                    label: "Full name (as on your ID)",
                    required: true,
                    autoComplete: "name",
                    half: true,
                  },
                  {
                    name: "phone",
                    label: "Phone number",
                    type: "tel",
                    required: true,
                    autoComplete: "tel",
                    half: true,
                  },
                  {
                    name: "email",
                    label: "Email address",
                    type: "email",
                    required: true,
                    autoComplete: "email",
                  },
                ],
              },
              {
                legend: "2. Your shop",
                fields: [
                  {
                    name: "shop",
                    label: "Shop name",
                    required: true,
                    hint: "Shown to customers on your listings.",
                    half: true,
                  },
                  {
                    name: "rc",
                    label: "CAC RC / BN number",
                    hint: "Registered businesses only.",
                    half: true,
                  },
                  {
                    name: "state",
                    label: "State",
                    type: "select",
                    required: true,
                    options: nigerianStates,
                    half: true,
                  },
                  {
                    name: "city",
                    label: "City or LGA",
                    required: true,
                    autoComplete: "address-level2",
                    half: true,
                  },
                  {
                    name: "sells",
                    label: "What will you sell?",
                    type: "select",
                    required: true,
                    options: [
                      "Phones",
                      "Laptops",
                      "Accessories",
                      "Gaming",
                      "Smart home",
                      "Office",
                      "Workstations",
                      "Cars",
                      "A mix of these",
                    ],
                  },
                ],
              },
              {
                legend: "3. Verification",
                fields: [
                  {
                    name: "idType",
                    label: "ID type",
                    type: "select",
                    required: true,
                    options: [
                      "NIN slip",
                      "International passport",
                      "Driver’s licence",
                      "Voter’s card",
                    ],
                    half: true,
                  },
                  {
                    name: "nin",
                    label: "NIN",
                    inputMode: "numeric",
                    required: true,
                    hint: "11 digits.",
                    minLength: 11,
                    half: true,
                  },
                  {
                    name: "idFile",
                    label: "Photo of your ID",
                    type: "file",
                    accept: "image/*,.pdf",
                    required: true,
                  },
                  {
                    name: "cac",
                    label: "CAC certificate",
                    type: "file",
                    accept: "image/*,.pdf",
                    hint: "Registered businesses only.",
                  },
                ],
              },
              {
                legend: "4. Payouts",
                fields: [
                  {
                    name: "bank",
                    label: "Bank",
                    type: "select",
                    required: true,
                    options: [
                      "Access Bank",
                      "First Bank",
                      "GTBank",
                      "Moniepoint MFB",
                      "OPay",
                      "UBA",
                      "Wema Bank",
                      "Zenith Bank",
                      "Other",
                    ],
                    half: true,
                  },
                  {
                    name: "account",
                    label: "Account number (NUBAN)",
                    inputMode: "numeric",
                    required: true,
                    minLength: 10,
                    half: true,
                    hint: "10 digits, in your name or your business’s name.",
                  },
                  {
                    name: "password",
                    label: "Create a password",
                    type: "password",
                    required: true,
                    autoComplete: "new-password",
                    minLength: 8,
                    half: true,
                  },
                  {
                    name: "confirm",
                    label: "Confirm password",
                    type: "password",
                    required: true,
                    autoComplete: "new-password",
                    matches: "password",
                    half: true,
                  },
                  {
                    name: "agreement",
                    label:
                      "I agree to the Seller agreement and Prohibited items policy",
                    type: "checkbox",
                    required: true,
                  },
                ],
              },
            ]}
            submitLabel="Submit application"
            successTitle="Application ready"
            successMessage="In the live Seller Centre we’d verify your ID and bank details and email you the result."
          />
        </div>
        <p className="text-sm text-muted">
          Already a seller?{" "}
          <Link href="/sign-in" className="link">
            Sign in
          </Link>{" "}
          · Read the{" "}
          <Link href="/legal/seller-agreement" className="link">
            Seller agreement
          </Link>
        </p>
      </div>
    </main>
  );
}
