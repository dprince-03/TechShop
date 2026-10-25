import type { Metadata } from "next";
import { connection } from "next/server";

import { NewShopTicketForm } from "@/features/aftersales/components/Forms";
import { requireShop } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";

export const metadata: Metadata = {
  title: "Ask seller support | TechShop Seller Centre",
};

export default async function NewTicketPage(
  props: PageProps<"/centre/[shopId]/support/new">,
) {
  await connection();
  const { shopId } = await props.params;
  const shop = await requireShop(shopId, `/centre/${shopId}/support/new`);
  return (
    <ShopPage shop={shop} section="support" title="Ask seller support">
      <section className="card card--roomy">
        <NewShopTicketForm shopId={shop.id} />
      </section>
    </ShopPage>
  );
}
