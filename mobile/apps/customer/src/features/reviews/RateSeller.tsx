import { useState } from "react";

import { Button, Notice, StarPicker } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";

import { problem } from "./problem";

/** Rate the seller of a delivered parcel (once). */
export function RateSeller({ fulfilmentId, seller }: { fulfilmentId: string; seller: string }) {
  const { session } = useSession();
  const [rating, setRating] = useState(0);
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<{ ok: boolean; text: string } | null>(null);
  if (result?.ok) return <Notice tone="success">{result.text}</Notice>;
  return (
    <>
      <StarPicker label={`How was ${seller}?`} value={rating} onChange={setRating} />
      {result ? <Notice tone="danger">{result.text}</Notice> : null}
      <Button
        label={`Rate ${seller}`}
        busy={busy}
        disabled={!rating}
        onPress={async () => {
          setBusy(true);
          try {
            await session.api.post("/api/v1/me/deliveries/{fulfilmentId}/rating", { path: { fulfilmentId }, body: { rating } });
            setResult({ ok: true, text: "Thanks for rating the seller." });
          } catch (e) {
            setResult({ ok: false, text: problem(e) });
          }
          setBusy(false);
        }}
      />
    </>
  );
}
