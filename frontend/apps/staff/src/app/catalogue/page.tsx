import type { Metadata } from "next";

import { ModulePage } from "@/components/ModulePage";
import { modules } from "@/lib/moduleData";

export const metadata: Metadata = { title: "Catalogue | TechShop Staff" };

export default async function CataloguePage(props: PageProps<"/catalogue">) {
  const sp = await props.searchParams;
  const one = (v: string | string[] | undefined) =>
    Array.isArray(v) ? v[0] : v;
  return (
    <ModulePage
      slug="catalogue"
      config={modules.catalogue}
      tab={one(sp.tab)}
      q={one(sp.q)}
    />
  );
}
