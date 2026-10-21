import { connection } from "next/server";

import { Table, Status, humanise } from "@techshop/ui/components";
import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { exitStaff } from "@/features/hr/actions";
import { InviteForm } from "@/features/hr/components/InviteForm";
import { can, requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";
import { formatDate } from "@/lib/format";

export const metadata = { title: "HR & staff | TechShop Staff" };

export default async function HrPage() {
  await connection();
  const me = await requireStaff("hr.view");
  const members =
    (await (
      await bff.api()
    ).get("/api/v1/staff/hr/members", { init: { cache: "no-store" } })) ?? [];
  const manage = can(me, "hr.manage");
  return (
    <div className="stack stack--lg">
      <PageHead
        title="HR & staff"
        description="Staff records. Exiting someone revokes all their access at once."
      />
      <Table
        caption="Staff members"
        rows={members}
        rowKey={(m) => m.userId}
        columns={[
          { label: "Name", cell: (m) => `${m.firstName} ${m.lastName}` },
          { label: "Employee no.", cell: (m) => m.employeeNo },
          { label: "Job", cell: (m) => m.jobTitle },
          { label: "Department", cell: (m) => m.department },
          { label: "Hired", cell: (m) => formatDate(m.hiredOn) },
          {
            label: "Status",
            cell: (m) => (
              <Status
                label={humanise(m.status)}
                tone={
                  m.status === "active"
                    ? "success"
                    : m.status === "exited"
                      ? "neutral"
                      : "warning"
                }
              />
            ),
          },
          {
            label: "",
            cell: (m) =>
              manage && m.status !== "exited" && m.userId !== me.id ? (
                <ActionButton
                  label="Exit"
                  variant="ghost"
                  confirm={`Exit ${m.firstName} ${m.lastName}? Their sessions end and every role is removed.`}
                  action={exitStaff.bind(null, m.userId)}
                />
              ) : null,
          },
        ]}
      />
      {manage ? (
        <section
          className="card card--roomy stack"
          aria-labelledby="invite-title"
        >
          <h2 id="invite-title" className="text-xl">
            Add a staff member
          </h2>
          <InviteForm />
        </section>
      ) : null}
    </div>
  );
}
