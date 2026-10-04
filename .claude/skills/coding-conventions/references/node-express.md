# Node.js / Express Conventions

## Layout
```
src/
  app.ts                  # express app, middleware, routes
  server.ts               # http server + graceful shutdown
  config/                 # env parsing with zod
  middleware/             # auth, tenant, error handler, rate limit, request id
  modules/
    <feature>/
      <feature>.routes.ts
      <feature>.controller.ts
      <feature>.service.ts
      <feature>.repository.ts
      <feature>.schema.ts     # zod schemas (request validation + types)
      <feature>.types.ts
  lib/                    # db, redis, logger, mailer
  utils/
tests/
```

## Style
- TypeScript preferred, `strict: true`. If the project is JS, use JSDoc types and match it.
- ESLint + Prettier; no unused vars; `async/await` over raw promises/callbacks.
- ES modules or CommonJS — match the project, don't mix.

## Validation
```ts
export const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
});
export type CreateUserInput = z.infer<typeof createUserSchema>;
```
- Validate body, params, query in a `validate(schema)` middleware. Pass only parsed data downstream (strips unknown fields).

## Errors
- Custom `AppError` class with `statusCode`, `code`, `message`; subclasses `NotFoundError`, `ValidationError`, `ForbiddenError`, `ConflictError`.
- Controllers wrapped by an `asyncHandler` (or Express 5 native async support).
- One central error middleware: logs unexpected errors, returns the standard error envelope.

## Controllers
```ts
export const create = asyncHandler(async (req, res) => {
  const user = await userService.create(req.tenantId, req.body);
  res.status(201).json({ data: toUserDto(user) });
});
```
- Never return DB documents/rows directly; map to DTOs.

## App setup
- `helmet()`, `cors({ origin: allowList, credentials: true })`, `express.json({ limit: '1mb' })`.
- `app.set('trust proxy', 1)` behind Nginx/Render so rate limiting sees real IPs.
- Disable `x-powered-by`.

## Logging
- `pino` (or the project's logger) with request ID; redact `authorization`, `password`, `token` fields.

## DB
- PostgreSQL: Prisma, Drizzle, Knex or `pg` — parameterized queries only.
- MongoDB: Mongoose with schema validation; sanitize operators from input (`express-mongo-sanitize` or zod).

## Async & process
- Handle `unhandledRejection` and `uncaughtException` (log + exit for a supervisor to restart).
- Graceful shutdown: stop accepting connections, close DB/Redis.
