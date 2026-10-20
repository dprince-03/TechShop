import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { Status, humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { Tabs } from "@/components/Tabs";
import { resetMfa } from "@/features/admin/actions";
import { RoleEditor } from "@/features/admin/components/RoleEditor";
import { adminData, auditLog, authEvents } from "@/features/admin/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Admin console | TechShop Staff" };

const TABS = [
  { key: "staff", label: "Staff & roles" },
  { key: "roles", label: "Roles" },
  { key: "people", label: "All people" },
  { key: "audit", label: "Audit log" },
  { key: "signins", label: "Sign-in events" },
];

export default async function AdminPage(props: PageProps<"/admin">) {
  await connection();
  const me = await requireStaff("admin.users");
  const sp = await props.searchParams;
  const wanted = param(sp.tab);
  const tab = TABS.some((t) => t.key === wanted) ? wanted! : "staff";
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Admin console"
        description="Who can do what, and a record of what was done."
      />
      <Tabs
        label="Admin sections"
        current={tab}
        items={TABS.map((t) => ({ ...t, href: `/admin?tab=${t.key}` }))}
      />
      {tab === "audit" ? (
        <Audit />
      ) : tab === "signins" ? (
        <SignIns />
      ) : (
        <Directory
          tab={tab}
          q={param(sp.q)}
          canRoles={can(me, "admin.roles")}
          canMfa={can(me, "admin.mfa_reset")}
          myId={me.id}
        />
      )}
    </div>
  );
}

async function Directory({
  tab,
  q,
  canRoles,
  canMfa,
  myId,
}: {
  tab: string;
  q?: string;
  canRoles: boolean;
  canMfa: boolean;
  myId: string;
}) {
  const { users, roles, permissions, members } = await adminData(
    q?.slice(0, 80),
  );
  const roleList = roles.map((r) => ({ key: r.key, name: r.name }));
  if (tab === "roles") {
    return (
      <Table
        caption="Roles"
        rows={roles}
        rowKey={(r) => r.id}
        columns={[
          { label: "Role", cell: (r) => r.name },
          {
            label: "Key",
            cell: (r) => <code className="text-xs">{r.key}</code>,
          },
          {
            label: "Permissions",
            cell: (r) => (
              <span className="text-xs">
                {(r.permissionKeys ?? []).join(", ")}
              </span>
            ),
          },
          { label: "Type", cell: (r) => (r.isSystem ? "System" : "Custom") },
        ]}
      />
    );
  }
  if (tab === "people") {
    return (
      <section className="stack">
        <form method="get" className="cluster" role="search">
          <input type="hidden" name="tab" value="people" />
          <label className="field" style={{ flex: "1 1 18rem" }}>
            <span className="field__label">Find a person</span>
            <input
              name="q"
              className="input"
              defaultValue={q}
              placeholder="Name, email or phone"
            />
          </label>
          <button
            className="btn btn--secondary"
            style={{ alignSelf: "flex-end" }}
          >
            Search
          </button>
        </form>
        <Table
          caption="People"
          rows={users}
          rowKey={(u) => u.id}
          columns={[
            { label: "Name", cell: (u) => u.name },
            { label: "Email", cell: (u) => u.email ?? "—" },
            { label: "Status", cell: (u) => humanise(u.status) },
            { label: "Joined", cell: (u) => formatDateTime(u.createdAt) },
          ]}
        />
        <p className="text-xs text-muted">
          {permissions.filter((p) => p.isSensitive).length} sensitive
          permissions are marked in the role catalogue.
        </p>
      </section>
    );
  }
  return (
    <Table
      caption="Staff and their roles"
      rows={members.filter((m) => m.status !== "exited")}
      rowKey={(m) => m.userId}
      columns={[
        { label: "Name", cell: (m) => `${m.firstName} ${m.lastName}` },
        { label: "Job", cell: (m) => `${m.jobTitle} · ${m.department}` },
        {
          label: "Authenticator",
          cell: (m) => (
            <Status
              label={m.mfaEnabled ? "Set up" : "Not set up"}
              tone={m.mfaEnabled ? "success" : "warning"}
            />
          ),
        },
        {
          label: "Roles",
          cell: (m) =>
            canRoles && m.userId !== myId ? (
              <RoleEditor
                userId={m.userId}
                current={m.roleKeys ?? []}
                roles={roleList}
              />
            ) : (
              (m.roleKeys ?? []).join(", ")
            ),
        },
        {
          label: "",
          cell: (m) =>
            canMfa && m.mfaEnabled && m.userId !== myId ? (
              <ActionButton
                label="Reset authenticator"
                variant="ghost"
                confirm={`Reset ${m.firstName}’s authenticator? They must set it up again.`}
                action={resetMfa.bind(null, m.userId)}
              />
            ) : null,
        },
      ]}
    />
  );
}

async function Audit() {
  const rows = await auditLog();
  return (
    <Table
      caption="Audit log"
      rows={rows}
      rowKey={(r) => String(r.id)}
      empty={{
        title: "Nothing recorded yet",
        body: "Changes made by staff and the system appear here.",
      }}
      columns={[
        { label: "When", cell: (r) => formatDateTime(r.createdAt) },
        { label: "Action", cell: (r) => humanise(r.action) },
        {
          label: "On",
          cell: (r) => `${humanise(r.entityType)} ${r.entityId.slice(0, 8)}`,
        },
        {
          label: "By",
          cell: (r) =>
            r.actorUserId ? r.actorUserId.slice(0, 8) : humanise(r.actorKind),
        },
        {
          label: "Request",
          cell: (r) => (
            <code className="text-xs">{r.requestId?.slice(0, 12) ?? ""}</code>
          ),
        },
      ]}
    />
  );
}

async function SignIns() {
  const rows = await authEvents();
  return (
    <Table
      caption="Sign-in events"
      rows={rows}
      rowKey={(r) => String(r.id)}
      empty={{
        title: "No events",
        body: "Sign-ins, failures and session changes appear here.",
      }}
      columns={[
        { label: "When", cell: (r) => formatDateTime(r.createdAt) },
        { label: "Event", cell: (r) => humanise(r.kind) },
        {
          label: "Person",
          cell: (r) => (r.userId ? r.userId.slice(0, 8) : "—"),
        },
        { label: "Network", cell: (r) => r.ip ?? "—" },
      ]}
    />
  );
}
