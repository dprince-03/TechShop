import { connection } from "next/server";

import { Status, Table, humanise } from "@techshop/ui/components";

import { PageHead } from "@/components/PageHead";
import {
  BlockDeviceForm,
  UnblockControl,
} from "@/features/risk/components/RiskControls";
import { RiskNav } from "@/features/risk/components/RiskNav";
import { blockedDevices, checkImei } from "@/features/risk/server";
import { can, requireStaff } from "@/features/session/server";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Device blocklist | TechShop Staff" };

export default async function DevicesPage(props: PageProps<"/risk/devices">) {
  await connection();
  const me = await requireStaff("risk.view");
  const imei = param((await props.searchParams).imei)
    ?.replace(/\s+/g, "")
    .slice(0, 20);
  const [check, rows] = await Promise.all([
    imei ? checkImei(imei).catch(() => null) : Promise.resolve(null),
    blockedDevices(),
  ]);
  const canBlock = can(me, "devices.block");
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Device blocklist"
        description="Stolen, fraudulent or counterfeit IMEIs can't be stocked, sold, traded in, returned or repaired."
      />
      <RiskNav current="devices" />
      <section className="card stack" aria-labelledby="imei-check">
        <h2 id="imei-check" className="text-xl">
          Check an IMEI
        </h2>
        <form className="cluster" role="search">
          <label htmlFor="imei" className="visually-hidden">
            IMEI
          </label>
          <input
            id="imei"
            name="imei"
            defaultValue={imei}
            className="input"
            inputMode="numeric"
            maxLength={20}
            placeholder="15-digit IMEI"
          />
          <button className="btn btn--secondary">Check</button>
        </form>
        {check ? (
          <p role="status" className="cluster">
            <Status
              label={
                check.luhnOk
                  ? "Check digit valid"
                  : "Check digit wrong: typo or fake IMEI"
              }
              tone={check.luhnOk ? "success" : "danger"}
            />
            <Status
              label={
                check.blocked
                  ? `Blocked: ${humanise(check.reason ?? "")}`
                  : "Not on the blocklist"
              }
              tone={check.blocked ? "danger" : "success"}
            />
          </p>
        ) : null}
      </section>
      {canBlock ? (
        <section className="card stack" aria-labelledby="imei-block">
          <h2 id="imei-block" className="text-xl">
            Block a device
          </h2>
          <BlockDeviceForm />
        </section>
      ) : null}
      <Table
        caption="Blocked devices"
        rows={rows ?? []}
        rowKey={(d) => d.imei}
        empty={{
          title: "No blocked devices",
          body: "Devices are added from reports and confirmed fraud.",
        }}
        columns={[
          {
            label: "IMEI",
            cell: (d) => <span className="tabular-nums">{d.imei}</span>,
          },
          { label: "Reason", cell: (d) => humanise(d.reason) },
          {
            label: "Reported by",
            cell: (d) => `${d.source}${d.reference ? ` · ${d.reference}` : ""}`,
          },
          {
            label: "Added",
            cell: (d) =>
              `${formatDateTime(d.createdAt)}${d.addedBy ? ` by ${d.addedBy}` : ""}`,
          },
          ...(canBlock
            ? [
                {
                  label: "",
                  cell: (d: { imei: string }) => (
                    <UnblockControl imei={d.imei} />
                  ),
                },
              ]
            : []),
        ]}
      />
    </div>
  );
}
