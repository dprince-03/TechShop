# <Project Name>

<One-sentence description of what this is and who it's for.>

## Tech stack
- Backend: 
- Frontend: 
- Database: 
- Hosting: 

## Getting started

### Prerequisites
- <runtime + version>
- Docker

### Setup
```bash
git clone <repo>
cd <repo>
cp .env.example .env      # fill in values (see "Environment variables")
docker compose up -d      # starts DB, Redis
<install command>
<migrate command>
<dev command>
```
App runs at http://localhost:<port>.

## Environment variables
| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | yes | PostgreSQL connection string |

## Common commands
| Command | Purpose |
|---|---|
| `<test>` | Run tests |
| `<lint>` | Lint |
| `<migrate>` | Run migrations |

## Project structure
```
<tree of key folders with one-line descriptions>
```

## Deployment
<Where it's hosted, how deploys happen, link to runbooks.>

## Documentation
- Architecture: `docs/architecture.md`
- ADRs: `docs/adr/`
- API: `openapi.yaml`

## Contributing
Branch naming, commit format, PR process.
