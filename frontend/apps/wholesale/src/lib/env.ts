import "server-only";

import { z } from "zod";

const envSchema = z.object({
  NODE_ENV: z
    .enum(["development", "test", "production"])
    .default("development"),
  API_URL: z.url(),
});

export const env = envSchema.parse(process.env);
