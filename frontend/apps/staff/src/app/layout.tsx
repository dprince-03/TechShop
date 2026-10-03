import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import { SampleDataNotice } from "@techshop/ui/components";
import "@techshop/ui/styles.css";

import { StaffShell } from "@/components/StaffShell";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "TechShop Staff",
  description: "Internal TechShop staff portal.",
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
        <SampleDataNotice />
        <StaffShell>{children}</StaffShell>
      </body>
    </html>
  );
}
