import "server-only";

import { createBff } from "@techshop/bff";

/** The staff portal's BFF: session cookie (12 h staff sessions), sign-in/out, API client. */
export const bff = createBff("staff", "staff");
