import { CameraView, useCameraPermissions } from "expo-camera";
import { useRef, useState } from "react";
import { StyleSheet, View } from "react-native";

import type { MobileSession } from "@techshop/mobile-auth";
import { Body, Button, Notice } from "@techshop/mobile-ui";
import { tokens } from "@/theme";

/**
 * Takes a proof photo and uploads it (presigned URL, then "complete"), handing back the file
 * ID for the job action. Needs signal: the photo is the proof, so it's sent at once.
 */
export function PhotoProof({ session, label, onPhoto, busy }: { session: MobileSession; label: string; onPhoto: (fileId: string) => void; busy?: boolean }) {
  const [permission, requestPermission] = useCameraPermissions();
  const [open, setOpen] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const camera = useRef<CameraView>(null);

  const shoot = async () => {
    setError(null);
    setSending(true);
    try {
      const pic = await camera.current?.takePictureAsync({ quality: 0.5 });
      if (!pic) throw new Error("no photo");
      const blob = await (await fetch(pic.uri)).blob();
      const up = await session.api.post("/api/v1/files/uploads", { body: { purpose: "delivery_proof", contentType: "image/jpeg", sizeBytes: blob.size || 1, name: "proof.jpg" } });
      const put = await fetch(up.uploadUrl, { method: "PUT", headers: { "Content-Type": "image/jpeg" }, body: blob });
      if (!put.ok) throw new Error(`upload ${put.status}`);
      await session.api.post("/api/v1/files/{id}/complete", { path: { id: up.fileId }, body: {} });
      setOpen(false);
      onPhoto(up.fileId);
    } catch {
      setError("The photo didn't send. Check your signal and try again.");
    } finally {
      setSending(false);
    }
  };

  return (
    <View style={styles.wrap}>
      {open && permission?.granted ? (
        <View style={styles.camera}>
          <CameraView ref={camera} style={StyleSheet.absoluteFill} />
        </View>
      ) : null}
      {open ? (
        <>
          <Button label="Take the photo" variant="primary" busy={sending || busy} onPress={shoot} />
          <Button label="Cancel" onPress={() => setOpen(false)} />
        </>
      ) : (
        <Button
          label={label}
          variant="primary"
          busy={busy}
          onPress={async () => {
            if (!permission?.granted && !(await requestPermission()).granted) return;
            setOpen(true);
          }}
        />
      )}
      {permission && !permission.granted && !permission.canAskAgain ? <Body muted small>Camera access is off. Turn it on in Settings to take the proof photo.</Body> : null}
      {error ? <Notice tone="danger">{error}</Notice> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { gap: tokens.space[3] },
  camera: { height: 320, borderRadius: tokens.radius.lg, overflow: "hidden" },
});
