import { Redirect, useRouter } from "expo-router";
import { useState } from "react";

import { Body, Button, Card, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";

/** The signed-in customer, and sign-out (orders and addresses come with the next slice). */
export default function Account() {
  const { status, me, signOut } = useSession();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  if (status === "signed-out") return <Redirect href="/sign-in" />;
  return (
    <Screen>
      <Title>{me ? `Hello, ${me.firstName}` : "Your account"}</Title>
      {me ? (
        <Card>
          <Body>{me.firstName} {me.lastName}</Body>
          {me.phone ? <Body muted>{`0${me.phone.slice(4)}`}</Body> : null}
          {me.email ? <Body muted>{me.email}</Body> : null}
        </Card>
      ) : null}
      <Body muted small>Orders, addresses and checkout in the app are coming soon.</Body>
      <Button
        label="Sign out"
        variant="danger"
        busy={busy}
        onPress={async () => {
          setBusy(true);
          await signOut();
          router.replace("/");
        }}
      />
    </Screen>
  );
}
