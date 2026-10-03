import type { Metadata } from "next";
import { notFound } from "next/navigation";

import {
  LegalDocument,
  legalDocs,
  type LegalDoc,
} from "@techshop/ui/components";

const DOCS: LegalDoc[] = ["terms", "privacy", "cookies"];

export const dynamicParams = false;

export function generateStaticParams() {
  return DOCS.map((doc) => ({ doc }));
}

export async function generateMetadata(
  props: PageProps<"/legal/[doc]">,
): Promise<Metadata> {
  const { doc } = await props.params;
  const d = legalDocs[doc as LegalDoc];
  return d ? { title: `${d.title} | TechShop`, description: d.summary } : {};
}

export default async function LegalPage(props: PageProps<"/legal/[doc]">) {
  const { doc } = await props.params;
  if (!DOCS.includes(doc as LegalDoc)) notFound();
  return (
    <main id="main" className="section section--page">
      <div className="container">
        <LegalDocument doc={doc as LegalDoc} />
      </div>
    </main>
  );
}
