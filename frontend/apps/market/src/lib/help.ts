import type { Faq, IconName } from "@techshop/ui/components";

/**
 * SAMPLE help content. Describes how things work in general terms; exact
 * policies (timelines, fees, windows) must be confirmed by the business.
 */
export type HelpTopic = {
  slug: string;
  title: string;
  icon: IconName;
  summary: string;
  sections: { heading: string; body: string[] }[];
  faqs?: Faq[];
};

export const helpTopics: HelpTopic[] = [
  {
    slug: "delivery",
    title: "Delivery",
    icon: "truck",
    summary: "Where we deliver, how long it takes and what it costs.",
    sections: [
      {
        heading: "Where we deliver",
        body: [
          "We deliver to all 36 states and the FCT. Delivery estimates and fees are shown on each product page and confirmed at checkout, before you pay.",
        ],
      },
      {
        heading: "Delivery times",
        body: [
          "Items sold by TechShop usually ship from our own warehouses. Items from marketplace sellers ship from the seller, so each seller in your cart has its own estimate.",
        ],
      },
      {
        heading: "On delivery day",
        body: [
          "You’ll get an SMS with your rider’s name and a delivery window. Check the item before you sign for it. If anything looks wrong, refuse it and contact us.",
        ],
      },
    ],
    faqs: [
      {
        q: "Can I collect my order instead?",
        a: "Pickup points are planned. Until then, all orders are delivered to the address you choose at checkout.",
      },
      {
        q: "What if I miss the delivery?",
        a: "The rider will call you. If they can’t reach you, we’ll arrange another attempt and send you an SMS.",
      },
    ],
  },
  {
    slug: "returns",
    title: "Returns & refunds",
    icon: "returns",
    summary: "How to return an item and when you’ll get your money back.",
    sections: [
      {
        heading: "When you can return an item",
        body: [
          "You can return an item that arrives faulty, damaged or not as described. Some categories may also allow change-of-mind returns; the product page will say so.",
        ],
      },
      {
        heading: "How to start a return",
        body: [
          "Go to your orders, choose the item and tell us what’s wrong. Keep the original box, accessories and receipt. We’ll arrange a pickup or tell you where to drop it.",
        ],
      },
      {
        heading: "Refunds",
        body: [
          "Once the item is checked, we refund to your original payment method. You’ll get an SMS and email when the refund is sent.",
        ],
      },
    ],
  },
  {
    slug: "warranty",
    title: "Warranty",
    icon: "shield",
    summary: "What’s covered, for how long, and how to make a claim.",
    sections: [
      {
        heading: "Types of warranty",
        body: [
          "New products come with the manufacturer’s warranty. Refurbished products sold by TechShop carry a TechShop warranty. Used items from marketplace sellers carry the seller’s warranty. Each product page shows which applies.",
        ],
      },
      {
        heading: "Making a claim",
        body: [
          "Have your order number and the device’s IMEI or serial number ready. We’ll check the fault and repair, replace or refund under the warranty terms.",
        ],
      },
    ],
  },
  {
    slug: "payments",
    title: "Payments",
    icon: "card",
    summary: "Ways to pay and what to do if a payment fails.",
    sections: [
      {
        heading: "Ways to pay",
        body: [
          "Pay securely with Paystack (cards, bank transfer, USSD), OPay or Moniepoint. Sellers never see your card details.",
        ],
      },
      {
        heading: "If a payment fails",
        body: [
          "You won’t be charged for a failed payment. If money left your account but the order didn’t go through, it’s usually reversed automatically. Contact us with the transaction reference if it isn’t.",
        ],
      },
    ],
  },
];

export const getHelpTopic = (slug: string) =>
  helpTopics.find((t) => t.slug === slug);
