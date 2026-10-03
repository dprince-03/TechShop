import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { ModulePage } from "@/components/ModulePage";
import { modules } from "@/lib/moduleData";

/** Modules with their own folders (sub-pages) are excluded here. */
const SLUGS = Object.keys(modules).filter(
  (m) => m !== "orders" && m !== "catalogue",
);

export const dynamicParams = false;

export function generateStaticParams() {
  return SLUGS.map((module) => ({ module }));
}

export async function generateMetadata(
  props: PageProps<"/[module]">,
): Promise<Metadata> {
  const m = modules[(await props.params).module];
  return m ? { title: `${m.title} | TechShop Staff` } : {};
}

export default async function Module(props: PageProps<"/[module]">) {
  const { module } = await props.params;
  if (!SLUGS.includes(module)) notFound();
  const sp = await props.searchParams;
  const one = (v: string | string[] | undefined) =>
    Array.isArray(v) ? v[0] : v;
  return (
    <ModulePage
      slug={module}
      config={modules[module]}
      tab={one(sp.tab)}
      q={one(sp.q)}
    />
  );
}
