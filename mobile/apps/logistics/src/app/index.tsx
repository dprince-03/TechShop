import { StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { tokens, useColors } from "@/theme";

export default function Home() {
  const colors = useColors();

  return (
    <SafeAreaView style={[styles.screen, { backgroundColor: colors.bg }]}>
      <View style={styles.content}>
        <Text style={[styles.title, { color: colors.text }]}>{APP_TITLE}</Text>
      </View>
    </SafeAreaView>
  );
}

const APP_TITLE = "TechShop Logistics";

const titleSize = tokens.fontSize["4xl"].min;

const styles = StyleSheet.create({
  screen: { flex: 1 },
  content: { padding: tokens.space[4] },
  title: {
    fontSize: titleSize,
    fontWeight: "600",
    letterSpacing: tokens.tracking.tight * titleSize,
    lineHeight: titleSize * tokens.leading.tight,
  },
});
