import "server-only";

import { createBff } from "@techshop/bff";

/** Seller Centre's BFF: session cookie, sign-in/out and the server-side API client. */
export const bff = createBff("seller", "seller");
