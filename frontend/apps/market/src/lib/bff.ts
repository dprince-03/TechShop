import "server-only";

import { createBff } from "@techshop/bff";

/** The market's BFF: session cookie, sign-in/out and the server-side API client. */
export const bff = createBff("market", "market");
