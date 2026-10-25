import { Redirect, useRouter } from "expo-router";
import { useState } from "react";

import { formatKobo } from "@techshop/api-client";
import { Body, Button, Card, Notice, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";

/** The signed-in customer: store credit, orders and after-sales, and sign-out. */
export default function Account() {
  const { status, me, signOut, session } = useSession();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const credit = useLoad(async () => (status === "signed-in" ? session.api.get("/api/v1/me/store-credit") : null), status);
  if (status === "signed-out") return <Redirect href="/sign-in" />;
  return (
    <Screen underHeader>
      <Title>{me ? `Hello, ${me.firstName}` : "Your account"}</Title>
      {me ? (
        <Card>
          <Body>{me.firstName} {me.lastName}</Body>
          {me.phone ? <Body muted>{`0${me.phone.slice(4)}`}</Body> : null}
          {me.email ? <Body muted>{me.email}</Body> : null}
        </Card>
      ) : null}
      {credit.data && credit.data.balanceKobo > 0 ? <Notice tone="success">{`You have ${formatKobo(credit.data.balanceKobo)} store credit. Use it at checkout when it covers the whole order.`}</Notice> : null}
      <Button label="Your orders" onPress={() => router.push("/orders")} />
      <Button label="Returns" onPress={() => router.push("/returns")} />
      <Button label="Warranty and repairs" onPress={() => router.push("/warranty")} />
      <Button label="Trade-ins" onPress={() => router.push("/trade-in")} />
      <Button label="Support" onPress={() => router.push("/support")} />
      <Body muted small>Addresses you add at checkout are saved to your account.</Body>
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
