import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

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
  title: "TechShop",
  description:
    "The TechShop company: who we are, what we do, and how to work with us.",
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
          nav={[
            { label: "About", href: "/about" },
            { label: "Our businesses", href: "/#businesses-title" },
            { label: "Partners", href: "/partners" },
            { label: "Careers", href: "/careers" },
            { label: "Newsroom", href: "/news" },
            { label: "Contact", href: "/contact" },
          ]}
          actions={
            <a href={sites.market} className="btn btn--inverse btn--sm">
              Shop now
            </a>
          }
        />
        {children}
        <SiteFooter
          tagline="TechShop sells, supplies and supports the technology Nigerian homes and businesses depend on."
          columns={[
            {
              title: "Company",
              links: [
                { label: "About", href: "/about" },
                { label: "Careers", href: "/careers" },
                { label: "Newsroom", href: "/news" },
              ],
            },
            {
              title: "Businesses",
              links: [
                { label: "TechShop Market", href: sites.market },
                { label: "Wholesale & retail", href: sites.wholesale },
                { label: "Sell on TechShop", href: sites.seller },
              ],
            },
            {
              title: "Partners",
              links: [
                { label: "Suppliers & brands", href: "/partners" },
                { label: "Logistics partners", href: "/partners#logistics" },
              ],
            },
            {
              title: "Contact",
              links: [
                { label: "Customer support", href: `${sites.market}/help` },
                { label: "Business sales", href: `${sites.wholesale}/quote` },
                { label: "Press", href: "/contact#press" },
              ],
            },
          ]}
        />
      </body>
    </html>
  );
}
