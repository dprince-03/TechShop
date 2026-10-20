import { ContentBlocks } from "./ContentBlocks";
import { PageHeader } from "./PageHeader";

/** A CMS page as a document: its single h1, last-updated date and the safe block renderer. */
export function PublishedDocument({
  title,
  body,
  updatedAt,
  breadcrumbs,
}: {
  title: string;
  body: unknown;
  updatedAt?: string | null;
  breadcrumbs?: { label: string; href?: string }[];
}) {
  const updated = updatedAt
    ? new Intl.DateTimeFormat("en-NG", {
        dateStyle: "long",
        timeZone: "Africa/Lagos",
      }).format(new Date(updatedAt))
    : undefined;
  return (
    <article className="stack stack--lg">
      <PageHeader
        title={title}
        lead={updated ? `Last updated ${updated}` : undefined}
        breadcrumbs={breadcrumbs}
      />
      <ContentBlocks body={body} />
    </article>
  );
}
