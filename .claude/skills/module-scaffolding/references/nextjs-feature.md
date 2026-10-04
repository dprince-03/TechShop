# Next.js Full-Stack Feature Template

```
app/(dashboard)/invoices/
  page.tsx               # Server Component: list (fetch via server/invoices)
  [id]/page.tsx          # detail
  new/page.tsx           # form page
  actions.ts             # 'use server' actions: create/update/delete
components/invoices/
  invoice-table.tsx
  invoice-form.tsx       # 'use client', RHF + zod
server/invoices/
  queries.ts             # import 'server-only'; tenant-scoped reads
  mutations.ts           # tenant-scoped writes
lib/validation/invoice.ts  # shared zod schemas
```

Server Action pattern:
```ts
'use server';
export async function createInvoice(formData: unknown) {
  const session = await requireSession();            // auth
  assertCan(session, 'invoices:write');               // authZ
  const input = createInvoiceSchema.parse(formData);  // validation
  const inv = await mutations.create(session.tenantId, input);
  revalidatePath('/invoices');
  return { ok: true, id: inv.id };
}
```
Never rely on middleware alone for auth; re-check in actions and route handlers.
