import { Tabs } from "@/components/Tabs";

/** Sections of the risk module. */
export function RiskNav({
  current,
}: {
  current: "cases" | "devices" | "rules";
}) {
  return (
    <Tabs
      label="Risk sections"
      current={current}
      items={[
        { key: "cases", label: "Cases", href: "/risk" },
        { key: "devices", label: "Device blocklist", href: "/risk/devices" },
        { key: "rules", label: "Rules", href: "/risk/rules" },
      ]}
    />
  );
}
