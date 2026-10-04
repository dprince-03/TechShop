---
name: testing-standards
description: Write, structure, and review automated tests — unit, integration, API, end-to-end, and contract tests — for Go, Node.js/TypeScript, React/Next.js, and Java projects, including fixtures, mocking strategy, test databases, coverage goals, and CI setup. Use this skill whenever the user asks to add tests, write test cases, fix failing tests, improve coverage, set up a test framework, or whenever new features or bug fixes are written that should ship with tests.
---

# Testing Standards

## Principles

- **Test behaviour, not implementation.** Assert on outputs and side effects users care about, so refactors don't break tests needlessly.
- **Testing pyramid:** many fast unit tests, a solid layer of integration/API tests, a few critical-path E2E tests.
- **Real dependencies where cheap:** use a real PostgreSQL/Mongo/Redis in containers for integration tests instead of mocking the DB; mock only external third-party services (payments, email, SMS).
- **Deterministic:** no reliance on wall clock, random values, test order, or network. Inject clocks and ID generators.
- **Every bug fix gets a regression test** that fails before the fix.
- **Match the project:** reuse existing frameworks, helpers, factories, and naming. Read `references/<stack>.md` for defaults.

## What to cover for a feature

1. Happy path
2. Validation errors (missing/invalid fields, boundaries: 0, negative, max length, empty)
3. Not found
4. **Authorization**: unauthenticated (401), wrong role (403), **other tenant's resource** (404/403)
5. Conflicts / duplicates
6. Edge cases in business rules (limits, state transitions, concurrency where relevant)
7. Idempotency for payment/order endpoints (same key twice → one effect)

## Naming

- Describe behaviour: `returns 404 when invoice belongs to another tenant`, `TestCreateInvoice_RejectsNegativeAmount`.
- Arrange / Act / Assert sections, one logical assertion focus per test.

## Test data

- Factories/builders (`makeUser({ role: 'admin' })`) over large shared fixtures.
- Each test creates what it needs; clean up via transactions rolled back per test or truncation between tests.
- Never use production data.

## Coverage

- Targets: ~80% lines on services/business logic; don't chase 100% on glue code.
- Critical modules (auth, payments, permissions, tenant scoping) should be close to fully covered.

## CI

- Unit tests on every push; integration tests on every PR (with service containers); E2E on merge to main or nightly.
- Fail the build on test failure; publish coverage report.
- Flaky tests are bugs: quarantine and fix, don't retry forever.

## Stack references

- Go → `references/go.md`
- Node/TypeScript backend → `references/node.md`
- React/Next.js frontend & E2E → `references/frontend-e2e.md`
- Java → `references/java.md`
