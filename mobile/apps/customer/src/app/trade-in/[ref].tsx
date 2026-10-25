import { Stack, useLocalSearchParams } from "expo-router";
import { useState } from "react";
import { ActivityIndicator } from "react-native";

import { deviceCondition, formatKobo, tradeInStatus } from "@techshop/api-client";
import { Body, Button, Card, Choice, Field, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { when } from "@/features/orders/status";
import { problem } from "@/features/reviews/problem";

/** A trade-in: estimate, where to bring it, and the offer to accept or decline. */
export default function TradeIn() {
  const { ref } = useLocalSearchParams<{ ref: string }>();
  const { session } = useSession();
  const r = useLoad(() => session.api.get("/api/v1/me/trade-ins/{ref}", { path: { ref } }), ref);
  const [payout, setPayout] = useState<"store_credit" | "bank">("store_credit");
  const [bankQuery, setBankQuery] = useState("");
  const [bank, setBank] = useState<string | null>(null);
  const [account, setAccount] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const banks = useLoad(async () => (payout === "bank" ? ((await session.api.get("/api/v1/me/banks")) ?? []) : []), payout);

  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading trade-in" />;
  if (!r.data) return <Screen underHeader><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const t = r.data;
  const s = tradeInStatus[t.status] ?? { label: t.status, tone: "neutral" as const };
  const matches = (banks.data ?? []).filter((b) => bankQuery.length >= 2 && b.name.toLowerCase().includes(bankQuery.toLowerCase())).slice(0, 6);

  const run = async (fn: () => Promise<unknown>, ok: string) => {
    setBusy(true);
    setMsg(null);
    try {
      await fn();
      setMsg({ ok: true, text: ok });
      await r.reload();
    } catch (e) {
      setMsg({ ok: false, text: problem(e) });
    } finally {
      setBusy(false);
    }
  };
  const accept = () => {
    if (payout === "bank" && (!bank || !/^\d{10}$/.test(account))) return setMsg({ ok: false, text: "Choose your bank and enter the 10-digit account number." });
    return run(
      () => session.api.post("/api/v1/me/trade-ins/{ref}/accept", { path: { ref }, body: payout === "bank" ? { payoutMethod: "bank", bankCode: bank!, accountNumber: account } : { payoutMethod: "store_credit" } }),
      payout === "bank" ? "Accepted. We'll send the money within 2 working days." : "Accepted. The credit is in your account now.",
    );
  };

  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Stack.Screen options={{ title: t.reference }} />
      <Title>{t.device}</Title>
      <StatusLine label={s.label} tone={s.tone} />
      {msg ? <Notice tone={msg.ok ? "success" : "danger"}>{msg.text}</Notice> : null}
      <Card>
        <Body small muted>{`${t.reference} · you said: ${deviceCondition[t.declaredCondition] ?? t.declaredCondition}`}</Body>
        <Body>{t.estimate != null ? `Estimate ${formatKobo(t.estimate)}` : "We're pricing it"}</Body>
        {t.inspectionNotes ? <Body small>{t.inspectionNotes}</Body> : null}
        {t.paidAt ? <Body small>{`Paid ${when(t.paidAt)}${t.payoutMethod === "store_credit" ? " as store credit" : ""}`}</Body> : null}
      </Card>
      {(t.dropOff ?? []).map((st) => (
        <Card key={st.id}>
          <Body>{`Bring it to ${st.name}`}</Body>
          <Body small muted>{`${st.address}, ${st.city}`}</Body>
        </Card>
      ))}
      {t.status === "offer_sent" && t.offer != null ? (
        <Card>
          <Body>{`Our offer: ${formatKobo(t.offer)}`}</Body>
          {t.offerExpiresAt ? <Body small muted>{`Answer by ${when(t.offerExpiresAt)}`}</Body> : null}
          <Choice<"store_credit" | "bank">
            label="How do you want to be paid?"
            value={payout}
            onChange={setPayout}
            options={[
              { value: "store_credit", label: "Store credit, straight away" },
              { value: "bank", label: "Bank transfer, within 2 working days" },
            ]}
          />
          {payout === "bank" ? (
            <>
              <Field label="Your bank" value={bankQuery} onChangeText={setBankQuery} placeholder="Type to search" />
              {matches.length ? <Choice<string> label="Choose your bank" value={bank} onChange={setBank} options={matches.map((b) => ({ value: b.code, label: b.name }))} /> : null}
              <Field label="Account number" value={account} onChangeText={(v) => setAccount(v.replace(/\D/g, ""))} keyboardType="number-pad" maxLength={10} hint="We check the name with your bank." />
            </>
          ) : null}
          <Button label="Accept the offer" variant="primary" busy={busy} onPress={accept} />
          <Button label="Decline and collect my device" busy={busy} onPress={() => run(() => session.api.post("/api/v1/me/trade-ins/{ref}/decline", { path: { ref } }), "Declined. Collect your device from the store.")} />
        </Card>
      ) : null}
      {t.status === "submitted" || t.status === "estimated" ? (
        <Button label="Cancel this trade-in" variant="danger" busy={busy} onPress={() => run(() => session.api.post("/api/v1/me/trade-ins/{ref}/cancel", { path: { ref } }), "Trade-in cancelled.")} />
      ) : null}
    </Screen>
  );
}
