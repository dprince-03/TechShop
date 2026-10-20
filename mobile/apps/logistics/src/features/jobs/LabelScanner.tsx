import { CameraView, useCameraPermissions } from "expo-camera";
import { useRef, useState } from "react";
import { StyleSheet, View } from "react-native";

import { Body, Button, Field } from "@techshop/mobile-ui";
import { tokens } from "@/theme";

/** Scans the parcel label (QR or Code 128), with typing as the fallback when the camera can't. */
export function LabelScanner({ onLabel, busy }: { onLabel: (label: string) => void; busy?: boolean }) {
  const [permission, requestPermission] = useCameraPermissions();
  const [typed, setTyped] = useState("");
  const [scanning, setScanning] = useState(false);
  const done = useRef(false);

  return (
    <View style={styles.wrap}>
      {scanning && permission?.granted ? (
        <View style={styles.camera}>
          <CameraView
            style={StyleSheet.absoluteFill}
            barcodeScannerSettings={{ barcodeTypes: ["qr", "code128"] }}
            onBarcodeScanned={({ data }) => {
              // The camera fires repeatedly; take the first read only.
              if (done.current) return;
              done.current = true;
              setScanning(false);
              onLabel(data.trim().toUpperCase());
            }}
          />
        </View>
      ) : null}
      {scanning ? (
        <Button label="Stop scanning" onPress={() => setScanning(false)} />
      ) : (
        <Button
          label="Scan parcel label"
          variant="primary"
          busy={busy}
          onPress={async () => {
            if (!permission?.granted && !(await requestPermission()).granted) return;
            done.current = false;
            setScanning(true);
          }}
        />
      )}
      {permission && !permission.granted && !permission.canAskAgain ? <Body muted small>Camera access is off. Type the label instead.</Body> : null}
      <Field label="Or type the label" value={typed} onChangeText={(t) => setTyped(t.toUpperCase())} autoCapitalize="characters" autoCorrect={false} placeholder="TSL-…" />
      <Button label="Confirm label" disabled={typed.trim().length < 4 || busy} onPress={() => onLabel(typed.trim())} />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { gap: tokens.space[3] },
  camera: { height: 260, borderRadius: tokens.radius.lg, overflow: "hidden" },
});
