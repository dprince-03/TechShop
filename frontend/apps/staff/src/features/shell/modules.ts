import type { IconName } from "@techshop/ui/components";

/**
 * Staff portal modules as they appear in the sidebar. `permission` hides a module from people
 * who can't use it (the API enforces it regardless). `live` modules use the API; the others are
 * previews with sample data until their backend milestone (M4–M7) ships.
 */
export type StaffModule = {
  label: string;
  href: string;
  icon: IconName;
  permission: string;
  live: boolean;
};
export type ModuleGroup = { title: string; modules: StaffModule[] };

export const moduleGroups: ModuleGroup[] = [
  {
    title: "Operations",
    modules: [
      {
        label: "Orders",
        href: "/orders",
        icon: "receipt",
        permission: "orders.view",
        live: true,
      },
      {
        label: "Catalogue",
        href: "/catalogue",
        icon: "tag",
        permission: "catalog.view",
        live: true,
      },
      {
        label: "Inventory",
        href: "/inventory",
        icon: "box",
        permission: "inventory.view",
        live: true,
      },
      {
        label: "Warehouse",
        href: "/warehouse",
        icon: "building",
        permission: "fulfilment.pick",
        live: true,
      },
      {
        label: "Dispatch",
        href: "/dispatch",
        icon: "truck",
        permission: "dispatch.view",
        live: true,
      },
      {
        label: "Customer support",
        href: "/support",
        icon: "support",
        permission: "support.view",
        live: false,
      },
    ],
  },
  {
    title: "Marketplace & B2B",
    modules: [
      {
        label: "Vendors",
        href: "/vendors",
        icon: "store",
        permission: "vendors.view",
        live: true,
      },
      {
        label: "B2B accounts",
        href: "/b2b",
        icon: "briefcase",
        permission: "business.view",
        live: false,
      },
      {
        label: "Purchasing",
        href: "/purchasing",
        icon: "file",
        permission: "purchasing.view",
        live: false,
      },
    ],
  },
  {
    title: "Finance & risk",
    modules: [
      {
        label: "Finance",
        href: "/finance",
        icon: "wallet",
        permission: "finance.view",
        live: true,
      },
      {
        label: "Risk & fraud",
        href: "/risk",
        icon: "shield",
        permission: "risk.view",
        live: false,
      },
    ],
  },
  {
    title: "Growth",
    modules: [
      {
        label: "Marketing",
        href: "/marketing",
        icon: "bolt",
        permission: "marketing.view",
        live: true,
      },
      {
        label: "Content",
        href: "/content",
        icon: "grid",
        permission: "content.edit",
        live: true,
      },
      {
        label: "Search",
        href: "/search-tuning",
        icon: "search",
        permission: "search.manage",
        live: true,
      },
      {
        label: "Analytics",
        href: "/analytics",
        icon: "chart",
        permission: "analytics.view",
        live: false,
      },
    ],
  },
  {
    title: "Services",
    modules: [
      {
        label: "Warranty & repairs",
        href: "/warranty",
        icon: "wrench",
        permission: "aftersales.manage",
        live: false,
      },
      {
        label: "Trade-ins",
        href: "/trade-ins",
        icon: "swap",
        permission: "aftersales.manage",
        live: false,
      },
      {
        label: "Car sales",
        href: "/cars",
        icon: "car",
        permission: "cars.view",
        live: false,
      },
      {
        label: "Point of sale",
        href: "/pos",
        icon: "card",
        permission: "pos.sell",
        live: false,
      },
    ],
  },
  {
    title: "Organisation",
    modules: [
      {
        label: "HR & staff",
        href: "/hr",
        icon: "users",
        permission: "hr.view",
        live: true,
      },
      {
        label: "Admin console",
        href: "/admin",
        icon: "lock",
        permission: "admin.users",
        live: true,
      },
      {
        label: "IT tools",
        href: "/it",
        icon: "settings",
        permission: "it.flags",
        live: false,
      },
    ],
  },
];

/** The groups and modules a person may open, given their permission list. */
export function visibleGroups(permissions: readonly string[]): ModuleGroup[] {
  const has = new Set(permissions);
  return moduleGroups
    .map((g) => ({
      ...g,
      modules: g.modules.filter((m) => has.has(m.permission)),
    }))
    .filter((g) => g.modules.length > 0);
}

export const moduleByHref = (href: string) =>
  moduleGroups.flatMap((g) => g.modules).find((m) => m.href === href);
