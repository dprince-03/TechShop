# Claude Skills Pack — 16 Project Skills

| Skill | Triggers on |
|---|---|
| `coding-conventions` | Writing/editing code in Go, Node, React/Next.js, Java |
| `module-scaffolding` | "Add a new module/endpoint/CRUD/feature" |
| `security-audit` | Audits, security reviews, compliance readiness (includes your 11 checklists + scan script) |
| `code-review` | Reviewing code or PRs |
| `testing-standards` | Writing/fixing tests, coverage, test setup |
| `database-migrations` | Schemas, migrations, indexes (includes risky-SQL checker script) |
| `api-design` | Endpoint design, OpenAPI specs (includes OpenAPI template) |
| `devops-deployment` | Docker, Nginx, CI/CD, Render, AWS (includes Dockerfiles, compose, nginx, CI templates) |
| `git-workflow` | Commits, branches, PRs, releases, changelogs |
| `debugging-runbooks` | Bugs, errors, outages (includes 13 runbooks) |
| `performance-optimization` | Slow APIs, queries, Web Vitals, load testing |
| `frontend-design-system` | UI, components, accessibility, responsive design |
| `mobile-development` | React Native, Flutter, store releases |
| `technical-docs` | READMEs, ADRs, architecture, runbooks, postmortems |
| `client-deliverables` | Proposals, estimates, progress reports, handovers |
| `project-domain-knowledge` | Project-specific business rules (fill in per project) |

## Installing

**Claude Code (per project, shared with the team):** copy the `.claude/skills/` folder into the root of your repo and commit it.

**Claude Code (personal, all projects):** copy the skill folders into `~/.claude/skills/`.

**Claude.ai:** open each file in `packaged/` (`.skill` files) and use the save/upload option, or upload them through your skills settings.

## Customize first

1. `project-domain-knowledge` — copy `references/_TEMPLATE.md` per project and fill it in. This is the skill that adds the most value over time.
2. `coding-conventions` — edit the reference files to match your actual house style.
3. `client-deliverables` — add your rates, payment terms, and branding.
