import type { Metadata } from "next";

import { categories } from "@techshop/fixtures";
import { Breadcrumbs, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = { title: "Add a product | TechShop Staff" };

export default function NewProductPage() {
  return (
    <div className="stack stack--lg">
      <Breadcrumbs
        items={[
          { label: "Dashboard", href: "/" },
          { label: "Catalogue", href: "/catalogue" },
          { label: "Add a product" },
        ]}
      />
      <h1 className="text-xl">Add a product</h1>
      <div className="card card--roomy">
        <PreviewForm
          id="product"
          groups={[
            {
              legend: "Basics",
              fields: [
                { name: "name", label: "Product name", required: true },
                { name: "brand", label: "Brand", required: true, half: true },
                {
                  name: "category",
                  label: "Category",
                  type: "select",
                  required: true,
                  options: categories
                    .filter((c) => c.slug !== "cars")
                    .map((c) => c.name),
                  half: true,
                },
                {
                  name: "condition",
                  label: "Condition",
                  type: "select",
                  required: true,
                  options: ["New", "UK-used", "Refurbished"],
                  half: true,
                },
                { name: "sku", label: "SKU", half: true },
                {
                  name: "spec",
                  label: "Key specs",
                  placeholder: "e.g. 12GB RAM · 256GB",
                  hint: "Shown on product cards. Separate with “·”.",
                },
              ],
            },
            {
              legend: "Price & stock",
              fields: [
                {
                  name: "price",
                  label: "Price (₦)",
                  inputMode: "numeric",
                  required: true,
                  half: true,
                },
                {
                  name: "compare",
                  label: "Previous price (₦)",
                  inputMode: "numeric",
                  half: true,
                  hint: "Only if genuinely discounted.",
                },
                {
                  name: "stock",
                  label: "Opening stock",
                  inputMode: "numeric",
                  required: true,
                  half: true,
                },
                {
                  name: "warranty",
                  label: "Warranty",
                  type: "select",
                  required: true,
                  options: [
                    "12 months manufacturer",
                    "6 months TechShop",
                    "3 months",
                    "None",
                  ],
                  half: true,
                },
              ],
            },
            {
              legend: "Media",
              fields: [
                {
                  name: "photos",
                  label: "Product photos",
                  type: "file",
                  accept: "image/*",
                  hint: "At least one photo of the actual product on a plain background.",
                },
              ],
            },
          ]}
          submitLabel="Save product"
          successTitle="Product ready"
          successMessage="In the live portal the product would be saved as a draft for review."
        />
      </div>
    </div>
  );
}
