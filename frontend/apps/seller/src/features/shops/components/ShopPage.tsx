import type { ReactNode } from "react";

import { PageHeader, Status } from "@techshop/ui/components";

import type { Shop } from "@/features/session/server";

import { shopStatus } from "../server";
import { ShopNav, type Section } from "./ShopNav";

type Props = {
  shop: Shop;
  section: Section;
  title: string;
  lead?: string;
  actions?: ReactNode;
  children: ReactNode;
};

/** Frame for every shop page: the shop's name and status, the section title and the shop nav. */
export function ShopPage({
  shop,
  section,
  title,
  lead,
  actions,
  children,
}: Props) {
  const status = shopStatus[shop.status];
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          eyebrow={shop.displayName}
          title={title}
          lead={lead}
          actions={
            actions ?? <Status label={status.label} tone={status.tone} />
          }
          breadcrumbs={[
            { label: "Seller Centre", href: "/centre" },
            { label: shop.displayName, href: `/centre/${shop.id}` },
            { label: title },
          ]}
        />
        <ShopNav shopId={shop.id} current={section} />
        {children}
      </div>
    </main>
  );
}
