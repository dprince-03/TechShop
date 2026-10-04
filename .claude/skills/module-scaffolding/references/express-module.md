# Express Module Template (TypeScript)

`src/modules/invoices/`:

```ts
// invoices.schema.ts
export const createInvoiceSchema = z.object({
  number: z.string().max(50),
  amountMinor: z.number().int().positive(),
  currency: z.string().length(3),
}).strict();
export const updateInvoiceSchema = createInvoiceSchema.partial();
export const listQuerySchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
});
```

```ts
// invoices.repository.ts — every query scoped
export const invoiceRepo = {
  findById: (tenantId: string, id: string) =>
    db.invoice.findFirst({ where: { id, tenantId, deletedAt: null } }),
  list: (tenantId: string, { page, limit }: ListQuery) => /* paginated + count */,
  create: (data: NewInvoice) => db.invoice.create({ data }),
  update: (tenantId: string, id: string, data: Partial<NewInvoice>) =>
    db.invoice.updateMany({ where: { id, tenantId, deletedAt: null }, data }),
  softDelete: (tenantId: string, id: string) =>
    db.invoice.updateMany({ where: { id, tenantId }, data: { deletedAt: new Date() } }),
};
```

```ts
// invoices.service.ts
export async function createInvoice(actor: Actor, input: CreateInvoiceInput) {
  assertCan(actor, 'invoices:write');
  const invoice = await invoiceRepo.create({ ...input, tenantId: actor.tenantId });
  await audit.log(actor, 'invoice.create', invoice.id);
  return invoice;
}
```

```ts
// invoices.routes.ts
const router = Router();
router.use(requireAuth, resolveTenant);
router.get('/', requirePerm('invoices:read'), validate({ query: listQuerySchema }), controller.list);
router.get('/:id', requirePerm('invoices:read'), controller.get);
router.post('/', requirePerm('invoices:write'), validate({ body: createInvoiceSchema }), controller.create);
router.patch('/:id', requirePerm('invoices:write'), validate({ body: updateInvoiceSchema }), controller.update);
router.delete('/:id', requirePerm('invoices:delete'), controller.remove);
export default router;
```

Tests: Vitest/Jest + Supertest against a test DB; include cross-tenant and forbidden cases.
