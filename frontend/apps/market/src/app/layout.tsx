import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import { SampleDataNotice } from "@techshop/ui/components";
import "@techshop/ui/styles.css";

import { MarketFooter } from "@/components/MarketFooter";
import { MarketHeader } from "@/components/MarketHeader";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "TechShop Market",
  description:
    "Shop phones, laptops, gaming, smart home and more from TechShop and trusted sellers.",
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
        <MarketHeader />
        {children}
        <MarketFooter />
      </body>
    </html>
  );
}
