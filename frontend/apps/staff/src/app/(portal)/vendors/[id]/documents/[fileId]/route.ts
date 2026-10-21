import { NextResponse, type NextRequest } from "next/server";

import { bff } from "@/lib/bff";

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/**
 * Opens a KYC document. The short-lived download link is requested only when a reviewer clicks,
 * so the API's audit log records real views, not page loads. The API checks the permission.
 */
export async function GET(
  req: NextRequest,
  ctx: RouteContext<"/vendors/[id]/documents/[fileId]">,
) {
  const { fileId } = await ctx.params;
  if (!UUID.test(fileId)) return new NextResponse("Not found", { status: 404 });
  if (!(await bff.getSession()))
    return NextResponse.redirect(new URL("/sign-in", req.url));
  try {
    const { url } = await (
      await bff.api()
    ).get("/api/v1/files/{id}/url", {
      path: { id: fileId },
      init: { cache: "no-store" },
    });
    return NextResponse.redirect(url, {
      headers: {
        "Cache-Control": "no-store",
        "Referrer-Policy": "no-referrer",
      },
    });
  } catch {
    return new NextResponse("This document isn’t available.", { status: 404 });
  }
}
