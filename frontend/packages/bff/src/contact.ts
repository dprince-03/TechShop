import "server-only";

import { headers } from "next/headers";

import { createApiClient } from "@techshop/api-client";

import { forwardedFor } from "./client-ip";
import { failure, type ActionState } from "./forms";

export type ContactSite = "corporate" | "market" | "wholesale" | "seller";

const TOPICS = [
  "order",
  "delivery",
  "payment",
  "refund",
  "return",
  "warranty",
  "trade_in",
  "account",
  "selling",
  "business",
  "other",
] as const;
type Topic = (typeof TOPICS)[number];

const text = (form: FormData, name: string, max: number) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);

/**
 * The public contact form for any site: checks the fields, then files a support ticket (the
 * API confirms it to the sender and rate-limits per IP and per address). "website" is a hidden
 * trap field that only bots fill in.
 */
export async function submitContact(
  site: ContactSite,
  form: FormData,
  env: Record<string, string | undefined> = process.env,
): Promise<ActionState> {
  const name = text(form, "name", 80);
  const email = text(form, "email", 254);
  const phone = text(form, "phone", 20);
  const topic = text(form, "topic", 20) as Topic;
  const subject = text(form, "subject", 140);
  const message = text(form, "message", 5000);
  const errors: Record<string, string> = {};
  if (name.length < 2) errors.name = "Tell us your name.";
  if (!email && !phone)
    errors.email = "Give us an email address or phone number so we can reply.";
  if (!TOPICS.includes(topic)) errors.topic = "Choose what it's about.";
  if (subject.length < 3) errors.subject = "Give it a short subject.";
  if (!message) errors.message = "Tell us what's happening.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  try {
    // No session needed (the corporate site has none), so only API_URL and the proxy's client-IP
    // header are read: the API's per-IP limit must see the visitor, not this server.
    if (!env.API_URL) throw new Error("API_URL is required");
    const ipHeader = env.CLIENT_IP_HEADER?.trim().toLowerCase();
    const client = createApiClient({
      baseUrl: env.API_URL,
      headers:
        ipHeader && /^[a-z0-9-]{1,64}$/.test(ipHeader)
          ? forwardedFor(await headers(), { clientIpHeader: ipHeader })
          : {},
    });
    const r = await client.post("/api/v1/contact", {
      body: {
        site,
        name,
        topic,
        subject,
        message,
        email: email || undefined,
        phone: phone || undefined,
        orderNumber: text(form, "orderNumber", 20) || undefined,
        website: text(form, "website", 200) || undefined,
      },
    });
    return {
      ok: true,
      message: r.ticketNumber
        ? `Thanks, we've got it (ticket ${r.ticketNumber}). We reply within 24 hours.`
        : "Thanks, we've got it.",
    };
  } catch (e) {
    return failure(e);
  }
}
