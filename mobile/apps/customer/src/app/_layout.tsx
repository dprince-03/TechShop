import { Stack, ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";

import { SessionProvider } from "@/features/auth/SessionProvider";
import { CartProvider } from "@/features/cart/CartProvider";
import { useNavigationTheme } from "@/theme";

export default function RootLayout() {
  const theme = useNavigationTheme();
  return (
    <SessionProvider>
    <CartProvider>
    <ThemeProvider value={theme}>
      <Stack>
        <Stack.Screen name="index" options={{ title: "TechShop" }} />
        <Stack.Screen name="search" options={{ title: "Search" }} />
        <Stack.Screen name="c/[slug]" options={{ title: "Category" }} />
        <Stack.Screen name="p/[slug]" options={{ title: "Product" }} />
        <Stack.Screen name="sign-in" options={{ title: "Sign in", presentation: "modal" }} />
        <Stack.Screen name="account" options={{ title: "Account" }} />
        <Stack.Screen name="cart" options={{ title: "Cart" }} />
        <Stack.Screen name="orders/index" options={{ title: "Your orders" }} />
        <Stack.Screen name="orders/[number]" options={{ title: "Order" }} />
        <Stack.Screen name="track" options={{ title: "Track an order" }} />
      </Stack>
      <StatusBar style="auto" />
    </ThemeProvider>
    </CartProvider>
    </SessionProvider>
  );
}
