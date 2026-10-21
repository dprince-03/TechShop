import type { Metadata } from "next";
import { connection } from "next/server";

import { Status, humanise } from "@techshop/ui/components";

import { RowAction } from "@/components/RowAction";
import { pauseListing } from "@/features/listings/actions";
import {
  PriceForm,
  ResubmitForm,
  StockForm,
} from "@/features/listings/components/ListingForms";
import { listingStatus } from "@/features/listings/options";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime, naira } from "@/lib/format";

export const metadata: Metadata = { title: "Listing | TechShop Seller Centre" };

export default async function ListingPage(
  props: PageProps<"/centre/[shopId]/listings/[listingId]">,
) {
  await connection();
  const { shopId, listingId } = await props.params;
  const here = `/centre/${shopId}/listings/${listingId}`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const l = await api.get(
    "/api/v1/seller/shops/{shopId}/listings/{listingId}",
    { path: { shopId, listingId }, init: { cache: "no-store" } },
  );
  const status = listingStatus[l.status];
  const manage = shop.role !== "finance";
  const editable =
    manage && (l.status === "draft" || l.status === "changes_requested");
  const priced = manage && !["suspended", "rejected"].includes(l.status);
  return (
    <ShopPage
      shop={shop}
      section="listings"
      title={`${l.productName} · ${l.variantName}`}
      actions={<Status label={status.label} tone={status.tone} />}
    >
      <div className="grid" style={{ ["--grid-min" as string]: "18rem" }}>
        <section aria-labelledby="sum-title" className="card stack">
          <h2 id="sum-title" className="text-xl">
            Summary
          </h2>
          <dl className="stack stack--sm text-sm">
            <div className="cluster cluster--between">
              <dt className="text-muted">Price</dt>
              <dd className="tabular-nums">{naira(l.price)}</dd>
            </div>
            <div className="cluster cluster--between">
              <dt className="text-muted">Condition</dt>
              <dd>
                {humanise(l.condition)}
                {l.grade ? ` · grade ${l.grade}` : ""}
              </dd>
            </div>
            <div className="cluster cluster--between">
              <dt className="text-muted">In stock</dt>
              <dd className="tabular-nums">{l.sellerStock ?? "—"}</dd>
            </div>
            <div className="cluster cluster--between">
              <dt className="text-muted">Warranty</dt>
              <dd>
                {l.warrantyMonths
                  ? `${l.warrantyMonths} months (${l.warrantyProvider})`
                  : "None"}
              </dd>
            </div>
            <div className="cluster cluster--between">
              <dt className="text-muted">SKU</dt>
              <dd>{l.sku}</dd>
            </div>
          </dl>
          {manage && (l.status === "active" || l.status === "paused") ? (
            <RowAction
              action={pauseListing.bind(
                null,
                shop.id,
                l.id,
                l.status === "active",
              )}
              label={
                l.status === "active" ? "Take off sale" : "Put back on sale"
              }
              pendingLabel="Saving…"
            />
          ) : null}
        </section>
        {priced ? (
          <section aria-labelledby="price-title" className="card stack">
            <h2 id="price-title" className="text-xl">
              Price
            </h2>
            <PriceForm
              shopId={shop.id}
              listingId={l.id}
              price={l.price}
              compareAt={l.compareAt}
            />
          </section>
        ) : null}
        {priced ? (
          <section aria-labelledby="stock-title" className="card stack">
            <h2 id="stock-title" className="text-xl">
              Stock
            </h2>
            <StockForm
              shopId={shop.id}
              listingId={l.id}
              stock={l.sellerStock}
            />
          </section>
        ) : null}
      </div>
      {editable ? (
        <section aria-labelledby="fix-title" className="card card--roomy stack">
          <h2 id="fix-title" className="text-xl">
            Fix and resubmit
          </h2>
          <ResubmitForm shopId={shop.id} listingId={l.id} listing={l} />
        </section>
      ) : null}
      <section aria-labelledby="hist-title" className="stack">
        <h2 id="hist-title" className="text-xl">
          Checks and moderation
        </h2>
        {(l.reviews ?? []).length === 0 ? (
          <p className="text-sm text-muted">No checks recorded yet.</p>
        ) : (
          <ol role="list" className="stack stack--sm">
            {(l.reviews ?? []).map((r) => (
              <li
                key={r.createdAt + r.decision}
                className="card stack stack--sm"
              >
                <div className="cluster cluster--between">
                  <strong>
                    {humanise(r.decision.replace("auto_", "automatic "))}
                  </strong>
                  <span className="text-sm text-muted">
                    {dateTime.format(new Date(r.createdAt))}
                  </span>
                </div>
                {(r.reasons ?? []).length ? (
                  <ul className="text-sm">
                    {(r.reasons ?? []).map((x) => (
                      <li key={x}>{x}</li>
                    ))}
                  </ul>
                ) : null}
                {r.note ? <p className="text-sm">{r.note}</p> : null}
              </li>
            ))}
          </ol>
        )}
      </section>
    </ShopPage>
  );
}
