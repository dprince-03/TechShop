import "server-only";

import { createBff } from "@techshop/bff";

/** The wholesale site's BFF (business sign-in arrives with B2B, backend M6). */
export const bff = createBff("wholesale", "wholesale");
