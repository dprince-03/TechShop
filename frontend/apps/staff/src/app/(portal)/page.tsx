import Link from "next/link";
import { connection } from "next/server";

import { Icon } from "@techshop/ui/components";

import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { can, requireStaff } from "@/features/session/server";
import { visibleGroups } from "@/features/shell/modules";
import { bff } from "@/lib/bff";
import { formatKobo } from "@/lib/format";

const today = new Intl.DateTimeFormat("en-NG", {
  weekday: "long",
  day: "numeric",
  month: "long",
  year: "numeric",
  timeZone: "Africa/Lagos",
});

/** Reads a figure only if the person may see it; a failed read shows nothing rather than breaking the page. */
async function safe<T>(
  allowed: boolean,
  read: () => Promise<T>,
): Promise<T | null> {
  if (!allowed) return null;
  try {
    return await read();
  } catch {
    return null;
  }
}

type Paged = { items?: unknown[] | null; page: { hasMore: boolean } };
const count = (page: Paged) =>
  `${page.items?.length ?? 0}${page.page.hasMore ? "+" : ""}`;

export default async function Dashboard() {
  await connection();
  const me = await requireStaff();
  const api = await bff.api();
  const noStore = { cache: "no-store" } as const;
  const [unpaid, paid, board, refunds, exceptions, finance] = await Promise.all(
    [
      safe(can(me, "orders.view"), () =>
        api.get("/api/v1/staff/orders", {
          query: { status: "pending_payment", limit: 100 },
          init: noStore,
        }),
      ),
      safe(can(me, "orders.view"), () =>
        api.get("/api/v1/staff/orders", {
          query: { status: "paid", limit: 100 },
          init: noStore,
        }),
      ),
      safe(can(me, "dispatch.view"), () =>
        api.get("/api/v1/dispatch/board", { init: noStore }),
      ),
      safe(can(me, "finance.view"), () =>
        api.get("/api/v1/staff/finance/refunds", {
          query: { status: "requested" },
          init: noStore,
        }),
      ),
      safe(can(me, "finance.view"), () =>
        api.get("/api/v1/staff/finance/reconciliation/exceptions", {
          query: { status: "open" },
          init: noStore,
        }),
      ),
      safe(can(me, "finance.view"), () =>
        api.get("/api/v1/staff/finance/dashboard", { init: noStore }),
      ),
    ],
  );

  const figures = [
    ...(unpaid ? [{ label: "Awaiting payment", value: count(unpaid) }] : []),
    ...(paid ? [{ label: "Paid, to fulfil", value: count(paid) }] : []),
    ...(board
      ? [
          {
            label: "Unassigned deliveries",
            value: String(board.counts.unassigned),
            note: `${board.counts.inTransit} in transit · ${board.counts.deliveredToday} delivered today`,
          },
        ]
      : []),
    ...(refunds
      ? [{ label: "Refunds to approve", value: String(refunds.length) }]
      : []),
    ...(exceptions
      ? [
          {
            label: "Open reconciliation issues",
            value: String(exceptions.length),
          },
        ]
      : []),
    ...(finance
      ? [{ label: "Revenue (ledger)", value: formatKobo(finance.revenueKobo) }]
      : []),
  ];
  const modules = visibleGroups(me.permissions ?? []).flatMap((g) => g.modules);

  return (
    <div className="stack stack--lg">
      <PageHead
        title={`Good day, ${me.firstName}`}
        description={today.format(new Date())}
      />
      {figures.length ? (
        <Figures items={figures} />
      ) : (
        <p className="text-sm text-muted">No figures for your role yet.</p>
      )}
      <section aria-labelledby="modules-title" className="stack">
        <h2 id="modules-title" className="text-xl">
          Your modules
        </h2>
        <ul
          role="list"
          className="grid"
          style={{ ["--grid-min" as string]: "14rem" }}
        >
          {modules.map((m) => (
            <li key={m.href}>
              <Link href={m.href} className="card card--interactive cluster">
                <Icon name={m.icon} size={22} />
                <span>{m.label}</span>
                {m.live ? null : (
                  <span className="badge badge--outline">Preview</span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
