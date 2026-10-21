import { connection } from "next/server";

import { Table, Status, humanise } from "@techshop/ui/components";
import { ActionButton } from "@/components/ActionButton";
import { PageHead } from "@/components/PageHead";
import { Tabs } from "@/components/Tabs";
import { publishPage } from "@/features/content/actions";
import { blocksToText } from "@/features/content/blocks";
import {
  BannerForm,
  HelpEditor,
  PageEditor,
} from "@/features/content/components/Editors";
import { can, requireStaff } from "@/features/session/server";
import { bff } from "@/lib/bff";
import { formatDateTime, param } from "@/lib/format";

export const metadata = { title: "Content | TechShop Staff" };

const TABS = [
  { key: "pages", label: "Pages" },
  { key: "help", label: "Help articles" },
  { key: "banners", label: "Banners" },
];
const SITES = ["market", "wholesale", "seller", "corporate"] as const;

export default async function ContentPage(props: PageProps<"/content">) {
  await connection();
  const me = await requireStaff("content.edit");
  const sp = await props.searchParams;
  const wanted = param(sp.tab);
  const tab = TABS.some((t) => t.key === wanted) ? wanted! : "pages";
  const site = SITES.find((s) => s === param(sp.site)) ?? "market";
  const edit = param(sp.edit);
  const api = await bff.api();
  const pages =
    tab === "pages"
      ? ((await api.get("/api/v1/staff/content/pages", {
          query: { site },
          init: { cache: "no-store" },
        })) ?? [])
      : [];
  const editing = pages.find((p) => p.id === edit);
  const banners =
    tab === "banners"
      ? ((await bff.publicApi().get("/api/v1/content/banners", {
          query: { site },
          init: { cache: "no-store" },
        })) ?? [])
      : [];
  const help =
    tab === "help" && site !== "corporate"
      ? ((await bff.publicApi().get("/api/v1/content/help/{site}", {
          path: { site },
          init: { cache: "no-store" },
        })) ?? [])
      : [];

  return (
    <div className="stack stack--lg">
      <PageHead
        title="Content"
        description="Pages, help articles and banners for every site. Text only: no HTML is stored."
      />
      <Tabs
        label="Content sections"
        current={tab}
        items={TABS.map((t) => ({
          ...t,
          href: `/content?tab=${t.key}&site=${site}`,
        }))}
      />
      <nav aria-label="Site" className="cluster">
        {SITES.map((s) => (
          <a
            key={s}
            href={`/content?tab=${tab}&site=${s}`}
            className={`btn btn--sm ${s === site ? "btn--inverse" : "btn--secondary"}`}
          >
            {humanise(s)}
          </a>
        ))}
      </nav>
      {tab === "pages" ? (
        <>
          <Table
            caption="Pages"
            rows={pages}
            rowKey={(p) => p.id}
            href={(p) => `/content?tab=pages&site=${site}&edit=${p.id}`}
            empty={{
              title: "No pages yet",
              body: "Create one below. Legal pages show a draft outline until they are published here.",
            }}
            columns={[
              { label: "Title", cell: (p) => p.title },
              { label: "Address", cell: (p) => `/${p.slug}` },
              {
                label: "Status",
                cell: (p) => (
                  <Status
                    label={humanise(p.status)}
                    tone={p.status === "published" ? "success" : "warning"}
                  />
                ),
              },
              { label: "Updated", cell: (p) => formatDateTime(p.updatedAt) },
              {
                label: "",
                cell: (p) =>
                  !can(me, "content.publish") ||
                  p.status === "published" ? null : p.updatedBy === me.id ? (
                    <span className="text-xs text-muted">
                      Someone else must publish your edit
                    </span>
                  ) : (
                    <ActionButton
                      label="Publish"
                      variant="inverse"
                      action={publishPage.bind(null, p.id)}
                    />
                  ),
              },
            ]}
          />
          <section
            className="card card--roomy stack"
            aria-labelledby="editor-title"
          >
            <h2 id="editor-title" className="text-xl">
              {editing ? `Edit “${editing.title}”` : "New page"}
            </h2>
            <PageEditor
              key={editing?.id ?? "new"}
              initial={
                editing
                  ? {
                      site: editing.site,
                      slug: editing.slug,
                      title: editing.title,
                      body: blocksToText(editing.body),
                    }
                  : undefined
              }
            />
          </section>
        </>
      ) : null}
      {tab === "help" ? (
        <>
          <Table
            caption="Published help articles"
            rows={help}
            rowKey={(h) => h.slug}
            empty={{
              title: "No published articles",
              body:
                site === "corporate"
                  ? "The corporate site has no help centre."
                  : "Write one below.",
            }}
            columns={[
              { label: "Article", cell: (h) => h.title },
              { label: "Topic", cell: (h) => h.topic },
              {
                label: "Published",
                cell: (h) => formatDateTime(h.publishedAt),
              },
            ]}
          />
          <section
            className="card card--roomy stack"
            aria-labelledby="help-title"
          >
            <h2 id="help-title" className="text-xl">
              Write or update an article
            </h2>
            <p className="text-sm text-muted">
              Saving to an existing web address replaces that article.
            </p>
            <HelpEditor />
          </section>
        </>
      ) : null}
      {tab === "banners" ? (
        <>
          <Table
            caption="Live banners"
            rows={banners}
            rowKey={(b) => b.id}
            empty={{
              title: "No live banners",
              body: "Only live banners are listed.",
            }}
            columns={[
              { label: "Banner", cell: (b) => b.title },
              { label: "Placement", cell: (b) => b.placement },
              { label: "Links to", cell: (b) => b.linkUrl },
              { label: "Ends", cell: (b) => formatDateTime(b.endsAt) },
            ]}
          />
          <section
            className="card card--roomy stack"
            aria-labelledby="banner-title"
          >
            <h2 id="banner-title" className="text-xl">
              New banner
            </h2>
            <BannerForm />
          </section>
        </>
      ) : null}
    </div>
  );
}
