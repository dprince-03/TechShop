import Link from "next/link";
import type { ReactNode } from "react";

import type { Crumb } from "@techshop/ui/components";

import { listProducts } from "./api";
import { ProductListing } from "./components/ProductListing";
import { ServiceUnavailable } from "./components/ServiceUnavailable";
import type { MarketCategory } from "./mappers";
import { toCard } from "./mappers";
import { listingHref, type ListingQuery } from "./query";

type Props = {
  kind?: "search" | "deals" | "best-sellers";
  title: string;
  description?: string;
  breadcrumbs: Crumb[];
  action: string;
  query: ListingQuery;
  /** Top-level category slug for the card art, when the listing is one category. */
  artCategory?: string;
  browse?: ReactNode;
  categories?: MarketCategory[];
};

/** Loads one page of a listing from the API and renders it (or an unavailable state). */
export async function ListingPage({
  kind = "search",
  artCategory,
  ...props
}: Props) {
  let result: Awaited<ReturnType<typeof listProducts>>;
  try {
    result = await listProducts(props.query, kind);
  } catch {
    return (
      <ServiceUnavailable retryHref={listingHref(props.action, props.query)} />
    );
  }
  // The API suggests a spelling when it understood the words differently ("iphon" → "iphone").
  const q = props.query.q;
  const suggestion =
    q &&
    result.didYouMean &&
    result.didYouMean.toLowerCase() !== q.toLowerCase()
      ? result.didYouMean
      : undefined;
  const notice = suggestion ? (
    <p className="text-sm text-muted" role="note">
      Did you mean{" "}
      <Link
        href={listingHref(props.action, { ...props.query, q: suggestion })}
        className="link"
      >
        {suggestion}
      </Link>
      ?
    </p>
  ) : null;
  return (
    <ProductListing
      {...props}
      products={(result.items ?? []).map((p) => toCard(p, artCategory))}
      total={result.total}
      brands={result.facets?.brand ?? []}
      nextCursor={result.page.hasMore ? result.page.nextCursor : undefined}
      notice={notice}
    />
  );
}
