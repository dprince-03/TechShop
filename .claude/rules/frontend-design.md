---
paths:
  - "frontend/**"
  - "mobile/**"
  - "shared/design-tokens/**"
---

# Frontend design rules: UI/UX laws and principles

These rules apply to every web and mobile interface in TechShop. The *how* (tokens, CSS primitives, file locations, reference research) lives in the `frontend-design` skill; load it before UI work. This file covers the *why*: the laws and principles every screen must satisfy. When a rule and a deadline conflict, the rule wins. When two rules conflict, accessibility wins, then clarity, then aesthetics.

## 1. Laws of UX (apply on every screen)

| Law | What it says | How we apply it |
|---|---|---|
| **Jakob's Law** | Users spend most of their time on other sites and expect yours to work the same way. | Use familiar commerce conventions: logo top-left links home, search at the top, cart top-right with a count, product cards ordered image → title → rating → price. Don't reinvent checkout. |
| **Fitts's Law** | Time to hit a target depends on its size and distance. | Primary actions are large (≥ 44px; `--control-lg` for buy/checkout) and close to where the eye already is. On mobile, put key actions in thumb reach (bottom of the screen). Keep destructive actions small and away from primary ones. |
| **Hick's Law** | Decision time grows with the number and complexity of choices. | Limit top-level navigation to about 7 items, and group the rest. Offer one primary CTA per view. Use progressive disclosure: filters collapse, specs expand, mega-menus reveal on demand. |
| **Miller's Law** | Working memory holds about 7 ± 2 chunks. | Chunk content: group specs, split long forms into steps, format numbers (`₦1,250,000`, phone `0803 123 4567`). |
| **Tesler's Law** | Complexity can't be removed, only moved. | The system absorbs the complexity: autofill addresses, remember choices, compute delivery fees, pre-select sensible defaults. Never push work onto the user that the system could do. |
| **Doherty Threshold** | Productivity soars when responses are under 400ms. | Give feedback within 100ms (pressed states, optimistic UI). Use skeletons for loads over 400ms and progress for anything over 1s. Never leave a blank screen. |
| **Aesthetic–Usability Effect** | Users perceive attractive designs as easier to use. | Polish matters: consistent spacing, alignment and imagery. But beauty never hides problems; usability testing still decides. |
| **Von Restorff (Isolation) Effect** | The item that differs is the one remembered. | Reserve contrast for what matters: the primary button, the sale badge, the error. If everything stands out, nothing does, so use sale colour sparingly. |
| **Serial Position Effect** | Users remember the first and last items best. | Put the most important nav items first and last (e.g. Search first, Cart last). Put the key info at the start and end of lists and pages. |
| **Law of Proximity** (Gestalt) | Items close together are seen as related. | Spacing defines grouping: tight gaps inside a group, larger gaps between groups. Labels sit closer to their own field than to the next one. |
| **Law of Similarity** (Gestalt) | Similar-looking items are seen as related. | Same function means same appearance everywhere. Links always look like links, and a sale badge always looks the same. |
| **Law of Common Region** (Gestalt) | Items inside a shared boundary are seen as a group. | Use cards and surface bands to group content, not extra lines. |
| **Law of Prägnanz** (Gestalt) | People read complex images in their simplest form. | Prefer simple shapes, clean layouts and few competing elements. |
| **Law of Uniform Connectedness** (Gestalt) | Visually connected items are seen as related. | Connect steps (e.g. a checkout progress line) and tie related controls together. |
| **Goal-Gradient Effect** | Motivation rises the closer people get to a goal. | Show progress in multi-step flows (checkout, seller onboarding, KYC). Celebrate completion. |
| **Zeigarnik Effect** | Unfinished tasks stick in memory. | Persist the cart, drafts and half-done onboarding, and gently remind users to finish (never with dark patterns). |
| **Peak–End Rule** | Experiences are judged by their peak and their end. | Invest in key moments: add-to-cart feedback, the order-confirmation page, delivery-completed notifications. |
| **Postel's Law** | Be liberal in what you accept, conservative in what you send. | Accept `08031234567`, `0803 123 4567` and `+2348031234567` alike, and accept pasted card numbers with spaces. Validate kindly and output in one clean format. |
| **Occam's Razor** | Prefer the simplest solution that works. | Remove any element, step or field that doesn't earn its place. |
| **Pareto Principle** | Roughly 80% of effects come from 20% of causes. | Optimise the critical paths first: search → product → cart → checkout → payment. |
| **Parkinson's Law** | A task expands to fill the time available. | Shorten flows: guest checkout, saved addresses, one-tap reorder. |
| **Choice Overload** | Too many options cause decision paralysis. | Use comparison tools, curated "best for…" picks and sensible default variants. |
| **Cognitive Load** | Mental effort should go to the task, not the interface. | Use plain language, recognisable icons with labels, and consistent layouts. |
| **Mental Model** | Users bring expectations from their past experience. | Match how Nigerian shoppers think: prices in naira, delivery by state/LGA, familiar payment methods (card, bank transfer, USSD, wallets). |

## 2. Nielsen's 10 usability heuristics

1. **Visibility of system status:** always show what's happening: loading, saving, payment processing, order status, delivery tracking.
2. **Match with the real world:** use the customer's language ("Add to cart", "Delivery fee"), not system jargon ("SKU", "fulfilment node").
3. **User control and freedom:** provide undo for removals, easy exits, cancellable flows and a back button that works.
4. **Consistency and standards:** the same words, icons and patterns across all five web apps and both mobile apps.
5. **Error prevention:** constrain inputs (date pickers, quantity steppers), confirm destructive actions, disable impossible choices (out-of-stock variants).
6. **Recognition over recall:** keep recently viewed items, visible filters, persistent cart and saved addresses.
7. **Flexibility and efficiency:** shortcuts for experts (staff keyboard shortcuts, bulk actions, saved filters) without burdening newcomers.
8. **Aesthetic and minimalist design:** every element must earn its place.
9. **Help users recover from errors:** say what went wrong, why, and how to fix it, next to the problem, in plain words.
10. **Help and documentation:** contextual help where it's needed (tooltips, inline hints, FAQs), never as a substitute for clarity.

## 3. Accessibility (WCAG 2.2 AA is the floor, not the goal)

- **Perceivable:**
  - Text contrast ≥ 4.5:1 (≥ 3:1 for large text and UI boundaries).
  - Never use colour as the only signal: pair it with text or an icon (sale = badge plus percentage, error = icon plus message).
  - Every image has meaningful `alt` text; decorative images use `alt=""`.
- **Operable:**
  - Everything works by keyboard, in a logical tab order with a visible focus ring.
  - Touch targets ≥ 44 × 44px (WCAG 2.5.8 minimum is 24px; we go higher).
  - Provide a skip link to the main content.
  - No keyboard traps; respect `prefers-reduced-motion`; nothing flashes more than 3 times per second.
- **Understandable:**
  - Set `lang` on `<html>`.
  - Use visible labels on every input; placeholders are not labels.
  - Write error messages that say how to fix the problem.
  - Keep navigation consistent across pages.
- **Robust:**
  - Use semantic HTML first (`<button>`, `<a>`, `<nav>`, `<main>`, `<header>`, `<footer>`, `<section aria-labelledby>`, headings in order).
  - Use ARIA only when HTML can't express it; icon-only buttons need `aria-label`.
- **One `<h1>` per page.** Never skip heading levels for styling; style with classes instead.
- **Live regions:** announce cart updates, form errors and toasts with `aria-live="polite"` (`assertive` only for critical errors).
- **Testing:** test with keyboard only, at 200% zoom, at 320px width, and with a screen reader (VoiceOver or TalkBack) before a feature is done.

## 4. Visual design principles

- **Hierarchy:** size, weight and colour establish one clear reading order per screen. Use one dominant element per section.
- **Alignment:** everything sits on the spacing scale and aligns to the container edges. No eyeballed offsets.
- **Rhythm:** consistent vertical spacing between sections (`--section-space`) and inside components (the token scale).
- **Contrast for meaning:** neutral by default. Colour is reserved for action (accent), price drops (sale), urgency (promo) and status (success, warning, danger).
- **Whitespace is a feature:** premium means space around content. Density is allowed in data-heavy areas (staff portal tables, spec sheets), never in marketing hero areas.
- **Imagery:**
  - Products are shown whole, on a neutral surface, at a consistent ratio.
  - Use real product photography in production.
  - No stock-photo clichés, and never copy other brands' assets.
- **Typography:**
  - At most 2 weights per component and 3 sizes per card.
  - Line length 45–75 characters.
  - Never justify text, and never use all-caps for more than a short label.
- **Consistency over novelty:** reuse existing primitives before creating new ones.

## 5. Layout and responsive behaviour

- **Mobile-first:** design for a 360px viewport first, then enhance at 640 / 768 / 1024 / 1280.
- **Content-out, not device-in:** add a breakpoint only where the layout breaks, using the fixed set.
- **No horizontal page scroll** at any width. Horizontal scrolling is only allowed inside `.scroller` shelves, where the next item visibly peeks in.
- **Thumb zone on mobile:** primary actions in the bottom half (sticky "Add to cart" bars on product pages).
- **Reading patterns:** F-pattern for text-heavy pages and listings; Z-pattern for hero and marketing sections.
- **Above the fold:** the main value proposition and a way to act (search or a primary CTA) must be visible without scrolling at 360 × 640.
- **Sticky elements:** only the header (and the buy bar on product pages). Never stack more than 2 sticky layers.

## 6. E-commerce UX rules

- **Search:**
  - Prominent on every storefront page and forgiving of typos, with suggestions and category scoping.
  - A search with no results suggests alternatives; never a dead end.
- **Navigation:** the category taxonomy matches how shoppers think (Phones → Android / iPhone / Refurbished), with breadcrumbs on category and product pages.
- **Product cards:**
  - Order: image → badge → title (2-line clamp) → key spec (for gadgets) → rating with count → price → old price and discount.
  - The whole card is clickable; secondary actions (wishlist) are separate buttons.
- **Prices:**
  - Always in naira, formatted `₦1,250,000`, with no kobo shown unless non-zero.
  - Show the total cost early: delivery fees are estimated before checkout, and there are no surprise fees.
  - The current price is the most prominent. The old price is struck through and muted; the discount is shown as a percentage in the sale badge.
- **Trust:**
  - Show seller identity ("Sold by TechShop" vs a vendor name with rating), warranty, return policy, delivery estimates and secure-payment cues near the buy button.
  - Make payment methods visible (Paystack, OPay, Moniepoint).
- **Stock and urgency:**
  - Show real stock levels and real countdowns only.
  - **No fake scarcity, fake timers, confirmshaming, pre-ticked add-ons, sneak-into-basket or hidden subscriptions.** Dark patterns are banned; they also break consumer protection law (Nigeria's FCCPA).
- **Cart and checkout:**
  - Guest checkout allowed; minimum fields; address autocomplete by state/LGA.
  - Show progress, an order summary that's always visible, and edits without losing data.
  - Show payment errors with the next step.
- **Gadget-specific:**
  - Show specs that matter (storage, RAM, condition: new / UK-used / refurbished) and comparison.
  - Show IMEI/serial and warranty information in after-sales.
- **Cars:** a listing-style flow (year, mileage, location, inspection report, documents, book a viewing, financing enquiry), never add-to-cart.
- **B2B (wholesale):**
  - Show tiered pricing tables, minimum order quantities, quote requests, VAT-inclusive/exclusive toggles and invoice downloads.
  - Business buyers value speed and precision over decoration.
- **Marketplace:**
  - Seller identity and rating sit on every product.
  - A multi-seller cart groups items by seller with separate delivery estimates.

## 7. Forms

- One column, with labels above the field, and the field width matching the expected input.
- Mark optional fields, not required ones; ask only for what is needed.
- Use the right input types and attributes: `type="email"`, `inputmode="numeric"`, `autocomplete="..."` on every field.
- Validate on blur, not on every keystroke. Keep the user's input after an error; never clear a form.
- The primary button describes the outcome ("Pay ₦250,000", not "Submit") and is disabled only with a stated reason.

## 8. Feedback, states and motion

- **Every component designs all its states:** default, hover, focus, active, disabled, loading, empty, error, success, and skeleton.
- **Empty states** explain what goes here and give a next action.
- **Motion has purpose:** show cause and effect, orient (where something came from), and give feedback.
  - Durations from the tokens (150–400ms); ease-out for entering, ease-in for leaving.
  - Animate only `transform` and `opacity`.
  - Respect `prefers-reduced-motion`.
- **Toasts** are only for non-critical confirmations. Errors that need action stay inline.

## 9. Content and microcopy

- Plain, friendly, direct English. Use the active voice and keep sentences short.
- Buttons are verbs ("Add to cart", "Track order"). Links say where they go; never "click here".
- Consistent terms across apps: one word per concept ("Cart", not sometimes "Bag").
- Localise to Nigeria: naira, +234 phone format, states/LGAs, local delivery language, WAT time zone, DD/MM/YYYY dates.
- Never fake content in production: no invented reviews, ratings, stock counts, statistics or testimonials. Placeholder content in development must be clearly marked as sample data.

## 10. Performance is UX (Core Web Vitals, data-light Nigeria)

- **Targets on a mid-range Android over 3G/4G:** LCP < 2.5s, INP < 200ms, CLS < 0.1.
- **Server-render by default** (React Server Components). Add `"use client"` only for real interactivity, and keep client components small and at the leaves.
- **Images:**
  - Use `next/image` with explicit sizes, modern formats, lazy-loading below the fold and `priority` only for the LCP image.
  - Always reserve space (aspect ratio) to prevent layout shift.
- **Fonts:** load through `next/font` only; no layout shift from late fonts.
- **Third-party scripts:** budget them, load after interaction or idle, and justify each one.
- **Respect data:** no autoplay video on mobile data, and compress everything.
- **Every list over about 50 items** is paginated or virtualised.

## 11. Platform-specific

- **Web apps:**
  - Use the shared design system (`@techshop/ui`). App-specific components use CSS Modules with tokens.
  - Internal apps (seller, staff) prioritise density, keyboard efficiency, tables, bulk actions and clear status over marketing polish, but follow the same tokens and accessibility rules.
- **Mobile apps:**
  - Follow platform conventions (iOS HIG and Material).
  - Native navigation patterns (tab bar for the customer app; a task-focused stack for logistics).
  - Large tap targets for riders, who may be wearing gloves or in sunlight: high contrast and ≥ 48px controls in the logistics app.
  - Design for offline and poor connectivity.
- **Consistency across platforms:** the same tokens, terminology and flows on web and mobile; the visual adapts to the platform, the behaviour doesn't change.

## 12. Definition of done for any UI

- [ ] Uses tokens and primitives only (no raw colours or sizes, no new breakpoints).
- [ ] Works at 360px, 768px, 1024px and 1440px without horizontal page scroll.
- [ ] Light and dark mode both checked.
- [ ] Keyboard-only usable, visible focus, logical heading order, labelled controls.
- [ ] All states designed (loading, empty, error, disabled).
- [ ] No dark patterns; sample data clearly marked as such.
- [ ] Lint, typecheck and build pass; screenshots reviewed.
- [ ] Plan and decisions logged in `docs/plan.md`, actions in `docs/log.md`.
