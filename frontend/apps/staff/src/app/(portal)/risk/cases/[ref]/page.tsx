import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ApiError } from "@techshop/api-client";
import { Status, humanise } from "@techshop/ui/components";

import { Details } from "@/components/Details";
import { PageHead } from "@/components/PageHead";
import { CaseActions } from "@/features/risk/components/RiskControls";
import { riskCase } from "@/features/risk/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo } from "@/lib/format";

export const metadata = { title: "Risk case | TechShop Staff" };

const tone = {
  open: "warning",
  in_review: "info",
  escalated: "danger",
  cleared: "success",
  blocked: "danger",
} as const;

/** Feature values as an analyst reads them (money in naira, flags as yes/no). */
function feature(name: string, v: number) {
  if (name.endsWith("_kobo")) return formatKobo(v);
  if (name.endsWith("_ok") || name.endsWith("_blocked"))
    return v ? "Yes" : "No";
  return Number.isInteger(v) ? String(v) : v.toFixed(2);
}

export default async function CasePage(props: PageProps<"/risk/cases/[ref]">) {
  await connection();
  const me = await requireStaff("risk.view");
  const { ref } = await props.params;
  const c = await riskCase(ref).catch((e) => {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      notFound();
    throw e;
  });
  return (
    <div className="stack stack--lg">
      <PageHead
        title={`Case ${c.reference}`}
        description={
          <Status
            label={humanise(c.status)}
            tone={tone[c.status as keyof typeof tone] ?? "neutral"}
          />
        }
      />
      <section className="card stack" aria-labelledby="case-sum">
        <h2 id="case-sum" className="text-xl">
          What was flagged
        </h2>
        <Details
          items={[
            [
              "Subject",
              `${c.subject || c.subjectId} (${humanise(c.subjectType)})`,
            ],
            ["Type", humanise(c.kind)],
            ["Reason", c.reason],
            ["Score", c.score ?? null],
            ["Money at risk", c.moneyAtRisk ? formatKobo(c.moneyAtRisk) : null],
            [
              "Review by",
              c.slaDueAt
                ? `${formatDateTime(c.slaDueAt)}${c.overdue ? " (overdue)" : ""}`
                : null,
            ],
            ["Analyst", c.assignee],
            ["Label", c.label ? humanise(c.label) : null],
            ["Resolution", c.resolutionNote],
          ]}
        />
        <CaseActions
          reference={c.reference}
          status={c.status}
          canResolve={can(me, "risk.resolve")}
        />
      </section>
      {c.trace ? (
        <section className="card stack" aria-labelledby="case-trace">
          <h2 id="case-trace" className="text-xl">
            Decision trace
          </h2>
          <p className="text-sm text-muted">
            {humanise(c.trace.checkpoint)} check, rules v
            {c.trace.ruleSetVersion ?? "—"}: score {c.trace.score},{" "}
            {humanise(c.trace.outcome)} · {formatDateTime(c.trace.createdAt)}
          </p>
          <ul role="list" className="stack stack--sm text-sm">
            {(c.trace.fired ?? []).map((f) => (
              <li key={f.key}>
                {f.label} ·{" "}
                {f.action === "score"
                  ? `${f.weight > 0 ? "+" : ""}${f.weight}`
                  : humanise(f.action)}
              </li>
            ))}
          </ul>
          <Details
            items={Object.entries(c.trace.features ?? {}).map(([k, v]) => [
              humanise(k),
              feature(k, v),
            ])}
          />
        </section>
      ) : null}
      <section className="card stack" aria-labelledby="case-timeline">
        <h2 id="case-timeline" className="text-xl">
          Timeline
        </h2>
        <ol className="stack stack--sm text-sm">
          {(c.events ?? []).map((e, i) => (
            <li key={i}>
              <strong>{humanise(e.action)}</strong>
              {e.actor ? ` by ${e.actor}` : ""} · {formatDateTime(e.createdAt)}
              {e.note ? <p className="text-muted">{e.note}</p> : null}
            </li>
          ))}
        </ol>
      </section>
    </div>
  );
}
