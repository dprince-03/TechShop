/**
 * Normalises Nigerian mobile numbers (Postel's Law): accepts 0803 123 4567,
 * 08031234567, +234 803 123 4567, 2348031234567. Returns +234XXXXXXXXXX or null.
 */
export function normaliseNigerianPhone(input: string): string | null {
  const digits = input.replace(/[^\d+]/g, "").replace(/^\+/, "");
  let national: string | undefined;
  if (/^234[789][01]\d{8}$/.test(digits)) national = digits.slice(3);
  else if (/^0[789][01]\d{8}$/.test(digits)) national = digits.slice(1);
  else if (/^[789][01]\d{8}$/.test(digits)) national = digits;
  return national ? `+234${national}` : null;
}
