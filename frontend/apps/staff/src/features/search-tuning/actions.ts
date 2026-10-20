"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

export type SynonymState = StaffResult;

/** "two_way": all terms mean the same; "one_way": the terms also find the target word. */
export async function createSynonym(
  _prev: SynonymState,
  form: FormData,
): Promise<SynonymState> {
  const kind = form.get("kind") === "one_way" ? "one_way" : "two_way";
  const target = String(form.get("target") ?? "")
    .trim()
    .toLowerCase()
    .slice(0, 60);
  const terms = String(form.get("terms") ?? "")
    .split(",")
    .map((t) => t.trim().toLowerCase())
    .filter(Boolean)
    .slice(0, 20);
  if (!target || !terms.length)
    return { ok: false, message: "Add a word and at least one other term." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post("/api/v1/staff/search/synonyms", {
        body: { kind, target, terms },
      }),
    "Synonym added. Searches use it right away.",
  );
}

export async function deleteSynonym(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown synonym." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).delete(
        "/api/v1/staff/search/synonyms/{id}",
        { path: { id } },
      ),
    "Synonym removed.",
  );
}
