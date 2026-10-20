import { requireStaff } from "@/features/session/server";
import { StaffShell } from "@/features/shell/StaffShell";

/**
 * Every portal page needs a signed-in staff member. Pages and actions still check their own
 * permissions (and the API checks again); this only keeps signed-out visitors out.
 */
export default async function PortalLayout({ children }: LayoutProps<"/">) {
  const me = await requireStaff();
  return <StaffShell me={me}>{children}</StaffShell>;
}
