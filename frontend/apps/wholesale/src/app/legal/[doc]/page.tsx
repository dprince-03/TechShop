import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { publishedPage } from "@techshop/api-client";
import {
  LegalPageView,
  legalDocs,
  type LegalDoc,
} from "@techshop/ui/components";

import { bff } from "@/lib/bff";

const DOCS: LegalDoc[] = ["terms", "privacy", "cookies"];

export async function generateMetadata(
  props: PageProps<"/legal/[doc]">,
): Promise<Metadata> {
  const d = legalDocs[(await props.params).doc as LegalDoc];
  return d
    ? { title: `${d.title} | TechShop Wholesale`, description: d.summary }
    : {};
}

/** Published text from the CMS when legal has signed it off; the draft outline until then. */
export default async function LegalPage(props: PageProps<"/legal/[doc]">) {
  await connection();
  const { doc } = await props.params;
  if (!DOCS.includes(doc as LegalDoc)) notFound();
  const page = await publishedPage(bff.publicApi(), "wholesale", doc);
  return (
    <main id="main" className="section section--page">
      <div className="container">
        <LegalPageView doc={doc as LegalDoc} page={page} />
      </div>
    </main>
  );
}
