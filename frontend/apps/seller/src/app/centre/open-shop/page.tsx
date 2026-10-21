import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { connection } from "next/server";

import { PageHeader } from "@techshop/ui/components";

import { myShops } from "@/features/session/server";
import { OpenShopForm } from "@/features/shops/components/OpenShopForm";
import { stateOptions } from "@/features/shops/server";

export const metadata: Metadata = {
  title: "Open a shop | TechShop Seller Centre",
};

export default async function OpenShopPage() {
  await connection();
  const shops = await myShops("/centre/open-shop");
  const owned = shops.find((s) => s.role === "owner" && s.status !== "closed");
  if (owned) redirect(`/centre/${owned.id}`); // one open shop per owner
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Open your shop"
          lead="Next you’ll verify your identity and add a payout account."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Open a shop" },
          ]}
        />
        <div className="card card--roomy">
          <OpenShopForm states={await stateOptions()} />
        </div>
      </div>
    </main>
  );
}
