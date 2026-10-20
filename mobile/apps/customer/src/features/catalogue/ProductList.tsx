import { ActivityIndicator, FlatList, RefreshControl, StyleSheet, View } from "react-native";

import type { Schema } from "@techshop/api-client";
import { Body, Notice, tokens } from "@techshop/mobile-ui";

import { ProductRow } from "./ProductRow";

/** A list of search results with loading, error and empty states. */
export function ProductList({ items, loading, error, onRefresh, header, empty = "No products match." }: { items: Schema<"ProductSummary">[]; loading: boolean; error: string | null; onRefresh: () => void; header?: React.ReactElement; empty?: string }) {
  return (
    <FlatList
      data={items}
      keyExtractor={(p) => p.id}
      renderItem={({ item }) => <ProductRow p={item} />}
      contentContainerStyle={styles.list}
      ItemSeparatorComponent={() => <View style={{ height: tokens.space[3] }} />}
      refreshControl={<RefreshControl refreshing={loading} onRefresh={onRefresh} />}
      ListHeaderComponent={
        <View style={styles.header}>
          {header}
          {error ? <Notice tone="danger">{error}</Notice> : null}
        </View>
      }
      ListEmptyComponent={loading ? <ActivityIndicator accessibilityLabel="Loading products" /> : error ? null : <Body muted>{empty}</Body>}
    />
  );
}

const styles = StyleSheet.create({
  list: { padding: tokens.space[4] },
  header: { gap: tokens.space[3], marginBottom: tokens.space[3] },
});
