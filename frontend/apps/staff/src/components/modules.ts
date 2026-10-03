import type { IconName } from "@techshop/ui/components";

export type StaffModule = { label: string; href: string; icon: IconName };
export type ModuleGroup = { title: string; modules: StaffModule[] };

/**
 * All staff portal modules, grouped as they appear in the sidebar.
 * Access will be filtered by role once RBAC exists (see docs/plan.md).
 */
export const moduleGroups: ModuleGroup[] = [
  {
    title: "Operations",
    modules: [
      { label: "Orders", href: "/orders", icon: "receipt" },
      { label: "Catalogue", href: "/catalogue", icon: "tag" },
      { label: "Inventory", href: "/inventory", icon: "box" },
      { label: "Dispatch", href: "/dispatch", icon: "truck" },
      { label: "Customer support", href: "/support", icon: "support" },
    ],
  },
  {
    title: "Marketplace & B2B",
    modules: [
      { label: "Vendors", href: "/vendors", icon: "store" },
      { label: "B2B accounts", href: "/b2b", icon: "briefcase" },
      { label: "Purchasing", href: "/purchasing", icon: "file" },
    ],
  },
  {
    title: "Finance & risk",
    modules: [
      { label: "Finance", href: "/finance", icon: "wallet" },
      { label: "Risk & fraud", href: "/risk", icon: "shield" },
    ],
  },
  {
    title: "Growth",
    modules: [
      { label: "Marketing", href: "/marketing", icon: "bolt" },
      { label: "Analytics", href: "/analytics", icon: "chart" },
      { label: "Content", href: "/content", icon: "grid" },
    ],
  },
  {
    title: "Services",
    modules: [
      { label: "Warranty & repairs", href: "/warranty", icon: "wrench" },
      { label: "Trade-ins", href: "/trade-ins", icon: "swap" },
      { label: "Car sales", href: "/cars", icon: "car" },
      { label: "Point of sale", href: "/pos", icon: "card" },
    ],
  },
  {
    title: "Organisation",
    modules: [
      { label: "HR & staff", href: "/hr", icon: "users" },
      { label: "Admin console", href: "/admin", icon: "lock" },
      { label: "IT tools", href: "/it", icon: "settings" },
    ],
  },
];
