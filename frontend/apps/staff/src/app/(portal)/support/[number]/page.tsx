import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ApiError, ticketStatus, ticketTopic } from "@techshop/api-client";
import { Status, humanise } from "@techshop/ui/components";

import { Details } from "@/components/Details";
import { PageHead } from "@/components/PageHead";
import { can, requireStaff } from "@/features/session/server";
import {
  AssignForm,
  ClaimTicketButton,
  ReplyForm,
  StatusForm,
} from "@/features/support/components/TicketControls";
import styles from "@/features/support/components/Thread.module.css";
import { agents, ticket } from "@/features/support/server";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Ticket | TechShop Staff" };

export default async function TicketPage(
  props: PageProps<"/support/[number]">,
) {
  await connection();
  const me = await requireStaff("support.view");
  const { number } = await props.params;
  const t = await ticket(number).catch((e) => {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      notFound();
    throw e;
  });
  const canReply = can(me, "support.reply");
  const canAssign = can(me, "support.assign");
  const people = canAssign
    ? ((await agents().catch(() => [])) ?? []).map((a) => ({
        value: a.id,
        label: a.name,
      }))
    : [];
  const status = ticketStatus[t.status] ?? {
    label: t.status,
    tone: "neutral" as const,
  };
  return (
    <div className="stack stack--lg">
      <PageHead
        title={t.subject}
        description={
          <>
            {t.number} · <Status label={status.label} tone={status.tone} />
          </>
        }
        actions={
          canReply && !t.assigneeId ? (
            <ClaimTicketButton number={t.number} />
          ) : null
        }
      />
      <div className={styles.layout}>
        <div className="stack stack--lg">
          <ol className={styles.thread} aria-label="Messages">
            {(t.messages ?? []).map((m) => (
              <li
                key={m.id}
                className={`${styles.message} ${m.internal ? styles.internal : m.authorKind === "staff" ? styles.staff : ""}`}
              >
                <p className={styles.meta}>
                  {m.internal ? "Internal note · " : ""}
                  {m.author} · {formatDateTime(m.createdAt)}
                </p>
                <p className={styles.body}>{m.body}</p>
              </li>
            ))}
          </ol>
          {canReply && t.status !== "closed" ? (
            <section className="card stack" aria-label="Reply">
              <ReplyForm number={t.number} />
            </section>
          ) : null}
        </div>
        <aside className="stack">
          <section className="card stack" aria-labelledby="t-about">
            <h2 id="t-about" className="text-xl">
              About
            </h2>
            <Details
              items={[
                ["From", `${t.requester} (${humanise(t.requesterKind)})`],
                ["Contact", t.contact],
                ["Topic", ticketTopic[t.topic] ?? t.topic],
                ["Order", t.orderNumber],
                [
                  "Channel",
                  `${humanise(t.channel)}${t.site ? ` · ${t.site}` : ""}`,
                ],
                ["Priority", humanise(t.priority)],
                ["Assignee", t.assignee],
                [
                  "Reply due",
                  t.firstReplyAt
                    ? `Answered ${formatDateTime(t.firstReplyAt)}`
                    : `${formatDateTime(t.slaDueAt)}${t.overdue ? " (overdue)" : ""}`,
                ],
                ["Opened", formatDateTime(t.createdAt)],
              ]}
            />
          </section>
          {canReply && t.status !== "closed" ? (
            <section className="card stack" aria-label="Status">
              <StatusForm
                number={t.number}
                status={t.status}
                priority={t.priority}
              />
            </section>
          ) : null}
          {canAssign && people.length ? (
            <section className="card stack" aria-label="Assignment">
              <AssignForm
                number={t.number}
                agents={people}
                current={t.assigneeId ?? undefined}
              />
            </section>
          ) : null}
        </aside>
      </div>
    </div>
  );
}
