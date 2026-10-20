import { PaymentMethods, SiteFooter } from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";

export function MarketFooter() {
  return (
    <SiteFooter
      tagline="Phones, laptops, gaming, smart home, office tech and cars — from TechShop and verified sellers across Nigeria."
      columns={[
        {
          title: "Shop",
          links: [
            { label: "Today’s deals", href: "/deals" },
            { label: "Phones", href: "/c/phones" },
            { label: "Laptops", href: "/c/laptops" },
            { label: "Gaming", href: "/c/gaming" },
            { label: "Cars", href: "/c/cars" },
          ],
        },
        {
          title: "Help",
          links: [
            { label: "Track an order", href: "/orders" },
            { label: "Delivery", href: "/help/delivery" },
            { label: "Returns & refunds", href: "/help/returns" },
            { label: "Warranty", href: "/help/warranty" },
            { label: "Contact us", href: "/help/contact" },
          ],
        },
        {
          title: "Sell & business",
          links: [
            { label: "Sell on TechShop", href: sites.seller },
            { label: "Wholesale & retail", href: sites.wholesale },
            { label: "Trade in your device", href: "/trade-in" },
          ],
        },
        {
          title: "TechShop",
          links: [
            { label: "About us", href: sites.corporate },
            { label: "Careers", href: `${sites.corporate}/careers` },
            { label: "Newsroom", href: `${sites.corporate}/news` },
          ],
        },
      ]}
      extra={<PaymentMethods />}
    />
  );
}
