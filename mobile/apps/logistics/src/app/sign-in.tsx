import { Redirect } from "expo-router";
import { useState } from "react";

import { Body, Button, Field, Notice, Screen, Title } from "@/components/ui";
import { useSession } from "@/features/auth/SessionProvider";

/** Phone → 6-digit code. Riders never type passwords; the phone is approved once by dispatch. */
export default function SignIn() {
  const { status, requestCode, verifyCode } = useSession();
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"phone" | "code">("phone");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  if (status === "signed-in") return <Redirect href="/" />;

  const run = async (fn: () => Promise<string | null>, next?: () => void) => {
    setBusy(true);
    const e = await fn();
    setBusy(false);
    setError(e);
    if (!e) next?.();
  };

  return (
    <Screen>
      <Title>TechShop Logistics</Title>
      <Body muted>Sign in with the phone number dispatch has for you.</Body>
      {step === "phone" ? (
        <>
          <Field label="Mobile number" value={phone} onChangeText={setPhone} keyboardType="phone-pad" autoComplete="tel" textContentType="telephoneNumber" placeholder="0803 123 4567" error={error} />
          <Button label="Send code" variant="primary" busy={busy} onPress={() => run(() => requestCode(phone), () => setStep("code"))} />
        </>
      ) : (
        <>
          <Notice>We sent a 6-digit code to {phone}.</Notice>
          <Field label="Code" value={code} onChangeText={(t) => setCode(t.replace(/\D/g, "").slice(0, 6))} keyboardType="number-pad" autoComplete="one-time-code" textContentType="oneTimeCode" maxLength={6} error={error} autoFocus />
          <Button label="Sign in" variant="primary" busy={busy} disabled={code.length !== 6} onPress={() => run(() => verifyCode(code))} />
          <Button label="Use a different number" onPress={() => { setStep("phone"); setCode(""); setError(null); }} />
        </>
      )}
    </Screen>
  );
}
