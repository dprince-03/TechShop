import { connection } from "next/server";

import { Table } from "@techshop/ui/components";
import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { revokeDevice, revokeOtherDevices } from "@/features/account/actions";
import { signOut } from "@/features/auth/actions";
import { requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Your account | TechShop Staff" };

export default async function AccountPage() {
  await connection();
  const me = await requireStaff();
  const sessions =
    (await (
      await bff.api()
    ).get("/api/v1/me/sessions", { init: { cache: "no-store" } })) ?? [];
  const groups = Object.entries(
    (me.permissions ?? []).reduce<Record<string, string[]>>((acc, p) => {
      const [area, action] = p.split(".");
      (acc[area] ??= []).push(action.replace(/_/g, " "));
      return acc;
    }, {}),
  ).sort(([a], [b]) => a.localeCompare(b));

  return (
    <div className="stack stack--lg">
      <PageHead
        title="Your account"
        description={`${me.firstName} ${me.lastName} · ${me.email ?? ""}`}
        actions={
          <form action={signOut}>
            <button className="btn btn--secondary">Sign out</button>
          </form>
        }
      />
      <section className="card stack" aria-labelledby="devices-title">
        <div className="cluster cluster--between">
          <h2 id="devices-title" className="text-xl">
            Signed-in devices
          </h2>
          {sessions.some((s) => !s.current) ? (
            <ActionButton
              label="Sign out of all other devices"
              action={revokeOtherDevices}
            />
          ) : null}
        </div>
        <Table
          caption="Signed-in devices"
          rows={sessions}
          rowKey={(s) => s.id}
          columns={[
            {
              label: "Device",
              cell: (s) =>
                `${s.deviceName || s.platform}${s.current ? " (this one)" : ""}`,
            },
            { label: "App", cell: (s) => s.audience },
            { label: "Last used", cell: (s) => formatDateTime(s.lastUsedAt) },
            {
              label: "",
              cell: (s) =>
                s.current ? null : (
                  <ActionButton
                    label="Sign out"
                    variant="ghost"
                    action={revokeDevice.bind(null, s.id)}
                  />
                ),
            },
          ]}
        />
      </section>
      <section className="card stack" aria-labelledby="perms-title">
        <h2 id="perms-title" className="text-xl">
          What you can do
        </h2>
        <p className="text-sm text-muted">
          From your roles. An administrator changes roles in the Admin console.
        </p>
        <dl className="stack stack--sm text-sm">
          {groups.map(([area, actions]) => (
            <div key={area}>
              <dt className="field__label">
                {area.charAt(0).toUpperCase() + area.slice(1)}
              </dt>
              <dd className="text-muted">{actions.join(", ")}</dd>
            </div>
          ))}
        </dl>
      </section>
    </div>
  );
}
