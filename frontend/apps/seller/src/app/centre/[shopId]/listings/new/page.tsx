import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";
import { connection } from "next/server";

import { EmptyState } from "@techshop/ui/components";

import { NewListingForm } from "@/features/listings/components/ListingForms";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "New listing | TechShop Seller Centre",
};

/** Two steps on one URL: find the product (?q=), then offer one of its variants (?product=). */
export default async function NewListingPage(
  props: PageProps<"/centre/[shopId]/listings/new">,
) {
  await connection();
  const { shopId } = await props.params;
  const sp = await props.searchParams;
  const here = `/centre/${shopId}/listings/new`;
  const shop = await requireShop(shopId, here);
  await sellerApi(here);
  if (shop.status !== "active" || shop.role === "finance")
    redirect(`/centre/${shopId}/listings`);
  const q = typeof sp.q === "string" ? sp.q.trim().slice(0, 120) : "";
  const slug = typeof sp.product === "string" ? sp.product : "";
  const api = bff.publicApi();
  const product = slug
    ? await api
        .get("/api/v1/products/{slug}", { path: { slug } })
        .catch(() => null)
    : null;
  const results =
    !product && q
      ? await api.get("/api/v1/search", { query: { q, limit: 12 } })
      : null;

  return (
    <ShopPage
      shop={shop}
      section="listings"
      title="New listing"
      lead="Offer a product from the TechShop catalogue. New shops’ first 10 listings are checked by our team."
    >
      {product ? (
        <section
          aria-labelledby="offer-title"
          className="card card--roomy stack"
        >
          <h2 id="offer-title" className="text-xl">
            {product.name}
          </h2>
          <p className="text-sm">
            <Link href={here} className="link">
              Choose a different product
            </Link>
          </p>
          <NewListingForm
            shopId={shop.id}
            variants={(product.variants ?? []).map((v) => ({
              value: v.id,
              label: v.name,
            }))}
          />
        </section>
      ) : (
        <section aria-labelledby="find-title" className="stack">
          <h2 id="find-title" className="text-xl">
            Find the product
          </h2>
          <form className="cluster" role="search">
            <label htmlFor="q" className="visually-hidden">
              Product name or model
            </label>
            <input
              id="q"
              name="q"
              defaultValue={q}
              className="input"
              placeholder="e.g. Nova X5 Pro"
              maxLength={120}
            />
            <button className="btn btn--secondary">Search</button>
          </form>
          {results ? (
            (results.items ?? []).length ? (
              <ul role="list" className="stack stack--sm">
                {(results.items ?? []).map((p) => (
                  <li key={p.id}>
                    <Link
                      href={`${here}?product=${encodeURIComponent(p.slug)}`}
                      className="card card--interactive cluster cluster--between"
                    >
                      <span>
                        <strong>{p.name}</strong>
                        {p.brandName ? (
                          <span className="text-sm text-muted">
                            {" "}
                            · {p.brandName}
                          </span>
                        ) : null}
                      </span>
                      <span className="link--arrow">Choose</span>
                    </Link>
                  </li>
                ))}
              </ul>
            ) : (
              <EmptyState icon="search" title="No matching products">
                Check the spelling, or search by brand and model. Products not
                in the catalogue can’t be listed yet.
              </EmptyState>
            )
          ) : null}
        </section>
      )}
    </ShopPage>
  );
}
