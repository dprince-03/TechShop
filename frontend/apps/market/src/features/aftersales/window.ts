/** Returns are open for 7 days after delivery (orders-fulfilment.md §7). */
const RETURN_WINDOW_MS = 7 * 24 * 60 * 60 * 1000;

/** Whether a parcel delivered at `deliveredAt` can still be returned. */
export function inReturnWindow(
  deliveredAt: string | null | undefined,
): boolean {
  return (
    Boolean(deliveredAt) &&
    Date.now() - new Date(deliveredAt!).getTime() <= RETURN_WINDOW_MS
  );
}
