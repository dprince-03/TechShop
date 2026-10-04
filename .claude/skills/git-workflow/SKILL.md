---
name: git-workflow
description: Git and GitHub workflow conventions — Conventional Commit messages, branch naming, pull request titles and descriptions, changelogs, semantic versioning, release notes, tagging, and resolving merge/rebase situations. Use this skill whenever the user asks to write a commit message, name a branch, draft or describe a PR, prepare a release, write release notes or a changelog, bump a version, or asks how to handle a Git situation (rebase, squash, revert, cherry-pick, conflict).
---

# Git Workflow

## Branches

`<type>/<short-kebab-description>` with optional ticket: `feat/invoice-export`, `fix/TRP-142-tenant-leak`, `chore/upgrade-gin`.
Types: `feat`, `fix`, `chore`, `refactor`, `docs`, `test`, `perf`, `ci`, `hotfix`.

Default model: trunk-based with short-lived branches off `main`, merged via PR. Use `develop`/release branches only if the project already does.

## Commit messages (Conventional Commits)

```
<type>(<scope>): <imperative summary, ≤ 72 chars, no period>

<body: what and why, wrapped at 72; not how>

<footer: BREAKING CHANGE: ..., Refs: TRP-142, Co-authored-by: ...>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`. Breaking change: `feat(api)!: ...` plus `BREAKING CHANGE:` footer.

Examples:
- `feat(invoices): add PDF export endpoint`
- `fix(auth): reject refresh tokens after password change`
- `refactor(db): extract tenant scoping into repository helper`
- `chore(deps): bump gin to v1.10.0`

When generating from a diff: summarise the *intent* (why), group unrelated changes into separate commits, and never invent ticket numbers.

## Pull requests

Title = the main Conventional Commit line. Body from `assets/pull_request_template.md`:
- What & why
- How (key decisions)
- Screenshots/recordings for UI
- Testing done
- Security/migration/config impact
- Checklist

Keep PRs small (ideally < 400 changed lines). Squash-merge by default so `main` history reads as one commit per PR.

## Versioning & releases (SemVer)

- MAJOR: breaking changes · MINOR: new backward-compatible features · PATCH: fixes.
- Tag `vX.Y.Z` on `main`; generate notes from commits since the last tag.
- Changelog (`CHANGELOG.md`, Keep a Changelog format): sections **Added, Changed, Deprecated, Removed, Fixed, Security**.
- Release notes for clients: plain language, user-facing impact first, no internal jargon.

## Common situations

| Situation | Command / approach |
|---|---|
| Update branch with main | `git fetch origin && git rebase origin/main` (or merge if branch is shared) |
| Squash local commits | `git rebase -i origin/main` |
| Undo a pushed commit safely | `git revert <sha>` |
| Undo last local commit, keep changes | `git reset --soft HEAD~1` |
| Bring one fix to another branch | `git cherry-pick <sha>` |
| Recover lost work | `git reflog` |
| Committed a secret | Revoke/rotate the secret first, then purge history (git filter-repo) — rotation matters more than deletion |

Never force-push to `main` or shared branches; use `--force-with-lease` on your own branches.
