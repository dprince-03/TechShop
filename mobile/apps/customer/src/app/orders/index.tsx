import { Redirect, useRouter } from "expo-router";
import { ActivityIndicator, FlatList, RefreshControl, Text, View } from "react-native";

import { formatKobo } from "@techshop/api-client";
import { Body, Card, Notice, tokens, useColors } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { orderStatus, when } from "@/features/orders/status";

/** The customer's orders, newest first. */
export default function Orders() {
  const { status, session } = useSession();
  const router = useRouter();
  const c = useColors();
  const r = useLoad(() => session.api.get("/api/v1/me/orders", { query: { limit: 30 } }), status);
  if (status === "signed-out") return <Redirect href="/sign-in" />;
  return (
    <FlatList
      data={r.data?.items ?? []}
      keyExtractor={(o) => o.id}
      contentContainerStyle={{ padding: tokens.space[4] }}
      ItemSeparatorComponent={() => <View style={{ height: tokens.space[3] }} />}
      refreshControl={<RefreshControl refreshing={r.loading} onRefresh={r.reload} />}
      ListHeaderComponent={r.error ? <Notice tone="danger">{r.error}</Notice> : null}
      ListEmptyComponent={r.loading ? <ActivityIndicator accessibilityLabel="Loading orders" /> : <Body muted>No orders yet.</Body>}
      renderItem={({ item: o }) => (
        <Card label={`Order ${o.orderNumber}, ${orderStatus[o.status] ?? o.status}, ${formatKobo(o.totalKobo)}`} onPress={() => router.push({ pathname: "/orders/[number]", params: { number: o.orderNumber } })}>
          <Text style={{ color: c.text, fontSize: tokens.fontSize.lg, fontWeight: "600" }}>{o.orderNumber}</Text>
          <Body>{orderStatus[o.status] ?? o.status}</Body>
          <Body muted small>{`${when(o.placedAt)} · ${formatKobo(o.totalKobo)}`}</Body>
        </Card>
      )}
    />
  );
}
