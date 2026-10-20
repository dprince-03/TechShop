import { useRouter } from "expo-router";
import { useEffect, useState } from "react";

import { Body, Button, Field, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";

/** Phone → code → (new customers) name. No passwords. */
export default function SignIn() {
  const s = useSession();
  const router = useRouter();
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [first, setFirst] = useState("");
  const [last, setLast] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const run = async (fn: () => Promise<string | null>) => {
    setBusy(true);
    const e = await fn();
    setBusy(false);
    setError(e);
  };

  // Signed in: return to where the shopper came from (e.g. the cart), or home.
  useEffect(() => {
    if (s.status !== "signed-in") return;
    if (router.canGoBack()) router.back();
    else router.replace("/");
  }, [s.status, router]);

  return (
    <Screen underHeader>
      <Title>Sign in or create an account</Title>
      {s.step === "phone" ? (
        <>
          <Body muted>We’ll text you a code. A new number creates your account.</Body>
          <Field label="Mobile number" value={phone} onChangeText={setPhone} keyboardType="phone-pad" autoComplete="tel" textContentType="telephoneNumber" placeholder="0803 123 4567" error={error} />
          <Button label="Send code" variant="primary" busy={busy} onPress={() => run(() => s.requestCode(phone))} />
        </>
      ) : s.step === "code" ? (
        <>
          <Field label="6-digit code" value={code} onChangeText={(t) => setCode(t.replace(/\D/g, "").slice(0, 6))} keyboardType="number-pad" autoComplete="one-time-code" textContentType="oneTimeCode" maxLength={6} error={error} autoFocus />
          <Button label="Continue" variant="primary" busy={busy} disabled={code.length !== 6} onPress={() => run(() => s.verifyCode(code))} />
          <Button label="Use a different number" onPress={() => (s.restart(), setCode(""), setError(null))} />
        </>
      ) : (
        <>
          <Body>Welcome to TechShop! What should we call you?</Body>
          <Field label="First name" value={first} onChangeText={setFirst} autoComplete="given-name" textContentType="givenName" />
          <Field label="Last name" value={last} onChangeText={setLast} autoComplete="family-name" textContentType="familyName" error={error} />
          <Button label="Create account" variant="primary" busy={busy} onPress={() => run(() => s.completeProfile(first, last))} />
        </>
      )}
    </Screen>
  );
}
