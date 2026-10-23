import { connection } from "next/server";

import { Status, humanise } from "@techshop/ui/components";

import { PageHead } from "@/components/PageHead";
import {
  DraftActions,
  RuleForm,
  StartDraft,
  ThresholdsForm,
} from "@/features/risk/components/RiskControls";
import { RiskNav } from "@/features/risk/components/RiskNav";
import { ruleSets } from "@/features/risk/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Risk rules | TechShop Staff" };

const CHECKPOINTS = ["checkout", "payout", "trade_in", "return"];

function condition(c: { all?: { f: string; op: string; v: number }[] | null }) {
  const ops: Record<string, string> = {
    lt: "<",
    lte: "≤",
    gt: ">",
    gte: "≥",
    eq: "=",
  };
  return (c.all ?? [])
    .map((x) => `${humanise(x.f)} ${ops[x.op] ?? x.op} ${x.v}`)
    .join(" and ");
}

export default async function RulesPage() {
  await connection();
  const me = await requireStaff("risk.view");
  const sets = (await ruleSets()) ?? [];
  const canEdit = can(me, "risk.rules");
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Risk rules"
        description="Each check adds up rule weights into a score from 0 to 100; the bands decide allow, step up (held for review at checkout), hold or block. Changes go through a draft, the test bench and a publish with your authenticator."
      />
      <RiskNav current="rules" />
      {CHECKPOINTS.map((cp) => {
        const active = sets.find(
          (s) => s.checkpoint === cp && s.status === "active",
        );
        const draft = sets.find(
          (s) => s.checkpoint === cp && s.status === "draft",
        );
        return (
          <section key={cp} className="card stack" aria-labelledby={`cp-${cp}`}>
            <h2 id={`cp-${cp}`} className="text-xl">
              {humanise(cp)}
            </h2>
            {active ? (
              <>
                <p className="text-sm text-muted">
                  Live: version {active.version}
                  {active.publishedAt
                    ? `, published ${formatDateTime(active.publishedAt)}`
                    : ""}{" "}
                  · step up {active.thresholds.step}, hold{" "}
                  {active.thresholds.hold}, block {active.thresholds.block}
                </p>
                <ul role="list" className="stack stack--sm text-sm">
                  {(active.rules ?? []).map((r) => (
                    <li key={r.id} className="cluster cluster--between">
                      <span>
                        {r.label}{" "}
                        <span className="text-muted">
                          ({condition(r.condition)})
                        </span>
                      </span>
                      <Status
                        label={
                          !r.enabled
                            ? "Off"
                            : r.action === "score"
                              ? `${r.weight > 0 ? "+" : ""}${r.weight}`
                              : humanise(r.action)
                        }
                        tone={
                          !r.enabled
                            ? "neutral"
                            : r.action === "block"
                              ? "danger"
                              : r.action === "hold"
                                ? "warning"
                                : "info"
                        }
                      />
                    </li>
                  ))}
                </ul>
              </>
            ) : (
              <p className="text-muted">
                No live rules: this check falls back by amount.
              </p>
            )}
            {canEdit && !draft && active ? (
              <StartDraft checkpoint={cp} />
            ) : null}
            {draft ? (
              <div className="stack">
                <h3 className="text-lg">
                  Draft version {draft.version}
                  {draft.notes ? ` · ${draft.notes}` : ""}
                </h3>
                {canEdit ? (
                  <>
                    <ThresholdsForm
                      setId={draft.id}
                      step={draft.thresholds.step}
                      hold={draft.thresholds.hold}
                      block={draft.thresholds.block}
                    />
                    {(draft.rules ?? [])
                      .filter((r) => r.action === "score")
                      .map((r) => (
                        <RuleForm
                          key={r.id}
                          setId={draft.id}
                          ruleId={r.id}
                          weight={r.weight}
                          enabled={r.enabled}
                          label={r.label}
                        />
                      ))}
                    <DraftActions setId={draft.id} />
                  </>
                ) : (
                  <p className="text-sm text-muted">
                    A rules manager is editing a draft.
                  </p>
                )}
              </div>
            ) : null}
          </section>
        );
      })}
    </div>
  );
}
