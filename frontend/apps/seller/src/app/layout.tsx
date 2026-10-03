import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";

import { SiteFooter, SiteHeader } from "@techshop/ui/components";
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
  title: "TechShop Seller Centre",
  description: "Sell on TechShop: manage your listings, orders, and payouts.",
  robots: { index: false, follow: false },
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
        <SiteHeader
          product="Seller Centre"
          nav={[
            { label: "How it works", href: "/#how-title" },
            { label: "Requirements", href: "/#req-title" },
            { label: "Fees", href: "/fees" },
            { label: "FAQ", href: "/#faq-title" },
            { label: "Help", href: "/help" },
          ]}
          actions={
            <>
              <Link
                href="/sign-in"
                className="btn btn--ghost btn--sm hide-mobile"
              >
                Sign in
              </Link>
              <Link href="/register" className="btn btn--inverse btn--sm">
                Start selling
              </Link>
            </>
          }
        />
        {children}
        <SiteFooter
          tagline="Seller Centre: everything you need to sell on TechShop Market."
          columns={[
            {
              title: "Selling",
              links: [
                { label: "Register", href: "/register" },
                { label: "Fees", href: "/fees" },
                { label: "Seller policies", href: "/policies" },
              ],
            },
            {
              title: "Support",
              links: [
                { label: "Seller help", href: "/help" },
                { label: "Contact seller support", href: "/help/contact" },
              ],
            },
            {
              title: "TechShop",
              links: [
                { label: "TechShop Market", href: sites.market },
                { label: "Wholesale & retail", href: sites.wholesale },
                { label: "About TechShop", href: sites.corporate },
              ],
            },
            {
              title: "Legal",
              links: [
                { label: "Seller agreement", href: "/legal/seller-agreement" },
                { label: "Prohibited items", href: "/policies/prohibited" },
              ],
            },
          ]}
        />
      </body>
    </html>
  );
}
