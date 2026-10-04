---
name: code-review
description: Review code, diffs, and pull requests for correctness, readability, maintainability, performance, test coverage, and consistency with project conventions, producing prioritized, actionable comments with suggested fixes. Use this skill whenever the user asks to review, critique, check, or improve code or a PR, asks "what's wrong with this", "is this good", or pastes code for feedback. For deep security-focused audits, also use the security-audit skill.
---

# Code Review

Review like a senior engineer who wants the author to succeed: specific, prioritized, and with fixes, not just complaints.

## Process

1. **Understand intent**: what is the change supposed to do? (PR description, ticket, or ask briefly if unclear.)
2. **Read the whole diff once** before commenting, to avoid comments the code later answers.
3. **Check in priority order**:
   1. **Correctness** — does it do what's intended? Edge cases, null/empty, off-by-one, error paths, concurrency, timezone, money precision.
   2. **Security quick-check** — authZ on every endpoint, tenant scoping, input validation, secrets, injection. Escalate to `security-audit` if anything serious appears.
   3. **Data integrity** — transactions, idempotency, race conditions, migrations safety.
   4. **Tests** — do tests cover the behaviour and failure cases? Would they catch a regression?
   5. **Design** — right layer, single responsibility, no duplication of existing helpers, sensible abstractions (not premature).
   6. **Readability** — naming, function length, clear control flow, comments explaining why.
   7. **Performance** — N+1 queries, unbounded queries, missing indexes, unnecessary work in loops, large payloads, unnecessary re-renders.
   8. **Consistency** — matches project conventions (see `coding-conventions`).
4. **Note what's good** — briefly, so the author knows what to keep doing.

## Comment labels

- **[blocker]** must fix before merge (bugs, security, data loss)
- **[major]** should fix (design problems, missing tests for critical logic)
- **[minor]** nice improvement
- **[nit]** style/preference — optional
- **[question]** clarification needed

## Output format

```
## Review summary
<1–3 sentences: overall assessment and merge recommendation: Approve / Approve with changes / Request changes>

## Blockers
- [blocker] `file.go:42` — <issue>. <why it matters>.
  Suggested fix:
  ```go
  ...
  ```

## Major
...
## Minor & nits
...
## What's good
...
```

## Rules

- Be concrete: cite file and line, show the fix.
- One issue per comment; don't bury blockers among nits.
- Don't rewrite the whole thing in a different style unless asked.
- If something is a matter of preference, say so.
- If you couldn't verify something (e.g., code not shown), say what you'd need.
