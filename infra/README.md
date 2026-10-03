# infra/

Infrastructure for running TechShop. Currently **local development only**; production deployment is not set up yet.

| Folder | Contents |
| --- | --- |
| `docker/` | `compose.yml`: local PostgreSQL and the nginx reverse proxy |
| `nginx/` | nginx config that routes local hostnames to each app's dev server |

## Local hostnames

Start the services with `make infra-up` from the repo root, then run the apps (`make dev-api`, `make dev-web`). Open:

| URL | App | Dev port |
| --- | --- | --- |
| http://techshop.localhost | Company site (corporate) | 3005 |
| http://market.techshop.localhost | Customer marketplace | 3001 |
| http://wholesale.techshop.localhost | Wholesale & retail | 3002 |
| http://seller.techshop.localhost | Seller centre | 3003 |
| http://staff.techshop.localhost | Staff portal | 3004 |
| http://api.techshop.localhost | Go API | 8080 |

Each app is also reachable directly at `http://localhost:<port>`. If port 80 is taken, set `NGINX_PORT` (e.g. `NGINX_PORT=8000 make infra-up`) and add `:8000` to the URLs.

## Changing an app's port

A web app's port appears in three places. Change all three:

1. **The app's scripts:** `frontend/apps/<app>/package.json`, in both `"dev": "next dev --port …"` and `"start": "next start --port …"`. Next.js can't read the port from `.env` files, because the server starts before env files load.
2. **nginx routing:** `infra/nginx/conf.d/techshop.conf`, in the `proxy_pass http://host.docker.internal:…;` line under that app's `server_name`. Then restart nginx: `make infra-down && make infra-up`.
3. **Backend CORS:** `http://localhost:<port>` in `CORS_ALLOWED_ORIGINS`, in your `backend/.env` (and in `backend/.env.example` / `internal/config/config.go` so others get it too). Then restart the API.

Find every reference with `grep -rn "<old port>" --exclude-dir=node_modules --exclude-dir=docs .`

Other ports:
- **Go API:** `PORT` in `backend/.env`. Also update the `api.techshop.localhost` block in nginx, and `API_URL` / `NEXT_PUBLIC_API_URL` in each `frontend/apps/*/.env.local` and `EXPO_PUBLIC_API_URL` in each `mobile/apps/*/.env.local`.
- **Postgres:** `POSTGRES_PORT=5433 make infra-up`. Then change the port in `DATABASE_URL` in `backend/.env`.
- **nginx:** `NGINX_PORT=8000 make infra-up`. Then add `:8000` to the local hostnames.

Taken on this machine: **3000** is used by AdGuard, so no app may use it.
