# Mockups and simulations

Interactive, self-contained HTML pages. Open any file in a browser; there's no build step. They use sample data only and send nothing anywhere.

Each system page has two tabs:
- **UI mockup**: the screens that system needs, fully clickable.
- **System simulation**: an animated diagram of how the parts communicate, with "what can go wrong" switches.

| Page | Plan doc |
|---|---|
| [`prototype.html`](prototype.html) | All seven apps sharing one state (web + mobile) |
| [`system-simulator.html`](system-simulator.html) | Whole-platform backend and recommender flows ([`../backend.md`](../backend.md), [`../recommendations.md`](../recommendations.md)) |
| [`messaging-marketing.html`](messaging-marketing.html) | [`../messaging-marketing.md`](../messaging-marketing.md) |
| [`identity-access.html`](identity-access.html) | [`../identity-access.md`](../identity-access.md) |
| [`payments-finance.html`](payments-finance.html) | [`../payments-finance.md`](../payments-finance.md) |
| [`search-catalogue.html`](search-catalogue.html) | [`../search-catalogue.md`](../search-catalogue.md) |
| [`orders-fulfilment.html`](orders-fulfilment.html) | [`../orders-fulfilment.md`](../orders-fulfilment.md) |
| [`trust-safety.html`](trust-safety.html) | [`../trust-safety.md`](../trust-safety.md) |

They're throwaway visual aids, not app code. The real implementation follows the plan docs and the rules in `CLAUDE.md`. The system pages were generated from a small shared kit (styles, a click framework and a simulation engine), so they look and behave the same. Private published copies on claude.ai are linked in [`../log.md`](../log.md).
