import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";

import {
  PaymentMethods,
  SampleDataNotice,
  SiteFooter,
  SiteHeader,
} from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";
import "@techshop/ui/styles.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "TechShop Wholesale & Retail",
  description:
    "Buy TechShop stock at retail, or in bulk at trade prices for your business.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en-NG"
      data-scroll-behavior="smooth"
      className={`${geistSans.variable} ${geistMono.variable}`}
    >
      <body>
        <a href="#main" className="skip-link">
          Skip to content
        </a>
        <SampleDataNotice />
        <SiteHeader
          product="Wholesale"
          nav={[
            { label: "Retail", href: "/retail" },
            { label: "Trade prices", href: "/wholesale" },
            { label: "Categories", href: "/categories" },
            { label: "Request a quote", href: "/quote" },
            { label: "Help", href: "/help" },
          ]}
          actions={
            <>
              <Link
                href="/business/sign-in"
                className="btn btn--ghost btn--sm hide-mobile"
              >
                Business sign in
              </Link>
              <Link
                href="/business/register"
                className="btn btn--inverse btn--sm"
              >
                Open account
              </Link>
            </>
          }
        />
        {children}
        <SiteFooter
          tagline="TechShop’s own stock, at retail or in bulk — with trade pricing, VAT invoices and credit terms for businesses."
          columns={[
            {
              title: "Buy",
              links: [
                { label: "Retail", href: "/retail" },
                { label: "Trade prices", href: "/wholesale" },
                { label: "Request a quote", href: "/quote" },
              ],
            },
            {
              title: "Business",
              links: [
                { label: "Open an account", href: "/business/register" },
                { label: "Credit terms", href: "/business/credit" },
                { label: "Invoices", href: "/business/invoices" },
              ],
            },
            {
              title: "More from TechShop",
              links: [
                { label: "TechShop Market", href: sites.market },
                { label: "Sell on TechShop", href: sites.seller },
              ],
            },
            {
              title: "Company",
              links: [
                { label: "About TechShop", href: sites.corporate },
                { label: "Contact sales", href: "/contact" },
              ],
            },
          ]}
          extra={<PaymentMethods />}
        />
      </body>
    </html>
  );
}
