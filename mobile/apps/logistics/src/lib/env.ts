// EXPO_PUBLIC_* values are inlined at build time. Never put secrets in them.
const apiUrl = process.env.EXPO_PUBLIC_API_URL;
if (!apiUrl) throw new Error("EXPO_PUBLIC_API_URL is not set (see .env.example)");

export const env = { apiUrl: apiUrl.replace(/\/$/, "") };
