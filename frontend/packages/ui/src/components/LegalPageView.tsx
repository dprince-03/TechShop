import { LegalDocument, type LegalDoc } from "./LegalDocument";
import { PublishedDocument } from "./PublishedDocument";

/** A legal page: the published CMS version when there is one, otherwise the draft outline. */
export function LegalPageView({
  doc,
  page,
}: {
  doc: LegalDoc;
  page: {
    title: string;
    body: unknown;
    updatedAt?: string | null;
    publishedAt?: string | null;
  } | null;
}) {
  return page ? (
    <PublishedDocument
      title={page.title}
      body={page.body}
      updatedAt={page.publishedAt ?? page.updatedAt}
    />
  ) : (
    <LegalDocument doc={doc} />
  );
}
