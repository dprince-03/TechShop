import { Prose } from "./Prose";

/**
 * Renders CMS content stored as JSON blocks (docs/messaging-marketing.md §3: no HTML is ever
 * stored). Everything is rendered as React text, so content can't inject markup; unknown block
 * types are skipped, and links must be relative or https.
 */
type Block =
  | { type: "heading"; text: string; level?: 2 | 3 }
  | { type: "paragraph" | "text"; text: string }
  | { type: "list"; items: string[] }
  | { type: "image"; url: string; alt?: string }
  | { type: "button"; label: string; href: string };

const safeHref = (href: unknown): string | undefined =>
  typeof href === "string" &&
  ((href.startsWith("/") && !href.startsWith("//")) ||
    href.startsWith("https://"))
    ? href
    : undefined;

const str = (v: unknown) => (typeof v === "string" ? v : "");

function toBlocks(body: unknown): Block[] {
  if (!Array.isArray(body)) return [];
  return body.filter(
    (b): b is Block =>
      Boolean(b) && typeof b === "object" && typeof b.type === "string",
  );
}

export function ContentBlocks({ body }: { body: unknown }) {
  const blocks = toBlocks(body);
  return (
    <Prose>
      {blocks.map((b, i) => {
        switch (b.type) {
          case "heading":
            return b.level === 3 ? (
              <h3 key={i}>{str(b.text)}</h3>
            ) : (
              <h2 key={i}>{str(b.text)}</h2>
            );
          case "paragraph":
          case "text":
            return <p key={i}>{str(b.text)}</p>;
          case "list":
            return Array.isArray(b.items) ? (
              <ul key={i}>
                {b.items.map((item, j) => (
                  <li key={j}>{str(item)}</li>
                ))}
              </ul>
            ) : null;
          case "image": {
            const src = safeHref(b.url);
            return src ? (
              // eslint-disable-next-line @next/next/no-img-element -- CMS assets come from the asset host
              <img key={i} src={src} alt={str(b.alt)} loading="lazy" />
            ) : null;
          }
          case "button": {
            const href = safeHref(b.href);
            return href ? (
              <p key={i}>
                <a href={href} className="btn btn--secondary">
                  {str(b.label)}
                </a>
              </p>
            ) : null;
          }
          default:
            return null;
        }
      })}
    </Prose>
  );
}
