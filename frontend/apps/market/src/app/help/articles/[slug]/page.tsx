import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ContentBlocks, PageHeader } from "@techshop/ui/components";

import { getArticle } from "@/features/help/api";

export async function generateMetadata(
  props: PageProps<"/help/articles/[slug]">,
): Promise<Metadata> {
  const a = await getArticle((await props.params).slug).catch(() => null);
  return { title: a ? `${a.title} | TechShop Help` : "Help | TechShop Market" };
}

export default async function HelpArticlePage(
  props: PageProps<"/help/articles/[slug]">,
) {
  await connection();
  const article = await getArticle((await props.params).slug);
  if (!article) notFound();
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          eyebrow={article.topic}
          title={article.title}
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Help", href: "/help" },
            { label: article.title },
          ]}
        />
        <ContentBlocks body={article.body} />
      </div>
    </main>
  );
}
