import { legalDocs, type LegalDoc } from "./legalOutlines";
import { PageHeader } from "./PageHeader";
import { Prose } from "./Prose";
import styles from "./LegalDocument.module.css";

/** Legal page shell. Shows an outline clearly marked as a draft until counsel provides the text. */
export function LegalDocument({ doc }: { doc: LegalDoc }) {
  const d = legalDocs[doc];
  return (
    <>
      <PageHeader
        title={d.title}
        lead={d.summary}
        breadcrumbs={[
          { label: "Home", href: "/" },
          { label: "Legal" },
          { label: d.title },
        ]}
      />
      <p className={styles.draft} role="note">
        <strong>Draft outline.</strong> This page lists what the final document
        must cover. The legal text will be written by TechShop’s legal counsel
        before launch. Nothing here is a binding term.
      </p>
      <Prose>
        <nav aria-label="On this page">
          <ol>
            {d.sections.map((s, i) => (
              <li key={s.heading}>
                <a href={`#s${i + 1}`}>{s.heading}</a>
              </li>
            ))}
          </ol>
        </nav>
        {d.sections.map((s, i) => (
          <section key={s.heading} aria-labelledby={`s${i + 1}`}>
            <h2 id={`s${i + 1}`}>
              {i + 1}. {s.heading}
            </h2>
            <p>This section will cover:</p>
            <ul>
              {s.covers.map((c) => (
                <li key={c}>{c}</li>
              ))}
            </ul>
          </section>
        ))}
      </Prose>
    </>
  );
}

export { legalDocs, type LegalDoc };
