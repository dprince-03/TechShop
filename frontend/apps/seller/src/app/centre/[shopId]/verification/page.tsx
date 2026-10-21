import type { Metadata } from "next";
import { connection } from "next/server";

import { Status, humanise } from "@techshop/ui/components";

import { KycForm } from "@/features/kyc/components/KycForm";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";

export const metadata: Metadata = {
  title: "Verification | TechShop Seller Centre",
};

const tone = {
  submitted: "info",
  in_review: "info",
  approved: "success",
  rejected: "danger",
  needs_more_info: "warning",
} as const;

export default async function VerificationPage(
  props: PageProps<"/centre/[shopId]/verification">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}/verification`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const { kyc } = await api.get("/api/v1/seller/shops/{shopId}/onboarding", {
    path: { shopId },
    init: { cache: "no-store" },
  });
  const canSubmit =
    !kyc || kyc.status === "rejected" || kyc.status === "needs_more_info";
  return (
    <ShopPage
      shop={shop}
      section="verification"
      title="Verification"
      lead="We check every seller’s identity before they can sell (Nigeria Data Protection Act: your ID is encrypted and access is logged)."
    >
      {kyc ? (
        <section aria-labelledby="kyc-status" className="card stack">
          <h2 id="kyc-status" className="text-xl">
            Your latest submission
          </h2>
          <dl className="stack stack--sm">
            <div className="cluster">
              <dt className="text-muted">Status</dt>
              <dd>
                <Status label={humanise(kyc.status)} tone={tone[kyc.status]} />
              </dd>
            </div>
            <div className="cluster">
              <dt className="text-muted">ID</dt>
              <dd>
                {humanise(kyc.idType)} ending {kyc.idNumberLast4}
              </dd>
            </div>
            <div className="cluster">
              <dt className="text-muted">Documents</dt>
              <dd>
                {(kyc.documents ?? []).map((d) => humanise(d.kind)).join(", ")}
              </dd>
            </div>
          </dl>
          {kyc.note ? (
            <p className="alert alert--warning">
              <strong>Reviewer’s note:</strong> {kyc.note}
            </p>
          ) : null}
        </section>
      ) : null}
      {canSubmit ? (
        <section aria-labelledby="kyc-form" className="card card--roomy stack">
          <h2 id="kyc-form" className="text-xl">
            {kyc ? "Submit again" : "Verify your identity"}
          </h2>
          <KycForm shopId={shop.id} business={shop.type === "business"} />
        </section>
      ) : null}
    </ShopPage>
  );
}
