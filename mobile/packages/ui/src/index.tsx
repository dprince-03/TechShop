// Building blocks shared by the TechShop mobile apps, all on the shared design tokens.
// Controls are 52 px (tokens.control.lg): riders work outdoors, often with gloves, and the
// customer app keeps the same large targets (above the 44 px minimum).
import { tokens, type ColorScheme } from "@techshop/design-tokens";
import type { ReactNode } from "react";
import { ActivityIndicator, Pressable, RefreshControl, ScrollView, StyleSheet, Text, TextInput, useColorScheme, View, type TextInputProps } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

export { tokens };

/** Current colour scheme's tokens (light/dark follows the OS setting). */
export function useColors(): ColorScheme {
  return useColorScheme() === "dark" ? tokens.color.dark : tokens.color.light;
}

const RIDER_TARGET = tokens.control.lg;

type ScreenProps = {
  children: ReactNode;
  scroll?: boolean;
  /** A navigation header already sits above (and covers the status bar): don't add the top inset again. */
  underHeader?: boolean;
  /** Pull-to-refresh, for screens showing server data that can change. */
  onRefresh?: () => void;
  refreshing?: boolean;
};

export function Screen({ children, scroll = true, underHeader, onRefresh, refreshing = false }: ScreenProps) {
  const c = useColors();
  const body = <View style={styles.content}>{children}</View>;
  return (
    <SafeAreaView style={[styles.screen, { backgroundColor: c.bg }]} edges={underHeader ? ["left", "right"] : ["top", "left", "right"]}>
      {scroll ? (
        <ScrollView
          contentContainerStyle={styles.scroll}
          keyboardShouldPersistTaps="handled"
          refreshControl={onRefresh ? <RefreshControl refreshing={refreshing} onRefresh={onRefresh} /> : undefined}
        >
          {body}
        </ScrollView>
      ) : (
        body
      )}
    </SafeAreaView>
  );
}

export function Title({ children }: { children: ReactNode }) {
  const c = useColors();
  const size = tokens.fontSize["3xl"].min;
  return (
    <Text accessibilityRole="header" style={{ color: c.text, fontSize: size, fontWeight: "700", letterSpacing: tokens.tracking.tight * size }}>
      {children}
    </Text>
  );
}

export function Body({ children, muted, small, selectable }: { children: ReactNode; muted?: boolean; small?: boolean; selectable?: boolean }) {
  const c = useColors();
  return <Text selectable={selectable} style={{ color: muted ? c.textMuted : c.text, fontSize: small ? tokens.fontSize.sm : tokens.fontSize.base, lineHeight: (small ? tokens.fontSize.sm : tokens.fontSize.base) * tokens.leading.normal }}>{children}</Text>;
}

type ButtonProps = {
  label: string;
  onPress: () => void;
  variant?: "primary" | "secondary" | "danger";
  disabled?: boolean;
  busy?: boolean;
  accessibilityHint?: string;
};

/** One primary action per screen state (docs/mobile.md §4.1). */
export function Button({ label, onPress, variant = "secondary", disabled, busy, accessibilityHint }: ButtonProps) {
  const c = useColors();
  const fill = variant === "primary" ? c.accent : variant === "danger" ? c.danger : c.bg;
  const text = variant === "secondary" ? c.text : c.onAccent;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityHint={accessibilityHint}
      accessibilityState={{ disabled: Boolean(disabled || busy), busy: Boolean(busy) }}
      disabled={disabled || busy}
      onPress={onPress}
      style={({ pressed }) => [
        styles.button,
        { backgroundColor: fill, borderColor: variant === "secondary" ? c.borderStrong : fill, opacity: disabled ? 0.5 : pressed ? 0.85 : 1 },
      ]}
    >
      {busy ? <ActivityIndicator color={text} /> : <Text style={[styles.buttonText, { color: text }]}>{label}</Text>}
    </Pressable>
  );
}

export function Field({ label, error, hint, ...input }: TextInputProps & { label: string; error?: string | null; hint?: string }) {
  const c = useColors();
  return (
    <View style={styles.field}>
      <Text style={{ color: c.text, fontSize: tokens.fontSize.sm, fontWeight: "600" }}>{label}</Text>
      <TextInput
        accessibilityLabel={label}
        placeholderTextColor={c.textSubtle}
        style={[styles.input, { color: c.text, borderColor: error ? c.danger : c.borderStrong, backgroundColor: c.bg }]}
        {...input}
      />
      {hint && !error ? <Text style={{ color: c.textMuted, fontSize: tokens.fontSize.xs }}>{hint}</Text> : null}
      {error ? (
        <Text accessibilityLiveRegion="polite" style={{ color: c.danger, fontSize: tokens.fontSize.sm }}>
          {error}
        </Text>
      ) : null}
    </View>
  );
}

/** Inline status next to what it is about; announced to screen readers. */
export function Notice({ tone = "neutral", children }: { tone?: "neutral" | "success" | "warning" | "danger"; children: ReactNode }) {
  const c = useColors();
  const color = tone === "success" ? c.success : tone === "warning" ? c.warning : tone === "danger" ? c.danger : c.textMuted;
  return (
    <View accessibilityLiveRegion="polite" style={[styles.notice, { borderColor: color }]}>
      <Text style={{ color, fontSize: tokens.fontSize.sm }}>{children}</Text>
    </View>
  );
}

export function Card({ children, onPress, label }: { children: ReactNode; onPress?: () => void; label?: string }) {
  const c = useColors();
  const style = [styles.card, { backgroundColor: c.surfaceRaised, borderColor: c.border }];
  return onPress ? (
    <Pressable accessibilityRole="button" accessibilityLabel={label} onPress={onPress} style={({ pressed }) => [...style, { opacity: pressed ? 0.85 : 1 }]}>
      {children}
    </Pressable>
  ) : (
    <View style={style}>{children}</View>
  );
}

/** 1–5 star picker: five 44px buttons announced as a radio group ("4 stars, selected"). */
export function StarPicker({ label, value, onChange }: { label: string; value: number; onChange: (n: number) => void }) {
  const c = useColors();
  return (
    <View style={styles.field}>
      <Text style={{ color: c.text, fontSize: tokens.fontSize.sm, fontWeight: "600" }}>{label}</Text>
      <View accessibilityRole="radiogroup" accessibilityLabel={label} style={styles.stars}>
        {[1, 2, 3, 4, 5].map((n) => (
          <Pressable
            key={n}
            accessibilityRole="radio"
            accessibilityLabel={`${n} star${n === 1 ? "" : "s"}`}
            accessibilityState={{ checked: value === n }}
            onPress={() => onChange(n)}
            style={[styles.star, { borderColor: n <= value ? c.rating : c.borderStrong, backgroundColor: n <= value ? c.surface : c.bg }]}
          >
            <Text style={{ color: n <= value ? c.rating : c.textSubtle, fontSize: tokens.fontSize.xl }}>★</Text>
          </Pressable>
        ))}
      </View>
    </View>
  );
}

/** Pick one of a few options: large rows announced as a radio group ("Good, selected"). */
export function Choice<T extends string>({
  label,
  options,
  value,
  onChange,
}: {
  label: string;
  options: { value: T; label: string; hint?: string }[];
  value: T | null;
  onChange: (v: T) => void;
}) {
  const c = useColors();
  return (
    <View style={styles.field}>
      <Text style={{ color: c.text, fontSize: tokens.fontSize.sm, fontWeight: "600" }}>{label}</Text>
      <View accessibilityRole="radiogroup" accessibilityLabel={label} style={styles.choices}>
        {options.map((o) => {
          const on = value === o.value;
          return (
            <Pressable
              key={o.value}
              accessibilityRole="radio"
              accessibilityLabel={o.label}
              accessibilityHint={o.hint}
              accessibilityState={{ checked: on }}
              onPress={() => onChange(o.value)}
              style={[styles.choice, { borderColor: on ? c.accent : c.borderStrong, backgroundColor: on ? c.accentSubtle : c.bg }]}
            >
              <View style={[styles.dot, { borderColor: on ? c.accent : c.borderStrong, backgroundColor: on ? c.accent : "transparent" }]} />
              <View style={{ flex: 1 }}>
                <Text style={{ color: c.text, fontSize: tokens.fontSize.base }}>{o.label}</Text>
                {o.hint ? <Text style={{ color: c.textMuted, fontSize: tokens.fontSize.xs }}>{o.hint}</Text> : null}
              </View>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

/** Status as a coloured dot plus words (never colour alone). */
export function StatusLine({ label, tone = "neutral" }: { label: string; tone?: "success" | "warning" | "info" | "danger" | "neutral" }) {
  const c = useColors();
  const color = tone === "success" ? c.success : tone === "warning" ? c.warning : tone === "danger" ? c.danger : tone === "info" ? c.accent : c.textMuted;
  return (
    <View style={styles.status}>
      <View style={[styles.statusDot, { backgroundColor: color }]} />
      <Text style={{ color: c.text, fontSize: tokens.fontSize.sm }}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  stars: { flexDirection: "row", gap: tokens.space[2] },
  star: { width: tokens.control.md, height: tokens.control.md, borderWidth: 1, borderRadius: tokens.radius.sm, alignItems: "center", justifyContent: "center" },
  screen: { flex: 1 },
  scroll: { flexGrow: 1 },
  content: { flex: 1, padding: tokens.space[4], gap: tokens.space[4] },
  button: { minHeight: RIDER_TARGET, borderRadius: tokens.radius.pill, borderWidth: 1, alignItems: "center", justifyContent: "center", paddingHorizontal: tokens.space[5] },
  buttonText: { fontSize: tokens.fontSize.lg, fontWeight: "600" },
  field: { gap: tokens.space[1] },
  input: { minHeight: RIDER_TARGET, borderWidth: 1, borderRadius: tokens.radius.sm, paddingHorizontal: tokens.space[3], fontSize: tokens.fontSize.lg },
  notice: { borderWidth: 1, borderRadius: tokens.radius.md, padding: tokens.space[3] },
  card: { borderWidth: 1, borderRadius: tokens.radius.lg, padding: tokens.space[4], gap: tokens.space[2] },
  choices: { gap: tokens.space[2] },
  choice: { minHeight: RIDER_TARGET, flexDirection: "row", alignItems: "center", gap: tokens.space[3], borderWidth: 1, borderRadius: tokens.radius.md, paddingHorizontal: tokens.space[3], paddingVertical: tokens.space[2] },
  dot: { width: tokens.space[5], height: tokens.space[5], borderRadius: tokens.radius.pill, borderWidth: 2 },
  status: { flexDirection: "row", alignItems: "center", gap: tokens.space[2] },
  statusDot: { width: tokens.space[2], height: tokens.space[2], borderRadius: tokens.radius.pill },
});
