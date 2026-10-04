# Frontend & E2E Testing

## Component tests
- React Testing Library + Vitest/Jest. Query by role/label (`getByRole('button', { name: /save/i })`), not test IDs or class names.
- Use `user-event` for interactions; assert visible outcomes.
- Mock network with `msw`.

## Next.js
- Test Server Actions and route handlers as functions (auth, validation, authZ).
- Avoid testing framework internals.

## E2E
- Playwright (preferred) or Cypress.
- Cover critical journeys only: signup/login, core create/edit flow, checkout/payment (sandbox), permissions boundaries.
- Seed data via API/DB helpers, not by clicking through the UI.
- Run against a production-like build; record traces on failure.

## Accessibility
- `@axe-core/playwright` or `jest-axe` checks on key pages.

## Visual regression (optional)
- Playwright screenshots for design-critical components.
