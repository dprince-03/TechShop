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
        <Stack.Screen name="checkout" options={{ title: "Checkout" }} />
        <Stack.Screen name="returns/index" options={{ title: "Returns" }} />
        <Stack.Screen name="returns/new" options={{ title: "Return items" }} />
        <Stack.Screen name="returns/[rma]" options={{ title: "Return" }} />
        <Stack.Screen name="warranty/index" options={{ title: "Warranty" }} />
        <Stack.Screen name="warranty/new" options={{ title: "Warranty claim" }} />
        <Stack.Screen name="warranty/[number]" options={{ title: "Claim" }} />
        <Stack.Screen name="trade-in/index" options={{ title: "Trade-ins" }} />
        <Stack.Screen name="trade-in/new" options={{ title: "Trade in" }} />
        <Stack.Screen name="trade-in/[ref]" options={{ title: "Trade-in" }} />
        <Stack.Screen name="support/index" options={{ title: "Support" }} />
        <Stack.Screen name="support/new" options={{ title: "Ask support" }} />
        <Stack.Screen name="support/[number]" options={{ title: "Ticket" }} />
      </Stack>
      <StatusBar style="auto" />
    </ThemeProvider>
    </CartProvider>
    </SessionProvider>
  );
}
