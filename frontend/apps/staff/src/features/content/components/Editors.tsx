"use client";

import { useActionState } from "react";

import { createBanner, saveHelp, savePage, type EditorState } from "../actions";

function Feedback({ state }: { state: EditorState }) {
  return state.message ? (
    <p
      className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
      role="alert"
    >
      {state.message}
    </p>
  ) : null;
}

const BODY_HINT =
  "Blank line between paragraphs. Start a line with “## ” for a heading, “- ” for a list item.";

export function PageEditor({
  initial,
}: {
  initial?: { site: string; slug: string; title: string; body: string };
}) {
  const [state, action, pending] = useActionState(savePage, {} as EditorState);
  return (
    <form action={action} className="stack" noValidate>
      <Feedback state={state} />
      <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
        <label className="field">
          <span className="field__label">Site</span>
          <select
            name="site"
            className="input"
            defaultValue={initial?.site ?? "market"}
          >
            <option value="market">Market</option>
            <option value="wholesale">Wholesale</option>
            <option value="seller">Seller Centre</option>
            <option value="corporate">Corporate</option>
          </select>
        </label>
        <label className="field">
          <span className="field__label">Web address</span>
          <input
            name="slug"
            className="input"
            defaultValue={initial?.slug}
            placeholder="e.g. terms"
            required
          />
        </label>
      </div>
      <label className="field">
        <span className="field__label">Title</span>
        <input
          name="title"
          className="input"
          defaultValue={initial?.title}
          required
        />
      </label>
      <label className="field">
        <span className="field__label">Text</span>
        <textarea
          name="body"
          className="input"
          rows={10}
          defaultValue={initial?.body}
          required
        />
        <span className="field__hint">{BODY_HINT}</span>
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Saving…" : "Save draft"}
      </button>
    </form>
  );
}

export function HelpEditor() {
  const [state, action, pending] = useActionState(saveHelp, {} as EditorState);
  return (
    <form action={action} className="stack" noValidate>
      <Feedback state={state} />
      <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
        <label className="field">
          <span className="field__label">Site</span>
          <select name="site" className="input">
            <option value="market">Market</option>
            <option value="wholesale">Wholesale</option>
            <option value="seller">Seller Centre</option>
          </select>
        </label>
        <label className="field">
          <span className="field__label">Web address</span>
          <input
            name="slug"
            className="input"
            placeholder="e.g. returns"
            required
          />
        </label>
        <label className="field">
          <span className="field__label">Topic</span>
          <input
            name="topic"
            className="input"
            placeholder="e.g. Orders"
            required
          />
        </label>
        <label className="field">
          <span className="field__label">Status</span>
          <select name="status" className="input">
            <option value="draft">Draft</option>
            <option value="in_review">In review</option>
            <option value="published">Published</option>
          </select>
        </label>
      </div>
      <label className="field">
        <span className="field__label">Question or title</span>
        <input name="title" className="input" required />
      </label>
      <label className="field">
        <span className="field__label">Answer</span>
        <textarea name="body" className="input" rows={8} required />
        <span className="field__hint">{BODY_HINT}</span>
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Saving…" : "Save article"}
      </button>
    </form>
  );
}

export function BannerForm() {
  const [state, action, pending] = useActionState(
    createBanner,
    {} as EditorState,
  );
  return (
    <form action={action} className="stack" noValidate>
      <Feedback state={state} />
      <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
        <label className="field">
          <span className="field__label">Site</span>
          <select name="site" className="input">
            <option value="market">Market</option>
            <option value="wholesale">Wholesale</option>
            <option value="seller">Seller Centre</option>
            <option value="corporate">Corporate</option>
          </select>
        </label>
        <label className="field">
          <span className="field__label">Placement</span>
          <input name="placement" className="input" defaultValue="home_hero" />
        </label>
        <label className="field">
          <span className="field__label">Status</span>
          <select name="status" className="input">
            <option value="draft">Draft</option>
            <option value="live">Live</option>
          </select>
        </label>
      </div>
      <label className="field">
        <span className="field__label">Title</span>
        <input name="title" className="input" required />
      </label>
      <label className="field">
        <span className="field__label">Link</span>
        <input name="linkUrl" className="input" placeholder="/deals" required />
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Saving…" : "Save banner"}
      </button>
    </form>
  );
}
