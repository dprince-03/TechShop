import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { Status, humanise, type Tone } from "@/components/Status";
import { Table } from "@/components/Table";
import { Tabs } from "@/components/Tabs";
import {
  campaignStep,
  journeyStep,
  promotionStep,
} from "@/features/marketing/actions";
import { marketingData } from "@/features/marketing/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDate, formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Marketing | TechShop Staff" };

const TABS = [
  { key: "promotions", label: "Promotions" },
  { key: "coupons", label: "Coupons" },
  { key: "campaigns", label: "Campaigns" },
  { key: "segments", label: "Audiences" },
  { key: "journeys", label: "Journeys" },
];

const tone = (s: string): Tone =>
  ["active", "sent", "approved"].includes(s)
    ? "success"
    : ["draft", "scheduled", "pending_approval", "submitted"].includes(s)
      ? "warning"
      : ["cancelled", "ended", "archived"].includes(s)
        ? "neutral"
        : "info";

const promoValue = (kind: string, value: number | null) =>
  value === null
    ? humanise(kind)
    : kind === "percentage"
      ? `${value}% off`
      : kind === "fixed_amount"
        ? `${formatKobo(value)} off`
        : humanise(kind);

export default async function MarketingPage(props: PageProps<"/marketing">) {
  await connection();
  const me = await requireStaff("marketing.view");
  const wanted = param((await props.searchParams).tab);
  const tab = TABS.some((t) => t.key === wanted) ? wanted! : "promotions";
  const data = await marketingData();
  const edit = can(me, "marketing.edit");
  const approve = can(me, "marketing.approve");
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Marketing"
        description="Promotions, coupons, audiences, campaigns and automated journeys."
      />
      <Tabs
        label="Marketing sections"
        current={tab}
        items={TABS.map((t) => ({ ...t, href: `/marketing?tab=${t.key}` }))}
      />
      {tab === "promotions" ? (
        <Table
          caption="Promotions"
          rows={data.promotions}
          rowKey={(p) => p.id}
          empty={{
            title: "No promotions",
            body: "Discounts applied automatically at checkout appear here.",
          }}
          columns={[
            { label: "Promotion", cell: (p) => p.name },
            { label: "Offer", cell: (p) => promoValue(p.kind, p.value) },
            { label: "Paid for by", cell: (p) => humanise(p.fundedBy) },
            {
              label: "Runs",
              cell: (p) =>
                `${formatDate(p.startsAt)} – ${formatDate(p.endsAt)}`,
            },
            {
              label: "Status",
              cell: (p) => (
                <Status label={humanise(p.status)} tone={tone(p.status)} />
              ),
            },
            {
              label: "",
              cell: (p) =>
                edit && (p.status === "active" || p.status === "scheduled") ? (
                  <ActionButton
                    label={p.status === "active" ? "End now" : "Cancel"}
                    variant="ghost"
                    confirm={`Stop “${p.name}”?`}
                    action={promotionStep.bind(
                      null,
                      p.id,
                      p.status === "active" ? "end" : "cancel",
                    )}
                  />
                ) : null,
            },
          ]}
        />
      ) : null}
      {tab === "coupons" ? (
        <Table
          caption="Coupons"
          rows={data.coupons}
          rowKey={(c) => c.id}
          empty={{
            title: "No coupons",
            body: "Codes customers type at checkout appear here.",
          }}
          columns={[
            {
              label: "Code",
              cell: (c) => (c.isUniqueCodes ? "Unique codes" : c.code),
            },
            { label: "Promotion", cell: (c) => c.promotionName },
            { label: "Offer", cell: (c) => promoValue(c.kind, c.value) },
            {
              label: "Minimum order",
              cell: (c) =>
                c.minOrderKobo ? formatKobo(c.minOrderKobo) : "None",
            },
            {
              label: "Used",
              cell: (c) =>
                `${c.redeemedCount}${c.maxRedemptions ? ` / ${c.maxRedemptions}` : ""}`,
              numeric: true,
            },
            {
              label: "Status",
              cell: (c) => (
                <Status
                  label={humanise(c.promotionStatus)}
                  tone={tone(c.promotionStatus)}
                />
              ),
            },
          ]}
        />
      ) : null}
      {tab === "campaigns" ? (
        <>
          <Table
            caption="Campaigns"
            rows={data.campaigns}
            rowKey={(c) => c.id}
            empty={{
              title: "No campaigns",
              body: "SMS, email and push campaigns to an audience appear here.",
            }}
            columns={[
              { label: "Campaign", cell: (c) => c.name },
              {
                label: "Channels",
                cell: (c) => (c.channels ?? []).map(humanise).join(", "),
              },
              { label: "Sends", cell: (c) => formatDateTime(c.scheduledAt) },
              {
                label: "Hold-out",
                cell: (c) => `${c.holdoutPct}%`,
                numeric: true,
              },
              {
                label: "Status",
                cell: (c) => (
                  <Status label={humanise(c.status)} tone={tone(c.status)} />
                ),
              },
              {
                label: "Workflow",
                cell: (c) => (
                  <div className="cluster">
                    {edit && c.status === "draft" ? (
                      <ActionButton
                        label="Submit"
                        action={campaignStep.bind(null, c.id, "submit")}
                      />
                    ) : null}
                    {approve &&
                    ["pending_approval", "submitted"].includes(c.status) ? (
                      <ActionButton
                        label="Approve"
                        variant="inverse"
                        action={campaignStep.bind(null, c.id, "approve")}
                      />
                    ) : null}
                    {edit &&
                    ["scheduled", "approved", "sending"].includes(c.status) ? (
                      <ActionButton
                        label="Pause"
                        variant="ghost"
                        action={campaignStep.bind(null, c.id, "pause")}
                      />
                    ) : null}
                    {edit && c.status === "paused" ? (
                      <ActionButton
                        label="Resume"
                        action={campaignStep.bind(null, c.id, "resume")}
                      />
                    ) : null}
                    {edit && !["sent", "cancelled"].includes(c.status) ? (
                      <ActionButton
                        label="Cancel"
                        variant="ghost"
                        confirm={`Cancel “${c.name}”?`}
                        action={campaignStep.bind(null, c.id, "cancel")}
                      />
                    ) : null}
                  </div>
                ),
              },
            ]}
          />
          <p className="text-xs text-muted">
            The person who created a campaign can’t approve it. Approving asks
            for a recent authenticator code.
          </p>
        </>
      ) : null}
      {tab === "segments" ? (
        <Table
          caption="Audiences"
          rows={data.segments}
          rowKey={(s) => s.id}
          empty={{
            title: "No audiences",
            body: "Saved customer groups for campaigns appear here.",
          }}
          columns={[
            { label: "Audience", cell: (s) => s.name },
            {
              label: "People (last count)",
              cell: (s) => s.lastCount?.toLocaleString("en-NG") ?? "—",
              numeric: true,
            },
            { label: "Counted", cell: (s) => formatDateTime(s.lastCountedAt) },
          ]}
        />
      ) : null}
      {tab === "journeys" ? (
        <Table
          caption="Journeys"
          rows={data.journeys}
          rowKey={(j) => j.id}
          empty={{
            title: "No journeys",
            body: "Automated message sequences (e.g. abandoned cart) appear here.",
          }}
          columns={[
            { label: "Journey", cell: (j) => j.name },
            { label: "Starts on", cell: (j) => humanise(j.trigger) },
            { label: "Goal", cell: (j) => humanise(j.goalEvent || "none") },
            {
              label: "Status",
              cell: (j) => (
                <Status label={humanise(j.status)} tone={tone(j.status)} />
              ),
            },
            {
              label: "",
              cell: (j) =>
                !can(me, "marketing.journeys") ? null : j.status ===
                  "active" ? (
                  <ActionButton
                    label="Pause"
                    variant="ghost"
                    action={journeyStep.bind(null, j.id, "pause")}
                  />
                ) : j.status !== "archived" ? (
                  <ActionButton
                    label="Activate"
                    action={journeyStep.bind(null, j.id, "activate")}
                  />
                ) : null,
            },
          ]}
        />
      ) : null}
    </div>
  );
}
