import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import "@techshop/ui/styles.css";

import { getCart, itemCount } from "@/features/cart/server";
import { MarketFooter } from "@/features/shell/MarketFooter";
import { MarketHeader } from "@/features/shell/MarketHeader";
import { bff } from "@/lib/bff";

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

export default async function RootLayout({ children }: LayoutProps<"/">) {
  const [session, cart] = await Promise.all([bff.getSession(), getCart()]);
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
        <MarketHeader signedIn={Boolean(session)} cartCount={itemCount(cart)} />
        {children}
        <MarketFooter />
      </body>
    </html>
  );
}
