import Link from "next/link";
import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { Status, humanise, type Tone } from "@/components/Status";
import { Table } from "@/components/Table";
import { Tabs } from "@/components/Tabs";
import {
  decideDevice,
  setRiderStatus,
  unassignJob,
} from "@/features/dispatch/actions";
import { AssignJob } from "@/features/dispatch/components/AssignJob";
import { ReturnJob } from "@/features/dispatch/components/ReturnJob";
import {
  board,
  JOB_STATUSES,
  pendingDevices,
  riders,
  zoneRates,
  zones,
  type JobStatus,
} from "@/features/dispatch/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo, param } from "@/lib/format";

export const metadata = { title: "Dispatch | TechShop Staff" };

const TABS = [
  { key: "board", label: "Board" },
  { key: "riders", label: "Riders" },
  { key: "phones", label: "Phone approvals" },
  { key: "zones", label: "Zones & rates" },
];

const jobTone: Record<string, Tone> = {
  unassigned: "warning",
  assigned: "info",
  picked_up: "info",
  en_route: "info",
  delivered: "success",
  failed: "danger",
  returned: "neutral",
  cancelled: "neutral",
};

export default async function DispatchPage(props: PageProps<"/dispatch">) {
  await connection();
  const me = await requireStaff("dispatch.view");
  const sp = await props.searchParams;
  const tab = TABS.some((t) => t.key === param(sp.tab))
    ? param(sp.tab)!
    : "board";
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Dispatch"
        description="Own-rider deliveries: assign jobs, manage riders and their phones, set zones and fees."
      />
      <Tabs
        label="Dispatch sections"
        current={tab}
        items={TABS.map((t) => ({ ...t, href: `/dispatch?tab=${t.key}` }))}
      />
      {tab === "board" ? (
        <Board
          status={JOB_STATUSES.find((s) => s === param(sp.status))}
          canAssign={can(me, "dispatch.assign")}
        />
      ) : null}
      {tab === "riders" ? (
        <Riders canManage={can(me, "logistics.manage")} />
      ) : null}
      {tab === "phones" ? (
        <Phones canDecide={can(me, "riders.manage_devices")} />
      ) : null}
      {tab === "zones" ? <Zones /> : null}
    </div>
  );
}

async function Board({
  status,
  canAssign,
}: {
  status?: JobStatus;
  canAssign: boolean;
}) {
  const [b, rs] = await Promise.all([board(status), riders()]);
  const riderOptions = rs
    .filter((r) => r.status === "active")
    .map((r) => ({
      id: r.userId,
      label: `${r.firstName} ${r.lastName} · ${r.activeJobs} jobs`,
      onShift: Boolean(r.openShiftId),
    }));
  return (
    <>
      <Figures
        items={[
          { label: "Unassigned", value: String(b.counts.unassigned) },
          { label: "Assigned", value: String(b.counts.assigned) },
          { label: "In transit", value: String(b.counts.inTransit) },
          { label: "Delivered today", value: String(b.counts.deliveredToday) },
          { label: "Failed", value: String(b.counts.failed) },
        ]}
      />
      <div className="cluster">
        <Link
          href="/dispatch"
          className={`btn btn--sm ${status ? "btn--secondary" : "btn--inverse"}`}
        >
          All open
        </Link>
        {JOB_STATUSES.map((s) => (
          <Link
            key={s}
            href={`/dispatch?status=${s}`}
            className={`btn btn--sm ${s === status ? "btn--inverse" : "btn--secondary"}`}
          >
            {humanise(s)}
          </Link>
        ))}
      </div>
      <Table
        caption="Delivery jobs"
        rows={b.jobs ?? []}
        rowKey={(j) => j.id}
        empty={{
          title: "No jobs here",
          body: "Packed parcels for own-rider zones appear as jobs.",
        }}
        columns={[
          { label: "Order", cell: (j) => j.orderNumber ?? "—" },
          { label: "Kind", cell: (j) => humanise(j.kind) },
          { label: "Recipient", cell: (j) => j.recipientName },
          { label: "Zone", cell: (j) => j.zoneName },
          {
            label: "Window",
            cell: (j) =>
              j.windowStart ? `${formatDateTime(j.windowStart)}` : "—",
          },
          { label: "Rider", cell: (j) => j.riderFirstName ?? "—" },
          {
            label: "Status",
            cell: (j) => (
              <Status label={humanise(j.status)} tone={jobTone[j.status]} />
            ),
          },
          {
            label: "Actions",
            cell: (j) =>
              !canAssign ? null : j.status === "unassigned" ? (
                <AssignJob jobId={j.id} riders={riderOptions} />
              ) : j.status === "assigned" ? (
                <ActionButton
                  label="Unassign"
                  variant="ghost"
                  action={unassignJob.bind(null, j.id)}
                />
              ) : j.status === "failed" ? (
                <ReturnJob jobId={j.id} />
              ) : null,
          },
        ]}
      />
    </>
  );
}

async function Riders({ canManage }: { canManage: boolean }) {
  const rs = await riders();
  return (
    <Table
      caption="Riders"
      rows={rs}
      rowKey={(r) => r.userId}
      empty={{
        title: "No riders",
        body: "Riders are staff members added by HR, then registered here.",
      }}
      columns={[
        { label: "Rider", cell: (r) => `${r.firstName} ${r.lastName}` },
        {
          label: "Vehicle",
          cell: (r) =>
            `${humanise(r.vehicleType)}${r.plateNumber ? ` · ${r.plateNumber}` : ""}`,
        },
        {
          label: "On shift",
          cell: (r) => (
            <Status
              label={r.openShiftId ? "On shift" : "Off shift"}
              tone={r.openShiftId ? "success" : "neutral"}
            />
          ),
        },
        { label: "Active jobs", cell: (r) => r.activeJobs, numeric: true },
        { label: "Last seen", cell: (r) => formatDateTime(r.lastSeenAt) },
        { label: "Status", cell: (r) => humanise(r.status) },
        {
          label: "",
          cell: (r) =>
            !canManage ? null : r.status === "active" ? (
              <ActionButton
                label="Suspend"
                variant="ghost"
                confirm={`Suspend ${r.firstName}? They can’t take jobs until reinstated.`}
                action={setRiderStatus.bind(null, r.userId, "suspend")}
              />
            ) : (
              <ActionButton
                label="Reinstate"
                action={setRiderStatus.bind(null, r.userId, "reinstate")}
              />
            ),
        },
      ]}
    />
  );
}

async function Phones({ canDecide }: { canDecide: boolean }) {
  const ds = await pendingDevices();
  return (
    <Table
      caption="Phones waiting for approval"
      rows={ds}
      rowKey={(d) => d.id}
      empty={{
        title: "No phones waiting",
        body: "A rider’s first sign-in on a phone appears here for approval.",
      }}
      columns={[
        { label: "Rider", cell: (d) => `${d.firstName} ${d.lastName}` },
        {
          label: "Phone",
          cell: (d) => `${humanise(d.platform)} · ${d.deviceId.slice(0, 12)}…`,
        },
        { label: "Registered", cell: (d) => formatDateTime(d.createdAt) },
        {
          label: "Decision",
          cell: (d) =>
            canDecide ? (
              <div className="cluster">
                <ActionButton
                  label="Approve"
                  variant="inverse"
                  action={decideDevice.bind(null, d.id, "approve")}
                />
                <ActionButton
                  label="Revoke"
                  variant="ghost"
                  action={decideDevice.bind(null, d.id, "revoke")}
                />
              </div>
            ) : null,
        },
      ]}
    />
  );
}

async function Zones() {
  const zs = await zones();
  const rates = await Promise.all(
    zs.map((z) => zoneRates(z.id).catch(() => [])),
  );
  return (
    <div className="stack">
      {zs.map((z, i) => (
        <section
          key={z.id}
          className="card stack"
          aria-label={`Zone ${z.name}`}
        >
          <div className="cluster cluster--between">
            <h2 className="text-xl">
              {z.name} · {z.stateCode}
            </h2>
            <Status
              label={z.isActive ? "Active" : "Paused"}
              tone={z.isActive ? "success" : "neutral"}
            />
          </div>
          <p className="text-sm text-muted">
            {(z.lgas ?? []).length
              ? `LGAs: ${z.lgas!.join(", ")}`
              : "Whole state"}
          </p>
          <Table
            caption={`Rates for ${z.name}`}
            rows={rates[i]}
            rowKey={(r) => r.id}
            empty={{
              title: "No rates",
              body: "Add weight bands to price deliveries in this zone.",
            }}
            columns={[
              {
                label: "Up to",
                cell: (r) =>
                  `${(r.maxWeightGrams / 1000).toLocaleString("en-NG")} kg`,
              },
              { label: "Shipped by", cell: (r) => humanise(r.fulfilledBy) },
              {
                label: "Time",
                cell: (r) => `${r.etaMinDays}–${r.etaMaxDays} days`,
              },
              {
                label: "Fee",
                cell: (r) => formatKobo(r.feeKobo),
                numeric: true,
              },
            ]}
          />
        </section>
      ))}
    </div>
  );
}
