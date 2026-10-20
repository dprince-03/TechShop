import { tokens, useColors } from "@techshop/mobile-ui";
import { DarkTheme, DefaultTheme } from "expo-router";
import { useColorScheme } from "react-native";

export { tokens, useColors };

/** React Navigation theme built from the shared design tokens. */
export function useNavigationTheme(): typeof DefaultTheme {
  const dark = useColorScheme() === "dark";
  const c = dark ? tokens.color.dark : tokens.color.light;
  const base = dark ? DarkTheme : DefaultTheme;
  return {
    ...base,
    colors: {
      ...base.colors,
      primary: c.accent,
      background: c.bg,
      card: c.bg,
      text: c.text,
      border: c.border,
      notification: c.sale,
    },
  };
}
