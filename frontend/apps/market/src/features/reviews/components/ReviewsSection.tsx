import Link from "next/link";

import { Rating } from "@techshop/ui/components";

import { bff } from "@/lib/bff";

import { ReportReviewForm } from "./ReviewForms";

const date = new Intl.DateTimeFormat("en-NG", {
  dateStyle: "medium",
  timeZone: "Africa/Lagos",
});

/** Ratings summary and reviews (verified purchases first) for a product page. */
export async function ReviewsSection({
  slug,
  verifiedOnly,
}: {
  slug: string;
  verifiedOnly: boolean;
}) {
  const [data, session] = await Promise.all([
    bff
      .publicApi()
      .get("/api/v1/products/{slug}/reviews", {
        path: { slug },
        query: { verified: verifiedOnly ? "true" : undefined, limit: 10 },
        init: { cache: "no-store" },
      })
      .catch(() => null),
    bff.getSession(),
  ]);
  if (!data) return null;
  const reviews = data.reviews.items ?? [];
  const { summary } = data;
  return (
    <section aria-labelledby="reviews-title" className="stack">
      <h2 id="reviews-title" className="text-2xl">
        Reviews
      </h2>
      {summary.total > 0 ? (
        <div className="cluster">
          <Rating average={summary.average} count={summary.total} />
          <span className="text-sm text-muted">
            {summary.verified} verified purchase
            {summary.verified === 1 ? "" : "s"}. Unverified reviews count half.
          </span>
          <Link
            href={
              verifiedOnly
                ? `/p/${slug}#reviews-title`
                : `/p/${slug}?reviews=verified#reviews-title`
            }
            className="link text-sm"
          >
            {verifiedOnly ? "Show all reviews" : "Only verified purchases"}
          </Link>
        </div>
      ) : (
        <p className="text-muted">
          No reviews yet. Bought it? Review it from your order page.
        </p>
      )}
      <ul role="list" className="stack">
        {reviews.map((r) => (
          <li key={r.id} className="card stack stack--sm">
            <div className="cluster cluster--between">
              <strong>{r.rating} out of 5</strong>
              <span className="text-xs text-muted">
                {date.format(new Date(r.createdAt))}
              </span>
            </div>
            {r.title ? (
              <p>
                <strong>{r.title}</strong>
              </p>
            ) : null}
            {r.body ? <p>{r.body}</p> : null}
            <p className="text-sm text-muted">
              {r.reviewer}
              {r.verified ? (
                <span className="badge badge--outline">Verified purchase</span>
              ) : null}
            </p>
            {r.sellerResponse ? (
              <p className="text-sm">
                <strong>Seller’s answer:</strong> {r.sellerResponse}
              </p>
            ) : null}
            {session ? <ReportReviewForm reviewId={r.id} /> : null}
          </li>
        ))}
      </ul>
    </section>
  );
}
