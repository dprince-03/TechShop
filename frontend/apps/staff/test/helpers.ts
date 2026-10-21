// Shared by the staff e2e test files: a browser-like client and the real two-step sign-in.
import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { account, form, totp, WebClient } from "@techshop/testing";

export const APP_DIR = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const BASE = process.env.STAFF_URL ?? "http://localhost:4105";
export const browser = () => new WebClient(BASE, APP_DIR);

// Each TOTP code works once; remember the last step used per account across tests.
const usedStep = new Map<string, number>();
async function freshCode(secret: string) {
  for (;;) {
    const step = Math.floor(Date.now() / 30_000);
    if (usedStep.get(secret) !== step) {
      usedStep.set(secret, step);
      return totp(secret);
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
}

/** Signs a seeded staff member in through the real two-step sign-in action. */
export async function signIn(role: string, next = "/") {
  const b = browser();
  const a = account(role);
  await b.get(`/sign-in?next=${encodeURIComponent(next)}`);
  const step1 = await b.action<{ step: string }>("/sign-in", "staffSignIn", [
    { step: "start" },
    form({ intent: "password", email: a.email, password: a.password, next }),
  ]);
  assert.equal(step1.value.step, "totp", JSON.stringify(step1.value));
  let step2 = await b.action<{ step: string; error?: string }>(
    "/sign-in",
    "staffSignIn",
    [
      { step: "totp" },
      form({ intent: "totp", code: await freshCode(a.totpSecret!), next }),
    ],
  );
  if (!step2.redirect && /already used/.test(step2.value?.error ?? "")) {
    // An earlier run used this 30-second code; take the next one.
    usedStep.set(a.totpSecret!, Math.floor(Date.now() / 30_000));
    step2 = await b.action("/sign-in", "staffSignIn", [
      { step: "totp" },
      form({ intent: "totp", code: await freshCode(a.totpSecret!), next }),
    ]);
  }
  assert.equal(step2.redirect, next, JSON.stringify(step2.value));
  return b;
}
