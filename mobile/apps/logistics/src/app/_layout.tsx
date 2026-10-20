import { Stack, ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";

import { SessionProvider } from "@/features/auth/SessionProvider";
import { useNavigationTheme } from "@/theme";

export default function RootLayout() {
  const theme = useNavigationTheme();
  return (
    <SessionProvider>
      <ThemeProvider value={theme}>
        <Stack screenOptions={{ headerShown: false }}>
          <Stack.Screen name="job/[id]" options={{ headerShown: true, title: "Job" }} />
          <Stack.Screen name="sync" options={{ headerShown: true, title: "Sync" }} />
          <Stack.Screen name="profile" options={{ headerShown: true, title: "Profile" }} />
        </Stack>
        <StatusBar style="auto" />
      </ThemeProvider>
    </SessionProvider>
  );
}
