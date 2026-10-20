import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function adminData(q?: string) {
  const api = await bff.api();
  const [users, roles, permissions, members] = await Promise.all([
    api.get("/api/v1/staff/admin/users", { query: { q }, init: NO_STORE }),
    api.get("/api/v1/staff/admin/roles", { init: NO_STORE }),
    api.get("/api/v1/staff/admin/permissions", { init: NO_STORE }),
    api.get("/api/v1/staff/hr/members", { init: NO_STORE }).catch(() => []),
  ]);
  return {
    users: users ?? [],
    roles: roles ?? [],
    permissions: permissions ?? [],
    members: members ?? [],
  };
}

export async function auditLog(entityType?: string) {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/admin/audit-log", {
      query: { entityType },
      init: NO_STORE,
    })) ?? []
  );
}

export async function authEvents(kind?: string) {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/admin/auth-events", {
      query: { kind },
      init: NO_STORE,
    })) ?? []
  );
}
