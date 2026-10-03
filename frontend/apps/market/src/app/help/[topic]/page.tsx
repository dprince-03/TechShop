import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import {
  FaqList,
  PageHeader,
  PreviewForm,
  Prose,
} from "@techshop/ui/components";

import { getHelpTopic, helpTopics } from "@/lib/help";

import styles from "../help.module.css";

const TOPICS = [...helpTopics.map((t) => t.slug), "contact"];

export const dynamicParams = false;

export function generateStaticParams() {
  return TOPICS.map((topic) => ({ topic }));
}

export async function generateMetadata(
  props: PageProps<"/help/[topic]">,
): Promise<Metadata> {
  const { topic } = await props.params;
  const t = getHelpTopic(topic);
  return { title: `${t?.title ?? "Contact us"} | TechShop Help` };
}

function TopicNav({ current }: { current: string }) {
  return (
    <nav aria-label="Help topics" className={styles.aside}>
      {[
        ...helpTopics.map((t) => ({ slug: t.slug, title: t.title })),
        { slug: "contact", title: "Contact us" },
      ].map((t) => (
        <Link
          key={t.slug}
          href={`/help/${t.slug}`}
          className={styles.asideLink}
          aria-current={t.slug === current ? "page" : undefined}
        >
          {t.title}
        </Link>
      ))}
    </nav>
  );
}

export default async function HelpTopicPage(props: PageProps<"/help/[topic]">) {
  const { topic } = await props.params;
  if (!TOPICS.includes(topic)) notFound();
  const t = getHelpTopic(topic);
  const crumbs = [
    { label: "Home", href: "/" },
    { label: "Help", href: "/help" },
    { label: t?.title ?? "Contact us" },
  ];

  return (
    <main id="main" className="section section--page">
      <div className="container">
        {t ? (
          <PageHeader title={t.title} lead={t.summary} breadcrumbs={crumbs} />
        ) : (
          <PageHeader
            title="Contact us"
            lead="Send us a message and we’ll reply by email or phone."
            breadcrumbs={crumbs}
          />
        )}
        <div className={styles.layout}>
          <div className="stack stack--lg">
            {t ? (
              <>
                <Prose>
                  {t.sections.map((s) => (
                    <section key={s.heading}>
                      <h2>{s.heading}</h2>
                      {s.body.map((p) => (
                        <p key={p}>{p}</p>
                      ))}
                    </section>
                  ))}
                </Prose>
                {t.faqs ? <FaqList items={t.faqs} /> : null}
                <p className="text-xs text-muted">
                  Sample help content — exact policies will be confirmed before
                  launch.
                </p>
              </>
            ) : (
              <div className="card card--roomy">
                <PreviewForm
                  id="contact"
                  groups={[
                    {
                      fields: [
                        {
                          name: "name",
                          label: "Full name",
                          required: true,
                          autoComplete: "name",
                          half: true,
                        },
                        {
                          name: "email",
                          label: "Email address",
                          type: "email",
                          required: true,
                          autoComplete: "email",
                          half: true,
                        },
                        {
                          name: "phone",
                          label: "Phone number",
                          type: "tel",
                          autoComplete: "tel",
                          half: true,
                        },
                        {
                          name: "order",
                          label: "Order number",
                          placeholder: "TS-10482",
                          half: true,
                        },
                        {
                          name: "topic",
                          label: "Topic",
                          type: "select",
                          required: true,
                          options: [
                            "Delivery",
                            "Returns & refunds",
                            "Warranty",
                            "Payments",
                            "My account",
                            "Something else",
                          ],
                        },
                        {
                          name: "message",
                          label: "How can we help?",
                          type: "textarea",
                          required: true,
                        },
                      ],
                    },
                  ]}
                  submitLabel="Send message"
                  successTitle="Message ready"
                  successMessage="In the live store this would open a support ticket and you’d get a reference by email."
                />
              </div>
            )}
          </div>
          <TopicNav current={topic} />
        </div>
      </div>
    </main>
  );
}
