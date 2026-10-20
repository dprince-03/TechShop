/*
 * SAMPLE data for every staff module. Shapes mirror what the Go API will
 * return; values are fictional. Replace module by module as endpoints land.
 */

export type Tone = "success" | "warning" | "info" | "danger" | "neutral";

export type ModuleConfig = {
  title: string;
  description: string;
  primary?: { label: string; href: string };
  kpis: { label: string; value: string; note?: string }[];
  /** Tab filters; rows match on `tab`. First tab shows everything. */
  tabs: { key: string; label: string }[];
  columns: string[];
  /** Column index rendered as a status dot + text. */
  statusColumn?: number;
  tones?: Record<string, Tone>;
  rows: { tab?: string; href?: string; cells: string[] }[];
};

const ALL = { key: "all", label: "All" };

export const modules: Record<string, ModuleConfig> = {
  orders: {
    title: "Orders",
    description: "Every order from Market, Wholesale and the app.",
    kpis: [
      { label: "Today", value: "186" },
      { label: "Awaiting payment", value: "12" },
      { label: "Awaiting dispatch", value: "23", note: "6 over SLA" },
      { label: "Refund requests", value: "5" },
    ],
    tabs: [
      ALL,
      { key: "pending", label: "Pending payment" },
      { key: "paid", label: "Paid" },
      { key: "delivery", label: "In delivery" },
      { key: "refund-requested", label: "Refunds" },
    ],
    columns: ["Order", "Customer", "Channel", "Total", "Status", "Placed"],
    statusColumn: 4,
    tones: {
      "Pending payment": "warning",
      Paid: "success",
      Packed: "info",
      "Out for delivery": "info",
      Delivered: "neutral",
      "Refund requested": "danger",
    },
    rows: [
      {
        tab: "paid",
        href: "/orders/TS-10482",
        cells: ["TS-10482", "Adaeze O.", "Market", "₦978,000", "Paid", "09:42"],
      },
      {
        tab: "pending",
        href: "/orders/TS-10481",
        cells: [
          "TS-10481",
          "Brightline Schools Ltd",
          "Wholesale",
          "₦8,480,000",
          "Pending payment",
          "09:15",
        ],
      },
      {
        tab: "delivery",
        href: "/orders/TS-10480",
        cells: [
          "TS-10480",
          "Musa K.",
          "App",
          "₦38,500",
          "Out for delivery",
          "08:57",
        ],
      },
      {
        tab: "paid",
        href: "/orders/TS-10479",
        cells: [
          "TS-10479",
          "Tolu A.",
          "Market",
          "₦1,507,500",
          "Packed",
          "08:31",
        ],
      },
      {
        tab: "refund-requested",
        href: "/orders/TS-10478",
        cells: [
          "TS-10478",
          "Chinedu E.",
          "Market",
          "₦89,000",
          "Refund requested",
          "Yesterday",
        ],
      },
      {
        tab: "delivery",
        href: "/orders/TS-10477",
        cells: [
          "TS-10477",
          "Halima B.",
          "App",
          "₦158,000",
          "Delivered",
          "Yesterday",
        ],
      },
    ],
  },
  catalogue: {
    title: "Catalogue",
    description:
      "Products, categories, specs and prices across TechShop stock and marketplace listings.",
    primary: { label: "Add a product", href: "/catalogue/new" },
    kpis: [
      { label: "Live products", value: "27" },
      { label: "Awaiting review", value: "8", note: "Vendor listings" },
      { label: "Missing photos", value: "27" },
      { label: "Price changes today", value: "4" },
    ],
    tabs: [
      ALL,
      { key: "live", label: "Live" },
      { key: "review", label: "Awaiting review" },
      { key: "draft", label: "Drafts" },
    ],
    columns: ["Product", "Category", "Seller", "Price", "Status"],
    statusColumn: 4,
    tones: {
      Live: "success",
      "Awaiting review": "warning",
      Draft: "neutral",
      Rejected: "danger",
    },
    rows: [
      {
        tab: "live",
        cells: [
          "Nova X5 Pro 5G",
          "Phones · Android",
          "TechShop",
          "₦899,000",
          "Live",
        ],
      },
      {
        tab: "live",
        cells: [
          "Kestrel Book 14 Business",
          "Laptops · Business",
          "TechShop",
          "₦1,390,000",
          "Live",
        ],
      },
      {
        tab: "review",
        cells: [
          "Orbit One 12 (UK-used)",
          "Phones · iOS",
          "Sabon Gari Mobile",
          "₦520,000",
          "Awaiting review",
        ],
      },
      {
        tab: "review",
        cells: [
          "Arcadia Nitro 15 (refurb)",
          "Laptops · Gaming",
          "Ikeja Gadget Hub",
          "₦1,450,000",
          "Awaiting review",
        ],
      },
      {
        tab: "draft",
        cells: [
          "Lumen Doorbell Cam",
          "Smart home · Security",
          "TechShop",
          "₦72,000",
          "Draft",
        ],
      },
    ],
  },
  inventory: {
    title: "Inventory",
    description:
      "Stock levels across warehouses and shops, transfers and receiving.",
    primary: { label: "Receive stock", href: "/inventory?tab=receiving" },
    kpis: [
      { label: "SKUs in stock", value: "412" },
      { label: "Below reorder level", value: "14" },
      { label: "Inbound shipments", value: "3" },
      { label: "Stock value", value: "₦486.2m" },
    ],
    tabs: [
      ALL,
      { key: "low", label: "Low stock" },
      { key: "receiving", label: "Receiving" },
      { key: "transfers", label: "Transfers" },
    ],
    columns: ["Product", "Warehouse", "On hand", "Reorder at", "Status"],
    statusColumn: 4,
    tones: {
      "In stock": "success",
      Low: "warning",
      "Out of stock": "danger",
      Inbound: "info",
    },
    rows: [
      { tab: "low", cells: ["Forge Tower W9", "Ikeja WH", "2", "5", "Low"] },
      { tab: "low", cells: ["Orbit One 15 Max", "Ikeja WH", "6", "10", "Low"] },
      { cells: ["Nova A3 Lite", "Ikeja WH", "142", "40", "In stock"] },
      {
        tab: "receiving",
        cells: ["Volt 20,000mAh Power Bank", "Abuja WH", "0", "50", "Inbound"],
      },
      {
        tab: "transfers",
        cells: ["Pulse Buds Pro", "Ikeja → Abuja", "60", "30", "Inbound"],
      },
      {
        tab: "low",
        cells: ["Forge Mobile WS 17", "Ikeja WH", "0", "3", "Out of stock"],
      },
    ],
  },
  dispatch: {
    title: "Dispatch",
    description:
      "Assign riders, manage delivery zones and track deliveries live.",
    primary: { label: "Assign riders", href: "/dispatch?tab=unassigned" },
    kpis: [
      { label: "Awaiting assignment", value: "23" },
      { label: "Riders on shift", value: "41" },
      { label: "Out for delivery", value: "96" },
      { label: "Failed attempts", value: "4" },
    ],
    tabs: [
      ALL,
      { key: "unassigned", label: "Unassigned" },
      { key: "transit", label: "In transit" },
      { key: "failed", label: "Failed" },
    ],
    columns: ["Order", "Zone", "Rider", "Window", "Status"],
    statusColumn: 4,
    tones: {
      Unassigned: "warning",
      "En route": "info",
      Delivered: "success",
      Failed: "danger",
    },
    rows: [
      {
        tab: "unassigned",
        cells: ["TS-10479", "Lekki", "—", "12:00–15:00", "Unassigned"],
      },
      {
        tab: "transit",
        cells: ["TS-10480", "Surulere", "Emeka N.", "09:00–12:00", "En route"],
      },
      {
        tab: "transit",
        cells: ["TS-10471", "Wuse II", "Aisha M.", "09:00–12:00", "En route"],
      },
      {
        tab: "failed",
        cells: ["TS-10466", "Ikorodu", "Tunde B.", "15:00–18:00", "Failed"],
      },
      { cells: ["TS-10462", "Yaba", "Kemi O.", "09:00–12:00", "Delivered"] },
    ],
  },
  support: {
    title: "Customer support",
    description: "Tickets from web, app, WhatsApp and email.",
    kpis: [
      { label: "Open tickets", value: "17", note: "4 urgent" },
      { label: "Avg first reply", value: "38 min" },
      { label: "Resolved today", value: "52" },
      { label: "Satisfaction", value: "92%" },
    ],
    tabs: [
      ALL,
      { key: "urgent", label: "Urgent" },
      { key: "open", label: "Open" },
      { key: "waiting", label: "Waiting on customer" },
    ],
    columns: ["Ticket", "Customer", "Topic", "Channel", "Status"],
    statusColumn: 4,
    tones: {
      Urgent: "danger",
      Open: "warning",
      "Waiting on customer": "neutral",
      Resolved: "success",
    },
    rows: [
      {
        tab: "urgent",
        cells: [
          "#5531",
          "Chinedu E.",
          "Refund not received",
          "WhatsApp",
          "Urgent",
        ],
      },
      {
        tab: "open",
        cells: ["#5530", "Fatima S.", "Change delivery address", "Web", "Open"],
      },
      {
        tab: "waiting",
        cells: [
          "#5527",
          "Ibrahim D.",
          "Warranty claim",
          "Email",
          "Waiting on customer",
        ],
      },
      {
        tab: "urgent",
        cells: ["#5524", "Grace U.", "Wrong item delivered", "App", "Urgent"],
      },
      { cells: ["#5519", "Bola A.", "Payment failed", "Web", "Resolved"] },
    ],
  },
  vendors: {
    title: "Vendors",
    description: "Seller applications, KYC, performance and commissions.",
    kpis: [
      { label: "Active sellers", value: "128" },
      { label: "Awaiting KYC review", value: "9" },
      { label: "Under review (performance)", value: "3" },
      { label: "Avg seller rating", value: "4.4" },
    ],
    tabs: [
      ALL,
      { key: "kyc", label: "KYC review" },
      { key: "active", label: "Active" },
      { key: "flagged", label: "Flagged" },
    ],
    columns: ["Seller", "Type", "Location", "Rating", "Status"],
    statusColumn: 4,
    tones: {
      "KYC review": "warning",
      Active: "success",
      Flagged: "danger",
      Suspended: "neutral",
    },
    rows: [
      {
        tab: "kyc",
        cells: [
          "Computer Village Deals",
          "Business",
          "Ikeja, Lagos",
          "—",
          "KYC review",
        ],
      },
      {
        tab: "kyc",
        cells: ["Uche N.", "Individual", "Enugu", "—", "KYC review"],
      },
      {
        tab: "active",
        cells: [
          "Ikeja Gadget Hub",
          "Business",
          "Ikeja, Lagos",
          "4.6",
          "Active",
        ],
      },
      {
        tab: "active",
        cells: ["Wuse Tech Store", "Business", "Abuja", "4.4", "Active"],
      },
      {
        tab: "flagged",
        cells: ["Phone Palace NG", "Business", "Onitsha", "3.1", "Flagged"],
      },
    ],
  },
  b2b: {
    title: "B2B accounts",
    description: "Business customers, quotes, credit limits and terms.",
    primary: { label: "New quote", href: "/b2b?tab=quotes" },
    kpis: [
      { label: "Business accounts", value: "64" },
      { label: "Open quotes", value: "11" },
      { label: "Credit outstanding", value: "₦38.4m" },
      { label: "Overdue invoices", value: "2" },
    ],
    tabs: [
      ALL,
      { key: "quotes", label: "Quotes" },
      { key: "accounts", label: "Accounts" },
      { key: "overdue", label: "Overdue" },
    ],
    columns: ["Business", "Type", "Account manager", "Value", "Status"],
    statusColumn: 4,
    tones: {
      "Quote sent": "info",
      "Awaiting approval": "warning",
      Active: "success",
      Overdue: "danger",
    },
    rows: [
      {
        tab: "quotes",
        cells: [
          "Brightline Schools Ltd",
          "School",
          "Ngozi A.",
          "₦8,480,000",
          "Quote sent",
        ],
      },
      {
        tab: "quotes",
        cells: [
          "Harbour Logistics",
          "Office / SME",
          "Ngozi A.",
          "₦3,120,000",
          "Awaiting approval",
        ],
      },
      {
        tab: "accounts",
        cells: [
          "Lagos State (sample agency)",
          "Government",
          "Dayo F.",
          "₦—",
          "Active",
        ],
      },
      {
        tab: "overdue",
        cells: [
          "Prime Clinics",
          "Healthcare",
          "Dayo F.",
          "₦1,450,000",
          "Overdue",
        ],
      },
    ],
  },
  purchasing: {
    title: "Purchasing",
    description: "Purchase orders and supplier management.",
    primary: { label: "New purchase order", href: "/purchasing?tab=draft" },
    kpis: [
      { label: "Open POs", value: "7" },
      { label: "Awaiting delivery", value: "3" },
      { label: "Spend this month", value: "₦212.6m" },
      { label: "Active suppliers", value: "18" },
    ],
    tabs: [
      ALL,
      { key: "draft", label: "Drafts" },
      { key: "open", label: "Open" },
      { key: "received", label: "Received" },
    ],
    columns: ["PO", "Supplier", "Items", "Value", "Status"],
    statusColumn: 4,
    tones: {
      Draft: "neutral",
      Sent: "info",
      "Partially received": "warning",
      Received: "success",
    },
    rows: [
      {
        tab: "open",
        cells: [
          "PO-2207",
          "Nova Distribution (sample)",
          "120",
          "₦96,000,000",
          "Sent",
        ],
      },
      {
        tab: "open",
        cells: [
          "PO-2206",
          "Volt Accessories (sample)",
          "500",
          "₦14,500,000",
          "Partially received",
        ],
      },
      {
        tab: "draft",
        cells: [
          "PO-2208",
          "Forge Systems (sample)",
          "6",
          "₦42,300,000",
          "Draft",
        ],
      },
      {
        tab: "received",
        cells: [
          "PO-2201",
          "Kestrel Computing (sample)",
          "40",
          "₦24,800,000",
          "Received",
        ],
      },
    ],
  },
  finance: {
    title: "Finance",
    description: "Payment reconciliation, vendor payouts, refunds and reports.",
    kpis: [
      { label: "Unreconciled payments", value: "31" },
      { label: "Payouts due this week", value: "₦18.9m" },
      { label: "Refunds pending", value: "5" },
      { label: "Revenue (month)", value: "₦312.4m" },
    ],
    tabs: [
      ALL,
      { key: "reconciliation", label: "Reconciliation" },
      { key: "payouts", label: "Payouts" },
      { key: "refunds", label: "Refunds" },
    ],
    columns: ["Reference", "Provider", "Amount", "Matched to", "Status"],
    statusColumn: 4,
    tones: {
      Unmatched: "warning",
      Matched: "success",
      Disputed: "danger",
      Scheduled: "info",
    },
    rows: [
      {
        tab: "reconciliation",
        cells: ["PSK-9F21C", "Paystack", "₦978,000", "TS-10482", "Matched"],
      },
      {
        tab: "reconciliation",
        cells: ["OPY-77A10", "OPay", "₦38,500", "—", "Unmatched"],
      },
      {
        tab: "reconciliation",
        cells: ["MNP-30B44", "Moniepoint", "₦1,507,500", "TS-10479", "Matched"],
      },
      {
        tab: "payouts",
        cells: [
          "PAY-0412",
          "Bank transfer",
          "₦2,340,000",
          "Ikeja Gadget Hub",
          "Scheduled",
        ],
      },
      {
        tab: "refunds",
        cells: ["RF-0088", "Paystack", "₦89,000", "TS-10478", "Disputed"],
      },
    ],
  },
  risk: {
    title: "Risk & fraud",
    description:
      "Suspicious orders, failed payments, chargebacks and seller risk.",
    kpis: [
      { label: "Flagged orders", value: "6" },
      { label: "Chargebacks open", value: "2" },
      { label: "Blocked devices (IMEI)", value: "3" },
      { label: "Seller alerts", value: "1" },
    ],
    tabs: [
      ALL,
      { key: "orders", label: "Orders" },
      { key: "devices", label: "Devices" },
      { key: "sellers", label: "Sellers" },
    ],
    columns: ["Case", "Type", "Reason", "Raised", "Status"],
    statusColumn: 4,
    tones: { "Needs review": "warning", Blocked: "danger", Cleared: "success" },
    rows: [
      {
        tab: "orders",
        cells: [
          "RK-301",
          "Order",
          "Five cards tried on one order",
          "08:12",
          "Needs review",
        ],
      },
      {
        tab: "devices",
        cells: [
          "RK-299",
          "Device",
          "IMEI reported stolen",
          "Yesterday",
          "Blocked",
        ],
      },
      {
        tab: "sellers",
        cells: [
          "RK-297",
          "Seller",
          "Price far below market on 12 listings",
          "Yesterday",
          "Needs review",
        ],
      },
      {
        tab: "orders",
        cells: ["RK-290", "Order", "Address mismatch", "2 days ago", "Cleared"],
      },
    ],
  },
  marketing: {
    title: "Marketing",
    description: "Campaigns, coupons, flash sales and banners.",
    primary: { label: "New campaign", href: "/marketing?tab=draft" },
    kpis: [
      { label: "Live campaigns", value: "3" },
      { label: "Coupons redeemed (week)", value: "412" },
      { label: "Email subscribers", value: "18,240" },
      { label: "Flash sale revenue", value: "₦21.7m" },
    ],
    tabs: [
      ALL,
      { key: "live", label: "Live" },
      { key: "scheduled", label: "Scheduled" },
      { key: "draft", label: "Drafts" },
    ],
    columns: ["Campaign", "Channel", "Audience", "Ends", "Status"],
    statusColumn: 4,
    tones: {
      Live: "success",
      Scheduled: "info",
      Draft: "neutral",
      Ended: "neutral",
    },
    rows: [
      {
        tab: "live",
        cells: ["Today’s deals", "Site + app", "Everyone", "Midnight", "Live"],
      },
      {
        tab: "live",
        cells: [
          "Back to school laptops",
          "Email",
          "Subscribers",
          "15 Oct",
          "Live",
        ],
      },
      {
        tab: "scheduled",
        cells: [
          "Independence weekend",
          "Site + SMS",
          "Everyone",
          "—",
          "Scheduled",
        ],
      },
      {
        tab: "draft",
        cells: ["Trade-in bonus", "WhatsApp", "Phone owners", "—", "Draft"],
      },
    ],
  },
  analytics: {
    title: "Analytics",
    description: "Sales, conversion, best sellers and seller performance.",
    kpis: [
      { label: "Revenue (30 days)", value: "₦312.4m" },
      { label: "Conversion rate", value: "2.8%" },
      { label: "Avg order value", value: "₦176,400" },
      { label: "Returning customers", value: "34%" },
    ],
    tabs: [
      ALL,
      { key: "products", label: "Products" },
      { key: "categories", label: "Categories" },
      { key: "sellers", label: "Sellers" },
    ],
    columns: ["Name", "Type", "Units (30d)", "Revenue (30d)", "Trend"],
    rows: [
      {
        tab: "products",
        cells: ["Nova A3 Lite", "Product", "612", "₦115.7m", "▲ 14%"],
      },
      {
        tab: "products",
        cells: [
          "Volt 20,000mAh Power Bank",
          "Product",
          "1,840",
          "₦70.8m",
          "▲ 6%",
        ],
      },
      {
        tab: "categories",
        cells: ["Phones", "Category", "1,920", "₦138.2m", "▲ 9%"],
      },
      {
        tab: "categories",
        cells: ["Laptops", "Category", "310", "₦96.4m", "▼ 3%"],
      },
      {
        tab: "sellers",
        cells: ["Ikeja Gadget Hub", "Seller", "420", "₦41.0m", "▲ 11%"],
      },
    ],
  },
  content: {
    title: "Content",
    description:
      "Pages, banners, help articles and company news across all sites.",
    primary: { label: "New page", href: "/content?tab=draft" },
    kpis: [
      { label: "Published pages", value: "42" },
      { label: "Drafts", value: "6" },
      { label: "Banners live", value: "5" },
      { label: "Help articles", value: "18" },
    ],
    tabs: [
      ALL,
      { key: "published", label: "Published" },
      { key: "draft", label: "Drafts" },
    ],
    columns: ["Title", "Site", "Type", "Updated", "Status"],
    statusColumn: 4,
    tones: { Published: "success", Draft: "neutral", "In review": "warning" },
    rows: [
      {
        tab: "published",
        cells: ["Delivery", "Market", "Help article", "Today", "Published"],
      },
      {
        tab: "draft",
        cells: ["Terms of use", "All sites", "Legal", "Today", "In review"],
      },
      {
        tab: "published",
        cells: ["Homepage hero", "Market", "Banner", "Yesterday", "Published"],
      },
      {
        tab: "draft",
        cells: ["Credit terms", "Wholesale", "Page", "2 days ago", "Draft"],
      },
    ],
  },
  warranty: {
    title: "Warranty & repairs",
    description:
      "Warranty claims, repair jobs and device history by IMEI/serial.",
    kpis: [
      { label: "Open claims", value: "14" },
      { label: "In repair", value: "9" },
      { label: "Avg turnaround", value: "4.2 days" },
      { label: "Replaced this month", value: "6" },
    ],
    tabs: [
      ALL,
      { key: "new", label: "New claims" },
      { key: "repair", label: "In repair" },
      { key: "ready", label: "Ready for collection" },
    ],
    columns: ["Claim", "Device", "Fault", "Received", "Status"],
    statusColumn: 4,
    tones: {
      New: "warning",
      "In repair": "info",
      Ready: "success",
      Rejected: "danger",
    },
    rows: [
      {
        tab: "new",
        cells: ["WR-1180", "Nova X5 Pro 5G", "Won’t charge", "Today", "New"],
      },
      {
        tab: "repair",
        cells: [
          "WR-1176",
          "Kestrel Book 14",
          "Keyboard fault",
          "3 days ago",
          "In repair",
        ],
      },
      {
        tab: "ready",
        cells: [
          "WR-1170",
          "Pulse Buds Pro",
          "Left bud silent",
          "5 days ago",
          "Ready",
        ],
      },
      {
        cells: [
          "WR-1166",
          "Orbit One 13",
          "Water damage (not covered)",
          "6 days ago",
          "Rejected",
        ],
      },
    ],
  },
  "trade-ins": {
    title: "Trade-ins",
    description: "Trade-in requests, inspections, valuations and payouts.",
    kpis: [
      { label: "Awaiting inspection", value: "7" },
      { label: "Offers sent", value: "12" },
      { label: "Accepted (month)", value: "48" },
      { label: "Avg value", value: "₦212,000" },
    ],
    tabs: [
      ALL,
      { key: "inspection", label: "Inspection" },
      { key: "offer", label: "Offer sent" },
      { key: "done", label: "Completed" },
    ],
    columns: [
      "Request",
      "Device",
      "Condition (declared)",
      "Estimate",
      "Status",
    ],
    statusColumn: 4,
    tones: {
      "Awaiting inspection": "warning",
      "Offer sent": "info",
      Completed: "success",
      Declined: "neutral",
    },
    rows: [
      {
        tab: "inspection",
        cells: [
          "TI-0731",
          "Orbit One 12, 128GB",
          "Good",
          "₦310,000",
          "Awaiting inspection",
        ],
      },
      {
        tab: "offer",
        cells: ["TI-0728", "Nova A2, 64GB", "Fair", "₦62,000", "Offer sent"],
      },
      {
        tab: "done",
        cells: [
          "TI-0720",
          "Arcadia Play 4",
          "Like new",
          "₦280,000",
          "Completed",
        ],
      },
    ],
  },
  cars: {
    title: "Car sales",
    description:
      "Car listings, inspections, viewing bookings and financing enquiries.",
    primary: { label: "New listing", href: "/cars?tab=listings" },
    kpis: [
      { label: "Live listings", value: "4" },
      { label: "Viewings this week", value: "11" },
      { label: "Inspections pending", value: "1" },
      { label: "Financing enquiries", value: "3" },
    ],
    tabs: [
      ALL,
      { key: "viewings", label: "Viewings" },
      { key: "listings", label: "Listings" },
      { key: "financing", label: "Financing" },
    ],
    columns: ["Reference", "Car", "Customer / dealer", "Date", "Status"],
    statusColumn: 4,
    tones: {
      Booked: "info",
      Confirmed: "success",
      "Inspection pending": "warning",
      Enquiry: "neutral",
    },
    rows: [
      {
        tab: "viewings",
        cells: [
          "VW-210",
          "2022 Meridian C300 Sedan",
          "Kunle A.",
          "Sat 10:00",
          "Confirmed",
        ],
      },
      {
        tab: "viewings",
        cells: [
          "VW-209",
          "2024 Kodo RX SUV",
          "Amaka E.",
          "Sat 13:00",
          "Booked",
        ],
      },
      {
        tab: "listings",
        cells: [
          "CL-044",
          "2021 Talon P4 Pickup",
          "Lekki Auto Lane",
          "—",
          "Inspection pending",
        ],
      },
      {
        tab: "financing",
        cells: [
          "FN-018",
          "2019 Astra City Hatchback",
          "Yusuf M.",
          "Yesterday",
          "Enquiry",
        ],
      },
    ],
  },
  pos: {
    title: "Point of sale",
    description: "In-store sales, tills and shift reports for physical shops.",
    kpis: [
      { label: "Shops", value: "—", note: "No shops set up" },
      { label: "Tills", value: "—" },
      { label: "Sales today", value: "—" },
      { label: "Open shifts", value: "—" },
    ],
    tabs: [ALL],
    columns: ["Shop", "Till", "Cashier", "Sales today", "Status"],
    rows: [],
  },
  hr: {
    title: "HR & staff",
    description: "Staff directory, departments and attendance.",
    primary: { label: "Add staff member", href: "/hr?tab=directory" },
    kpis: [
      { label: "Staff", value: "96" },
      { label: "On shift now", value: "54" },
      { label: "On leave", value: "4" },
      { label: "Open roles", value: "0" },
    ],
    tabs: [
      ALL,
      { key: "directory", label: "Directory" },
      { key: "leave", label: "Leave" },
    ],
    columns: ["Name", "Department", "Role", "Location", "Status"],
    statusColumn: 4,
    tones: { Active: "success", "On leave": "neutral", "On shift": "info" },
    rows: [
      {
        tab: "directory",
        cells: ["Ngozi A.", "Sales", "Account manager", "Lagos", "Active"],
      },
      {
        tab: "directory",
        cells: ["Emeka N.", "Logistics", "Dispatch rider", "Lagos", "On shift"],
      },
      {
        tab: "leave",
        cells: [
          "Bisi O.",
          "Customer experience",
          "Support lead",
          "Lagos",
          "On leave",
        ],
      },
      {
        tab: "directory",
        cells: ["Dayo F.", "Sales", "B2B manager", "Abuja", "Active"],
      },
    ],
  },
  admin: {
    title: "Admin console",
    description:
      "Staff accounts, roles and permissions, settings and the audit log.",
    primary: { label: "Invite staff", href: "/admin?tab=users" },
    kpis: [
      { label: "Staff accounts", value: "96" },
      { label: "Roles", value: "—", note: "Role model not defined yet" },
      { label: "Admins", value: "3" },
      { label: "Audit events (24h)", value: "1,204" },
    ],
    tabs: [
      ALL,
      { key: "users", label: "Users" },
      { key: "audit", label: "Audit log" },
    ],
    columns: ["Event", "User", "Module", "When", "Status"],
    statusColumn: 4,
    tones: { Success: "success", Denied: "danger" },
    rows: [
      {
        tab: "audit",
        cells: [
          "Changed price of Nova X5 Pro 5G",
          "Tolu (Catalogue)",
          "Catalogue",
          "09:51",
          "Success",
        ],
      },
      {
        tab: "audit",
        cells: [
          "Approved refund RF-0087",
          "Bisi (Support)",
          "Orders",
          "09:20",
          "Success",
        ],
      },
      {
        tab: "audit",
        cells: [
          "Tried to open Finance",
          "Emeka (Logistics)",
          "Finance",
          "08:44",
          "Denied",
        ],
      },
      {
        tab: "users",
        cells: [
          "Invited Kemi O.",
          "Admin",
          "Admin console",
          "Yesterday",
          "Success",
        ],
      },
    ],
  },
  it: {
    title: "IT tools",
    description: "Feature flags, integrations and system health.",
    kpis: [
      { label: "API status", value: "—", note: "Connect to Go API /health" },
      { label: "Feature flags", value: "4" },
      { label: "Integrations", value: "3", note: "Paystack, OPay, Moniepoint" },
      { label: "Incidents (30d)", value: "0" },
    ],
    tabs: [
      ALL,
      { key: "flags", label: "Feature flags" },
      { key: "integrations", label: "Integrations" },
    ],
    columns: ["Name", "Type", "Environment", "Updated", "Status"],
    statusColumn: 4,
    tones: { On: "success", Off: "neutral", "Not connected": "warning" },
    rows: [
      {
        tab: "flags",
        cells: ["Guest checkout", "Feature flag", "Production", "—", "Off"],
      },
      {
        tab: "flags",
        cells: ["Trade-in estimates", "Feature flag", "Staging", "—", "On"],
      },
      {
        tab: "integrations",
        cells: ["Paystack", "Payments", "—", "—", "Not connected"],
      },
      {
        tab: "integrations",
        cells: ["OPay", "Payments", "—", "—", "Not connected"],
      },
      {
        tab: "integrations",
        cells: ["Moniepoint", "Payments", "—", "—", "Not connected"],
      },
    ],
  },
};
