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
| http://techshop.localhost | Company site (corporate) | 3000 |
| http://market.techshop.localhost | Customer marketplace | 3001 |
| http://wholesale.techshop.localhost | Wholesale & retail | 3002 |
| http://seller.techshop.localhost | Seller centre | 3003 |
| http://staff.techshop.localhost | Staff portal | 3004 |
| http://api.techshop.localhost | Go API | 8080 |

Each app is also reachable directly at `http://localhost:<port>`. If port 80 is taken, set `NGINX_PORT` (e.g. `NGINX_PORT=8000 make infra-up`) and add `:8000` to the URLs.
