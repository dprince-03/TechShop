import { connection } from "next/server";

import { PageHead } from "@/components/PageHead";
import { Table } from "@/components/Table";
import { requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Notifications | TechShop Staff" };

export default async function NotificationsPage() {
  await connection();
  await requireStaff();
  const inbox = await (
    await bff.api()
  ).get("/api/v1/me/notifications", {
    query: { limit: 50 },
    init: { cache: "no-store" },
  });
  const items = inbox.data ?? [];
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Notifications"
        description="Messages sent to you by the system."
      />
      <Table
        caption="Notifications"
        rows={items}
        rowKey={(n) => n.id}
        empty={{
          title: "No notifications",
          body: "Alerts for your work appear here.",
        }}
        columns={[
          { label: "Message", cell: (n) => n.body ?? n.template },
          { label: "Received", cell: (n) => formatDateTime(n.createdAt) },
          { label: "Read", cell: (n) => (n.readAt ? "Yes" : "No") },
        ]}
      />
    </div>
  );
}
