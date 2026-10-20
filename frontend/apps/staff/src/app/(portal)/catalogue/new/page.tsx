import { connection } from "next/server";

import { PageHead } from "@/components/PageHead";
import { NewProductForm } from "@/features/catalogue/components/NewProductForm";
import { categories } from "@/features/catalogue/server";
import { requireStaff } from "@/features/session/server";

export const metadata = { title: "Add a product | TechShop Staff" };

export default async function NewProductPage() {
  await connection();
  await requireStaff("catalog.edit");
  const cats = (await categories()).map((c) => ({
    id: c.id,
    slug: c.slug,
    name: c.name,
    parentId: c.parentId || null,
    kind: c.kind,
  }));
  return (
    <div className="stack stack--lg">
      <PageHead
        title="Add a product"
        description="Creates the product and its first variant. Stock is received in Inventory."
      />
      <NewProductForm categories={cats} />
    </div>
  );
}
