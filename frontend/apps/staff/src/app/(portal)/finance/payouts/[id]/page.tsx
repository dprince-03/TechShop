import { Status, Table, humanise } from "@techshop/ui/components";
import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import {
  approveBatch,
  cancelBatch,
  submitBatch,
} from "@/features/payouts/actions";
import { ExcludePayout } from "@/features/payouts/components/ExcludePayout";
import { getBatch } from "@/features/payouts/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDate, formatDateTime, formatKobo } from "@/lib/format";

export const metadata = { title: "Payout batch | TechShop Staff" };

const payoutTone = {
  scheduled: "neutral",
  processing: "info",
  paid: "success",
  failed: "danger",
  on_hold: "warning",
} as const;

export default async function PayoutBatchPage(
  props: PageProps<"/finance/payouts/[id]">,
) {
  await connection();
  const me = await requireStaff("finance.view");
  const { id } = await props.params;
  const { batch: b, payouts } = await getBatch(id);
  const draft = b.status === "draft";
  const prepare = can(me, "payouts.prepare");
  const approve = can(me, "payouts.approve");
  const awaitingSecond = b.status === "submitted" && b.firstApprovedBy;

  return (
    <div className="stack stack--lg">
      <PageHead
        title="Payout batch"
        description={`Prepared ${formatDateTime(b.createdAt)}. Whoever prepared it can’t approve it${b.needsTwoApprovers ? "; above the threshold, two managers approve" : ""}.`}
        actions={
          <div className="cluster">
            {draft && prepare && b.count > 0 ? (
              <ActionButton
                label="Submit for approval"
                variant="inverse"
                action={submitBatch.bind(null, b.id)}
              />
            ) : null}
            {b.status === "submitted" && approve ? (
              <ActionButton
                label={awaitingSecond ? "Give the second approval" : "Approve"}
                variant="inverse"
                confirm={`Approve paying ${formatKobo(b.total)} to ${b.count} seller${b.count === 1 ? "" : "s"}? The transfers start straight away.`}
                action={approveBatch.bind(null, b.id)}
              />
            ) : null}
            {(draft || b.status === "submitted") && prepare ? (
              <ActionButton
                label="Cancel batch"
                variant="ghost"
                confirm="Cancel this batch? Its earnings go into the next run."
                action={cancelBatch.bind(null, b.id)}
              />
            ) : null}
          </div>
        }
      />
      <Figures
        items={[
          { label: "Status", value: humanise(b.status) },
          { label: "Total", value: formatKobo(b.total) },
          { label: "Payouts", value: String(b.count) },
          {
            label: "Approvals",
            value: b.approvedBy
              ? "Approved"
              : awaitingSecond
                ? "1 of 2"
                : "None yet",
            note: b.approvedAt ? formatDateTime(b.approvedAt) : undefined,
          },
        ]}
      />
      <Table
        caption="Payouts in this batch"
        rows={payouts ?? []}
        rowKey={(p) => p.id}
        empty={{
          title: "No payouts",
          body: "No seller could be paid when this batch was prepared.",
        }}
        columns={[
          { label: "Shop", cell: (p) => p.shop },
          { label: "Amount", numeric: true, cell: (p) => formatKobo(p.amount) },
          {
            label: "Account",
            cell: (p) =>
              `${p.accountName} · ${p.bankCode} ••••${p.accountLast4}`,
          },
          { label: "For", cell: (p) => formatDate(p.scheduledFor) },
          {
            label: "Status",
            cell: (p) => (
              <span className="stack">
                <Status
                  label={humanise(p.status)}
                  tone={payoutTone[p.status]}
                />
                {p.failureReason ? (
                  <span className="text-xs text-muted">{p.failureReason}</span>
                ) : null}
              </span>
            ),
          },
          ...(draft && prepare
            ? [
                {
                  label: "Leave out",
                  cell: (p: NonNullable<typeof payouts>[number]) => (
                    <ExcludePayout batchId={b.id} payoutId={p.id} />
                  ),
                },
              ]
            : []),
        ]}
      />
    </div>
  );
}
