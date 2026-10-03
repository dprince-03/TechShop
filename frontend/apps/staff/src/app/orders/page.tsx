import type { Metadata } from "next";

import { ModulePage } from "@/components/ModulePage";
import { modules } from "@/lib/moduleData";

export const metadata: Metadata = { title: "Orders | TechShop Staff" };

export default async function OrdersPage(props: PageProps<"/orders">) {
  const sp = await props.searchParams;
  const one = (v: string | string[] | undefined) =>
    Array.isArray(v) ? v[0] : v;
  return (
    <ModulePage
      slug="orders"
      config={modules.orders}
      tab={one(sp.tab)}
      q={one(sp.q)}
    />
  );
}
