import { Stack, ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";

import { useNavigationTheme } from "@/theme";

export default function RootLayout() {
  const theme = useNavigationTheme();

  return (
    <ThemeProvider value={theme}>
      <Stack screenOptions={{ headerShown: false }} />
      <StatusBar style="auto" />
    </ThemeProvider>
  );
}
