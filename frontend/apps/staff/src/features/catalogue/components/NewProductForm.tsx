"use client";

import { useActionState } from "react";

import { createProduct, type CreateState } from "../actions";

type Category = {
  id: string;
  slug: string;
  name: string;
  parentId: string | null;
  kind: string;
};

function Field({
  state,
  name,
  label,
  children,
  hint,
}: {
  state: CreateState;
  name: string;
  label: string;
  children: React.ReactNode;
  hint?: string;
}) {
  const error = state.errors?.[name];
  return (
    <label className="field">
      <span className="field__label">{label}</span>
      {children}
      {hint ? <span className="field__hint">{hint}</span> : null}
      {error ? <span className="field__error">{error}</span> : null}
    </label>
  );
}

export function NewProductForm({ categories }: { categories: Category[] }) {
  const [state, action, pending] = useActionState(
    createProduct,
    {} as CreateState,
  );
  const parents = categories.filter((c) => !c.parentId && c.kind === "product");
  const inv = (n: string) => (state.errors?.[n] ? true : undefined);
  return (
    <form action={action} className="stack stack--lg" noValidate>
      {state.message ? (
        <p
          className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
          role="alert"
        >
          {state.message}
        </p>
      ) : null}
      <fieldset className="card stack">
        <legend className="field__label">Product</legend>
        <Field state={state} name="name" label="Name">
          <input
            name="name"
            className="input"
            required
            maxLength={150}
            aria-invalid={inv("name")}
          />
        </Field>
        <Field
          state={state}
          name="slug"
          label="Web address (optional)"
          hint="Made from the name if left empty, e.g. orbit-one-15"
        >
          <input
            name="slug"
            className="input"
            maxLength={80}
            aria-invalid={inv("slug")}
          />
        </Field>
        <Field state={state} name="categorySlug" label="Category">
          <select
            name="categorySlug"
            className="input"
            defaultValue=""
            required
            aria-invalid={inv("categorySlug")}
          >
            <option value="" disabled>
              Choose a category
            </option>
            {parents.map((p) => (
              <optgroup key={p.slug} label={p.name}>
                <option value={p.slug}>{p.name} (general)</option>
                {categories
                  .filter((c) => c.parentId === p.id)
                  .map((c) => (
                    <option key={c.slug} value={c.slug}>
                      {c.name}
                    </option>
                  ))}
              </optgroup>
            ))}
          </select>
        </Field>
        <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
          <Field state={state} name="brandName" label="Brand (optional)">
            <input name="brandName" className="input" maxLength={80} />
          </Field>
          <Field state={state} name="model" label="Model (optional)">
            <input name="model" className="input" maxLength={80} />
          </Field>
        </div>
        <Field state={state} name="description" label="Description (optional)">
          <textarea
            name="description"
            className="input"
            rows={4}
            maxLength={4000}
          />
        </Field>
        <Field state={state} name="status" label="Status">
          <select name="status" className="input" defaultValue="draft">
            <option value="draft">Draft (not on the site)</option>
            <option value="active">Active</option>
          </select>
        </Field>
      </fieldset>
      <fieldset className="card stack">
        <legend className="field__label">First variant</legend>
        <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
          <Field state={state} name="sku" label="SKU">
            <input
              name="sku"
              className="input"
              required
              maxLength={60}
              aria-invalid={inv("sku")}
            />
          </Field>
          <Field state={state} name="variantName" label="Variant name">
            <input
              name="variantName"
              className="input"
              placeholder="e.g. 256GB · Black"
              maxLength={100}
            />
          </Field>
          <Field
            state={state}
            name="weightGrams"
            label="Weight in grams (optional)"
            hint="Used for delivery fees and packing checks."
          >
            <input
              name="weightGrams"
              className="input"
              inputMode="numeric"
              aria-invalid={inv("weightGrams")}
            />
          </Field>
        </div>
        <label className="check">
          <input type="checkbox" name="isSerialised" /> Has an IMEI or serial
          number (picked by scanning)
        </label>
      </fieldset>
      <fieldset className="card stack">
        <legend className="field__label">TechShop’s listing (optional)</legend>
        <p className="text-sm text-muted">
          Add a price to sell it from TechShop’s own stock. Leave empty to list
          later.
        </p>
        <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
          <Field state={state} name="price" label="Price (₦)">
            <input
              name="price"
              className="input"
              inputMode="numeric"
              placeholder="e.g. 250000"
              aria-invalid={inv("price")}
            />
          </Field>
          <Field state={state} name="condition" label="Condition">
            <select name="condition" className="input" defaultValue="new">
              <option value="new">New</option>
              <option value="uk_used">UK-used</option>
              <option value="refurbished">Refurbished</option>
              <option value="open_box">Open box</option>
            </select>
          </Field>
          <Field state={state} name="warrantyMonths" label="Warranty (months)">
            <input
              name="warrantyMonths"
              className="input"
              inputMode="numeric"
              defaultValue="12"
              aria-invalid={inv("warrantyMonths")}
            />
          </Field>
        </div>
      </fieldset>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Creating…" : "Create product"}
      </button>
    </form>
  );
}
