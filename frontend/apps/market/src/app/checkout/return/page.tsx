import { redirect } from "next/navigation";

/**
 * Where payment providers send the customer back (`PAYMENT_CALLBACK_URL?order=TS-…`). The
 * order page shows the confirmed status; the provider's webhook is what actually marks it paid.
 */
export default async function PaymentReturnPage(
  props: PageProps<"/checkout/return">,
) {
  const order = (await props.searchParams).order;
  const number =
    typeof order === "string" && /^TS-\d{1,12}$/.test(order)
      ? order
      : undefined;
  redirect(number ? `/orders/${number}?placed=1` : "/orders");
}
