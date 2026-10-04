---
name: client-deliverables
description: Produce professional client-facing deliverables for freelance and agency software work — project proposals, scopes of work, cost and time estimates, milestone plans, weekly/monthly progress reports, status update emails, change requests, handover documents, invoices, and meeting notes with action items. Use this skill whenever the user needs to communicate with or deliver something to a client, quote or estimate a project, report progress, request approval for scope changes, or hand over a finished project — even if they just say "write an update for my client" or "how much should I charge for this".
---

# Client Deliverables

## Principles

- **Plain language for non-technical clients**: translate technical work into business outcomes ("customers can now pay with transfer" not "integrated webhook handler").
- **No surprises**: surface risks, delays, and scope changes early and in writing.
- **Everything traceable**: deliverables reference agreed scope, milestones, and change requests.
- **Confidentiality**: never include secrets, credentials, or one client's details in another client's documents.
- **Format**: if the user wants a file, use the document skills available (Word/PDF/Docs) for formal deliverables; emails and short updates go inline.

## Templates (in `assets/`)

| Deliverable | Template |
|---|---|
| Proposal / statement of work | `proposal.md` |
| Estimate breakdown | `estimate.md` |
| Progress report | `progress-report.md` |
| Change request | `change-request.md` |
| Handover document | `handover.md` |
| Meeting notes | `meeting-notes.md` |

## Estimating

1. Break work into tasks of ≤ 2 days each (backend, frontend, mobile, QA, DevOps, PM/communication).
2. Estimate each as **optimistic / likely / pessimistic**; use (O + 4L + P) / 6.
3. Add explicit buffers: integration/unknowns 15–25%, QA/bug fixing 15–20%, project management/communication 10–15%.
4. Always list **assumptions** and **exclusions** — they protect both sides.
5. Present price as fixed-price per milestone, or rate × estimated hours with a cap — state which.
6. Ask the user for their rate and currency; never assume pricing on their behalf.

## Progress report essentials

- Overall status: 🟢 On track / 🟡 At risk / 🔴 Off track — with one-line reason
- Completed this period (outcomes, with demo links/screenshots)
- In progress & next period
- Blockers and what's needed from the client (with dates)
- Risks & mitigations
- Budget/time used vs planned

## Change requests

Any new feature, changed requirement, or extra platform after sign-off → written change request with impact on cost and timeline, approved before work starts.

## Handover

Credentials transferred via a password manager share (never email/chat), repo ownership, hosting/domain ownership in client's name, documentation, warranty/support terms.
