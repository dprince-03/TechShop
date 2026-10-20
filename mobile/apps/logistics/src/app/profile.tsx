import { useRouter } from "expo-router";
import { useEffect, useState } from "react";

import { Body, Button, Card, Screen, Title } from "@techshop/mobile-ui";
import { useSession } from "@/features/auth/SessionProvider";
import { clearJobCache } from "@/features/jobs/useJobs";
import { deviceId } from "@/lib/storage";

/** Who is signed in, this phone's ID (dispatch approves it once), and sign-out. */
export default function Profile() {
  const { me, signOut } = useSession();
  const [id, setId] = useState("");
  const [busy, setBusy] = useState(false);
  const router = useRouter();
  useEffect(() => {
    void deviceId().then(setId);
  }, []);
  return (
    <Screen underHeader>
      <Title>{me ? `${me.firstName} ${me.lastName}` : "Profile"}</Title>
      <Card>
        <Body muted small>This phone’s ID</Body>
        <Body selectable>{id}</Body>
        <Body muted small>Dispatch approves each phone once before it can take jobs.</Body>
      </Card>
      <Button
        label="Sign out"
        variant="danger"
        busy={busy}
        onPress={async () => {
          setBusy(true);
          await clearJobCache();
          await signOut();
          // Today sends a signed-out rider to sign-in.
          router.replace("/");
        }}
      />
    </Screen>
  );
}
