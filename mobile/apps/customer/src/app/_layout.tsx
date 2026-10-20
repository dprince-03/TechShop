import { Stack, ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";

import { useNavigationTheme } from "@/theme";

export default function RootLayout() {
  const theme = useNavigationTheme();
  return (
    <ThemeProvider value={theme}>
      <Stack>
        <Stack.Screen name="index" options={{ title: "TechShop" }} />
        <Stack.Screen name="search" options={{ title: "Search" }} />
        <Stack.Screen name="c/[slug]" options={{ title: "Category" }} />
        <Stack.Screen name="p/[slug]" options={{ title: "Product" }} />
      </Stack>
      <StatusBar style="auto" />
    </ThemeProvider>
  );
}
