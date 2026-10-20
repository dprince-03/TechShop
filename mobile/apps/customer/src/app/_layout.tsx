import { Stack, ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";

import { SessionProvider } from "@/features/auth/SessionProvider";
import { useNavigationTheme } from "@/theme";

export default function RootLayout() {
  const theme = useNavigationTheme();
  return (
    <SessionProvider>
    <ThemeProvider value={theme}>
      <Stack>
        <Stack.Screen name="index" options={{ title: "TechShop" }} />
        <Stack.Screen name="search" options={{ title: "Search" }} />
        <Stack.Screen name="c/[slug]" options={{ title: "Category" }} />
        <Stack.Screen name="p/[slug]" options={{ title: "Product" }} />
        <Stack.Screen name="sign-in" options={{ title: "Sign in", presentation: "modal" }} />
        <Stack.Screen name="account" options={{ title: "Account" }} />
      </Stack>
      <StatusBar style="auto" />
    </ThemeProvider>
    </SessionProvider>
  );
}
