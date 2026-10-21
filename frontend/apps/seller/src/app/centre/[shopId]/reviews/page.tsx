import type { Metadata } from "next";
import { connection } from "next/server";

import { EmptyState } from "@techshop/ui/components";

import { RespondForm } from "@/features/reviews/components/RespondForm";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime } from "@/lib/format";

export const metadata: Metadata = { title: "Reviews | TechShop Seller Centre" };

export default async function ReviewsPage(
  props: PageProps<"/centre/[shopId]/reviews">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}/reviews`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const page = await api.get("/api/v1/seller/shops/{shopId}/reviews", {
    path: { shopId },
    init: { cache: "no-store" },
  });
  const reviews = page.items ?? [];
  const canAnswer = shop.role === "owner" || shop.role === "manager";
  return (
    <ShopPage
      shop={shop}
      section="reviews"
      title="Reviews"
      lead="Reviews of items you shipped. You can answer each one once; to report one that breaks the rules, use the link on the product page."
    >
      {reviews.length === 0 ? (
        <EmptyState icon="star" title="No reviews yet">
          Customers can review items after delivery.
        </EmptyState>
      ) : (
        <ul role="list" className="stack">
          {reviews.map((r) => (
            <li key={r.id} className="card stack">
              <div className="cluster cluster--between">
                <strong>{r.productName}</strong>
                <span className="text-sm text-muted">
                  {dateTime.format(new Date(r.createdAt))}
                </span>
              </div>
              <p className="text-sm">
                <strong>{r.rating} out of 5</strong>
              </p>
              {r.title ? (
                <p>
                  <strong>{r.title}</strong>
                </p>
              ) : null}
              {r.body ? <p>{r.body}</p> : null}
              <p className="text-sm text-muted">— {r.reviewer}</p>
              {r.sellerResponse ? (
                <p className="alert alert--success">
                  <strong>Your answer:</strong> {r.sellerResponse}
                </p>
              ) : canAnswer ? (
                <RespondForm shopId={shop.id} reviewId={r.id} />
              ) : null}
            </li>
          ))}
        </ul>
      )}
    </ShopPage>
  );
}
