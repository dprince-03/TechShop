/*
 * SAMPLE DATA — fictional brands, products, sellers, and figures for UI
 * development only. Replace with Go API calls (see README).
 */
import {
  naira,
  type CarListing,
  type Category,
  type CategorySlug,
  type Product,
  type Seller,
  type WholesaleProduct,
} from "@techshop/api-client";

// ---------- Sellers ----------

export const techshop: Seller = {
  id: "s-techshop",
  name: "TechShop",
  type: "techshop",
};
const ikejaHub: Seller = {
  id: "s-ikeja",
  name: "Ikeja Gadget Hub",
  type: "vendor",
  rating: 4.6,
};
const abujaTech: Seller = {
  id: "s-abuja",
  name: "Wuse Tech Store",
  type: "vendor",
  rating: 4.4,
};
const kanoMobile: Seller = {
  id: "s-kano",
  name: "Sabon Gari Mobile",
  type: "vendor",
  rating: 4.3,
};
const autoLane: Seller = {
  id: "s-autolane",
  name: "Lekki Auto Lane",
  type: "vendor",
  rating: 4.5,
};

// ---------- Categories ----------

export const categories: Category[] = [
  {
    slug: "phones",
    name: "Phones",
    description:
      "Android, iOS, and feature phones — new, UK-used, and refurbished.",
    subcategories: [
      { slug: "android", name: "Android" },
      { slug: "ios", name: "iOS" },
      { slug: "refurbished", name: "Refurbished" },
      { slug: "feature-phones", name: "Feature phones" },
    ],
  },
  {
    slug: "laptops",
    name: "Laptops",
    description: "Business, gaming, and workstation laptops.",
    subcategories: [
      { slug: "business", name: "Business" },
      { slug: "gaming", name: "Gaming" },
      { slug: "workstation", name: "Workstation" },
    ],
  },
  {
    slug: "accessories",
    name: "Accessories",
    description: "Audio, chargers, power banks, cases, and storage.",
    subcategories: [
      { slug: "audio", name: "Audio" },
      { slug: "power", name: "Power & charging" },
      { slug: "storage", name: "Storage" },
    ],
  },
  {
    slug: "gaming",
    name: "Gaming",
    description: "Consoles, controllers, and gaming gear.",
    subcategories: [
      { slug: "consoles", name: "Consoles" },
      { slug: "controllers", name: "Controllers" },
    ],
  },
  {
    slug: "smart-home",
    name: "Smart home",
    description: "Speakers, lighting, security cameras, and plugs.",
    subcategories: [
      { slug: "security", name: "Security" },
      { slug: "lighting", name: "Lighting" },
    ],
  },
  {
    slug: "office",
    name: "Office",
    description: "Printers, monitors, networking, and power backup.",
    subcategories: [
      { slug: "printers", name: "Printers" },
      { slug: "power-backup", name: "Power backup" },
    ],
  },
  {
    slug: "workstations",
    name: "Workstations",
    description: "Desktop towers and pro displays for demanding work.",
    subcategories: [
      { slug: "towers", name: "Towers" },
      { slug: "displays", name: "Pro displays" },
    ],
  },
  {
    slug: "cars",
    name: "Cars",
    description: "Inspected brand-new and used cars from verified dealers.",
    subcategories: [
      { slug: "brand-new", name: "Brand new" },
      { slug: "foreign-used", name: "Foreign used" },
    ],
  },
];

export const categoryBySlug = Object.fromEntries(
  categories.map((c) => [c.slug, c]),
) as Record<CategorySlug, Category>;

// ---------- Products ----------

type P = Omit<Product, "slug" | "stock" | "condition"> &
  Partial<Pick<Product, "stock" | "condition">>;

const slugify = (s: string) =>
  s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");

const product = (p: P): Product => ({
  stock: "in-stock",
  condition: "new",
  ...p,
  slug: slugify(p.name),
});

export const products: Product[] = [
  // Phones
  product({
    id: "p1",
    name: "Nova X5 Pro 5G",
    brand: "Nova",
    category: "phones",
    subcategory: "android",
    keySpec: "12GB RAM · 256GB",
    price: naira(899_000),
    compareAtPrice: naira(1_050_000),
    rating: { average: 4.7, count: 312 },
    seller: techshop,
    badges: ["deal"],
  }),
  product({
    id: "p2",
    name: "Orbit One 15 Max",
    brand: "Orbit",
    category: "phones",
    subcategory: "ios",
    keySpec: "256GB · Titanium",
    price: naira(2_150_000),
    rating: { average: 4.8, count: 540 },
    seller: techshop,
    badges: ["best-seller"],
  }),
  product({
    id: "p3",
    name: "Nova A3 Lite",
    brand: "Nova",
    category: "phones",
    subcategory: "android",
    keySpec: "6GB RAM · 128GB",
    price: naira(189_000),
    compareAtPrice: naira(215_000),
    rating: { average: 4.4, count: 1204 },
    seller: ikejaHub,
  }),
  product({
    id: "p4",
    name: "Orbit One 13 (UK-used)",
    brand: "Orbit",
    category: "phones",
    subcategory: "ios",
    keySpec: "128GB · Battery 89%",
    condition: "uk-used",
    price: naira(640_000),
    rating: { average: 4.3, count: 87 },
    seller: kanoMobile,
    stock: "low-stock",
  }),
  product({
    id: "p5",
    name: "Vela Fold Z",
    brand: "Vela",
    category: "phones",
    subcategory: "android",
    keySpec: "12GB RAM · 512GB",
    price: naira(2_480_000),
    rating: { average: 4.6, count: 64 },
    seller: techshop,
    badges: ["new"],
  }),
  product({
    id: "p6",
    name: "Nova C1 Refurbished",
    brand: "Nova",
    category: "phones",
    subcategory: "refurbished",
    keySpec: "4GB RAM · 64GB",
    condition: "refurbished",
    price: naira(99_500),
    compareAtPrice: naira(124_000),
    rating: { average: 4.1, count: 233 },
    seller: abujaTech,
  }),

  // Laptops
  product({
    id: "p7",
    name: "Kestrel Book 14 Business",
    brand: "Kestrel",
    category: "laptops",
    subcategory: "business",
    keySpec: "Core i7 · 16GB · 512GB SSD",
    price: naira(1_390_000),
    compareAtPrice: naira(1_520_000),
    rating: { average: 4.6, count: 148 },
    seller: techshop,
    badges: ["deal"],
  }),
  product({
    id: "p8",
    name: "Arcadia Strix 16 Gaming",
    brand: "Arcadia",
    category: "laptops",
    subcategory: "gaming",
    keySpec: "RTX-class GPU · 32GB · 1TB",
    price: naira(3_250_000),
    rating: { average: 4.8, count: 92 },
    seller: techshop,
    badges: ["best-seller"],
  }),
  product({
    id: "p9",
    name: "Forge Mobile WS 17",
    brand: "Forge",
    category: "laptops",
    subcategory: "workstation",
    keySpec: "Pro GPU · 64GB · 2TB",
    price: naira(4_890_000),
    rating: { average: 4.9, count: 21 },
    seller: techshop,
    stock: "low-stock",
  }),
  product({
    id: "p10",
    name: "Kestrel Air 13",
    brand: "Kestrel",
    category: "laptops",
    subcategory: "business",
    keySpec: "Ryzen 5 · 8GB · 256GB SSD",
    price: naira(689_000),
    rating: { average: 4.3, count: 410 },
    seller: abujaTech,
  }),
  product({
    id: "p11",
    name: "Arcadia Nitro 15",
    brand: "Arcadia",
    category: "laptops",
    subcategory: "gaming",
    keySpec: "RTX-class GPU · 16GB · 512GB",
    price: naira(1_850_000),
    compareAtPrice: naira(2_100_000),
    rating: { average: 4.5, count: 176 },
    seller: ikejaHub,
    badges: ["deal"],
  }),

  // Accessories
  product({
    id: "p12",
    name: "Pulse Buds Pro",
    brand: "Pulse",
    category: "accessories",
    subcategory: "audio",
    keySpec: "ANC · 30h battery",
    price: naira(79_000),
    compareAtPrice: naira(99_000),
    rating: { average: 4.5, count: 865 },
    seller: techshop,
    badges: ["deal"],
  }),
  product({
    id: "p13",
    name: "Volt 20,000mAh Power Bank",
    brand: "Volt",
    category: "accessories",
    subcategory: "power",
    keySpec: "65W fast charge",
    price: naira(38_500),
    rating: { average: 4.6, count: 2104 },
    seller: kanoMobile,
    badges: ["best-seller"],
  }),
  product({
    id: "p14",
    name: "Pulse Studio Headphones",
    brand: "Pulse",
    category: "accessories",
    subcategory: "audio",
    keySpec: "Over-ear · Wireless",
    price: naira(145_000),
    rating: { average: 4.4, count: 312 },
    seller: techshop,
  }),
  product({
    id: "p15",
    name: "Volt 1TB Portable SSD",
    brand: "Volt",
    category: "accessories",
    subcategory: "storage",
    keySpec: "USB-C · 1,050MB/s",
    price: naira(112_000),
    compareAtPrice: naira(134_000),
    rating: { average: 4.7, count: 198 },
    seller: ikejaHub,
  }),

  // Gaming
  product({
    id: "p16",
    name: "Arcadia Play 5 Console",
    brand: "Arcadia",
    category: "gaming",
    subcategory: "consoles",
    keySpec: "1TB · 4K",
    price: naira(985_000),
    rating: { average: 4.9, count: 433 },
    seller: techshop,
    badges: ["best-seller"],
  }),
  product({
    id: "p17",
    name: "Arcadia Pro Controller",
    brand: "Arcadia",
    category: "gaming",
    subcategory: "controllers",
    keySpec: "Wireless · Haptic",
    price: naira(89_000),
    compareAtPrice: naira(105_000),
    rating: { average: 4.6, count: 290 },
    seller: abujaTech,
    badges: ["deal"],
  }),
  product({
    id: "p18",
    name: "Arcadia Handheld S",
    brand: "Arcadia",
    category: "gaming",
    subcategory: "consoles",
    keySpec: '7" OLED · 512GB',
    price: naira(720_000),
    rating: { average: 4.5, count: 58 },
    seller: techshop,
    badges: ["new"],
  }),

  // Smart home
  product({
    id: "p19",
    name: "Lumen Home Speaker",
    brand: "Lumen",
    category: "smart-home",
    keySpec: "Voice assistant · Wi-Fi",
    price: naira(68_000),
    rating: { average: 4.3, count: 154 },
    seller: techshop,
  }),
  product({
    id: "p20",
    name: "Lumen Guard Camera 2K",
    brand: "Lumen",
    category: "smart-home",
    subcategory: "security",
    keySpec: "Night vision · App alerts",
    price: naira(54_000),
    compareAtPrice: naira(65_000),
    rating: { average: 4.5, count: 377 },
    seller: ikejaHub,
    badges: ["deal"],
  }),
  product({
    id: "p21",
    name: "Lumen Smart Bulb (4-pack)",
    brand: "Lumen",
    category: "smart-home",
    subcategory: "lighting",
    keySpec: "Colour · Schedules",
    price: naira(29_000),
    rating: { average: 4.2, count: 96 },
    seller: techshop,
  }),

  // Office
  product({
    id: "p22",
    name: "Fieldwork Laser Printer M2",
    brand: "Fieldwork",
    category: "office",
    subcategory: "printers",
    keySpec: "Mono · Wi-Fi · Duplex",
    price: naira(245_000),
    rating: { average: 4.4, count: 122 },
    seller: techshop,
  }),
  product({
    id: "p23",
    name: "Fieldwork 2kVA Inverter + UPS",
    brand: "Fieldwork",
    category: "office",
    subcategory: "power-backup",
    keySpec: "Pure sine wave",
    price: naira(410_000),
    compareAtPrice: naira(465_000),
    rating: { average: 4.6, count: 211 },
    seller: abujaTech,
    badges: ["deal"],
  }),
  product({
    id: "p24",
    name: 'Vantage 27" QHD Monitor',
    brand: "Vantage",
    category: "office",
    keySpec: "IPS · 75Hz · USB-C",
    price: naira(285_000),
    rating: { average: 4.5, count: 167 },
    seller: techshop,
  }),

  // Workstations
  product({
    id: "p25",
    name: "Forge Tower W9",
    brand: "Forge",
    category: "workstations",
    subcategory: "towers",
    keySpec: "24-core · 128GB · Pro GPU",
    price: naira(7_950_000),
    rating: { average: 4.9, count: 12 },
    seller: techshop,
    stock: "low-stock",
  }),
  product({
    id: "p26",
    name: 'Vantage Pro Display 32" 4K',
    brand: "Vantage",
    category: "workstations",
    subcategory: "displays",
    keySpec: "Colour-accurate · HDR",
    price: naira(1_150_000),
    rating: { average: 4.7, count: 34 },
    seller: techshop,
  }),
  product({
    id: "p27",
    name: "Forge Compact WS",
    brand: "Forge",
    category: "workstations",
    subcategory: "towers",
    keySpec: "12-core · 64GB · 2TB",
    price: naira(3_400_000),
    compareAtPrice: naira(3_750_000),
    rating: { average: 4.6, count: 19 },
    seller: techshop,
    badges: ["deal"],
  }),
];

export const productsByCategory = (slug: CategorySlug) =>
  products.filter((p) => p.category === slug);

/** Discounted items, for the flash-deals shelf. */
export const deals = products.filter((p) => p.compareAtPrice);

export const bestSellers = products.filter((p) =>
  p.badges?.includes("best-seller"),
);

// ---------- Cars ----------

export const cars: CarListing[] = [
  {
    id: "c1",
    slug: "meridian-c300-2022",
    title: "Meridian C300 Sedan",
    year: 2022,
    mileageKm: 18_400,
    location: "Lekki, Lagos",
    condition: "foreign-used",
    transmission: "automatic",
    price: naira(38_500_000),
    inspected: true,
    seller: autoLane,
  },
  {
    id: "c2",
    slug: "kodo-rx-2024",
    title: "Kodo RX SUV",
    year: 2024,
    mileageKm: 0,
    location: "Victoria Island, Lagos",
    condition: "brand-new",
    transmission: "automatic",
    price: naira(62_000_000),
    inspected: true,
    seller: techshop,
  },
  {
    id: "c3",
    slug: "astra-city-2019",
    title: "Astra City Hatchback",
    year: 2019,
    mileageKm: 64_200,
    location: "Wuse II, Abuja",
    condition: "nigerian-used",
    transmission: "manual",
    price: naira(9_800_000),
    inspected: true,
    seller: autoLane,
  },
  {
    id: "c4",
    slug: "talon-p4-2021",
    title: "Talon P4 Pickup",
    year: 2021,
    mileageKm: 41_000,
    location: "GRA, Port Harcourt",
    condition: "foreign-used",
    transmission: "automatic",
    price: naira(29_900_000),
    inspected: false,
    seller: autoLane,
  },
];

// ---------- Wholesale ----------

const wholesale = (
  p: Product,
  moq: number,
  tiers: [number, number][],
): WholesaleProduct => ({
  ...p,
  // Wholesale is TechShop's own stock.
  seller: techshop,
  minOrderQuantity: moq,
  tiers: tiers.map(([minQuantity, unit]) => ({
    minQuantity,
    unitPrice: naira(unit),
  })),
});

const byId = (id: string) => products.find((p) => p.id === id)!;

export const wholesaleProducts: WholesaleProduct[] = [
  wholesale(byId("p3"), 5, [
    [5, 176_000],
    [20, 169_000],
    [50, 162_000],
  ]),
  wholesale(byId("p10"), 3, [
    [3, 655_000],
    [10, 632_000],
    [25, 610_000],
  ]),
  wholesale(byId("p13"), 10, [
    [10, 34_000],
    [50, 31_500],
    [200, 29_000],
  ]),
  wholesale(byId("p22"), 2, [
    [2, 232_000],
    [10, 221_000],
    [30, 212_000],
  ]),
  wholesale(byId("p24"), 4, [
    [4, 268_000],
    [15, 259_000],
    [40, 249_000],
  ]),
  wholesale(byId("p12"), 10, [
    [10, 71_000],
    [50, 66_500],
    [150, 62_000],
  ]),
];

// ---------- Staff portal (sample operations data) ----------

export type SampleOrder = {
  id: string;
  customer: string;
  channel: "Market" | "Wholesale" | "App";
  items: number;
  total: ReturnType<typeof naira>;
  /** Product id, quantity and the unit price charged (trade price for wholesale). */
  lines: {
    productId: string;
    quantity: number;
    unitPrice: ReturnType<typeof naira>;
  }[];
  status:
    | "Pending payment"
    | "Paid"
    | "Packed"
    | "Out for delivery"
    | "Delivered"
    | "Refund requested";
  placedAt: string;
};

export const sampleOrders: SampleOrder[] = [
  {
    id: "TS-10482",
    customer: "Adaeze O.",
    channel: "Market",
    items: 2,
    total: naira(978_000),
    status: "Paid",
    placedAt: "09:42",
    lines: [
      { productId: "p1", quantity: 1, unitPrice: naira(899_000) },
      { productId: "p12", quantity: 1, unitPrice: naira(79_000) },
    ],
  },
  {
    id: "TS-10481",
    customer: "Brightline Schools Ltd",
    channel: "Wholesale",
    items: 40,
    total: naira(8_480_000),
    status: "Pending payment",
    placedAt: "09:15",
    lines: [{ productId: "p22", quantity: 40, unitPrice: naira(212_000) }],
  },
  {
    id: "TS-10480",
    customer: "Musa K.",
    channel: "App",
    items: 1,
    total: naira(38_500),
    status: "Out for delivery",
    placedAt: "08:57",
    lines: [{ productId: "p13", quantity: 1, unitPrice: naira(38_500) }],
  },
  {
    id: "TS-10479",
    customer: "Tolu A.",
    channel: "Market",
    items: 3,
    total: naira(1_507_500),
    status: "Packed",
    placedAt: "08:31",
    lines: [
      { productId: "p7", quantity: 1, unitPrice: naira(1_390_000) },
      { productId: "p12", quantity: 1, unitPrice: naira(79_000) },
      { productId: "p13", quantity: 1, unitPrice: naira(38_500) },
    ],
  },
  {
    id: "TS-10478",
    customer: "Chinedu E.",
    channel: "Market",
    items: 1,
    total: naira(89_000),
    status: "Refund requested",
    placedAt: "Yesterday",
    lines: [{ productId: "p17", quantity: 1, unitPrice: naira(89_000) }],
  },
  {
    id: "TS-10477",
    customer: "Halima B.",
    channel: "App",
    items: 2,
    total: naira(158_000),
    status: "Delivered",
    placedAt: "Yesterday",
    lines: [{ productId: "p12", quantity: 2, unitPrice: naira(79_000) }],
  },
];

export const sampleKpis = [
  {
    label: "Revenue today",
    value: naira(14_207_500),
    change: "+12% vs. last Thursday",
  },
  { label: "Orders today", value: 186, change: "+8%" },
  { label: "Awaiting dispatch", value: 23, change: "6 over SLA" },
  { label: "Open tickets", value: 17, change: "4 urgent" },
] as const;

export const sampleAttention = [
  {
    label: "Vendor applications awaiting KYC review",
    count: 9,
    module: "Vendors",
  },
  { label: "Refund requests to approve", count: 5, module: "Orders" },
  { label: "Products below reorder level", count: 14, module: "Inventory" },
  { label: "Payments not yet reconciled", count: 31, module: "Finance" },
  {
    label: "Trade-in devices awaiting inspection",
    count: 7,
    module: "Trade-ins",
  },
];

// ---------- Lookups & queries (mirror future API endpoints) ----------

export const getProduct = (slug: string) =>
  products.find((p) => p.slug === slug);

export const getCar = (slug: string) => cars.find((c) => c.slug === slug);

export const getSubcategory = (category: CategorySlug, sub: string) =>
  categoryBySlug[category]?.subcategories.find((s) => s.slug === sub);

export type SortKey =
  "relevance" | "price-asc" | "price-desc" | "rating" | "discount";

export const sortOptions: { value: SortKey; label: string }[] = [
  { value: "relevance", label: "Most relevant" },
  { value: "price-asc", label: "Price: low to high" },
  { value: "price-desc", label: "Price: high to low" },
  { value: "rating", label: "Top rated" },
  { value: "discount", label: "Biggest discount" },
];

export type ProductQuery = {
  q?: string;
  category?: CategorySlug;
  subcategory?: string;
  condition?: Product["condition"][];
  seller?: Seller["type"][];
  /** Naira (major units). */
  min?: number;
  max?: number;
  sort?: SortKey;
};

const discountOf = (p: Product) =>
  p.compareAtPrice
    ? (p.compareAtPrice.amount - p.price.amount) / p.compareAtPrice.amount
    : 0;

/** Filters and sorts the catalogue — the shape a future `GET /api/v1/products` will accept. */
export function queryProducts(query: ProductQuery): Product[] {
  const q = query.q?.trim().toLowerCase();
  let list = products.filter((p) => {
    if (query.category && p.category !== query.category) return false;
    if (query.subcategory && p.subcategory !== query.subcategory) return false;
    if (query.condition?.length && !query.condition.includes(p.condition))
      return false;
    if (query.seller?.length && !query.seller.includes(p.seller.type))
      return false;
    if (query.min !== undefined && p.price.amount < query.min * 100)
      return false;
    if (query.max !== undefined && p.price.amount > query.max * 100)
      return false;
    if (q) {
      const haystack =
        `${p.name} ${p.brand} ${p.keySpec ?? ""} ${p.category} ${p.subcategory ?? ""}`.toLowerCase();
      if (!q.split(/\s+/).every((word) => haystack.includes(word)))
        return false;
    }
    return true;
  });

  switch (query.sort) {
    case "price-asc":
      list = [...list].sort((a, b) => a.price.amount - b.price.amount);
      break;
    case "price-desc":
      list = [...list].sort((a, b) => b.price.amount - a.price.amount);
      break;
    case "rating":
      list = [...list].sort(
        (a, b) => (b.rating?.average ?? 0) - (a.rating?.average ?? 0),
      );
      break;
    case "discount":
      list = [...list].sort((a, b) => discountOf(b) - discountOf(a));
      break;
  }
  return list;
}

/** Other products from the same category, best rated first. */
export const relatedProducts = (product: Product, limit = 8) =>
  products
    .filter((p) => p.category === product.category && p.id !== product.id)
    .sort((a, b) => (b.rating?.average ?? 0) - (a.rating?.average ?? 0))
    .slice(0, limit);

const conditionNames = {
  new: "New",
  "uk-used": "UK-used",
  refurbished: "Refurbished",
} as const;

/** Spec rows for the product page. Derived until the API returns full specifications. */
export function productSpecs(p: Product): { label: string; value: string }[] {
  const category = categoryBySlug[p.category];
  const sub = p.subcategory
    ? getSubcategory(p.category, p.subcategory)?.name
    : undefined;
  return [
    { label: "Brand", value: p.brand },
    { label: "Model", value: p.name },
    {
      label: "Category",
      value: sub ? `${category.name} · ${sub}` : category.name,
    },
    { label: "Condition", value: conditionNames[p.condition] },
    ...(p.keySpec ? [{ label: "Highlights", value: p.keySpec }] : []),
    { label: "Warranty", value: warrantyFor(p) },
  ];
}

/** SAMPLE warranty wording — real terms come from the product/seller record. */
export const warrantyFor = (p: Product) =>
  p.condition === "new"
    ? "12 months manufacturer warranty"
    : p.condition === "refurbished"
      ? "6 months TechShop warranty"
      : "3 months seller warranty";

/** SAMPLE delivery estimate — real estimates come from the logistics API. */
export const deliveryEstimate = (p: Product) =>
  p.seller.type === "techshop"
    ? { lagos: "Tomorrow", elsewhere: "2–4 working days", fee: naira(2_500) }
    : {
        lagos: "1–2 working days",
        elsewhere: "3–5 working days",
        fee: naira(3_500),
      };

// ---------- Sample cart (interface preview only) ----------

export type CartLine = { product: Product; quantity: number };

export const sampleCart: CartLine[] = [
  { product: byId("p1"), quantity: 1 },
  { product: byId("p12"), quantity: 2 },
  { product: byId("p11"), quantity: 1 },
  { product: byId("p13"), quantity: 1 },
];

// ---------- Reference data ----------

/** Nigeria's 36 states and the FCT, for address and delivery forms. */
export const nigerianStates = [
  "Abia",
  "Adamawa",
  "Akwa Ibom",
  "Anambra",
  "Bauchi",
  "Bayelsa",
  "Benue",
  "Borno",
  "Cross River",
  "Delta",
  "Ebonyi",
  "Edo",
  "Ekiti",
  "Enugu",
  "FCT (Abuja)",
  "Gombe",
  "Imo",
  "Jigawa",
  "Kaduna",
  "Kano",
  "Katsina",
  "Kebbi",
  "Kogi",
  "Kwara",
  "Lagos",
  "Nasarawa",
  "Niger",
  "Ogun",
  "Ondo",
  "Osun",
  "Oyo",
  "Plateau",
  "Rivers",
  "Sokoto",
  "Taraba",
  "Yobe",
  "Zamfara",
];

/** TechShop's own stock (retail side of the wholesale & retail site). */
export const ownStock = products.filter((p) => p.seller.type === "techshop");

export const getWholesaleProduct = (slug: string) =>
  wholesaleProducts.find((p) => p.slug === slug);

export const getProductById = (id: string) => products.find((p) => p.id === id);
