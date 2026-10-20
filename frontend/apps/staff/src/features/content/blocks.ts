// Plain-text editing for CMS blocks (no HTML is ever stored): blank lines separate blocks,
// "## " starts a heading, lines starting "- " form a list, anything else is a paragraph.
export type Block =
  | { type: "heading"; text: string; level: 2 }
  | { type: "paragraph"; text: string }
  | { type: "list"; items: string[] };

export function textToBlocks(text: string): Block[] {
  return text
    .replace(/\r/g, "")
    .split(/\n{2,}/)
    .map((chunk) => chunk.trim())
    .filter(Boolean)
    .map((chunk): Block => {
      if (chunk.startsWith("## "))
        return { type: "heading", text: chunk.slice(3).trim(), level: 2 };
      const lines = chunk.split("\n");
      if (lines.every((l) => l.trim().startsWith("- ")))
        return {
          type: "list",
          items: lines.map((l) => l.trim().slice(2).trim()),
        };
      return { type: "paragraph", text: lines.join(" ") };
    });
}

export function blocksToText(body: unknown): string {
  if (!Array.isArray(body)) return "";
  return body
    .map((b: { type?: string; text?: string; items?: string[] }) =>
      b.type === "heading"
        ? `## ${b.text ?? ""}`
        : b.type === "list"
          ? (b.items ?? []).map((i) => `- ${i}`).join("\n")
          : (b.text ?? ""),
    )
    .filter(Boolean)
    .join("\n\n");
}
