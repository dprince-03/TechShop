"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });
const text = (form: FormData, name: string, max = 200) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);
const needNote = (note: string): StaffResult | null =>
  note.trim().length < 3
    ? { ok: false, message: "Write a note (it's part of the audit trail)." }
    : null;

export async function claimCase(ref: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/cases/{ref}/claim", {
        path: { ref },
      }),
    "You're reviewing this case.",
  );
}

export async function noteCase(
  ref: string,
  note: string,
): Promise<StaffResult> {
  return (
    needNote(note) ??
    runStaffAction(
      async () =>
        (await api()).post("/api/v1/staff/risk/cases/{ref}/notes", {
          path: { ref },
          body: { note: note.trim().slice(0, 1000) },
        }),
      "Note added.",
    )
  );
}

export async function escalateCase(
  ref: string,
  note: string,
): Promise<StaffResult> {
  return (
    needNote(note) ??
    runStaffAction(
      async () =>
        (await api()).post("/api/v1/staff/risk/cases/{ref}/escalate", {
          path: { ref },
          body: { note: note.trim().slice(0, 1000) },
        }),
      "Escalated to a manager.",
    )
  );
}

export async function resolveCase(
  ref: string,
  decision: "clear" | "block",
  label: "legit" | "unclear" | "",
  note: string,
): Promise<StaffResult> {
  return (
    needNote(note) ??
    runStaffAction(
      async () =>
        (await api()).post("/api/v1/staff/risk/cases/{ref}/resolution", {
          path: { ref },
          body: {
            decision,
            label: decision === "clear" ? label || "legit" : undefined,
            note: note.trim().slice(0, 1000),
          },
        }),
      decision === "clear"
        ? "Cleared. Anything held has been released."
        : "Blocked. What was held is stopped (orders are cancelled and refunded).",
    )
  );
}

const REASONS = ["reported_stolen", "fraud", "counterfeit"] as const;

export async function blockDevice(
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const imei = text(form, "imei", 20).replace(/\s+/g, "");
  const reason = REASONS.find((r) => r === text(form, "reason"));
  const source = text(form, "source", 80);
  if (!/^\d{15}$/.test(imei) || !reason || source.length < 2)
    return {
      ok: false,
      message: "Enter the 15-digit IMEI, the reason and who reported it.",
    };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/devices", {
        body: {
          imei,
          reason,
          source,
          reference: text(form, "reference", 80) || undefined,
        },
      }),
    "Blocked. It can't be sold, traded in, returned or repaired now.",
  );
}

export async function unblockDevice(
  imei: string,
  note: string,
): Promise<StaffResult> {
  return (
    needNote(note) ??
    runStaffAction(
      async () =>
        (await api()).post("/api/v1/staff/risk/devices/{imei}/unblock", {
          path: { imei },
          body: { note: note.trim().slice(0, 1000) },
        }),
      "Removed from the blocklist.",
    )
  );
}

export async function draftRules(
  checkpoint: string,
  notes: string,
): Promise<StaffResult> {
  const cp = (["checkout", "payout", "trade_in", "return"] as const).find(
    (c) => c === checkpoint,
  );
  if (!cp) return { ok: false, message: "Unknown checkpoint." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/rule-sets/drafts", {
        body: {
          checkpoint: cp,
          notes: notes.trim().slice(0, 500) || undefined,
        },
      }),
    "Draft started.",
  );
}

export async function updateRule(
  setId: string,
  ruleId: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const weight = Number.parseInt(text(form, "weight", 5), 10);
  if (
    !UUID.test(setId) ||
    !UUID.test(ruleId) ||
    !Number.isFinite(weight) ||
    weight < -100 ||
    weight > 100
  )
    return { ok: false, message: "Weight must be between −100 and 100." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/rule-sets/{id}/rules/{ruleId}", {
        path: { id: setId, ruleId },
        body: { weight, enabled: form.get("enabled") === "on" },
      }),
    "Rule saved in the draft.",
  );
}

export async function setThresholds(
  setId: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const n = (k: string) => Number.parseInt(text(form, k, 3), 10);
  const step = n("step"),
    hold = n("hold"),
    block = n("block");
  if (!(step >= 1 && step < hold && hold < block && block <= 100))
    return {
      ok: false,
      message: "Bands must rise: step up < hold < block, up to 100.",
    };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/rule-sets/{id}/thresholds", {
        path: { id: setId },
        body: { step, hold, block },
      }),
    "Bands saved in the draft.",
  );
}

export async function testRules(setId: string): Promise<StaffResult> {
  try {
    const r = await (
      await api()
    ).post("/api/v1/staff/risk/rule-sets/{id}/test", { path: { id: setId } });
    return {
      ok: true,
      message: `Replayed ${r.decisions} decisions from the last 30 days: ${r.changed} would change. Confirmed fraud caught ${r.fraudCaught}, missed ${r.fraudMissed}; cleared cases still stopped ${r.legitStopped}.`,
    };
  } catch {
    return { ok: false, message: "The test bench couldn't run. Try again." };
  }
}

export async function publishRules(setId: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/rule-sets/{id}/publish", {
        path: { id: setId },
      }),
    "Published. New decisions use these rules.",
  );
}

export async function discardRules(setId: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/risk/rule-sets/{id}/discard", {
        path: { id: setId },
      }),
    "Draft discarded.",
  );
}
