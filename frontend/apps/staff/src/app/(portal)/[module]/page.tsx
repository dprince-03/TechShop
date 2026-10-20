import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { SampleDataNotice } from "@techshop/ui/components";

import { ModulePage } from "@/features/preview/ModulePage";
import { modules } from "@/features/preview/moduleData";
import { requireStaff } from "@/features/session/server";
import { moduleGroups } from "@/features/shell/modules";

/** Which backend milestone brings each preview module its API (docs/backend.md §8). */
const MILESTONE: Record<string, string> = {
  vendors: "M4 (marketplace)",
  support: "M5 (after-sales and support)",
  risk: "M5 (after-sales and support)",
  warranty: "M5 (after-sales and support)",
  "trade-ins": "M5 (after-sales and support)",
  b2b: "M6 (B2B)",
  purchasing: "M6 (B2B)",
  analytics: "M7",
  cars: "M7",
  pos: "M7",
  it: "M7",
};

/** Preview modules only: live modules have their own folders and use the API. */
const PREVIEWS = moduleGroups
  .flatMap((g) => g.modules)
  .filter((m) => !m.live)
  .map((m) => ({ slug: m.href.slice(1), permission: m.permission }))
  .filter((m) => m.slug in modules);

export async function generateMetadata(
  props: PageProps<"/[module]">,
): Promise<Metadata> {
  const m = modules[(await props.params).module];
  return m ? { title: `${m.title} (preview) | TechShop Staff` } : {};
}

export default async function Module(props: PageProps<"/[module]">) {
  const { module } = await props.params;
  const preview = PREVIEWS.find((p) => p.slug === module);
  if (!preview) notFound();
  await requireStaff(preview.permission);
  const sp = await props.searchParams;
  const one = (v: string | string[] | undefined) =>
    Array.isArray(v) ? v[0] : v;
  return (
    <div className="stack">
      <SampleDataNotice>
        Preview: everything on this page is sample data. The live module arrives
        with backend {MILESTONE[module] ?? "a later milestone"}.
      </SampleDataNotice>
      <ModulePage
        slug={module}
        config={modules[module]}
        tab={one(sp.tab)}
        q={one(sp.q)}
      />
    </div>
  );
}
