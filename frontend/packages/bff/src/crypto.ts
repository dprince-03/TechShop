// AES-256-GCM sealing for the session cookie (identity-access.md §5.1). Uses Web Crypto so it
// runs in Server Components, Server Functions, Route Handlers and proxy alike.
const VERSION = "v1";
const encoder = new TextEncoder();
const decoder = new TextDecoder();

const keys = new Map<string, Promise<CryptoKey>>();

/** Derives the AES key from the app secret (already high-entropy, so a hash is enough). */
function keyFor(secret: string): Promise<CryptoKey> {
  let key = keys.get(secret);
  if (!key) {
    key = crypto.subtle
      .digest("SHA-256", encoder.encode(secret))
      .then((raw) =>
        crypto.subtle.importKey("raw", raw, "AES-GCM", false, [
          "encrypt",
          "decrypt",
        ]),
      );
    keys.set(secret, key);
  }
  return key;
}

function toBase64Url(bytes: Uint8Array): string {
  let s = "";
  for (const b of bytes) s += String.fromCharCode(b);
  return btoa(s).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function fromBase64Url(text: string): Uint8Array {
  const s = atob(text.replace(/-/g, "+").replace(/_/g, "/"));
  const out = new Uint8Array(s.length);
  for (let i = 0; i < s.length; i++) out[i] = s.charCodeAt(i);
  return out;
}

/** Encrypts a JSON value into `v1.<base64url(iv | ciphertext | tag)>`. */
export async function seal(value: unknown, secret: string): Promise<string> {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const data = encoder.encode(JSON.stringify(value));
  const cipher = new Uint8Array(
    await crypto.subtle.encrypt(
      { name: "AES-GCM", iv },
      await keyFor(secret),
      data,
    ),
  );
  const out = new Uint8Array(iv.length + cipher.length);
  out.set(iv);
  out.set(cipher, iv.length);
  return `${VERSION}.${toBase64Url(out)}`;
}

/** Decrypts a sealed value; returns null when it is malformed, tampered with or from another key. */
export async function unseal<T>(
  token: string | undefined,
  secret: string,
): Promise<T | null> {
  if (!token?.startsWith(`${VERSION}.`)) return null;
  try {
    const bytes = fromBase64Url(token.slice(VERSION.length + 1));
    if (bytes.length < 13) return null;
    const plain = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv: bytes.slice(0, 12) },
      await keyFor(secret),
      bytes.slice(12),
    );
    return JSON.parse(decoder.decode(plain)) as T;
  } catch {
    return null;
  }
}
