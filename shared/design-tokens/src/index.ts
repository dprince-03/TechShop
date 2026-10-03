/*
 * TechShop design tokens — the single source of truth for web and mobile.
 *
 * Sizes are px numbers. The web CSS generator (scripts/build-css.ts) turns
 * them into rem; React Native uses them as-is. Fluid values are { min, max }
 * pairs: clamp() on the web, `min` on mobile.
 */

export type Fluid = { min: number; max: number };

const light = {
  bg: "#ffffff",
  surface: "#f4f5f7", // section bands, media backdrops
  surfaceRaised: "#ffffff", // cards on a surface band
  surfaceInverse: "#111214", // dark promo bands, inverse buttons

  text: "#16171a",
  textMuted: "#5f6368", // secondary copy, meta
  textSubtle: "#6b7078", // captions, placeholders
  textInverse: "#f7f8fa",

  border: "#d9dbe0", // decorative hairlines and dividers
  borderStrong: "#8e939b", // form controls — 3:1 against bg

  accent: "#1a56f0", // primary action fill
  accentHover: "#1546c8",
  accentSubtle: "#e8efff",
  onAccent: "#ffffff",
  link: "#1a56f0",

  sale: "#d1242f", // discount text
  promo: "#c2410c", // deal text
  saleFill: "#d1242f", // badge backgrounds — white text in both themes
  promoFill: "#c2410c",
  onFill: "#ffffff", // text on sale/promo fills — passes on both themes' fills
  rating: "#f5a524", // stars (non-text)
  success: "#1a7f37",
  warning: "#9a6700",
  danger: "#d1242f",

  focus: "#1a56f0",
  overlay: "rgb(10 11 13 / 0.5)",
};

export type ColorScheme = typeof light;

const dark: ColorScheme = {
  ...light,
  bg: "#0b0c0e",
  surface: "#16181c",
  surfaceRaised: "#1d2025",
  surfaceInverse: "#f4f5f7",

  text: "#f2f3f5",
  textMuted: "#a3a8b0",
  textSubtle: "#8d939b",
  textInverse: "#16171a",

  border: "#2c3036",
  borderStrong: "#5b616a",

  accent: "#2563eb",
  accentHover: "#3b74f5",
  accentSubtle: "#16244a",
  link: "#6b9bff",

  sale: "#ff6b6b",
  promo: "#fb8c4b",
  success: "#4ac26b",
  warning: "#e3b341",
  danger: "#ff6b6b",

  focus: "#6b9bff",
};

export const tokens = {
  color: { light, dark },

  font: {
    // Web loads Geist via next/font; mobile uses the platform system font
    // until a custom font is bundled.
    sans: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    mono: "ui-monospace, SFMono-Regular, monospace",
  },

  fontSize: {
    xs: 12,
    sm: 14,
    base: 16,
    lg: 18,
    xl: 21,
    "2xl": 24,
    "3xl": { min: 28, max: 32 },
    "4xl": { min: 32, max: 40 },
    "5xl": { min: 40, max: 56 },
    "6xl": { min: 48, max: 72 },
  } satisfies Record<string, number | Fluid>,

  fontWeight: { regular: 400, medium: 500, semibold: 600, bold: 700 },

  // Unitless multipliers
  leading: { tight: 1.1, snug: 1.25, normal: 1.5, relaxed: 1.65 },

  // em — on mobile multiply by the font size to get letterSpacing in px
  tracking: { tight: -0.022, snug: -0.011, normal: 0, wide: 0.08 },

  space: {
    0: 0,
    1: 4,
    2: 8,
    3: 12,
    4: 16,
    5: 24,
    6: 32,
    7: 48,
    8: 64,
    9: 96,
    10: 128,
  },

  gutter: { min: 16, max: 40 } satisfies Fluid, // page side padding
  sectionSpace: { min: 48, max: 96 } satisfies Fluid, // vertical rhythm between sections

  container: { narrow: 720, md: 1024, default: 1280, wide: 1440 },

  radius: { xs: 4, sm: 8, md: 12, lg: 18, xl: 28, pill: 999 },

  borderWidth: 1,

  // Web-only CSS shadows; mobile elevation is defined per platform later.
  shadow: {
    light: {
      sm: "0 1px 2px rgb(0 0 0 / 0.06)",
      md: "0 4px 16px rgb(0 0 0 / 0.08)",
      lg: "0 12px 32px rgb(0 0 0 / 0.12)",
    },
    dark: {
      sm: "0 1px 2px rgb(0 0 0 / 0.4)",
      md: "0 4px 16px rgb(0 0 0 / 0.45)",
      lg: "0 12px 32px rgb(0 0 0 / 0.55)",
    },
  },

  duration: { fast: 150, base: 250, slow: 400 }, // ms
  easing: {
    standard: [0.25, 0.1, 0.25, 1],
    emphasized: [0.4, 0, 0.2, 1],
  } satisfies Record<string, [number, number, number, number]>,

  zIndex: { dropdown: 100, sticky: 200, overlay: 300, modal: 400, toast: 500 },

  control: { sm: 36, md: 44, lg: 52 }, // heights; md is the minimum touch target

  breakpoint: { sm: 640, md: 768, lg: 1024, xl: 1280 },

  measure: "65ch", // web-only max line length
} as const;

export type Tokens = typeof tokens;
