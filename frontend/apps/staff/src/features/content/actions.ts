"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

import { textToBlocks } from "./blocks";

type Site = "corporate" | "market" | "wholesale" | "seller";
const SITES: Site[] = ["corporate", "market", "wholesale", "seller"];
const SLUG = /^[a-z0-9][a-z0-9-]{0,79}$/;

export type EditorState = StaffResult;

export async function savePage(
  _prev: EditorState,
  form: FormData,
): Promise<EditorState> {
  const site = SITES.find((s) => s === form.get("site"));
  const slug = String(form.get("slug") ?? "")
    .trim()
    .toLowerCase();
  const title = String(form.get("title") ?? "")
    .trim()
    .slice(0, 150);
  const blocks = textToBlocks(String(form.get("body") ?? ""));
  if (!site || !SLUG.test(slug))
    return {
      ok: false,
      message:
        "Choose a site and a web address (lowercase letters, numbers, dashes).",
    };
  if (title.length < 2 || !blocks.length)
    return { ok: false, message: "Add a title and some text." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).put(
        "/api/v1/staff/content/pages/{site}/{slug}",
        { path: { site, slug }, body: { title, blocks } },
      ),
    "Saved as a draft. Publish it to put it live.",
  );
}

export async function publishPage(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown page." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/content/pages/{id}/publish",
        { path: { id } },
      ),
    "Published. It’s live on the site.",
  );
}

export async function saveHelp(
  _prev: EditorState,
  form: FormData,
): Promise<EditorState> {
  const site = (["market", "wholesale", "seller"] as const).find(
    (s) => s === form.get("site"),
  );
  const slug = String(form.get("slug") ?? "")
    .trim()
    .toLowerCase();
  const title = String(form.get("title") ?? "")
    .trim()
    .slice(0, 150);
  const topic = String(form.get("topic") ?? "")
    .trim()
    .slice(0, 60);
  const status =
    (["draft", "in_review", "published"] as const).find(
      (s) => s === form.get("status"),
    ) ?? "draft";
  const blocks = textToBlocks(String(form.get("body") ?? ""));
  if (!site || !SLUG.test(slug))
    return { ok: false, message: "Choose a site and a web address." };
  if (title.length < 2 || !topic || !blocks.length)
    return { ok: false, message: "Add a title, topic and some text." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).put(
        "/api/v1/staff/content/help/{site}/{slug}",
        { path: { site, slug }, body: { title, topic, status, blocks } },
      ),
    status === "published" ? "Article published." : "Article saved.",
  );
}

export async function createBanner(
  _prev: EditorState,
  form: FormData,
): Promise<EditorState> {
  const site = SITES.find((s) => s === form.get("site"));
  const title = String(form.get("title") ?? "")
    .trim()
    .slice(0, 120);
  const linkUrl = String(form.get("linkUrl") ?? "").trim();
  const placement = String(form.get("placement") ?? "home_hero")
    .trim()
    .slice(0, 40);
  const status = form.get("status") === "live" ? "live" : "draft";
  if (!site || title.length < 2)
    return { ok: false, message: "Choose a site and add a title." };
  if (
    !(linkUrl.startsWith("/") && !linkUrl.startsWith("//")) &&
    !linkUrl.startsWith("https://")
  )
    return {
      ok: false,
      message:
        "The link must be a path on the site (/deals) or an https:// address.",
    };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post("/api/v1/staff/content/banners", {
        body: { site, title, linkUrl, placement, status },
      }),
    "Banner saved.",
  );
}
