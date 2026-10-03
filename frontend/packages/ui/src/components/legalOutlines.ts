/*
 * Outlines for TechShop's legal documents. These are NOT legal text: they list
 * what each section must cover so counsel can draft the real documents.
 */

export type LegalDoc = "terms" | "privacy" | "cookies" | "seller-agreement";

type Section = { heading: string; covers: string[] };

export const legalDocs: Record<
  LegalDoc,
  { title: string; summary: string; sections: Section[] }
> = {
  terms: {
    title: "Terms of use",
    summary:
      "The rules for using TechShop’s websites and apps, and for buying from TechShop and marketplace sellers.",
    sections: [
      {
        heading: "Who we are",
        covers: [
          "TechShop’s registered company name, RC number and registered address",
          "How to contact us",
        ],
      },
      {
        heading: "Your account",
        covers: [
          "Eligibility and age",
          "Keeping login details secure",
          "Suspending or closing accounts",
        ],
      },
      {
        heading: "Buying on TechShop",
        covers: [
          "Who the seller is: TechShop or a marketplace seller",
          "Prices in naira, availability and pricing errors",
          "Order acceptance and cancellation",
        ],
      },
      {
        heading: "Payments",
        covers: [
          "Accepted methods (Paystack, OPay, Moniepoint, bank transfer)",
          "When payment is taken and refunded",
          "Failed and disputed payments",
        ],
      },
      {
        heading: "Delivery",
        covers: [
          "Delivery areas, estimates and fees",
          "Failed deliveries and collection",
        ],
      },
      {
        heading: "Returns, refunds and warranty",
        covers: [
          "Return windows and condition",
          "Refund timelines",
          "Manufacturer vs. TechShop vs. seller warranty",
          "Consumer rights under the Federal Competition and Consumer Protection Act 2018",
        ],
      },
      {
        heading: "Used, refurbished and traded-in devices",
        covers: [
          "Condition grading",
          "Proof of ownership and IMEI/serial checks",
        ],
      },
      {
        heading: "Liability and disputes",
        covers: [
          "Limits of liability",
          "Complaints process",
          "Governing law (Nigeria) and dispute resolution",
        ],
      },
      {
        heading: "Changes to these terms",
        covers: ["How and when customers are notified"],
      },
    ],
  },
  privacy: {
    title: "Privacy policy",
    summary:
      "How TechShop collects, uses, shares and protects personal data, and the rights people have over it.",
    sections: [
      {
        heading: "Data controller",
        covers: [
          "TechShop’s identity and contact details",
          "Data Protection Officer contact, as required by the Nigeria Data Protection Act 2023",
        ],
      },
      {
        heading: "What we collect",
        covers: [
          "Account and contact details",
          "Order, payment and delivery data",
          "Seller KYC data (ID, NIN, CAC, bank details)",
          "Device and usage data",
        ],
      },
      {
        heading: "Why we use it and on what basis",
        covers: [
          "Lawful basis for each purpose (contract, consent, legal obligation, legitimate interest)",
          "Marketing and how to opt out",
        ],
      },
      {
        heading: "Who we share it with",
        covers: [
          "Payment providers, logistics partners and marketplace sellers",
          "Regulators and law enforcement",
          "Transfers outside Nigeria and safeguards",
        ],
      },
      {
        heading: "How long we keep it",
        covers: ["Retention periods per data type"],
      },
      {
        heading: "Your rights",
        covers: [
          "Access, correction, deletion, objection and portability",
          "How to make a request and response times",
          "Complaints to the Nigeria Data Protection Commission",
        ],
      },
      {
        heading: "Security",
        covers: ["How data is protected", "Breach notification"],
      },
    ],
  },
  cookies: {
    title: "Cookie policy",
    summary:
      "What cookies and similar technologies TechShop uses, and how to control them.",
    sections: [
      {
        heading: "What cookies are",
        covers: ["Plain explanation of cookies and local storage"],
      },
      {
        heading: "Cookies we use",
        covers: [
          "Strictly necessary (sign-in, cart, security)",
          "Preferences",
          "Analytics",
          "Marketing — only with consent",
        ],
      },
      {
        heading: "Third parties",
        covers: [
          "Payment, analytics and advertising providers and their policies",
        ],
      },
      {
        heading: "Your choices",
        covers: [
          "Consent banner and how to change choices",
          "Browser controls",
        ],
      },
    ],
  },
  "seller-agreement": {
    title: "Seller agreement",
    summary:
      "The agreement between TechShop and marketplace sellers (businesses and individuals).",
    sections: [
      {
        heading: "Eligibility and verification",
        covers: [
          "KYC requirements for individuals and businesses",
          "Ongoing verification and account suspension",
        ],
      },
      {
        heading: "Listings",
        covers: [
          "Accurate descriptions, condition grading and photos",
          "Prohibited and restricted items",
          "Proof of ownership for used devices",
        ],
      },
      {
        heading: "Fees and payouts",
        covers: [
          "Commission by category",
          "Payout schedule and holds",
          "Taxes and invoices",
        ],
      },
      {
        heading: "Orders and fulfilment",
        covers: [
          "Dispatch times",
          "Using TechShop logistics or self-delivery",
          "Cancellations",
        ],
      },
      {
        heading: "Returns, refunds and disputes",
        covers: [
          "Seller responsibilities",
          "Chargebacks",
          "Dispute resolution",
        ],
      },
      {
        heading: "Performance and conduct",
        covers: ["Service standards and ratings", "Penalties and removal"],
      },
      {
        heading: "Liability, termination and governing law",
        covers: [
          "Indemnities and limits",
          "Ending the agreement",
          "Nigerian law",
        ],
      },
    ],
  },
};
