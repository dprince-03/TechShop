import { connection } from "next/server";

import { ActionButton } from "@/components/ActionButton";
import { Figures } from "@/components/Figures";
import { PageHead } from "@/components/PageHead";
import { humanise } from "@/components/Status";
import { Table } from "@/components/Table";
import { deleteSynonym } from "@/features/search-tuning/actions";
import { SynonymForm } from "@/features/search-tuning/components/SynonymForm";
import { requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";

export const metadata = { title: "Search | TechShop Staff" };

export default async function SearchTuningPage() {
  await connection();
  await requireStaff("search.manage");
  const api = await bff.api();
  const noStore = { cache: "no-store" } as const;
  const [insights, synonyms] = await Promise.all([
    api.get("/api/v1/staff/search/insights", { init: noStore }),
    api.get("/api/v1/staff/search/synonyms", { init: noStore }),
  ]);
  const s = insights.stats;
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Search"
        description="What shoppers search for, what finds nothing, and the synonyms that help."
      />
      <Figures
        items={[
          {
            label: "Searches (30 days)",
            value: s.searches.toLocaleString("en-NG"),
          },
          {
            label: "Found nothing",
            value: `${(s.zeroRate * 100).toFixed(1)}%`,
          },
          {
            label: "Slowest 5% took",
            value: `${Math.round(s.p95LatencyMs)} ms`,
          },
        ]}
      />
      <div className="grid" style={{ ["--grid-min" as string]: "24rem" }}>
        <section className="stack" aria-labelledby="top-title">
          <h2 id="top-title" className="text-xl">
            Top searches
          </h2>
          <Table
            caption="Top searches"
            rows={insights.top ?? []}
            rowKey={(r) => r.queryNorm}
            empty={{
              title: "No searches yet",
              body: "Searches appear as shoppers use the site.",
            }}
            columns={[
              { label: "Search", cell: (r) => r.queryNorm },
              { label: "Times", cell: (r) => r.searches, numeric: true },
              {
                label: "Avg results",
                cell: (r) => Math.round(r.avgResults),
                numeric: true,
              },
            ]}
          />
        </section>
        <section className="stack" aria-labelledby="zero-title">
          <h2 id="zero-title" className="text-xl">
            Searches that found nothing
          </h2>
          <Table
            caption="Zero-result searches"
            rows={insights.zeroResults ?? []}
            rowKey={(r) => r.queryNorm}
            empty={{ title: "None", body: "Every search found something." }}
            columns={[
              { label: "Search", cell: (r) => r.queryNorm },
              { label: "Times", cell: (r) => r.searches, numeric: true },
            ]}
          />
        </section>
      </div>
      {(insights.gaps ?? []).length ? (
        <section className="stack" aria-labelledby="gaps-title">
          <h2 id="gaps-title" className="text-xl">
            Catalogue gaps
          </h2>
          <Table
            caption="Catalogue gaps"
            rows={insights.gaps!}
            rowKey={(g) => g.queryNorm}
            columns={[
              { label: "Search", cell: (g) => g.queryNorm },
              {
                label: "Searches (30 days)",
                cell: (g) => g.searches30d,
                numeric: true,
              },
              { label: "Status", cell: (g) => humanise(g.status) },
              { label: "Note", cell: (g) => g.note || "—" },
            ]}
          />
        </section>
      ) : null}
      <section className="stack" aria-labelledby="syn-title">
        <h2 id="syn-title" className="text-xl">
          Synonyms
        </h2>
        <Table
          caption="Synonyms"
          rows={synonyms ?? []}
          rowKey={(x) => x.id}
          empty={{
            title: "No synonyms",
            body: "Add words shoppers use that mean the same thing.",
          }}
          columns={[
            { label: "Word", cell: (x) => x.target },
            { label: "Also", cell: (x) => (x.terms ?? []).join(", ") },
            {
              label: "Type",
              cell: (x) => (x.kind === "two_way" ? "Both ways" : "One way"),
            },
            {
              label: "",
              cell: (x) => (
                <ActionButton
                  label="Remove"
                  variant="ghost"
                  action={deleteSynonym.bind(null, x.id)}
                />
              ),
            },
          ]}
        />
        <div className="card card--roomy">
          <SynonymForm />
        </div>
      </section>
    </div>
  );
}
